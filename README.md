# k6s - Kubernetes Controller CLI

A Golang CLI application for managing Kubernetes deployments with informers and controllers.

## Features

- CLI interface built with Cobra
- Structured logging with zerolog
- FastHTTP server for API endpoints
- Kubernetes client-go integration
- Real-time deployment informers
- JSON API for deployment data
- Controller-runtime based controllers
- Leader election support
- Metrics endpoint

## Installation

```bash
# Install via Go
go install github.com/roman-povoroznyk/kubernetes-controller/k6s@latest

# Install via Helm
helm repo add k6s https://roman-povoroznyk.github.io/kubernetes-controller
helm install k6s k6s/k6s

# Install specific chart version
helm install k6s k6s/k6s --version 0.1.0

# Install from source
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller/k6s
make build
```

## Usage

```bash
# Show help
k6s --help

# Show version
k6s version

# List deployments (multiple ways to do it)
k6s deployment list
k6s deployment list --all-namespaces

# Watch deployments for changes (shows logs for events)
k6s deployment list --watch
k6s deployment list --watch --namespace=prod

# Watch with custom logic for detailed change analysis
k6s deployment list --watch --custom-logic
k6s deployment list --watch --custom-logic --namespace=prod

# Create deployments
k6s deployment create my-app --image=nginx --replicas=3
k6s deployment create api --image=httpd:alpine --replicas=1 --namespace=prod

# Delete deployments
k6s deployment delete my-app
k6s deployment delete api --namespace=prod

# Start server
k6s server
k6s server --port 9090

# Use environment variables
K6S_LOG_LEVEL=debug k6s version
K6S_SERVER_PORT=8081 k6s server

# Deploy to Kubernetes with Helm
helm install k6s ./charts/k6s
kubectl port-forward svc/k6s 8080:8080
curl http://localhost:8080/health
```

## Custom Logic for Deployment Analysis

The `--custom-logic` flag provides advanced analysis of deployment changes when watching events. It uses the informer's cache to compare previous and current states and provides detailed insights about what changed.

### Features

- **Detailed Change Analysis**: Tracks specific fields that changed (replicas, image, labels, resources, strategy)
- **Cache-based Comparison**: Uses informer cache to compare old vs new deployment states
- **Rich Logging**: Provides structured logs with change details including old/new values
- **Event Context**: Shows creation timestamps, generation changes, and cache status

### Usage Examples

```bash
# Basic watch mode (shows standard deployment events)
k6s deployment list --watch

# Advanced watch with custom logic analysis
k6s deployment list --watch --custom-logic

# Custom logic with specific namespace
k6s deployment list --watch --custom-logic --namespace=production

# Watch all namespaces with custom analysis
k6s deployment list --watch --custom-logic --all-namespaces
```

### Custom Logic Output

When using `--custom-logic`, you'll see additional structured logs for each deployment event:

```json
// ADD event
{"level":"info","namespace":"default","name":"my-app","replicas":3,"handler":"custom_logic","message":"Deployment added with custom analysis"}

// UPDATE event with detailed changes
{"level":"debug","namespace":"default","name":"my-app","change_type":"spec","field":"replicas","old_value":3,"new_value":5,"description":"Replicas changed from 3 to 5","message":"Deployment change detail"}
{"level":"debug","namespace":"default","name":"my-app","change_type":"spec","field":"containers[0].image","old_value":"nginx:1.14","new_value":"nginx:1.16","description":"Container my-app image changed from nginx:1.14 to nginx:1.16","message":"Deployment change detail"}
{"level":"info","namespace":"default","name":"my-app","handler":"custom_logic","changes_count":2,"change_types":["spec","spec"],"change_fields":["replicas","containers[0].image"],"message":"Deployment updated with custom analysis"}

// DELETE event
{"level":"info","namespace":"default","name":"my-app","handler":"custom_logic","cache_status":"not_found","message":"Deployment deleted with custom analysis"}
```

## Development Roadmap

### Core Infrastructure
- [x] **Step 1**: Golang CLI application using cobra-cli
- [x] **Step 2**: zerolog for log levels - info, debug, trace, warn, error
- [x] **Step 3**: pflag with flags for logs level
- [x] **Step 3+**: Use Viper to add env vars
- [x] **Step 4**: fasthttp with cobra command "server" and flags for server port
- [x] **Step 4+**: Add http requests logging
- [x] **Step 5**: makefile, distroless dockerfile, github workflow and initial tests, Trivy vulnerabilities check

### Kubernetes Integration
- [x] **Step 6**: k8s.io/client-go to list Kubernetes deployment resources in default namespace, auth via kubeconfig, list cli command
- [x] **Step 6+**: Add create/delete command, refactor command structure to kubectl-like deployment subcommands
- [x] **Step 7**: k8s.io/client-go create list/watch informer for Kubernetes deployments, envtest unit tests
- [x] **Step 7+**: add custom logic function for update/delete events using informers cache search
- [ ] **Step 7++**: use config to setup informers start configuration
- [ ] **Step 8**: json api handler to request list deployment resources in informer cache storage

### Controller Runtime
- [ ] **Step 9**: sigs.k8s.io/controller-runtime and controller with logic to report in log each event received
- [ ] **Step 9+**: multi-cluster informers, dynamically created informers
- [ ] **Step 10**: controller mgr to control informer and controller, leader election with lease resource, flag to disable leader election, flag for mgr metrics port

### Custom Resources & Platform Engineering
- [ ] **Step 11**: custom CRD Frontendpage with additional informer, controller with reconciliation logic
- [ ] **Step 11++**: add multi-project client configuration for management clusters
- [ ] **Step 12**: platform engineering integration based on Port.io API, handler for actions to CRUD custom resource
- [ ] **Step 12+**: Add Update action support for IDP and controller
- [ ] **Step 12++**: Discord notifications integration

### MCP Integration & Authentication
- [ ] **Step 13**: github.com/mark3labs/mcp-go/mcp to create mcp server for api handlers as mcp tools, flag to specify mcp port
- [ ] **Step 13+**: Add delete/update MCP tool
- [ ] **Step 13++**: Add OIDC auth to MCP
- [ ] **Step 14**: JWT authentication and authorisation for api and mcp
- [ ] **Step 14+**: add JWT auth for MCP

### Observability & Quality
- [ ] **Step 15**: basic OpenTelemetry code instrumentation
- [ ] **Step 15++**: 90% test coverage

## Development

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...

# Run specific package tests
go test -v ./pkg/kubernetes/

# Run with coverage
go test -v -race -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Troubleshooting

#### envtest Issues on macOS

If you encounter `etcd: no such file or directory` errors when running integration tests:

```bash
# Install setup-envtest tool
go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

# Download and setup required binaries
setup-envtest use -p path

# Set environment variable and run tests
export KUBEBUILDER_ASSETS="$(setup-envtest use -p path)"
go test -v ./pkg/kubernetes/
```

The integration tests use envtest to run a local Kubernetes API server for testing informers and controllers.

## License

MIT License - see [LICENSE](LICENSE) file for details.
