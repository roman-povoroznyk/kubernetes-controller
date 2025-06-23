# Kubernetes Controller

A lightweight command-line tool for interacting with Kubernetes clusters with a high-performance HTTP server.

## Features

- **Pod Management**: Create, delete, and list pods in a Kubernetes cluster
- **High-Performance HTTP Server**: FastHTTP-based server with health endpoint
- **Structured Logging**: Detailed request logging with unique request IDs
- **Graceful Shutdown**: Clean shutdown with configurable timeout
- **kubectl-like Output**: Familiar output format for Kubernetes operations

## Installation

```bash
# Clone the repository
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git

# Change to the project directory
cd kubernetes-controller

# Build the binary
go build -o k8s-ctrl main.go
```

## Usage

### Kubernetes Operations

```bash
# List pods
./k8s-ctrl list pod

# Create a pod
./k8s-ctrl create pod nginx-pod

# Delete a pod
./k8s-ctrl delete pod nginx-pod

# Set log level
./k8s-ctrl list pod --log-level debug
```

### HTTP Server

```bash
# Start HTTP server on port 8080
./k8s-ctrl server --server-port 8080

# Start with detailed logging
./k8s-ctrl server --server-port 8080 --log-level debug
```

### Available Endpoints

- `GET /health` - Health check endpoint that returns "OK"
- `GET /` - Welcome page with a greeting message
- All other paths return a 404 Not Found response

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run server tests
go test ./internal/server/...

# Run with code coverage
go test -cover ./...
```

### Project Structure

```
.
├── cmd/                     # CLI commands
│   ├── kubernetes/          # Kubernetes interaction commands
│   │   ├── create.go        # Pod creation command
│   │   ├── delete.go        # Pod deletion command
│   │   └── list.go          # Pod listing command
│   ├── server/              # HTTP server command
│   └── root.go              # Root CLI command
├── internal/                # Internal logic
│   ├── kubernetes/          # Kubernetes operations
│   │   └── pods.go          # Pod-related operations
│   └── server/              # HTTP server
│       ├── middleware/      # HTTP middleware
│       │   └── logging.go   # Request logging middleware
│       ├── handler.go       # HTTP request handlers
│       └── server.go        # FastHTTP server
└── main.go                  # Entry point
```

## Key Components

- **FastHTTP Server**: High-performance HTTP server with middleware support
- **Request Logging**: Detailed logging with unique request IDs for traceability
- **Kubernetes Client**: Client-go based Kubernetes API interactions
- **Cobra CLI**: Command-line interface for user interaction

## License

This project is licensed under the MIT License - see the LICENSE file for details.
