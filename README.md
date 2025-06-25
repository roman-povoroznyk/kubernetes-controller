# k6s - Kubernetes Controller CLI

A powerful and lightweight CLI tool for managing Kubernetes deployments with advanced controller functionality.

## Features

- 🚀 **CLI Interface**: Built with Cobra for intuitive command-line operations
- 📝 **Structured Logging**: Advanced logging with multiple levels and environment-specific configs
- ⚙️ **Configuration Management**: Flexible configuration with Viper supporting files and environment variables
- 🌐 **HTTP Server**: FastHTTP-based server with request logging and health endpoints
- 🏷️ **Build-time Versioning**: Dynamic version information injected at build time
- ⚓ **Kubernetes Integration**: Native k8s.io/client-go integration for deployment management
- 🔧 **CRUD Operations**: Create, read, update, and delete Kubernetes deployments

## Architecture

The project follows a modular architecture with clear separation of concerns:

- **cmd/**: CLI commands and flags
- **pkg/**: Core business logic
  - **config/**: Configuration management with Viper
  - **logger/**: Structured logging with Zerolog
  - **server/**: HTTP server with FastHTTP
  - **version/**: Build-time version management
  - **kubernetes/**: Kubernetes client and operations
- **configs/**: Optional configuration files
- **examples/**: Demo and testing code

## Quick Start

```bash
# Build the CLI
go build -o k6s

# Run k6s
./k6s

# Check version
./k6s --version

# List deployments
./k6s list deployments

# List deployments in specific namespace
./k6s list deployments -n kube-system

# List deployments in all namespaces
./k6s list deployments --all-namespaces

# Create a deployment
./k6s create deployment nginx --image=nginx:latest --replicas=3 --port=80

# Delete a deployment
./k6s delete deployment nginx

# Start HTTP server
./k6s server --port 8080

# Test server endpoints
curl http://localhost:8080/
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/info

# Build with version information
go build -ldflags "-X github.com/roman-povoroznyk/k6s/pkg/version.Version=0.2.0-dev -X github.com/roman-povoroznyk/k6s/pkg/version.GitCommit=$(git rev-parse --short HEAD)" -o k6s
```

## Kubernetes Commands

k6s provides intuitive commands for managing Kubernetes deployments:

### List Deployments

```bash
# List deployments in default namespace
./k6s list deployments

# List deployments in specific namespace
./k6s list deployments --namespace kube-system
./k6s list deployments -n kube-system

# List deployments across all namespaces
./k6s list deployments --all-namespaces

# Output in JSON format
./k6s list deployments -o json
```

### Create Deployments

```bash
# Basic deployment
./k6s create deployment my-app --image=nginx:latest

# Deployment with custom replicas
./k6s create deployment my-app --image=nginx:latest --replicas=3

# Deployment with port and health checks
./k6s create deployment my-app --image=nginx:latest --replicas=3 --port=80

# Deployment in specific namespace
./k6s create deployment my-app --image=nginx:latest --namespace my-namespace
```

### Delete Deployments

```bash
# Delete deployment from default namespace
./k6s delete deployment my-app

# Delete deployment from specific namespace
./k6s delete deployment my-app --namespace my-namespace
```

### Kubeconfig Support

```bash
# Use specific kubeconfig file
./k6s list deployments --kubeconfig=/path/to/kubeconfig

# Use current context from default kubeconfig
./k6s list deployments

# In-cluster authentication (when running inside pod)
./k6s list deployments
```

### Watch Deployments

Monitor Kubernetes deployment events in real-time using informers:

```bash
# Watch deployment events in all namespaces
./k6s watch deployments

# Watch deployment events in specific namespace
./k6s watch deployments --namespace kube-system
./k6s watch deployments -n kube-system

# Watch with custom resync period
./k6s watch deployments --resync 60s

# Use specific kubeconfig
./k6s watch deployments --kubeconfig=/path/to/kubeconfig
```

The watch command uses Kubernetes informers to monitor ADD, UPDATE, and DELETE events on deployments and logs them with structured logging. This is useful for:

- **Monitoring**: Real-time visibility into deployment changes
- **Debugging**: Understanding deployment lifecycle events
- **Automation**: Building event-driven workflows
- **Compliance**: Auditing deployment modifications

## HTTP API Endpoints

k6s provides RESTful API endpoints for accessing deployment information from the informer cache:

### API Endpoints

```bash
# Health check - verify API and cache status
curl http://localhost:8080/api/v1/health

# List all deployments from cache
curl http://localhost:8080/api/v1/deployments

# Get specific deployment from cache
curl "http://localhost:8080/api/v1/deployment?namespace=default&name=my-app"

# Application info
curl http://localhost:8080/api/v1/info
```

### API Response Format

The API returns JSON responses with structured deployment information:

```json
{
  "items": [
    {
      "name": "my-app",
      "namespace": "default",
      "labels": {"app": "my-app"},
      "replicas": 3,
      "ready_replicas": 3,
      "updated_replicas": 3,
      "available_replicas": 3,
      "image": "nginx:latest",
      "creation_timestamp": "2025-06-21T10:00:00Z",
      "status": "Ready"
    }
  ],
  "total": 1
}
```

### Cache Benefits

- **High Performance**: Responses served from local cache, no API server calls
- **Real-time Updates**: Cache automatically updated via informer events
- **Reduced Load**: Minimizes load on Kubernetes API server
- **Consistent View**: Provides consistent snapshot of deployment state

## Logging System

k6s includes advanced structured logging with environment-specific configurations:

### Development Mode (Default)
- Pretty console output with emojis
- Debug level and above
- Human-readable timestamps
- Color-coded log levels

### Production Mode
- JSON format for log aggregation
- Info level and above
- RFC3339 timestamps
- Structured fields for parsing

```bash
# Set environment for production logging
export K6S_APP_ENV=production
./k6s
```

## Configuration

k6s supports multiple configuration sources in order of precedence:

1. **Environment Variables** (prefixed with `K6S_`)
2. **Configuration Files** (YAML format, optional)
3. **Default Values**

### Environment Variables

```bash
export K6S_APP_ENV=production        # Application environment
export K6S_LOG_LEVEL=debug           # Log level
```

### Configuration File

Create `configs/config.yaml` (optional):

```yaml
app:
  name: k6s
  env: development

log:
  level: info
```

**Note**: The `configs/` directory is useful but not required. The application works with defaults and environment variables alone.

## Versioning

Version information is injected at build time:

```bash
# Standard build (uses default version)
go build -o k6s

# Build with version information
go build -ldflags "-X github.com/roman-povoroznyk/k6s/pkg/version.Version=0.2.0-dev -X github.com/roman-povoroznyk/k6s/pkg/version.GitCommit=$(git rev-parse --short HEAD) -X github.com/roman-povoroznyk/k6s/pkg/version.BuildDate=$(date -u +%Y-%m-%dT%H:%M:%SZ)" -o k6s

# Check version
./k6s --version
```

## Current Implementation

✅ **Step 1**: Basic CLI structure with Cobra
✅ **Step 2**: Zerolog integration with environment-specific logging
✅ **Step 3**: Environment variables and configuration with Viper
✅ **Step 4**: FastHTTP server with request logging
✅ **Step 5**: Production deployment
✅ **Step 6**: Kubernetes client-go integration
✅ **Step 7**: Informers and controllers
⏳ **Step 8**: REST API endpoints
⏳ **Step 9**: Controller runtime
⏳ **Step 10**: Leader election and metrics

## Architecture

The project follows a modular architecture with clear separation of concerns:

- **cmd/**: CLI commands and flags
- **pkg/**: Core business logic
- **examples/**: Demo and testing code

## Development

```bash
# Install dependencies
go mod tidy

# Run tests
go test ./...

# Build
go build -o k6s
```

## License

MIT License - see LICENSE file for details.
# README update
# README update
