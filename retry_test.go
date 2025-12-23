package overpass

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	t.Parallel()

	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("expected MaxRetries=3, got %d", config.MaxRetries)
	}

	if config.InitialBackoff != time.Second {
		t.Errorf("expected InitialBackoff=1s, got %v", config.InitialBackoff)
	}

	if config.MaxBackoff != 30*time.Second {
		t.Errorf("expected MaxBackoff=30s, got %v", config.MaxBackoff)
	}

	if config.BackoffMultiplier != 2.0 {
		t.Errorf("expected BackoffMultiplier=2.0, got %f", config.BackoffMultiplier)
	}

	if !config.Jitter {
		t.Error("expected Jitter=true")
	}
}

func TestIsRetryableStatus(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		status    int
		retryable bool
	}{
		{200, false},
		{400, false},
		{401, false},
		{403, false},
		{404, false},
		{429, true},
		{500, true},
		{502, true},
		{503, true},
		{504, true},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(fmt.Sprintf("status_%d", tc.status), func(t *testing.T) {
			t.Parallel()

			if got := isRetryableStatus(tc.status); got != tc.retryable {
				t.Errorf("status %d: expected %v, got %v", tc.status, tc.retryable, got)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	t.Parallel()

	config := RetryConfig{
		InitialBackoff:    time.Second,
		MaxBackoff:        10 * time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	testCases := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},
		{1, 2 * time.Second},
		{2, 4 * time.Second},
		{3, 8 * time.Second},
		{4, 10 * time.Second}, // capped at MaxBackoff
		{5, 10 * time.Second}, // still capped
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(fmt.Sprintf("attempt_%d", tc.attempt), func(t *testing.T) {
			t.Parallel()

			got := calculateBackoff(tc.attempt, config)
			if got != tc.expected {
				t.Errorf("attempt %d: expected %v, got %v", tc.attempt, tc.expected, got)
			}
		})
	}
}

func TestCalculateBackoffWithJitter(t *testing.T) {
	t.Parallel()

	config := DefaultRetryConfig()

	backoff := calculateBackoff(0, config)

	// Should be within 0-25% above initial backoff
	minExpected := config.InitialBackoff
	maxExpected := time.Duration(float64(config.InitialBackoff) * 1.25)

	if backoff < minExpected || backoff > maxExpected {
		t.Errorf("jittered backoff %v outside expected range [%v, %v]", backoff, minExpected, maxExpected)
	}
}

// Mock client that fails N times then succeeds.
type failingMockClient struct {
	failCount   int
	currentFail int
	statusCode  int
}

func (m *failingMockClient) Do(req *http.Request) (*http.Response, error) {
	m.currentFail++

	if m.currentFail <= m.failCount {
		body := []byte(fmt.Sprintf("error %d", m.currentFail))

		return &http.Response{
			StatusCode: m.statusCode,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	}

	// Success after failCount attempts
	successBody := []byte(`{"osm3s":{},"elements":[]}`)

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(bytes.NewReader(successBody)),
	}, nil
}

func TestRetrySuccess(t *testing.T) {
	t.Parallel()

	mock := &failingMockClient{failCount: 2, statusCode: 503}

	config := RetryConfig{
		MaxRetries:        3,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	client := NewWithSettings(apiEndpoint, 1, mock)
	client.retryConfig = config

	start := time.Now()
	result, err := client.QueryContext(context.Background(), "[out:json];node(1);out;")
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}

	if mock.currentFail != 3 {
		t.Errorf("expected 3 attempts, got %d", mock.currentFail)
	}

	// Should have waited ~10ms + ~20ms = ~30ms
	if elapsed < 25*time.Millisecond || elapsed > 50*time.Millisecond {
		t.Logf("warning: elapsed time %v outside expected range (test timing may be flaky)", elapsed)
	}

	if result.Count != 0 {
		t.Errorf("expected empty result, got Count=%d", result.Count)
	}
}

func TestRetryExhaustion(t *testing.T) {
	t.Parallel()

	mock := &failingMockClient{failCount: 10, statusCode: 503}

	config := RetryConfig{
		MaxRetries:        2,
		InitialBackoff:    10 * time.Millisecond,
		MaxBackoff:        100 * time.Millisecond,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	client := NewWithSettings(apiEndpoint, 1, mock)
	client.retryConfig = config

	_, err := client.QueryContext(context.Background(), "[out:json];node(1);out;")
	if err == nil {
		t.Fatal("expected error after exhausting retries")
	}

	if !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("expected 'max retries exceeded' error, got: %v", err)
	}

	if mock.currentFail != 3 {
		t.Errorf("expected 3 attempts (initial + 2 retries), got %d", mock.currentFail)
	}
}

func TestNoRetryOnNonRetryableStatus(t *testing.T) {
	t.Parallel()

	mock := &failingMockClient{failCount: 10, statusCode: 400}

	client := NewWithSettings(apiEndpoint, 1, mock)
	client.retryConfig = DefaultRetryConfig()

	_, err := client.QueryContext(context.Background(), "[out:json];node(1);out;")
	if err == nil {
		t.Fatal("expected error")
	}

	if mock.currentFail != 1 {
		t.Errorf("expected only 1 attempt (no retries for 400), got %d", mock.currentFail)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	t.Parallel()

	mock := &failingMockClient{failCount: 10, statusCode: 503}

	config := RetryConfig{
		MaxRetries:        5,
		InitialBackoff:    100 * time.Millisecond,
		MaxBackoff:        time.Second,
		BackoffMultiplier: 2.0,
		Jitter:            false,
	}

	client := NewWithSettings(apiEndpoint, 1, mock)
	client.retryConfig = config

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
	defer cancel()

	_, err := client.QueryContext(ctx, "[out:json];node(1);out;")
	if err == nil {
		t.Fatal("expected error from context cancellation")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got: %v", err)
	}

	// Should have attempted 1-2 times before context cancelled
	if mock.currentFail < 1 || mock.currentFail > 2 {
		t.Logf("warning: unexpected attempt count: %d (test timing may be flaky)", mock.currentFail)
	}
}

func TestDisableRetry(t *testing.T) {
	t.Parallel()

	mock := &failingMockClient{failCount: 2, statusCode: 503}

	config := RetryConfig{MaxRetries: 0}
	client := NewWithSettings(apiEndpoint, 1, mock)
	client.retryConfig = config

	_, err := client.QueryContext(context.Background(), "[out:json];node(1);out;")
	if err == nil {
		t.Fatal("expected error")
	}

	if mock.currentFail != 1 {
		t.Errorf("expected only 1 attempt when retries disabled, got %d", mock.currentFail)
	}
}

func TestRetryDifferentStatusCodes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name         string
		statusCode   int
		shouldRetry  bool
		maxRetries   int
		expectedFail int
	}{
		{"429 Too Many Requests", 429, true, 2, 3},
		{"500 Internal Server Error", 500, true, 2, 3},
		{"502 Bad Gateway", 502, true, 2, 3},
		{"503 Service Unavailable", 503, true, 2, 3},
		{"504 Gateway Timeout", 504, true, 2, 3},
		{"400 Bad Request", 400, false, 2, 1},
		{"401 Unauthorized", 401, false, 2, 1},
		{"403 Forbidden", 403, false, 2, 1},
		{"404 Not Found", 404, false, 2, 1},
	}

	for _, tc := range testCases {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mock := &failingMockClient{failCount: 10, statusCode: tc.statusCode}

			config := RetryConfig{
				MaxRetries:        tc.maxRetries,
				InitialBackoff:    1 * time.Millisecond,
				MaxBackoff:        10 * time.Millisecond,
				BackoffMultiplier: 2.0,
				Jitter:            false,
			}

			client := NewWithSettings(apiEndpoint, 1, mock)
			client.retryConfig = config

			_, err := client.QueryContext(context.Background(), "[out:json];node(1);out;")
			if err == nil {
				t.Fatal("expected error")
			}

			if mock.currentFail != tc.expectedFail {
				t.Errorf("expected %d attempts, got %d", tc.expectedFail, mock.currentFail)
			}

			if tc.shouldRetry && !strings.Contains(err.Error(), "max retries exceeded") {
				t.Errorf("expected 'max retries exceeded' for retryable status, got: %v", err)
			}
		})
	}
}
