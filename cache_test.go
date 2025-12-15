package overpass

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestDefaultCacheConfig(t *testing.T) {
	config := DefaultCacheConfig()

	if config.Enabled {
		t.Error("cache should be disabled by default")
	}
	if config.TTL != 5*time.Minute {
		t.Errorf("expected TTL=5m, got %v", config.TTL)
	}
	if config.MaxEntries != 1000 {
		t.Errorf("expected MaxEntries=1000, got %d", config.MaxEntries)
	}
}

func TestCacheDisabledByDefault(t *testing.T) {
	c := newCache(DefaultCacheConfig())

	// Set should be no-op when disabled
	c.set("endpoint", "query", Result{Count: 42})

	// Get should return miss
	_, hit := c.get("endpoint", "query")
	if hit {
		t.Error("cache hit when cache disabled")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        time.Minute,
		MaxEntries: 100,
	}
	c := newCache(config)

	result := Result{Count: 42, Timestamp: time.Now()}

	// Cache miss initially
	_, hit := c.get("endpoint", "query1")
	if hit {
		t.Error("unexpected cache hit")
	}

	// Set and retrieve
	c.set("endpoint", "query1", result)

	retrieved, hit := c.get("endpoint", "query1")
	if !hit {
		t.Fatal("expected cache hit")
	}

	if retrieved.Count != result.Count {
		t.Errorf("expected Count=%d, got %d", result.Count, retrieved.Count)
	}
}

func TestCacheExpiration(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        100 * time.Millisecond,
		MaxEntries: 100,
	}
	c := newCache(config)

	result := Result{Count: 42}
	c.set("endpoint", "query1", result)

	// Should be cached immediately
	_, hit := c.get("endpoint", "query1")
	if !hit {
		t.Fatal("expected cache hit")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, hit = c.get("endpoint", "query1")
	if hit {
		t.Error("expected cache miss after expiration")
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	config := CacheConfig{Enabled: true, TTL: time.Minute, MaxEntries: 100}
	c := newCache(config)

	// Different queries should have different cache entries
	c.set("endpoint", "query1", Result{Count: 1})
	c.set("endpoint", "query2", Result{Count: 2})

	result1, hit1 := c.get("endpoint", "query1")
	result2, hit2 := c.get("endpoint", "query2")

	if !hit1 || !hit2 {
		t.Fatal("expected cache hits")
	}

	if result1.Count != 1 || result2.Count != 2 {
		t.Error("cache entries mixed up")
	}

	// Different endpoints should have different cache entries
	c.set("endpoint1", "query", Result{Count: 10})
	c.set("endpoint2", "query", Result{Count: 20})

	result1, _ = c.get("endpoint1", "query")
	result2, _ = c.get("endpoint2", "query")

	if result1.Count != 10 || result2.Count != 20 {
		t.Error("endpoint differentiation failed")
	}
}

func TestCacheMaxEntries(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        time.Hour,
		MaxEntries: 3,
	}
	c := newCache(config)

	// Fill cache beyond capacity
	c.set("e", "q1", Result{Count: 1})
	time.Sleep(time.Millisecond) // Ensure different timestamps
	c.set("e", "q2", Result{Count: 2})
	time.Sleep(time.Millisecond)
	c.set("e", "q3", Result{Count: 3})
	time.Sleep(time.Millisecond)
	c.set("e", "q4", Result{Count: 4}) // Should evict oldest (q1)

	// Check size
	if size := c.size(); size != 3 {
		t.Errorf("expected size=3, got %d", size)
	}

	// q1 should be evicted
	_, hit := c.get("e", "q1")
	if hit {
		t.Error("q1 should have been evicted")
	}

	// q2-q4 should exist
	_, hit = c.get("e", "q4")
	if !hit {
		t.Error("q4 should exist")
	}
}

