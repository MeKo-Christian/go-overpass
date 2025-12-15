# AI Coding Agent Instructions for go-overpass

## Project Overview
go-overpass is a Go library for querying OpenStreetMap data via the Overpass API. It provides a simple, idiomatic Go interface with zero external dependencies, focusing on rate-limited HTTP requests and structured JSON response parsing.

## Architecture
- **Client** ([client.go](client.go)): Manages HTTP communication with semaphore-based rate limiting (default 1 concurrent request). Supports custom endpoints and parallelism via `NewWithSettings()`.
- **Query Processing** ([query.go](query.go)): Handles HTTP POST requests, JSON unmarshaling, and element linking. Transforms Overpass API responses into typed Go structs with pointer relationships.
- **Type System** ([types.go](types.go)): Defines OSM elements (Node, Way, Relation) sharing a common `Meta` struct. `Result` organizes elements in maps by ID with cross-references.
- **Helpers** ([helpers.go](helpers.go)): Lazy-initialization functions for Result maps to ensure elements exist before modification.

## Data Flow
1. User calls `QueryContext(ctx, "[out:json];...")` with OverpassQL query
2. Client acquires semaphore token for rate limiting
3. HTTP POST to endpoint with `data=query` form parameter
4. JSON response unmarshaled into internal structs, then transformed to public types
5. Elements linked by pointers (e.g., Way.Nodes references actual Node objects in Result.Nodes)
6. Result returned with timestamp, count, and organized element maps

## Key Patterns & Conventions
- **Queries must include `[out:json]`** for correct JSON parsing
- **Context-first**: Prefer `QueryContext(ctx, query)` over deprecated `Query(query)`
- **Error wrapping**: Use `fmt.Errorf("context: %w", err)` for error chains
- **Pointer linking**: Elements in Result are interconnected via pointers, not copies
- **Optional fields**: `Bounds` and `Geometry` may be nil if not requested
- **Rate limiting**: Semaphore channel controls concurrency; default allows 1 parallel request
- **Testing**: Table-driven tests with mock HTTP clients; integration tests use `-tags=integration`

## Developer Workflows
- **Build & Test**: `go test -v ./...` (unit), `go test -v -tags=integration ./...` (integration)
- **Coverage**: `go test -v -cover ./...`
- **Lint**: `golangci-lint run` (CI uses timeout 5m)
- **Benchmarks**: `go test -bench=. ./...`
- **CI Matrix**: Tests Go 1.21+ on Ubuntu/macOS/Windows with race detection

## Common Pitfalls
- Forgetting `[out:json]` in queries leads to unmarshaling errors
- Accessing nil `Bounds` without checking causes panics
- Using `Query()` instead of `QueryContext()` ignores cancellation/timeouts
- Assuming element pointers are independent; modifications affect shared references

## Examples
- **Basic query**: `client.QueryContext(ctx, "[out:json];node(1);out;")`
- **Custom client**: `overpass.NewWithSettings("https://custom-endpoint.com/api", 5, http.DefaultClient)`
- **Error handling**: `if errors.Is(err, context.DeadlineExceeded) { /* timeout */ }`
- **Accessing results**: `for id, node := range result.Nodes { fmt.Printf("%d: %.6f,%.6f", id, node.Lat, node.Lon) }`</content>
<parameter name="filePath">/mnt/projekte/Code/go-overpass/.github/copilot-instructions.md