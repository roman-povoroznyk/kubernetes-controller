# k6s - Kubernetes Controller CLI

A simple CLI tool for monitoring Kubernetes deployments with real-time event logging.

## Features

- Real-time deployment monitoring
- Structured JSON logging
- Kubernetes controller-runtime integration
- Prometheus metrics endpoint

## Installation

```bash
# Install from source
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller/k6s
go build -o k6s .

# Install via Helm
helm repo add k6s https://roman-povoroznyk.github.io/kubernetes-controller
helm install k6s k6s/k6s
```

## Usage

### Basic Commands

```bash
# Show help and version
k6s --help
k6s version

# List deployments
k6s deployment list
k6s deployment list --all-namespaces

# Watch deployments for changes
k6s deployment list --watch

# Create/delete deployments
k6s deployment create my-app --image=nginx --replicas=3
k6s deployment delete my-app
```

### Controller

```bash
# Start controller (monitors all deployment events)
k6s controller start

# Start with custom settings
k6s controller start --metrics-port 9090 --namespace production

# Monitor metrics
curl http://localhost:8080/metrics
```

### Server (API endpoints)

```bash
# Start server with API
k6s server --enable-informer --port 8080

# API endpoints
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/deployments
```

## Configuration

### Environment Variables

```bash
export K6S_LOG_LEVEL=debug
export K6S_CONTROLLER_METRICS_PORT=9090

# Use different kubeconfig file
export KUBECONFIG=/path/to/your/kubeconfig
```

### Kubernetes Context

The `--context` flag refers to context names from your kubeconfig file:

```bash
# See available contexts
kubectl config get-contexts

# Example output:
# CURRENT   NAME       CLUSTER    AUTHINFO   NAMESPACE
# *         minikube   minikube   minikube   default
#           prod-k8s   prod       prod-user  production

# Use specific context
k6s controller start --context prod-k8s
```

## Development

```bash
# Run tests
go test -v ./...

# Build
go build -o k6s .
```

## License

MIT License