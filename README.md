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
make build
```

## Usage

### Kubernetes Operations

```bash
# List pods
./k8s-ctrl list pod

# List deployments
./k8s-ctrl list deployment

# List resources in specific namespace
./k8s-ctrl list deployment -n kube-system

# Create a pod
./k8s-ctrl create pod nginx-pod

# Create a deployment
./k8s-ctrl create deployment nginx-deployment

# Delete a pod
./k8s-ctrl delete pod nginx-pod

# Delete a deployment
./k8s-ctrl delete deployment nginx-deployment
```

### HTTP Server

```bash
# Start HTTP server on port 8080
./k8s-ctrl server --server-port 8080

# Start with detailed logging
./k8s-ctrl server --server-port 8080 --log-level debug
```

### Version information

```bash
# Display version
./k8s-ctrl version
```

### Available Endpoints

- `GET /health` - Health check endpoint that returns "OK"
- `GET /` - Welcome page with a greeting message
- All other paths return a 404 Not Found response

## Development

### Makefile Commands

```bash
# Build the binary
make build

# Run all tests
make test

# Generate code coverage report
make coverage

# Build Docker image
make docker-build

# Clean up artifacts
make clean
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run server tests
go test ./internal/server/...

# Run with code coverage
go test -cover ./...
```

### Docker Container


```bash
# Build Docker image
make docker-build

# Run container
docker run -p 8080:8080 k8s-ctrl:latest
```

### Project Structure

```
.
├── cmd/                         # CLI commands
│   ├── kubernetes/              # Kubernetes interaction commands
│   │   ├── create.go            # Pod/Deployment creation command
│   │   ├── delete.go            # Pod/Deployment deletion command
│   │   └── list.go              # Pod/Deployment listing command
│   ├── server/                  # HTTP server command
│   ├── root.go                  # Root CLI command
│   └── version.go               # Version command
├── internal/                    # Internal logic (not importable by external packages)
│   ├── kubernetes/              # Kubernetes operations
│   │   ├── pods.go              # Pod-related operations
│   │   ├── pods_test.go         # Pod operation tests
│   │   ├── deployments.go       # Deployment-related operations
│   │   ├── deployments_test.go  # Deployment operation tests
│   │   └── util.go              # Shared utility functions
│   └── server/                  # HTTP server
│       ├── middleware/
│       │   ├── logging.go       # Request logging middleware
│       │   └── logging_test.go  # Middleware tests
│       ├── handler.go           # HTTP request handlers
│       └── server.go            # FastHTTP server
│       └── server_test.go       # Server tests
├── .github/workflows/           # CI/CD pipelines
│   └── ci.yaml                  # GitHub Actions workflow
├── Dockerfile                   # Distroless container definition
├── Makefile                     # Build automation
└── main.go                      # Entry point
```

## Key Components

- **Cobra CLI**: Command-line interface for user interaction
- **Distroless Container**: Minimal, secure container image
- **FastHTTP Server**: High-performance HTTP server with middleware support
- **Kubernetes Client**: Client-go based Kubernetes API interactions
- **Request Logging**: Detailed logging with unique request IDs for traceability

## Environment Variables

### All command-line flags can also be set via environment variables:

- K8S_CTRL_KUBECONFIG - Path to kubeconfig file (equivalent to --kubeconfig)
- K8S_CTRL_LOG_LEVEL - Log level (equivalent to --log-level)
- K8S_CTRL_SERVER_PORT - HTTP server port (equivalent to --server-port)

## License

This project is licensed under the MIT License - see the LICENSE file for details.
