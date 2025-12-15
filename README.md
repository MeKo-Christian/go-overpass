# go-overpass

Go library for accessing the [Overpass API](http://wiki.openstreetmap.org/wiki/Overpass_API)

[![GoDoc](https://godoc.org/github.com/serjvanilla/go-overpass?status.svg)](https://godoc.org/github.com/serjvanilla/go-overpass/v2)
[![Go Report Card](https://goreportcard.com/badge/github.com/serjvanilla/go-overpass)](https://goreportcard.com/report/github.com/serjvanilla/go-overpass)
[![CI](https://github.com/MeKo-Christian/go-overpass/workflows/CI/badge.svg)](https://github.com/MeKo-Christian/go-overpass/actions)

## Features

- **Simple, idiomatic Go API** - Clean and intuitive interface
- **Context support** - Cancellation and timeouts via `context.Context`
- **Built-in rate limiting** - Respects server rate limits with configurable concurrency
- **Comprehensive error handling** - Detailed error messages and error wrapping
- **Full OpenStreetMap type support** - Nodes, Ways, Relations with all metadata
- **Zero external dependencies** - Only uses Go standard library
- **Well-tested** - 89.5% test coverage with comprehensive test suite

## Installation

```bash
go get github.com/serjvanilla/go-overpass/v2
```

## Quick Start

### Basic Query

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/serjvanilla/go-overpass/v2"
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

## Migration from v1

v2.0.0 introduces context support and breaking changes:

### Breaking Changes

1. **HTTPClient interface** now requires `Do(*http.Request)` method instead of `PostForm()`
   - `http.DefaultClient` already implements this
   - Custom implementations need to update

2. **Module path** changed to `/v2` suffix

### Migration Steps

1. **Update import:**
   ```go
   // Old
   import "github.com/serjvanilla/go-overpass"

   // New
   import "github.com/serjvanilla/go-overpass/v2"
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
