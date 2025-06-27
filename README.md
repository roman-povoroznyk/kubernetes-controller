# Kubernetes Controller

A lightweight command-line tool for interacting with Kubernetes clusters with a high-performance HTTP server.

## Features

- **Pod Management**: Create, delete, and list pods in a Kubernetes cluster
- **Deployment Management**: Create, delete, and list deployments in a Kubernetes cluster
- **Deployment Informer**: Real-time monitoring of deployment events with structured logging
- **High-Performance HTTP Server**: FastHTTP-based server with health endpoint and informer integration
- **Structured Logging**: Detailed request logging with unique request IDs and event tracking
- **Graceful Shutdown**: Clean shutdown with configurable timeout
- **kubectl-like Output**: Familiar output format for Kubernetes operations
- **Flexible Authentication**: Support for both kubeconfig and in-cluster authentication

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

# Watch deployment events (with kubeconfig)
./k8s-ctrl watch deployment -n default

# Watch pod events (with kubeconfig)
./k8s-ctrl watch pod -n default

# Watch deployment events (in-cluster mode)
./k8s-ctrl watch deployment --in-cluster

# Watch with custom resync period
./k8s-ctrl watch deployment --resync-period 60s

# Watch pods in specific namespace
./k8s-ctrl watch pod -n kube-system
```

### HTTP Server

```bash
# Start HTTP server on port 8080 with deployment informer
./k8s-ctrl server --server-port 8080

# Start with detailed logging and custom namespace
./k8s-ctrl server --server-port 8080 --log-level debug --namespace kube-system

# Start with in-cluster authentication (for running inside Kubernetes)
./k8s-ctrl server --in-cluster --namespace default

# Start with in-cluster authentication and selective informers
./k8s-ctrl server --in-cluster --namespace default --enable-deployment-informer --no-enable-pod-informer

# Start with custom kubeconfig and only pod informer
./k8s-ctrl server --kubeconfig /path/to/config --enable-pod-informer --no-enable-deployment-informer
```

### API Endpoints

Once the HTTP server is running, you can access the following REST API endpoints:

#### Health Check
```bash
# Health check endpoint
curl http://localhost:8080/health
# Response: OK
```

#### Deployment Information
```bash
# Get deployment names only
curl http://localhost:8080/deployments/names
# Response: ["nginx", "api-server"]

# Get full deployment information
curl http://localhost:8080/deployments
# Response: Detailed deployment objects with status, replicas, etc.
```

#### Pod Information
```bash
# Get pod names only
curl http://localhost:8080/pods/names
# Response: ["nginx-7584b6f84c-vnbv8", "api-server-abc123-xyz"]

# Get full pod information
curl http://localhost:8080/pods
# Response: Detailed pod objects with status, node, containers, etc.
```

#### Individual Resource Endpoints
```bash
# Get specific deployment by name
curl http://localhost:8080/deployments/nginx
# Response: Detailed deployment object for 'nginx'

# Get specific pod by name
curl http://localhost:8080/pods/nginx-7584b6f84c-vnbv8
# Response: Detailed pod object for the specified pod
```

#### Response Format

All endpoints return JSON with structured data:

**Deployment Object:**
```json
{
  "name": "nginx",
  "namespace": "default",
  "replicas": 3,
  "ready": 3,
  "updated": 3,
  "available": 3,
  "age": "2h",
  "image": "nginx:latest",
  "labels": {
    "app": "nginx"
  }
}
```

**Pod Object:**
```json
{
  "name": "nginx-7584b6f84c-vnbv8",
  "namespace": "default",
  "phase": "Running",
  "ready": "1/1",
  "restarts": 0,
  "age": "2h",
  "image": "nginx:latest",
  "node": "minikube",
  "labels": {
    "app": "nginx",
    "pod-template-hash": "7584b6f84c"
  }
}
```

**Note:** API endpoints only return data from resources that are actively watched by the informers. Make sure the appropriate informers are enabled (--enable-deployment-informer, --enable-pod-informer) when starting the server.


### Version information

```bash
# Display version
./k8s-ctrl version
```

### Available Endpoints

- `GET /health` - Health check endpoint that returns "OK"
- `GET /` - Welcome page with a greeting message
- All other paths return a 404 Not Found response

The server automatically starts deployment and pod informers (configurable) that monitor Kubernetes events in real-time and log them with structured JSON format including:
- **Deployment events**: Creation, updates, and deletion events with replica counts, status changes, container images, generation and namespace details
- **Pod events**: Pod lifecycle events (creation, updates, deletion) with phase changes, readiness status, restart counts, and node assignments

## Development

### Makefile Commands

```bash
# Build the binary
make build

# Run all tests
make test

# Generate code coverage report
make coverage

# Run tests with envtest (recommended for Kubernetes controller tests)
make test

# Run tests with coverage and export to XML
make test-coverage

# Set up envtest environment
make envtest

# Format code
make format

# Run linter (requires golangci-lint)
make lint

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

# Run container with kubeconfig mounted
docker run -p 8080:8080 \
  -v ~/.kube/config:/root/.kube/config:ro \
  k8s-ctrl:latest server --kubeconfig /root/.kube/config
```

### Project Structure

```
.
├── cmd/                         # CLI commands
│   ├── kubernetes/              # Kubernetes interaction commands
│   │   ├── create.go            # Pod/Deployment creation command
│   │   ├── delete.go            # Pod/Deployment deletion command
│   │   ├── list.go              # Pod/Deployment listing command
│   │   └── watch.go             # Watch deployments and pods for events
│   ├── server/                  # HTTP server command
│   ├── root.go                  # Root CLI command
│   └── version.go               # Version command
├── internal/                    # Internal logic (not importable by external packages)
│   ├── informer/                # Kubernetes informers
│   │   ├── deployments_informer.go      # Deployment informer implementation
│   │   ├── deployments_informer_test.go # Deployment informer tests
│   │   ├── pods_informer.go             # Pod informer implementation
│   │   ├── pods_informer_test.go        # Pod informer tests
│   │   └── interface.go         # Informer interfaces and manager
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

## Environment Variables

### All command-line flags can also be set via environment variables:

- **K8S_CTRL_KUBECONFIG** - Path to kubeconfig file (equivalent to --kubeconfig)
- **K8S_CTRL_LOG_LEVEL** - Log level (equivalent to --log-level)
- **K8S_CTRL_SERVER_PORT** - HTTP server port (equivalent to --server-port)
- **K8S_CTRL_IN_CLUSTER** - Use in-cluster authentication (equivalent to --in-cluster)
- **K8S_CTRL_NAMESPACE** - Default namespace to watch (equivalent to --namespace)
- **K8S_CTRL_RESYNC_PERIOD** - Informer resync period (equivalent to --resync-period)
- **K8S_CTRL_ENABLE_DEPLOYMENT_INFORMER** - Enable deployment informer (equivalent to --enable-deployment-informer)
- **K8S_CTRL_ENABLE_POD_INFORMER** - Enable pod informer (equivalent to --enable-pod-informer)

## Key Components

- **Cobra CLI**: Command-line interface for user interaction
- **Distroless Container**: Minimal, secure container image
- **FastHTTP Server**: High-performance HTTP server with middleware support and informer integration
- **Kubernetes Client**: Client-go based Kubernetes API interactions with flexible authentication
- **Deployment Informer**: Real-time monitoring of Kubernetes deployment events
- **Request Logging**: Detailed logging with unique request IDs for traceability
- **Envtest Support**: Comprehensive testing with Kubernetes API server simulation

## License

This project is licensed under the MIT License - see the LICENSE file for details.
```
