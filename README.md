# k6s - Kubernetes Deployment Controller

[![Build Status](https://img.shields.io/badge/build-passing-green)](https://github.com/roman-povoroznyk/kubernetes-controller/actions)
[![Go Version](https://img.shields.io/badge/go-1.24+-blue)](https://golang.org/dl/)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![Helm Chart](https://img.shields.io/badge/helm-v0.2.0-blue)](charts/k6s)

**k6s** is Kubernetes controller for monitoring Deployment resources. It supports both single-cluster and multi-cluster modes with Prometheus metrics, health checks, and leader election.

## Features

- **Real-time Deployment monitoring** - track changes in Deployment resources
- **Multi-cluster support** - manage multiple Kubernetes clusters
- **Prometheus metrics** - built-in metrics for monitoring
- **Health checks** - readiness and liveness probes
- **Leader election** - high availability support
- **Structured logging** - JSON and text logging formats
- **Security-first** - non-root containers, RBAC automation

## Quick Start

### Using Helm

```bash
# Install the controller
helm install k6s-controller charts/k6s

# Check status
kubectl get pods -l app.kubernetes.io/name=k6s
```

### Using Docker

```bash
# Run locally
docker run -d \
  --name k6s-controller \
  -p 8080:8080 \
  -p 8081:8081 \
  k6s-controller:latest controller start

# Check health
curl http://localhost:8081/healthz
```

### From Source

```bash
# Clone and build
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller/k6s
go build -o k6s .

# Run
./k6s controller start
```

## Usage

### Single Cluster Mode

```bash
# Basic usage
k6s controller start

# With options
k6s controller start \
  --namespace production \
  --metrics-port 9090 \
  --log-level debug
```

### Multi-cluster Mode

```yaml
# clusters.yaml
multi_cluster:
  clusters:
    - name: prod-us
      kubeconfig: ~/.kube/prod-us
      namespace: production
    - name: prod-eu
      kubeconfig: ~/.kube/prod-eu
      namespace: production
```

```bash
k6s controller start --mode multi --config-file clusters.yaml
```

### Monitoring

```bash
# Health endpoints
curl http://localhost:8081/healthz    # Health check
curl http://localhost:8081/readyz     # Readiness check

# Metrics
curl http://localhost:8080/metrics    # Prometheus metrics
```

## Development

### Development Roadmap

#### Core Infrastructure
- [x] **Step 1**: Golang CLI application using cobra-cli
- [x] **Step 2**: zerolog for log levels - info, debug, trace, warn, error
- [x] **Step 3**: pflag with flags for logs level
- [x] **Step 3+**: Use Viper to add env vars
- [x] **Step 4**: fasthttp with cobra command "server" and flags for server port
- [x] **Step 4+**: Add http requests logging
- [x] **Step 5**: makefile, distroless dockerfile, github workflow and initial tests, Trivy vulnerabilities check

#### Kubernetes Integration
- [x] **Step 6**: k8s.io/client-go to list Kubernetes deployment resources in default namespace, auth via kubeconfig, list cli command
- [x] **Step 6+**: Add create/delete command
- [x] **Step 7**: k8s.io/client-go create list/watch informer for Kubernetes deployments, envtest unit tests
- [x] **Step 7+**: add custom logic function for update/delete events using informers cache search
- [x] **Step 7++**: use config to setup informers start configuration
- [x] **Step 8**: json api handler to request list deployment resources in informer cache storage

#### Controller Runtime
- [x] **Step 9**: sigs.k8s.io/controller-runtime and controller with logic to report in log each event received
- [x] **Step 9+**: multi-cluster informers, dynamically created informers
- [x] **Step 10**: controller mgr to control informer and controller, leader election with lease resource, flag to disable leader election, flag for mgr metrics port

#### Custom Resources & Platform Engineering
- [ ] **Step 11**: custom CRD Frontendpage with additional informer, controller with reconciliation logic
- [ ] **Step 11++**: add multi-project client configuration for management clusters
- [ ] **Step 12**: platform engineering integration based on Port.io API, handler for actions to CRUD custom resource
- [ ] **Step 12+**: Add Update action support for IDP and controller
- [ ] **Step 12++**: Discord notifications integration

#### MCP Integration & Authentication
- [ ] **Step 13**: github.com/mark3labs/mcp-go/mcp to create mcp server for api handlers as mcp tools, flag to specify mcp port
- [ ] **Step 13+**: Add delete/update MCP tool
- [ ] **Step 13++**: Add OIDC auth to MCP
- [ ] **Step 14**: JWT authentication and authorisation for api and mcp
- [ ] **Step 14+**: add JWT auth for MCP

#### Observability & Quality
- [ ] **Step 15**: basic OpenTelemetry code instrumentation
- [ ] **Step 15++**: 90% test coverage

### Local Development

```bash
# Clone and setup
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller/k6s

# Install dependencies
go mod download

# Run tests
go test ./...

# Build and run
go build -o k6s .
./k6s controller start
```

## Configuration

### Environment Variables

```bash
export K6S_LOG_LEVEL=debug
export K6S_CONTROLLER_METRICS_PORT=9090
export KUBECONFIG=/path/to/your/kubeconfig
```

### Configuration Files

```yaml
# ~/.k6s/config.yaml
log:
  level: "info"
  format: "json"

controller:
  single:
    namespace: ""
    metrics_port: 8080
    health_port: 8081
    leader_election:
      enabled: true
      id: "k6s-controller"
      namespace: "default"
```

## Changelog

### v0.10.0 (2025-07-07)

**Production Ready Release**

**Features:**
- Complete Helm chart with production-ready defaults
- Full Prometheus metrics integration
- Health and readiness endpoints
- Leader election for high availability
- Multi-cluster support with custom informers
- Structured logging with JSON/text formats
- Configuration validation and migration
- Security-first design (non-root, read-only)
- Graceful shutdown handling
- RBAC automation

**Testing:**
- Validated on Minikube
- Helm chart deployment verified
- All endpoints functional
- Multi-cluster scenarios tested

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- Issues: [GitHub Issues](https://github.com/roman-povoroznyk/kubernetes-controller/issues)
- Discussions: [GitHub Discussions](https://github.com/roman-povoroznyk/kubernetes-controller/discussions)
