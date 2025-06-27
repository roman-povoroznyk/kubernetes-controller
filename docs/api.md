# Deployment API Documentation

## Overview

The k6s server provides REST API endpoints for accessing deployment resources when the informer is enabled. All deployment data is served from the informer's cache, providing fast and efficient access to Kubernetes deployment information.

## Base URL

```
http://localhost:8080/api/v1
```

## Authentication

Currently, no authentication is required. The server uses the same kubeconfig credentials as the CLI commands.

## Endpoints

### List Deployments

**GET** `/api/v1/deployments`

Returns a list of all deployments from the informer cache.

#### Query Parameters

- `namespace` (optional): Filter deployments by namespace

#### Response

```json
{
  "items": [
    {
      "name": "my-app",
      "namespace": "default",
      "replicas": 3,
      "ready": 2,
      "updated": 3,
      "available": 2,
      "age": "5m",
      "image": "nginx:1.21",
      "labels": {
        "app": "my-app",
        "tier": "frontend"
      }
    }
  ],
  "count": 1
}
```

#### Example Requests

```bash
# List all deployments
curl http://localhost:8080/api/v1/deployments

# List deployments in specific namespace
curl http://localhost:8080/api/v1/deployments?namespace=production

# Pretty print with jq
curl -s http://localhost:8080/api/v1/deployments | jq
```

### Get Deployment

**GET** `/api/v1/deployments/{name}`
**GET** `/api/v1/deployments/{namespace}/{name}`

Returns a single deployment by name. If namespace is not specified, "default" is assumed.

#### Response

```json
{
  "name": "my-app",
  "namespace": "default", 
  "replicas": 3,
  "ready": 2,
  "updated": 3,
  "available": 2,
  "age": "5m",
  "image": "nginx:1.21",
  "labels": {
    "app": "my-app",
    "tier": "frontend"
  }
}
```

#### Example Requests

```bash
# Get deployment from default namespace
curl http://localhost:8080/api/v1/deployments/my-app

# Get deployment from specific namespace
curl http://localhost:8080/api/v1/deployments/production/api-server
```

## Status Codes

- **200 OK**: Request successful
- **400 Bad Request**: Invalid request format
- **404 Not Found**: Deployment not found
- **405 Method Not Allowed**: HTTP method not supported
- **503 Service Unavailable**: Informer not configured or not ready

## Error Response Format

```json
{
  "error": "not found",
  "message": "Deployment my-app/production not found"
}
```

## Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Deployment name |
| `namespace` | string | Kubernetes namespace |
| `replicas` | integer | Desired number of replicas |
| `ready` | integer | Number of ready replicas |
| `updated` | integer | Number of updated replicas |
| `available` | integer | Number of available replicas |
| `age` | string | Time since deployment creation (e.g., "5m", "2h", "3d") |
| `image` | string | Container image of first container |
| `labels` | object | Deployment labels |

## Server Configuration

### Enable Informer

To enable the deployment API endpoints, start the server with the `--enable-informer` flag:

```bash
# Start server with informer (all namespaces)
k6s server --enable-informer

# Start server with informer for specific namespace
k6s server --enable-informer --namespace=production

# Custom port and resync period
k6s server --port 9090 --enable-informer --resync-period=5m
```

### Configuration File

You can also configure the informer through a configuration file:

```yaml
# ~/.k6s/k6s.yaml
informer:
  namespace: "production"
  resync_period: "5m"
  enable_custom_logic: true
  label_selector: "app=web"

log_level: "info"
```

```bash
# Start server with config file
k6s server --enable-informer --config ~/.k6s/k6s.yaml
```

## Examples

### Monitor Deployments

```bash
# Check if server is ready
curl http://localhost:8080/health

# List all deployments
curl -s http://localhost:8080/api/v1/deployments | jq '.items[] | {name, namespace, replicas, ready}'

# Monitor specific deployment
watch 'curl -s http://localhost:8080/api/v1/deployments/my-app | jq "{name, replicas, ready, available}"'

# Filter by namespace
curl -s "http://localhost:8080/api/v1/deployments?namespace=kube-system" | jq '.count'
```

### Integration with Other Tools

```bash
# Use with kubectl for comparison
kubectl get deployments --all-namespaces -o json | jq '.items[].metadata.name' | sort > kubectl-deployments.txt
curl -s http://localhost:8080/api/v1/deployments | jq '.items[].name' | sort > api-deployments.txt
diff kubectl-deployments.txt api-deployments.txt

# Export to file
curl -s http://localhost:8080/api/v1/deployments > deployments-backup.json

# Check deployment status
DEPLOYMENT="my-app"
STATUS=$(curl -s http://localhost:8080/api/v1/deployments/$DEPLOYMENT | jq '.ready == .replicas')
if [ "$STATUS" = "true" ]; then
  echo "Deployment $DEPLOYMENT is ready"
else
  echo "Deployment $DEPLOYMENT is not ready"
fi
```

## Performance Notes

- Data is served from the informer's in-memory cache, providing fast response times
- Cache is kept up-to-date with real-time Kubernetes events
- No direct API calls to Kubernetes API server for each request
- Suitable for high-frequency polling and monitoring applications

## Limitations

- Read-only access (no create, update, delete operations)
- Requires informer to be running and synced
- Only supports deployment resources (no pods, services, etc.)
- No authentication or authorization (inherits from kubeconfig)
