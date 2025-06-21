# Build variables
APP_NAME := k6s
VERSION := $(shell git describe --tags 2>/dev/null || echo "v0.5.0-dev")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS := -ldflags "-X github.com/roman-povoroznyk/$(APP_NAME)/pkg/version.Version=$(VERSION) -X github.com/roman-povoroznyk/$(APP_NAME)/pkg/version.GitCommit=$(GIT_COMMIT) -X github.com/roman-povoroznyk/$(APP_NAME)/pkg/version.BuildDate=$(BUILD_DATE)"

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOCLEAN := $(GOCMD) clean
GOTEST := $(GOCMD) test
GOGET := $(GOCMD) get
GOMOD := $(GOCMD) mod
BINARY_DIR := bin
BINARY_NAME := $(APP_NAME)

# Docker parameters
DOCKER_IMAGE := $(APP_NAME)
DOCKER_TAG := $(VERSION)

.PHONY: all build clean test deps lint fmt vet help docker-build docker-run

## Default target
all: clean deps test build

## Build the binary
build:
	@echo "Building $(BINARY_NAME) version $(VERSION)..."
	@mkdir -p $(BINARY_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_DIR)/$(BINARY_NAME) .

## Build for production
build-prod: clean deps test
	@echo "Building $(BINARY_NAME) for production..."
	@mkdir -p $(BINARY_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_DIR)/$(BINARY_NAME) .

## Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BINARY_DIR)

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...

## Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## Install/update dependencies
deps:
	@echo "Installing dependencies..."
	$(GOMOD) tidy
	$(GOMOD) download

## Run linter
lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run

## Format code
fmt:
	@echo "Formatting code..."
	$(GOCMD) fmt ./...

## Run vet
vet:
	@echo "Running vet..."
	$(GOCMD) vet ./...

## Run quality checks
quality: fmt vet lint test

## Build Docker image
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

## Run Docker container
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG) server

## Run Trivy vulnerability scan
security-scan:
	@echo "Running Trivy security scan..."
	@which trivy > /dev/null || (echo "Please install Trivy: https://aquasecurity.github.io/trivy/" && exit 1)
	trivy fs .
	trivy image $(DOCKER_IMAGE):$(DOCKER_TAG)

## Install development tools
dev-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

## Run development server
dev:
	@echo "Starting development server..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) . && ./$(BINARY_NAME) server

## Show help
help:
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

## Show version info
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"
