# Kubernetes Controller

[![CI](https://github.com/roman-povoroznyk/kubernetes-controller/actions/workflows/ci.yaml/badge.svg)](https://github.com/roman-povoroznyk/kubernetes-controller/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/roman-povoroznyk/kubernetes-controller)](https://goreportcard.com/report/github.com/roman-povoroznyk/kubernetes-controller)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A Kubernetes CLI tool and HTTP server for managing resources with real-time event monitoring using controller-runtime.

## Features

- **Resource Management**: Create, delete, and list pods and deployments
- **Real-time Monitoring**: Watch Kubernetes events with structured logging
- **Controller Runtime**: Production-ready event handling using `sigs.k8s.io/controller-runtime`
- **High Availability**: Leader election with automatic failover for production deployments
- **REST API**: Query resource information via HTTP endpoints with structured error handling
- **Multiple Authentication**: Supports kubeconfig files and in-cluster authentication
- **Observability**: Structured logging, metrics endpoint (/metrics), health checks (/health), and unique request IDs
- **Performance**: FastHTTP server with optimized handlers and connection pooling
- **Production Ready**: Comprehensive error handling, resource management, and deployment automation

## Installation

### From Source

```bash
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller
make build
```

### Using Docker

```bash
# Build image
make docker-build

# Run with mounted kubeconfig
docker run -p 8080:8080 -p 8081:8081 \
  -v ~/.kube/config:/root/.kube/config:ro \
  k8s-ctrl:latest server --kubeconfig /root/.kube/config

# Production run with leader election
docker run -p 8080:8080 -p 8081:8081 \
  -v ~/.kube/config:/root/.kube/config:ro \
  k8s-ctrl:latest server \
  --kubeconfig /root/.kube/config \
  --enable-leader-election \
  --metrics-port 8081
```

### Using Helm

```bash
# Deploy to Kubernetes
helm install k8s-ctrl ./charts/k8s-ctrl
```

## Usage

### CLI Commands

```bash
# Resource operations
./k8s-ctrl list pod -n default
./k8s-ctrl list deployment -n kube-system
./k8s-ctrl create pod my-pod
./k8s-ctrl create deployment my-app
./k8s-ctrl delete pod my-pod
./k8s-ctrl delete deployment my-app

# Watch events (with informers)
./k8s-ctrl watch deployment -n default
./k8s-ctrl watch pod --resync-period 60s

# Version information
./k8s-ctrl version
```
### HTTP Server

```bash
# Basic server
./k8s-ctrl server --server-port 8080

# With custom configuration
./k8s-ctrl server \
  --server-port 8080 \
  --log-level debug \
  --namespace kube-system \
  --enable-deployment-informer \
  --enable-pod-informer \
  --enable-leader-election \
  --metrics-port 8081

# In-cluster mode (for Kubernetes deployment)
./k8s-ctrl server --in-cluster --namespace default

# Development mode without leader election
./k8s-ctrl server --enable-leader-election=false
```

### Configuration

The application supports configuration via command-line flags, environment variables, or config files:

```bash
# Environment variables (K8S_CTRL_ prefix)
export K8S_CTRL_LOG_LEVEL=debug
export K8S_CTRL_SERVER_PORT=8080
export K8S_CTRL_NAMESPACE=production

# Command-line flags
./k8s-ctrl server --log-level debug --server-port 8080 --namespace production
```

**Available Configuration:**
- `--kubeconfig`: Path to kubeconfig file
- `--log-level`: Logging level (trace, debug, info, warn, error)
- `--server-port`: HTTP server port (default: 8080)
- `--namespace`: Default namespace to watch (default: default)
- `--in-cluster`: Use in-cluster authentication
- `--enable-deployment-informer`: Enable deployment informer (default: true)
- `--enable-pod-informer`: Enable pod informer (default: true)
- `--enable-controller`: Enable controller-runtime controller (default: true)
- `--resync-period`: Informer resync period (default: 10m)
- `--enable-leader-election`: Enable leader election for high availability (default: true)
- `--leader-election-namespace`: Namespace for leader election lease (default: default)
- `--metrics-port`: Port for controller manager metrics (default: 8081)

## Event Monitoring

### Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Health check for liveness/readiness probes |
| `/metrics` | GET | Prometheus metrics (port 8081) |
| `/deployments` | GET | List all deployments |
| `/deployments/names` | GET | Get deployment names only |
| `/deployments/{name}` | GET | Get specific deployment |
| `/pods` | GET | List all pods |
| `/pods/names` | GET | Get pod names only |
| `/pods/{name}` | GET | Get specific pod |

The application includes a production-ready Kubernetes controller that watches for Deployment events and logs them with structured format:

```json
{
  "level": "info",
  "namespace": "default",
  "name": "nginx-deployment",
  "component": "deployment-controller",
  "created": "2025-06-27T12:00:00Z",
  "replicas": 3,
  "ready_replicas": 0,
  "available_replicas": 0,
  "generation": "123",
  "message": "Deployment CREATE event received"
}
```

**Event Types:**
- **CREATE**: New deployments (ObservedGeneration = 0)
- **UPDATE**: Modified deployments (ObservedGeneration > 0)
- **DELETE**: Removed deployments (resource not found)

**Controller Manager Integration:**
- **Unified Lifecycle**: All controllers and informers managed by single `controller-runtime` manager
- **Leader Election**: Automatic coordination for high availability deployments
- **Graceful Shutdown**: Proper cleanup and resource management on termination
- **Metrics Collection**: Built-in Prometheus metrics on dedicated endpoint

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│  CLI Commands   │    │  HTTP Server     │    │   Controller    │
│                 │    │                  │    │   Runtime       │
│                 │    │                  │    │   Manager       │
│ • list          │    │ • REST API       │    │ • Event Watch   │
│ • create        │────│ • Health checks  │────│ • Reconcile     │
│ • delete        │    │ • Middleware     │    │ • Leader Elect  │
│ • watch         │    │ • Informers      │    │ • Metrics       │
└─────────────────┘    └──────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                        ┌──────────────────┐
                        │   Kubernetes     │
                        │   API Server     │
                        └──────────────────┘
```

**Key Components:**
- **CLI**: Cobra-based command interface with kubectl-like operations
- **HTTP Server**: FastHTTP server with middleware and real-time endpoints
- **Controller Manager**: `sigs.k8s.io/controller-runtime` manager with leader election
- **Controllers**: Deployment controller with structured event handling
- **Informers**: Real-time event watchers integrated with manager lifecycle
- **Observability**: Structured logging, metrics endpoint, and health monitoring

## Development

### Prerequisites

- Go 1.24+
- Kubernetes cluster (local or remote)
- kubectl configured

### Building and Testing

```bash
# Build the application
make build

# Run all tests
make test

# Run only unit tests
make test-unit

# Generate coverage report
make test-coverage

# Set up test environment
make envtest

# Code formatting and linting
make format
make lint

# Clean artifacts
make clean
```

### Project Structure

```
.
├── cmd/                                       # CLI commands
│   ├── kubernetes/                            # Kubernetes interaction commands
│   │   ├── create.go                          # Pod/Deployment creation command
│   │   ├── delete.go                          # Pod/Deployment deletion command
│   │   ├── list.go                            # Pod/Deployment listing command
│   │   └── watch.go                           # Watch deployments and pods for events
│   ├── server/                                # HTTP server command
│   │   ├── server.go                          # Server command implementation
│   │   └── server_test.go                     # Server command tests
│   ├── root.go                                # Root CLI command
│   └── version.go                             # Version command
├── internal/                                  # Internal logic (not importable by external packages)
│   ├── controller/                            # Controller-runtime implementations
│   │   ├── deployment_controller.go           # Deployment controller implementation
│   │   └── informer_manager.go                # Manager-integrated informers
│   ├── informer/                              # Kubernetes informers
│   │   ├── deployment_informer.go             # Deployment informer implementation
│   │   ├── pod_informer.go                    # Pod informer implementation
│   │   └── interface.go                       # Informer interfaces and manager
│   ├── kubernetes/                            # Kubernetes operations
│   │   ├── deployments.go                     # Deployment-related operations
│   │   ├── pods.go                            # Pod-related operations
│   │   └── util.go                            # Shared utility functions
│   └── server/                                # HTTP server
│       ├── middleware/                        # HTTP middleware components
│       │   └── logging.go                     # Request logging middleware
│       ├── handler.go                         # HTTP request handlers
│       └── server.go                          # FastHTTP server implementation
├── charts/                                    # Helm charts
│   └── k8s-ctrl/                              # Kubernetes deployment chart
├── .github/workflows/                         # CI/CD pipelines
│   └── ci.yaml                                # GitHub Actions workflow
├── Dockerfile                                 # Distroless container definition
├── Makefile                                   # Build automation
└── main.go                                    # Entry point
```

### Running Tests

The project includes comprehensive testing:

- **Unit Tests**: Fast tests with mocked dependencies
- **Integration Tests**: Tests with real Kubernetes API using envtest
- **Controller Tests**: Controller-runtime specific testing

```bash
# Quick unit tests
go test -short ./...

# Full test suite with integration tests
make test

# Watch mode for development
make test | grep -v "PASS"
```

### Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Run linting: `make lint`
6. Commit your changes: `git commit -m 'Add amazing feature'`
7. Push to the branch: `git push origin feature/amazing-feature`
8. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
