# go-overpass justfile

# Install development dependencies (formatters and linters)
setup-deps:
    #!/bin/bash
    echo "Installing development dependencies..."

    # Install treefmt (required for formatting)
    command -v treefmt >/dev/null 2>&1 || { echo "Installing treefmt..."; curl -fsSL https://github.com/numtide/treefmt/releases/download/v2.1.1/treefmt_2.1.1_linux_amd64.tar.gz | sudo tar -C /usr/local/bin -xz treefmt; }

    # Install prettier (Node.js formatter for markdown)
    command -v prettier >/dev/null 2>&1 || { echo "Installing prettier..."; npm install -g prettier || echo "Prettier installation failed - npm not found. Please install Node.js/npm manually."; }

    # Install gofumpt (Go formatter)
    command -v gofumpt >/dev/null 2>&1 || { echo "Installing gofumpt..."; go install mvdan.cc/gofumpt@latest; }

    # Install gci (Go import formatter)
    command -v gci >/dev/null 2>&1 || { echo "Installing gci..."; go install github.com/daixiang0/gci@latest; }

    # Install golangci-lint (Go linter)
    command -v golangci-lint >/dev/null 2>&1 || { echo "Installing golangci-lint..."; curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.61.0; }

    echo "Development dependencies installation complete!"
    echo "Note: Ensure $(go env GOPATH)/bin is in your PATH for Go-based tools"

# Default target
default: test

# Format code using treefmt
fmt:
    treefmt --allow-missing-formatter

# Run linter
lint:
    golangci-lint run --config ./.golangci.toml --timeout 2m

# Run linter (with fix)
lint-fix:
    golangci-lint run --config ./.golangci.toml --timeout 2m --fix

# Run tests
test:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -cover ./...

# Build (verify compilation)
build:
    go build

# Run all checks
check: check-formatted lint test check-tidy

# Check if go.mod is tidy
check-tidy:
    #!/bin/bash
    set -e
    cp go.mod go.mod.bak
    cp go.sum go.sum.bak
    go mod tidy
    if ! diff -q go.mod go.mod.bak || ! diff -q go.sum go.sum.bak; then
        rm go.mod.bak go.sum.bak
        echo "Error: go.mod or go.sum is not tidy. Run 'go mod tidy'"
        exit 1
    fi
    rm go.mod.bak go.sum.bak

# Check if code is formatted
check-formatted:
    #!/bin/bash
    set -e
    # Create a temporary directory for comparison
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Copy current state
    cp -r . "$TMP_DIR/"

    # Format in temp directory
    cd "$TMP_DIR"
    treefmt --allow-missing-formatter > /dev/null 2>&1 || true

    # Compare
    cd - > /dev/null
    if ! diff -r --exclude='.git' --exclude='*.bak' . "$TMP_DIR/"; then
        echo "Error: Code is not formatted. Run 'just fmt'"
        exit 1
    fi

# Run go vet
vet:
    go vet ./...
