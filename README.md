# Kubernetes Controller CLI

A CLI utility for creating, listing, and deleting Pods in a Kubernetes cluster.

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
```

---

## Usage

> By default, the CLI uses the current context from `~/.kube/config`.

### Create a Pod

```sh
go run main.go create pod my-nginx --namespace=default
```

### List Pods

```sh
go run main.go list pod --namespace=default
```

### Delete a Pod

```sh
go run main.go delete pod my-nginx --namespace=default
```

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
