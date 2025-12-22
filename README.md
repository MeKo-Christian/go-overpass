# go-overpass

Go library for accessing the [Overpass API](http://wiki.openstreetmap.org/wiki/Overpass_API)

[![GoDoc](https://godoc.org/github.com/MeKo-Christian/go-overpass?status.svg)](https://godoc.org/github.com/MeKo-Christian/go-overpass)
[![Go Report Card](https://goreportcard.com/badge/github.com/MeKo-Christian/go-overpass)](https://goreportcard.com/report/github.com/MeKo-Christian/go-overpass)
[![CI](https://github.com/MeKo-Christian/go-overpass/workflows/CI/badge.svg)](https://github.com/MeKo-Christian/go-overpass/actions)

## Features

- **Simple, idiomatic Go API** - Clean and intuitive interface
- **Context support** - Cancellation and timeouts via `context.Context`
- **Automatic retry with exponential backoff** - Resilient queries with configurable retry logic
- **In-memory caching** - Optional response caching with TTL for faster repeated queries
- **Query builder** - Fluent API for constructing Overpass QL queries
- **Feature categorization** - Helper methods for classifying OSM elements by tags
- **Built-in rate limiting** - Respects server rate limits with configurable concurrency
- **Comprehensive error handling** - Detailed error messages and error wrapping
- **Full OpenStreetMap type support** - Nodes, Ways, Relations with all metadata
- **Zero external dependencies** - Only uses Go standard library
- **Well-tested** - Comprehensive test suite with extensive coverage

## Installation

```bash
go get github.com/MeKo-Christian/go-overpass
```

## Quick Start

### Basic Query

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/MeKo-Christian/go-overpass"
)

func main() {
    client := overpass.New()

    result, err := client.QueryContext(context.Background(),
        `[out:json];node(1);out;`)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d elements\n", result.Count)
}
```

### Query with Timeout

```go
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

result, err := client.QueryContext(ctx,
    `[out:json];relation(1673881);>>;out body;`)
if err != nil {
    log.Fatal(err)
}
```

### Custom Settings

```go
client := overpass.NewWithSettings(
    "http://api.openstreetmap.fr/oapi/interpreter/",
    5, // max 5 concurrent requests
    http.DefaultClient,
)
```

## Usage

### Client Creation

**Default client:**

```go
client := overpass.New()
// Uses overpass-api.de with rate limit of 1 concurrent request
```

**Custom client:**

```go
client := overpass.NewWithSettings(
    "https://custom-endpoint.com/api",
    3, // maxParallel
    customHTTPClient,
)
```

### Making Queries

**With context (recommended):**

```go
ctx := context.Background()
result, err := client.QueryContext(ctx, "[out:json];node(1);out;")
```

**Without context (deprecated but still works):**

```go
result, err := client.Query("[out:json];node(1);out;")
// Internally uses context.Background()
```

**Package-level function:**

```go
result, err := overpass.QueryContext(ctx, "[out:json];node(1);out;")
// Uses default client
```

### Overpass Turbo Macro Expansion (Subset)

The `turbo` subpackage provides a small, pure-Go preprocessor for common Overpass Turbo
macros so you can paste Turbo queries and run them against the Overpass API.

```go
import "github.com/MeKo-Christian/go-overpass/turbo"

res, err := turbo.Expand(`node({{bbox}});out;`, turbo.Options{
    BBox: &turbo.BBox{South: 52.5, West: 13.4, North: 52.51, East: 13.41},
})
if err != nil {
    log.Fatal(err)
}

result, err := overpass.QueryContext(ctx, res.Query)
```

Supported macros in this initial subset: `{{bbox}}`, `{{center}}`, `{{date}}`,
`{{date:<n unit>}}`, and custom shortcuts `{{key=value}}`.

If the query includes `{{data:overpass,server=...}}`, the parsed `Result` exposes
`EndpointOverride` so you can switch endpoints if desired. Use
`turbo.ApplyEndpointOverride` to prefer the override when present.

### Working with Results

```go
result, err := client.QueryContext(ctx, query)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Timestamp: %s\n", result.Timestamp)
fmt.Printf("Total elements: %d\n", result.Count)

// Access nodes
for id, node := range result.Nodes {
    fmt.Printf("Node %d: lat=%f, lon=%f\n", id, node.Lat, node.Lon)
    for k, v := range node.Tags {
        fmt.Printf("  %s=%s\n", k, v)
    }
}

// Access ways
for id, way := range result.Ways {
    fmt.Printf("Way %d: %d nodes\n", id, len(way.Nodes))
    if way.Bounds != nil {
        fmt.Printf("  Bounds: (%f,%f) to (%f,%f)\n",
            way.Bounds.Min.Lat, way.Bounds.Min.Lon,
            way.Bounds.Max.Lat, way.Bounds.Max.Lon)
    }
}

// Access relations
for id, relation := range result.Relations {
    fmt.Printf("Relation %d: %d members\n", id, len(relation.Members))
}
```

## Advanced Features

### Retry Logic with Exponential Backoff

The library automatically retries failed requests with exponential backoff (enabled by default):

```go
// Default retry configuration (3 retries with exponential backoff)
client := overpass.New()

// Custom retry configuration
client := overpass.NewWithRetry(
    "https://overpass-api.de/api/interpreter",
    2, // max parallel requests
    http.DefaultClient,
    overpass.RetryConfig{
        MaxRetries:        5,
        InitialBackoff:    2 * time.Second,
        MaxBackoff:        60 * time.Second,
        BackoffMultiplier: 2.0,
        Jitter:            true,
    },
)

// Disable retries
client.SetRetryConfig(overpass.RetryConfig{MaxRetries: 0})
```

**Retry behavior:**

- Automatically retries on server errors: 429, 500, 502, 503, 504
- Does not retry client errors: 400, 401, 403, 404
- Respects context cancellation during backoff waits
- Adds jitter to prevent thundering herd

### In-Memory Caching

Optional response caching with TTL (disabled by default):

```go
client := overpass.New()

// Enable caching
client.SetCacheConfig(overpass.CacheConfig{
    Enabled:    true,
    TTL:        5 * time.Minute,
    MaxEntries: 1000,
})

// Query (first call hits API, second call hits cache)
result1, _ := client.QueryContext(ctx, query)
result2, _ := client.QueryContext(ctx, query) // From cache

// Cache management
fmt.Printf("Cache size: %d\n", client.CacheSize())
client.ClearCache()

// Cleanup resources when done
defer client.Close()
```

**Cache features:**

- Thread-safe with automatic background cleanup
- Configurable TTL and maximum entries
- Simple FIFO eviction when max entries exceeded
- Cache key based on endpoint + query string

### Query Builder

Fluent API for constructing Overpass QL queries:

```go
// Build a query
query := overpass.NewQueryBuilder().
    Node().
    Way().
    BBox(52.5, 13.4, 52.51, 13.41).
    Tag("amenity", "restaurant").
    Tag("cuisine", "italian").
    OutputCenter().
    Timeout(60).
    Build()

// Execute with builder
result, err := client.QueryWithBuilder(ctx, query)

// Helper functions for common patterns
restaurants := overpass.FindRestaurants(52.5, 13.4, 52.51, 13.41)
highways := overpass.FindHighways(52.5, 13.4, 52.51, 13.41, "primary")
cafes := overpass.FindAmenity(52.5, 13.4, 52.51, 13.41, "cafe")

result, err := client.QueryContext(ctx, restaurants.Build())
```

**Query builder features:**

- Tag filtering (exact, exists, not equal, regex)
- Bounding box queries
- Multiple element types (node, way, relation)
- Output modes (body, geom, center, meta)
- Timeout configuration
- Helper functions for common patterns

### Feature Categorization

Helper methods for classifying OSM elements by their tags:

```go
result, err := client.QueryContext(ctx, query)

for _, node := range result.Nodes {
    // Get high-level category
    category := node.GetCategory() // "amenity", "transportation", "natural", etc.
    subcategory := node.GetSubcategory() // "restaurant", "primary", "tree", etc.
    name := node.GetName()

    fmt.Printf("%s: %s - %s\n", category, subcategory, name)

    // Category helpers
    if node.IsAmenity() && node.IsFoodRelated() {
        fmt.Println("Found a restaurant/cafe/bar")
    }

    if node.IsTransportation() && node.IsRoad() {
        highway := node.GetSubcategory()
        fmt.Printf("Found a %s road\n", highway)
    }

    // Tag utilities
    if node.HasTag("wheelchair") {
        accessibility := node.GetTag("wheelchair", "unknown")
        fmt.Printf("Wheelchair accessible: %s\n", accessibility)
    }

    if node.MatchesFilter("amenity", "restaurant") {
        fmt.Println("This is a restaurant")
    }
}
```

**Categorization features:**

- Recognize standard OSM tag categories
- Priority-based categorization (highway > building > amenity)
- Helper methods for common categories (food, education, healthcare)
- Tag utility methods (HasTag, GetTag, MatchesFilter)

## Rate Limiting

The library respects server rate limits using a semaphore-based approach:

```go
// Allow only 1 concurrent request (default)
client := overpass.New()

// Allow up to 5 concurrent requests
client := overpass.NewWithSettings(endpoint, 5, http.DefaultClient)
```

Requests are automatically queued and executed according to the `maxParallel` setting.

## Error Handling

The library provides detailed error information with error wrapping:

```go
result, err := client.QueryContext(ctx, query)
if err != nil {
    // Check for specific error types
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Query timed out")
    } else if errors.Is(err, context.Canceled) {
        log.Println("Query was cancelled")
    } else {
        log.Printf("Query failed: %v", err)
    }
    return
}
```

### Server Errors

HTTP errors from the Overpass API are wrapped in `ServerError`:

```go
var serverErr *overpass.ServerError
if errors.As(err, &serverErr) {
    fmt.Printf("Server returned %d: %s\n",
        serverErr.StatusCode,
        string(serverErr.Body))
}
```

## Examples

See the [examples/](./examples/) directory for more usage patterns:

- **[basic](./examples/basic/)** - Simple query demonstrating core functionality
- **[timeout](./examples/timeout/)** - Using context for timeout control
- **[custom](./examples/custom/)** - Custom endpoint and rate limiting configuration

## Migration from original repo

This fork introduces context support and breaking changes:

### Breaking Changes

1. **HTTPClient interface** now requires `Do(*http.Request)` method instead of `PostForm()`
   - `http.DefaultClient` already implements this
   - Custom implementations need to update

### Migration Steps

1. **Update import:**

   ```go
   // Old
   import "github.com/serjvanilla/go-overpass"

   // New
   import "github.com/MeKo-Christian/go-overpass"
   ```

2. **Use QueryContext (recommended):**

   ```go
   // Old (still works but deprecated)
   result, err := client.Query("[out:json];node(1);out;")

   // New (preferred)
   ctx := context.Background()
   result, err := client.QueryContext(ctx, "[out:json];node(1);out;")
   ```

3. **Update custom HTTPClient if needed:**

   ```go
   // Old interface
   type HTTPClient interface {
       PostForm(url string, data url.Values) (*http.Response, error)
   }

   // New interface
   type HTTPClient interface {
       Do(req *http.Request) (*http.Response, error)
   }
   ```

The old `Query()` method still works but is deprecated. It internally calls `QueryContext()` with `context.Background()`.

## Requirements

- Go 1.21 or higher
- Queries must include `[out:json]` for correct JSON response format

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details.

## Development

```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -v -cover ./...

# Run integration tests (requires network)
go test -v -tags=integration ./...

# Run benchmarks
go test -bench=. ./...

# Run linter
golangci-lint run
```

## License

MIT License - see [LICENSE](./LICENSE) for details

## Acknowledgments

Original library by [serjvanilla](https://github.com/serjvanilla/go-overpass)
