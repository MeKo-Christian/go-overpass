//go:build integration

package overpass

import (
	"context"
	"testing"
	"time"
)

// TestRealAPIQuery tests a simple query against the real Overpass API
// Run with: go test -tags=integration -v
func TestRealAPIQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Query for a well-known node (ID 1 is a historic node in OSM)
	result, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	// We expect at least some data back
	if result.Count == 0 {
		t.Error("expected at least one element in result")
	}

	t.Logf("Query returned %d elements", result.Count)
	t.Logf("Timestamp: %s", result.Timestamp)

	// If we got nodes, log the first one
	if len(result.Nodes) > 0 {
		for id, node := range result.Nodes {
			t.Logf("Node %d: lat=%.6f, lon=%.6f, tags=%d", id, node.Lat, node.Lon, len(node.Tags))
			break
		}
	}
}

// TestRealAPI_Timeout tests timeout behavior with real API
func TestRealAPI_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := New()

	// Create a context with very short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This should timeout
	_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

// TestRealAPI_InvalidQuery tests error handling with real API
func TestRealAPI_InvalidQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Send an invalid query
	_, err := client.QueryContext(ctx, `invalid query syntax`)
	if err == nil {
		t.Error("expected error for invalid query, got nil")
	}

	t.Logf("Got expected error: %v", err)
}

// TestRealAPI_ComplexQuery tests a more complex query
func TestRealAPI_ComplexQuery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	client := New()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Query for nodes in a small area (example: around Berlin coordinates)
	// Using very small bounding box to keep query fast
	result, err := client.QueryContext(ctx, `
		[out:json];
		node(52.5,13.3,52.51,13.31);
		out 10;
	`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	t.Logf("Complex query returned %d elements", result.Count)
	t.Logf("Nodes: %d, Ways: %d, Relations: %d",
		len(result.Nodes), len(result.Ways), len(result.Relations))
}

// TestRealAPI_RateLimiting tests that rate limiting works with real endpoint
func TestRealAPI_RateLimiting(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Create client with maxParallel=1 to test rate limiting
	client := NewWithSettings(apiEndpoint, 1, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Make multiple requests - should be serialized
	numRequests := 3
	for i := 0; i < numRequests; i++ {
		_, err := client.QueryContext(ctx, `[out:json];node(1);out;`)
		if err != nil {
			t.Errorf("request %d failed: %v", i+1, err)
		}
		t.Logf("Completed request %d/%d", i+1, numRequests)
	}
}

// TestRealAPI_PackageFunction tests package-level function with real API
func TestRealAPI_PackageFunction(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use package-level QueryContext
	result, err := QueryContext(ctx, `[out:json];node(1);out;`)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if result.Count == 0 {
		t.Error("expected at least one element")
	}

	t.Logf("Package-level query returned %d elements", result.Count)
}
