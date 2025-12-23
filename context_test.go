package overpass

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestQueryContext_Success(t *testing.T) {
	client := NewWithSettings(apiEndpoint, 1, &mockHttpClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[{"type":"node","id":1,"lat":1.0,"lon":2.0}]}`),
		},
	})

	ctx := context.Background()

	result, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 1 {
		t.Errorf("expected count 1, got %d", result.Count)
	}

	if len(result.Nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(result.Nodes))
	}
}

func TestQueryContext_Cancellation(t *testing.T) {
	// Create a client with a slow mock that checks for context cancellation
	slowClient := &mockCancellableHttpClient{
		delay: 200 * time.Millisecond,
	}
	client := NewWithSettings(apiEndpoint, 1, slowClient)

	// Create a context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Should fail with context cancelled error
	_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got: %v", err)
	}
}

func TestQueryContext_Timeout(t *testing.T) {
	// Create a client with a slow mock
	slowClient := &mockCancellableHttpClient{
		delay: 200 * time.Millisecond,
	}
	client := NewWithSettings(apiEndpoint, 1, slowClient)

	// Create a context with a very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Should fail with deadline exceeded
	_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected context.DeadlineExceeded error, got: %v", err)
	}
}

func TestQueryContext_Background(t *testing.T) {
	client := NewWithSettings(apiEndpoint, 1, &mockHttpClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[]}`),
		},
	})

	// Using background context should work fine
	result, err := client.QueryContext(context.Background(), `[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
}

func TestQuery_UsesBackgroundContext(t *testing.T) {
	client := NewWithSettings(apiEndpoint, 1, &mockHttpClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[]}`),
		},
	})

	// Old Query method should still work (uses context.Background internally)
	result, err := client.Query(`[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
}

func TestPackageLevelQueryContext(t *testing.T) {
	// Test package-level QueryContext function
	// Note: This uses DefaultClient which is initialized with http.DefaultClient
	// In a real test environment, this would make actual HTTP requests
	// For unit testing, we would need to replace DefaultClient

	// Save original DefaultClient
	originalClient := DefaultClient

	// Replace with mock
	DefaultClient = NewWithSettings(apiEndpoint, 1, &mockHttpClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[]}`),
		},
	})

	// Restore after test
	defer func() {
		DefaultClient = originalClient
	}()

	ctx := context.Background()

	result, err := QueryContext(ctx, `[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
}

func TestPackageLevelQuery(t *testing.T) {
	// Save original DefaultClient
	originalClient := DefaultClient

	// Replace with mock
	DefaultClient = NewWithSettings(apiEndpoint, 1, &mockHttpClient{
		res: &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[]}`),
		},
	})

	// Restore after test
	defer func() {
		DefaultClient = originalClient
	}()

	result, err := Query(`[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Count != 0 {
		t.Errorf("expected count 0, got %d", result.Count)
	}
}

// mockCancellableHttpClient simulates HTTP client that respects context cancellation.
type mockCancellableHttpClient struct {
	delay time.Duration
}

func (m *mockCancellableHttpClient) Do(req *http.Request) (*http.Response, error) {
	// Check if context is already cancelled
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	default:
	}

	// Simulate slow operation with context awareness
	timer := time.NewTimer(m.delay)
	defer timer.Stop()

	select {
	case <-timer.C:
		// Completed successfully
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       newTestBody(`{"elements":[]}`),
		}, nil
	case <-req.Context().Done():
		// Context was cancelled during delay
		return nil, req.Context().Err()
	}
}
