.PHONY: help test test-coverage lint lint-fix fmt vet build clean example install-tools
GOLANGCI_LINT_VERSION := v2.12.2

# Default target
help:
	@echo "Available targets:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run golangci-lint (bundles staticcheck, errcheck, govet, gocyclo, misspell)"
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
	@if ! golangci-lint --version 2>/dev/null | grep -qE 'version v?2\.'; then \
		echo "  Installing golangci-lint $(GOLANGCI_LINT_VERSION)..."; \
		GOWORK=off go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION); \
	fi

# Add Go bin to PATH
GOPATH ?= $(shell go env GOPATH)
export PATH := $(GOPATH)/bin:$(PATH)

# Run all checks
check: fmt vet lint test

# Development setup
dev-setup: install-tools
	go mod download
	@echo "Development environment ready!"

# ----------------------------
# Security
# ----------------------------
# Pinned so a local scan and a CI scan judge the same code the same way; an
# unpinned scanner turns a green build red on someone else's machine.
GOVULNCHECK_VERSION := v1.8.0
GITLEAKS_VERSION := v8.30.1

.PHONY: security security-tools security-sast security-vuln security-secrets

security-tools:
	@if ! command -v govulncheck >/dev/null 2>&1; then \
		echo "  Installing govulncheck $(GOVULNCHECK_VERSION)..."; \
		GOWORK=off go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION); \
	fi
	@if ! command -v gitleaks >/dev/null 2>&1; then \
		echo "  Installing gitleaks $(GITLEAKS_VERSION)..."; \
		GOWORK=off go install github.com/zricethezav/gitleaks/v8@$(GITLEAKS_VERSION); \
	fi

# gosec also runs as part of `make lint`; this target isolates it so a security
# regression is readable without the rest of the linter output.
security-sast:
	@echo "[INFO] gosec (static analysis)..."
	@GOWORK=off $$(go env GOPATH)/bin/golangci-lint run --enable-only=gosec ./...

security-vuln: security-tools
	@echo "[INFO] govulncheck (known CVEs, incl. stdlib)..."
	@GOWORK=off $$(go env GOPATH)/bin/govulncheck ./...

security-secrets: security-tools
	@echo "[INFO] gitleaks (secret scan)..."
	@$$(go env GOPATH)/bin/gitleaks dir . --no-banner --redact

security: security-sast security-vuln security-secrets
	@echo "[INFO] All security checks passed!"
