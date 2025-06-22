# k8s Makefile
# Production-ready build and development workflow

APP_NAME = k8s
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME = $(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT_HASH = $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Go parameters
GOCMD = go
GOBUILD = $(GOCMD) build
GOCLEAN = $(GOCMD) clean
GOTEST = $(GOCMD) test
GOGET = $(GOCMD) get
GOMOD = $(GOCMD) mod

# Build flags
LDFLAGS = -ldflags "\
	-X github.com/roman-povoroznyk/k8s/cmd.Version=$(VERSION) \
	-X github.com/roman-povoroznyk/k8s/cmd.BuildTime=$(BUILD_TIME) \
	-X github.com/roman-povoroznyk/k8s/cmd.CommitHash=$(COMMIT_HASH)"

# Binary output
BINARY = ./bin/$(APP_NAME)

# Default target
.PHONY: all
all: clean lint test build

# Build the application
.PHONY: build
build:
	@echo "Building $(APP_NAME) version $(VERSION)..."
	@mkdir -p bin
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BINARY) ./main.go
	@echo "Build complete: $(BINARY)"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@echo "Clean complete"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	@echo "Tests complete"

# Run tests with coverage report
.PHONY: test-coverage
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...
	@echo "Format complete"

# Lint code
.PHONY: lint
lint: fmt
	@echo "Running linter..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}
	golangci-lint run --timeout=5m
	@echo "Lint complete"

# Tidy modules
.PHONY: tidy
tidy:
	@echo "Tidying modules..."
	$(GOMOD) tidy
	@echo "Tidy complete"

# Install dependencies
.PHONY: deps
deps:
	@echo "Installing dependencies..."
	$(GOMOD) download
	@echo "Dependencies installed"

# Run the application
.PHONY: run
run: build
	@echo "Running $(APP_NAME)..."
	$(BINARY)

# Run the server
.PHONY: run-server
run-server: build
	@echo "Running $(APP_NAME) server..."
	$(BINARY) server --log-level debug

# Development watch (requires entr)
.PHONY: watch
watch:
	@echo "Starting development watch..."
	@command -v entr >/dev/null 2>&1 || { \
		echo "Please install entr for watch functionality"; \
		echo "macOS: brew install entr"; \
		echo "Linux: apt-get install entr"; \
		exit 1; \
	}
	find . -name "*.go" | entr -r make run

# Create release
.PHONY: release
release: clean lint test
	@echo "Creating release $(VERSION)..."
	@mkdir -p releases
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o releases/$(APP_NAME)-linux-amd64 ./main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o releases/$(APP_NAME)-darwin-amd64 ./main.go
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o releases/$(APP_NAME)-darwin-arm64 ./main.go
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o releases/$(APP_NAME)-windows-amd64.exe ./main.go
	@echo "Release $(VERSION) created in releases/"

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all         - Clean, lint, test, and build"
	@echo "  build       - Build the application"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  test-coverage - Run tests with coverage report"
	@echo "  fmt         - Format code"
	@echo "  lint        - Run linter"
	@echo "  tidy        - Tidy modules"
	@echo "  deps        - Install dependencies"
	@echo "  run         - Build and run the application"
	@echo "  run-server  - Build and run the server"
	@echo "  watch       - Development watch mode"
	@echo "  release     - Create multi-platform release"
	@echo "  help        - Show this help"
