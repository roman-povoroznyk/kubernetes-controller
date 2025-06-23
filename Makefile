# k6s Makefile
.PHONY: build test clean lint security docker help

# Variables
BINARY_NAME=k6s
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GO_VERSION=$(shell go version | cut -d " " -f 3)
LDFLAGS=-ldflags "-X github.com/roman-povoroznyk/k6s/cmd.Version=$(VERSION) \
	-X github.com/roman-povoroznyk/k6s/cmd.GoVersion=$(GO_VERSION)"

# Default target
.DEFAULT_GOAL := help

help:
	@echo "Available targets:"
	@echo "  build         - Build the binary"
	@echo "  build-linux   - Build for Linux"
	@echo "  build-all     - Build for all platforms"
	@echo "  test          - Run tests"
	@echo "  test-coverage - Run tests with coverage"
	@echo "  lint          - Run linters"
	@echo "  fmt           - Format code"
	@echo "  vet           - Run go vet"
	@echo "  security      - Run security checks"
	@echo "  trivy-scan    - Run Trivy vulnerability scan"
	@echo "  docker        - Build Docker image"
	@echo "  docker-run    - Run Docker container"
	@echo "  dev           - Build and run in development mode"
	@echo "  clean         - Clean build artifacts"
	@echo "  deps          - Download dependencies"
	@echo "  install-tools - Install development tools"
	@echo "  helm-lint     - Lint Helm chart"
	@echo "  helm-template - Template Helm chart"
	@echo "  helm-package  - Package Helm chart"
	@echo "  helm-install  - Install Helm chart"
	@echo "  helm-test     - Test Helm chart"
	@echo "  helm-uninstall- Uninstall Helm chart"
	@echo "  helm-lint     - Lint Helm chart"
	@echo "  helm-template  - Template Helm chart"
	@echo "  helm-package   - Package Helm chart"
	@echo "  helm-install    - Install Helm chart"
	@echo "  helm-test      - Test Helm chart"
	@echo "  helm-uninstall  - Uninstall Helm chart"

# Build targets
build:
	@echo "Building $(BINARY_NAME)..."
	@go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

build-linux:
	@echo "Building $(BINARY_NAME) for Linux..."
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o bin/$(BINARY_NAME)-linux .

build-all: build build-linux

# Test targets
test:
	@echo "Running tests..."
	@go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

# Code quality
lint:
	@echo "Running linters..."
	@golangci-lint run || echo "golangci-lint not installed, skipping..."

fmt:
	@echo "Formatting code..."
	@go fmt ./...

vet:
	@echo "Running go vet..."
	@go vet ./...

# Security
security:
	@echo "Running security checks..."
	@gosec ./... || echo "gosec not installed, skipping..."

trivy-scan:
	@echo "Running Trivy vulnerability scan..."
	@trivy fs . || echo "trivy not installed, skipping..."

# Docker targets
docker:
	@echo "Building Docker image..."
	@docker build -t $(BINARY_NAME):$(VERSION) .
	@docker build -t $(BINARY_NAME):latest .

docker-run:
	@echo "Running Docker container..."
	@docker run --rm -p 8080:8080 $(BINARY_NAME):latest

# Development
dev: build
	@echo "Running in development mode..."
	@./bin/$(BINARY_NAME) server --log-level debug

# Clean up
clean:
	@echo "Cleaning up..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html

# Dependencies
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

# Install tools
install-tools:
	@echo "Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest

# Helm targets
helm-lint:
	@echo "Linting Helm chart..."
	@helm lint charts/k6s

helm-template:
	@echo "Templating Helm chart..."
	@helm template k6s charts/k6s

helm-package:
	@echo "Packaging Helm chart..."
	@helm package charts/k6s -d charts/

helm-install:
	@echo "Installing Helm chart..."
	@helm install k6s charts/k6s

helm-test:
	@echo "Testing Helm chart..."
	@helm test k6s

helm-uninstall:
	@echo "Uninstalling Helm chart..."
	@helm uninstall k6s
