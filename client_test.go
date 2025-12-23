package overpass

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	t.Parallel()

	client := New()

	if client.apiEndpoint != apiEndpoint {
		t.Errorf("expected endpoint %s, got %s", apiEndpoint, client.apiEndpoint)
	}

	if client.httpClient != http.DefaultClient {
		t.Error("expected http.DefaultClient")
	}

	if cap(client.semaphore) != 1 {
		t.Errorf("expected semaphore capacity 1, got %d", cap(client.semaphore))
	}

	if len(client.semaphore) != 1 {
		t.Errorf("expected semaphore length 1, got %d", len(client.semaphore))
	}
}

func TestNewWithSettings(t *testing.T) {
	t.Parallel()

	customEndpoint := "https://custom.example.com/api"
	maxParallel := 5
	customClient := &mockHTTPClient{}

	client := NewWithSettings(customEndpoint, maxParallel, customClient)

	if client.apiEndpoint != customEndpoint {
		t.Errorf("expected endpoint %s, got %s", customEndpoint, client.apiEndpoint)
	}

	if client.httpClient != customClient {
		t.Error("expected custom HTTP client")
	}

	if cap(client.semaphore) != maxParallel {
		t.Errorf("expected semaphore capacity %d, got %d", maxParallel, cap(client.semaphore))
	}

	if len(client.semaphore) != maxParallel {
		t.Errorf("expected semaphore length %d, got %d", maxParallel, len(client.semaphore))
	}
}

func TestClientRateLimiting(t *testing.T) {
	t.Parallel()

	maxParallel := 2
	requestCount := int32(0)
	maxConcurrent := int32(0)
	currentConcurrent := int32(0)

	// Custom client that simulates slow requests
	slowClient := &mockSlowHTTPClient{
		delay: 100 * time.Millisecond,
		onRequest: func() {
			current := atomic.AddInt32(&currentConcurrent, 1)
			atomic.AddInt32(&requestCount, 1)

			// Track max concurrent
			for {
				maxCur := atomic.LoadInt32(&maxConcurrent)
				if current <= maxCur || atomic.CompareAndSwapInt32(&maxConcurrent, maxCur, current) {
					break
				}
			}
		},
		onResponse: func() {
			atomic.AddInt32(&currentConcurrent, -1)
		},
	}

	client := NewWithSettings(apiEndpoint, maxParallel, slowClient)

	// Launch multiple concurrent requests
	numRequests := 5
	var waitGroup sync.WaitGroup
	waitGroup.Add(numRequests)

	for i := 0; i < numRequests; i++ {
		go func() {
			defer waitGroup.Done()

			_, _ = client.Query(`[out:json];node(1);out;`)
		}()
	}

	waitGroup.Wait()

	if atomic.LoadInt32(&requestCount) != int32(numRequests) {
		t.Errorf("expected %d requests, got %d", numRequests, requestCount)
	}

	maxCur := atomic.LoadInt32(&maxConcurrent)
	if maxCur > int32(maxParallel) {
		t.Errorf("rate limiting failed: max concurrent %d exceeds limit %d", maxCur, maxParallel)
	}

	t.Logf("Max concurrent requests: %d (limit: %d)", maxCur, maxParallel)
}

func TestClientConcurrency(t *testing.T) {
	t.Parallel()

	// Test that multiple goroutines can safely use the client
	// Use mockConcurrentHTTPClient that creates fresh response body for each request
	client := NewWithSettings(apiEndpoint, 3, &mockConcurrentHTTPClient{})

	numGoroutines := 10
	var waitGroup sync.WaitGroup
	waitGroup.Add(numGoroutines)

	errors := make(chan error, numGoroutines*5)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer waitGroup.Done()

			for j := 0; j < 5; j++ {
				_, err := client.Query(`[out:json];node(1);out;`)
				if err != nil {
					errors <- err
				}
			}
		}()
	}

	waitGroup.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("unexpected error: %v", err)
	}
}

// mockSlowHTTPClient simulates slow HTTP responses for rate limiting tests.
type mockSlowHTTPClient struct {
	delay      time.Duration
	onRequest  func()
	onResponse func()
}

func (m *mockSlowHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	if m.onRequest != nil {
		m.onRequest()
	}

	time.Sleep(m.delay)

	if m.onResponse != nil {
		m.onResponse()
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       newTestBody(`{"elements":[]}`),
	}, nil
}

// mockConcurrentHTTPClient creates fresh response bodies for each request (concurrency-safe).
type mockConcurrentHTTPClient struct{}

func (m *mockConcurrentHTTPClient) Do(_ *http.Request) (*http.Response, error) {
	// Create a fresh body for each request to avoid shared state issues
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       newTestBody(`{"elements":[]}`),
	}, nil
}