func TestCacheClear(t *testing.T) {
	config := CacheConfig{Enabled: true, TTL: time.Hour, MaxEntries: 100}
	c := newCache(config)

	c.set("e", "q1", Result{Count: 1})
	c.set("e", "q2", Result{Count: 2})

	if size := c.size(); size != 2 {
		t.Errorf("expected size=2, got %d", size)
	}

	c.clear()

	if size := c.size(); size != 0 {
		t.Errorf("expected size=0 after clear, got %d", size)
	}

	_, hit := c.get("e", "q1")
	if hit {
		t.Error("cache should be empty after clear")
	}
}

func TestCacheCleanup(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        50 * time.Millisecond,
		MaxEntries: 100,
	}
	c := newCache(config)

	c.set("e", "q1", Result{Count: 1})
	c.set("e", "q2", Result{Count: 2})

	if size := c.size(); size != 2 {
		t.Errorf("expected size=2, got %d", size)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Manual cleanup
	c.cleanup()

	if size := c.size(); size != 0 {
		t.Errorf("expected size=0 after cleanup, got %d", size)
	}
}

func TestCacheCleanupRoutine(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        50 * time.Millisecond,
		MaxEntries: 100,
	}
	c := newCache(config)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c.startCleanupRoutine(ctx)

	c.set("e", "q1", Result{Count: 1})
	c.set("e", "q2", Result{Count: 2})

	// Wait for automatic cleanup
	time.Sleep(150 * time.Millisecond)

	if size := c.size(); size != 0 {
		t.Errorf("expected automatic cleanup, got size=%d", size)
	}
}

func TestClientCacheIntegration(t *testing.T) {
	successBody := []byte(`{"osm3s":{},"elements":[{"type":"node","id":1}]}`)
	mock := &mockHttpClient{
		res: &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(successBody)),
		},
	}

	client := NewWithSettings(apiEndpoint, 1, mock)

	// Enable caching
	client.SetCacheConfig(CacheConfig{
		Enabled:    true,
		TTL:        time.Minute,
		MaxEntries: 100,
	})

	query := "[out:json];node(1);out;"

	// First query - should hit API
	result1, err := client.QueryContext(context.Background(), query)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// Create new mock that would return different result
	mock.res = &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader([]byte(`{"osm3s":{},"elements":[{"type":"node","id":999}]}`))),
	}

	// Second query - should hit cache, not API
	result2, err := client.QueryContext(context.Background(), query)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// Results should be identical (from cache)
	if result1.Count != result2.Count {
		t.Error("cache not working - got different results")
	}

	// Verify cache was used
	if client.CacheSize() != 1 {
		t.Errorf("expected cache size=1, got %d", client.CacheSize())
	}
}

func TestClientClearCache(t *testing.T) {
	client := New()
	client.SetCacheConfig(CacheConfig{Enabled: true, TTL: time.Hour, MaxEntries: 100})

	// Populate cache
	client.cache.set(client.apiEndpoint, "query1", Result{Count: 1})
	client.cache.set(client.apiEndpoint, "query2", Result{Count: 2})

	if size := client.CacheSize(); size != 2 {
		t.Errorf("expected size=2, got %d", size)
	}

	client.ClearCache()

	if size := client.CacheSize(); size != 0 {
		t.Errorf("expected size=0 after clear, got %d", size)
	}
}

func TestCacheDisabledSkipsStorage(t *testing.T) {
	config := CacheConfig{
		Enabled:    false,
		TTL:        time.Hour,
		MaxEntries: 100,
	}
	c := newCache(config)

	// Attempt to set
	c.set("endpoint", "query", Result{Count: 42})

	// Size should be 0 since cache is disabled
	if size := c.size(); size != 0 {
		t.Errorf("expected size=0 when disabled, got %d", size)
	}
}

func TestCacheWithZeroMaxEntries(t *testing.T) {
	config := CacheConfig{
		Enabled:    true,
		TTL:        time.Minute,
		MaxEntries: 0, // Unlimited
	}
	c := newCache(config)

	// Add many entries
	for i := 0; i < 100; i++ {
		c.set("e", string(rune(i)), Result{Count: i})
	}

	// All should be stored (no eviction)
	if size := c.size(); size != 100 {
		t.Errorf("expected size=100, got %d", size)
	}
}
