.PHONY: help test test-coverage lint lint-fix fmt vet build clean example install-tools

# Default target
help:
	@echo "Available targets:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run all quality checks (gofmt, vet, staticcheck, misspell, gocyclo, errcheck)"
	@echo "  make lint-fix       - Fix auto-fixable lint issues and format code"
	@echo "  make fmt            - Format code"
	@echo "  make vet            - Run go vet"
	@echo "  make build          - Build the plugin"
	@echo "  make clean          - Clean build artifacts"
	@echo "  make example        - Run example application"
	@echo "  make install-tools  - Install development tools"

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run linter
lint:
	golangci-lint run

# Fix auto-fixable lint issues and format code
lint-fix:
	gofmt -s -w .
	golangci-lint run --fix

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Build the plugin
build:
	go build -v ./...

# Clean build artifacts
clean:
	rm -f coverage.out coverage.html
	go clean

# Run example application
example:
	cd examples/basic && go run main.go

# Install development tools
install-tools:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install github.com/gordonklaus/ineffassign@latest
	go install github.com/client9/misspell/cmd/misspell@latest
	go install github.com/kisielk/errcheck@latest
	go install github.com/fzipp/gocyclo/cmd/gocyclo@latest

# Add Go bin to PATH
GOPATH ?= $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

# Run all checks
check: fmt vet lint test

# Development setup
dev-setup: install-tools
	go mod download
	@echo "Development environment ready!"
