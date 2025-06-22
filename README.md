# Kubernetes Controller CLI (k8s-ctrl)

A lightweight CLI utility for creating, listing, and deleting Pods in a Kubernetes cluster.

---

## Requirements

- Go 1.21+
- A running Kubernetes cluster (e.g., [minikube](https://minikube.sigs.k8s.io/docs/))
- A valid kubeconfig with access to your cluster

---

## Installation

```sh
git clone https://github.com/roman-povoroznyk/kubernetes-controller.git
cd kubernetes-controller
go mod tidy
go build -o k8s-ctrl main.go
```

---

## Usage

> By default, the CLI uses the current context from `~/.kube/config`.

### Create a Pod

```sh
./k8s-ctrl create pod nginx
```

### List Pods

```sh
./k8s-ctrl list pod
```

### Delete a Pod

```sh
./k8s-ctrl delete pod nginx
```

### Namespace Selection

Specify a namespace with the `-n` or `--namespace` flag:

```sh
./k8s-ctrl list pod --namespace=kube-system
./k8s-ctrl create pod nginx -n custom-namespace
```

### Log Level Control

Control the log level using the `--log-level` or `-l` flag:

```sh
./k8s-ctrl list pod --log-level=debug
./k8s-ctrl create pod nginx -l trace
```

Available log levels: `trace`, `debug`, `info`, `warn`, `error`

### Kubeconfig Path

Specify a custom kubeconfig path:

```sh
./k8s-ctrl list pod --kubeconfig=/path/to/custom/config
```

### Environment Variables

The CLI supports configuration through environment variables with the `K8S_CTRL_` prefix:

```sh
# Set log level
export K8S_CTRL_LOG_LEVEL=debug

# Set custom kubeconfig
export K8S_CTRL_KUBECONFIG=/path/to/kubeconfig

# Run command using environment settings
./k8s-ctrl list pod
```

**Note**: Command-line flags take precedence over environment variables.

---

## Project Structure

```
kubernetes-controller/
├── cmd/                # CLI commands (create, delete, list)
├── internal/kubeops/   # Business logic for Kubernetes API operations
├── main.go             # Entry point
├── go.mod, go.sum      # Go dependencies
```

---

## Testing

To run unit tests:

```sh
go test ./internal/kubeops
```

For verbose output:

```sh
go test -v ./internal/kubeops
```

---

## Extending

- To add support for other resources (e.g., Deployment, StatefulSet), add corresponding functions in `internal/kubeops/` and subcommands in `cmd/`.
- To work with a different cluster, change the context using:
  ```sh
  kubectl config use-context <context-name>
  ```

---

## License

MIT License
