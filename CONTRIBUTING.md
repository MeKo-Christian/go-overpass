# Contributing to go-overpass

Thank you for your interest in contributing to go-overpass!

## Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/MeKo-Christian/go-overpass.git
   cd go-overpass
   ```

2. **Install Go 1.21+**
   - Download from [golang.org](https://golang.org/dl/)

3. **Run tests:**
   ```bash
   go test -v ./...
   ```

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test -v ./...

# Run with coverage
go test -v -cover ./...

# Run with race detector
go test -v -race ./...
```

### Integration Tests
Integration tests require network access to the real Overpass API:

```bash
go test -v -tags=integration ./...
```

### Benchmarks
```bash
go test -bench=. ./...
```

## Code Quality

### Linting
We use golangci-lint for code quality:

```bash
golangci-lint run
```

### Code Style
- Follow standard Go conventions
- Use `gofmt` for formatting (done automatically by most editors)
- Write clear, descriptive variable and function names
- Add comments for exported functions and types

## Making Changes

### Branch Strategy
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes
4. Write or update tests
5. Ensure all tests pass
6. Commit with clear messages

### Commit Messages
Write clear, concise commit messages:

```
Add context support to Query method

- Implement QueryContext with context.Context parameter
- Keep Query method for backward compatibility
- Update tests to cover context cancellation
```

### Pull Request Process
1. **Ensure all tests pass**
2. **Update documentation** if needed
3. **Add tests** for new features
4. **Update CHANGELOG.md** with your changes
5. **Push and create PR** with clear description

## Coding Guidelines

### Error Handling
- Use error wrapping with `fmt.Errorf("context: %w", err)`
- Return detailed error messages
- Don't panic in library code

### Context Usage
- Always accept `context.Context` for I/O operations
- Respect context cancellation
- Use `context.Background()` as default

### Testing
- Write table-driven tests when possible
- Test error cases
- Use meaningful test names
- Keep tests focused and simple

### Documentation
- Document all exported functions and types
- Include usage examples in documentation
- Update README for user-facing changes

## Questions or Problems?

- Open an issue for bugs or feature requests
- Use discussions for questions
- Be respectful and constructive

## License

By contributing, you agree that your contributions will be licensed under the MIT License.
