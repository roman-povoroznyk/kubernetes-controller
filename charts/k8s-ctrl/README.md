# k8s-ctrl Helm Chart

This Helm chart deploys the `k8s-ctrl` application—a lightweight Kubernetes controller and HTTP server—into your cluster.

## Features

- Configurable image repository, tag, and pull policy
- Scalable deployment (replicaCount)
- Customizable service type, port, and protocol
- Resource requests and limits
- Health and readiness probes
- Environment variables and extra args support
- Node selectors, tolerations, and affinity
- Optional RBAC (ClusterRole/Binding and ServiceAccount)
- Support for custom volumes and mounts

## Usage

### Install

```sh
helm install k8s-ctrl ./charts/k8s-ctrl \
  --set image.repository=ghcr.io/your-org/k8s-ctrl \
  --set image.tag=v1.2.3
```

### Upgrade

```sh
helm upgrade k8s-ctrl ./charts/k8s-ctrl \
  --set image.tag=v1.2.4
```

### Uninstall

```sh
helm uninstall k8s-ctrl
```

## License

MIT
