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
go install github.com/roman-povoroznyk/k6s@latest
```

## Usage

```bash
# Show help
k6s --help

# Show version
k6s version

# Start server
k6s server
k6s server --port 9090

# Use environment variables
K6S_LOG_LEVEL=debug k6s version
K6S_SERVER_PORT=8081 k6s server
```

## Development Roadmap

### Core Infrastructure
- [x] **Step 1**: Golang CLI application using cobra-cli
- [x] **Step 2**: zerolog for log levels - info, debug, trace, warn, error
- [x] **Step 3**: pflag with flags for logs level
- [x] **Step 3+**: Use Viper to add env vars
- [x] **Step 4**: fasthttp with cobra command "server" and flags for server port
- [ ] **Step 4+**: Add http requests logging
- [ ] **Step 5**: makefile, distroless dockerfile, github workflow and initial tests, Trivy vulnerabilities check

### Kubernetes Integration
- [ ] **Step 6**: k8s.io/client-go to list Kubernetes deployment resources in default namespace, auth via kubeconfig, list cli command
- [ ] **Step 6+**: Add create/delete command
- [ ] **Step 7**: k8s.io/client-go create list/watch informer for Kubernetes deployments, envtest unit tests
- [ ] **Step 7+**: add custom logic function for update/delete events using informers cache search
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

## License

MIT License - see [LICENSE](LICENSE) file for details.
