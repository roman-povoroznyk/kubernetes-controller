# k8s - Kubernetes Controller

[![Go Report Card](https://goreportcard.com/badge/github.com/roman-povoroznyk/k8s)](https://goreportcard.com/report/github.com/roman-povoroznyk/k8s)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Release](https://img.shields.io/github/release/roman-povoroznyk/k8s.svg)](https://github.com/roman-povoroznyk/k8s/releases)

A production-ready Kubernetes controller and CLI tool for managing Kubernetes resources with advanced features including multi-cluster support, business rules validation, and comprehensive monitoring.

## Features

- **CLI Interface**: Command-line tool for managing Kubernetes deployments
- **HTTP API**: RESTful API with OpenAPI specification
- **Controller Runtime**: Built on controller-runtime framework for robust Kubernetes integration
- **Multi-cluster Support**: Manage resources across multiple Kubernetes clusters
- **Business Rules**: Custom validation and business logic for deployments
- **Informer Pattern**: Efficient resource watching and caching
- **Leader Election**: High availability with leader election
- **Monitoring**: Prometheus metrics and health checks
- **Structured Logging**: JSON structured logging with Zerolog
- **Configuration**: Environment-based configuration with Viper
- **Security**: Built-in security best practices and scanning

## Quick Start

### Prerequisites

- Go 1.21 or later
- Kubernetes cluster (local or remote)
- kubectl configured

### Installation

```bash
# Clone the repository
git clone https://github.com/roman-povoroznyk/k8s.git
cd k8s

# Build the binary
make build

# Or run directly
go run main.go --help
```

### Usage

#### CLI Commands

```bash
# List deployments
./k8s list

# Create a deployment
./k8s create --name my-app --image nginx:latest --replicas 3

# Start the HTTP server
./k8s server --port 8080

# Start the controller
./k8s controller

# Watch deployments
./k8s watch
```

#### HTTP API

Start the server:
```bash
./k8s server --port 8080
```

Available endpoints:
- `GET /health` - Health check
- `GET /metrics` - Prometheus metrics
- `GET /api/v1/deployments` - List deployments
- `POST /api/v1/deployments` - Create deployment
- `DELETE /api/v1/deployments/{name}` - Delete deployment

## Configuration

The application can be configured using environment variables:

```bash
# Server configuration
export SERVER_PORT=8080
export SERVER_HOST=0.0.0.0

# Kubernetes configuration
export KUBECONFIG=/path/to/kubeconfig
export NAMESPACE=default

# Logging
export LOG_LEVEL=info
export LOG_FORMAT=json

# Multi-cluster
export CLUSTERS_CONFIG=/path/to/clusters.yaml
```

## Development

### Prerequisites

- Go 1.21+
- Docker
- Kubernetes cluster (minikube, kind, or remote)

### Building

```bash
# Build binary
make build

# Build Docker image
make docker-build

# Run tests
make test

# Run linter
make lint

# Generate mocks
make generate
```

### Running locally

```bash
# Start minikube
minikube start

# Run the controller
go run main.go controller

# In another terminal, test CLI
go run main.go list
go run main.go create --name test-app --image nginx --replicas 2
```

## Docker

### Build and run

```bash
# Build image
docker build -t k8s:latest .

# Run CLI
docker run --rm k8s:latest list

# Run server
docker run -p 8080:8080 k8s:latest server
```

## Architecture

The controller follows the standard Kubernetes controller pattern:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Client    │    │   HTTP API      │    │   Controller    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                        ┌─────────────────┐
                        │   Kubernetes    │
                        │      API        │
                        └─────────────────┘
```

### Components

- **CLI**: Command-line interface using Cobra
- **HTTP API**: RESTful API using FastHTTP
- **Controller**: Kubernetes controller using controller-runtime
- **Informer**: Watches Kubernetes resources
- **Business Rules**: Custom validation logic
- **Multi-cluster**: Support for multiple Kubernetes clusters

## Monitoring

### Prometheus Metrics

The controller exposes Prometheus metrics at `/metrics`:

- `k8s_deployments_total` - Total number of deployments
- `k8s_controller_reconcile_duration` - Reconciliation duration
- `k8s_api_requests_total` - HTTP API requests count
- `k8s_business_rules_violations` - Business rules violations

### Health Checks

Health check endpoint at `/health` returns:
- Controller status
- Kubernetes connectivity
- Leader election status

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Code Style

- Follow Go best practices
- Use `gofmt` and `golint`
- Add tests for new features
- Update documentation

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- Create an issue on GitHub
- Check the [documentation](docs/)
- Review [examples](examples/)

