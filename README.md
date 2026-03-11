<p align="center">
  <img src="https://raw.githubusercontent.com/kubernetes/kubernetes/master/logo/logo.svg" width="100" alt="Kubernetes Logo"/>
</p>

<h1 align="center">k8s-autoheal-operator</h1>

<p align="center">
  <em>A Kubernetes operator that automatically detects and heals failing pods by performing rolling restarts on their parent deployments.</em>
</p>

<p align="center">
  <a href="https://github.com/zeldebro/k8s-autoheal-operator/actions/workflows/ci.yml"><img src="https://github.com/zeldebro/k8s-autoheal-operator/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://github.com/zeldebro/k8s-autoheal-operator/actions/workflows/release.yml"><img src="https://github.com/zeldebro/k8s-autoheal-operator/actions/workflows/release.yml/badge.svg" alt="Release"></a>
  <a href="https://goreportcard.com/report/github.com/zeldebro/k8s-autoheal-operator"><img src="https://goreportcard.com/badge/github.com/zeldebro/k8s-autoheal-operator" alt="Go Report Card"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/License-Apache%202.0-blue.svg" alt="License"></a>
  <a href="https://pkg.go.dev/github.com/zeldebro/k8s-autoheal-operator"><img src="https://pkg.go.dev/badge/github.com/zeldebro/k8s-autoheal-operator.svg" alt="Go Reference"></a>
</p>

---

## 🚀 Overview

**k8s-autoheal-operator** is a lightweight Kubernetes operator that watches for unhealthy pods and automatically heals them by performing a rolling restart on their parent Deployment. It detects pods in the following failure states:

- **CrashLoopBackOff** — Container is repeatedly crashing
- **OOMKilled** — Container was terminated due to out-of-memory
- **Error** — Container exited with an error

When a failing pod is detected, the operator traces it back to its parent Deployment (via ReplicaSet owner references) and triggers a rolling restart by annotating the Deployment's pod template — the same mechanism as `kubectl rollout restart`.

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────┐
│                   Kubernetes Cluster                     │
│                                                         │
│  ┌──────────────┐    ┌──────────────┐    ┌───────────┐ │
│  │  Pod Watcher  │───▶│  Rate-Limited │───▶│  Healer   │ │
│  │  (Informer)   │    │    Queue      │    │  Worker   │ │
│  └──────────────┘    └──────────────┘    └───────────┘ │
│        │                                       │        │
│        ▼                                       ▼        │
│  Watches all pods                    Performs rolling    │
│  for failure states                  restart on parent   │
│  (CrashLoopBackOff,                 Deployment via      │
│   OOMKilled, Error)                 annotation update    │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Components

| Component | File | Description |
|-----------|------|-------------|
| **Cluster Connect** | `internal/clusterConnect.go` | Establishes connection to the cluster (in-cluster or kubeconfig fallback) |
| **Pod Watcher** | `internal/watcher.go` | Uses SharedInformer to watch pod status changes and detect failures |
| **Healer** | `internal/healer.go` | Traces failed pods to Deployments, queues them, and performs rolling restarts |
| **Manager** | `cmd/main.go` | Kubebuilder-based manager with health checks, metrics, and leader election |

## ✨ Features

- 🔍 **Real-time pod monitoring** — Uses Kubernetes informers for efficient, event-driven pod watching
- 🔄 **Automatic healing** — Performs rolling restart on Deployments with failing pods
- 🛡️ **Duplicate prevention** — Tracks restarted deployments to avoid restart loops
- ⚡ **Rate-limited queue** — Prevents overwhelming the API server with restart requests
- 📊 **Secure metrics** — Exposes Prometheus-compatible metrics with optional TLS and RBAC
- 🏥 **Health probes** — Built-in health and readiness endpoints
- 👑 **Leader election** — Supports HA deployments with leader election
- 🐳 **Minimal footprint** — Distroless container image for security and small size

## 📋 Prerequisites

- Go 1.24+
- Docker 17.03+
- kubectl v1.11.3+
- Access to a Kubernetes v1.11.3+ cluster
- (Optional) [Kind](https://kind.sigs.k8s.io/) for local development

## 🚀 Quick Start

### Run Locally (outside the cluster)

```bash
# Clone the repository
git clone https://github.com/zeldebro/k8s-autoheal-operator.git
cd k8s-autoheal-operator

# Ensure you have a valid kubeconfig
kubectl cluster-info

# Run the operator
make run
```

### Deploy to a Cluster

```bash
# Build and push the container image
make docker-build docker-push IMG=<your-registry>/k8s-autoheal-operator:latest

# Deploy to the cluster
make deploy IMG=<your-registry>/k8s-autoheal-operator:latest
```

### Verify it's Running

```bash
# Check the operator pod
kubectl get pods -n k8s-autoheal-operator-system

# Check logs
kubectl logs -n k8s-autoheal-operator-system -l control-plane=controller-manager -f
```

### Test the Auto-Healing

```bash
# Create a deployment that will crash
kubectl create deployment crasher --image=busybox -- /bin/sh -c "exit 1"

# Watch the operator logs — it will detect the CrashLoopBackOff and restart the deployment
kubectl logs -n k8s-autoheal-operator-system -l control-plane=controller-manager -f
```

## ⚙️ Configuration

The operator accepts the following command-line flags:

| Flag | Default | Description |
|------|---------|-------------|
| `--metrics-bind-address` | `0` | Address for the metrics endpoint (`:8443` for HTTPS, `:8080` for HTTP, `0` to disable) |
| `--health-probe-bind-address` | `:8081` | Address for health/readiness probes |
| `--leader-elect` | `false` | Enable leader election for HA deployments |
| `--metrics-secure` | `true` | Serve metrics over HTTPS |
| `--enable-http2` | `false` | Enable HTTP/2 for metrics and webhook servers |
| `--metrics-cert-path` | `""` | Directory containing metrics server TLS certificate |
| `--webhook-cert-path` | `""` | Directory containing webhook TLS certificate |

## 🛠️ Development

### Build

```bash
make build
```

### Run Tests

```bash
# Unit tests
make test

# End-to-end tests (requires Kind)
make test-e2e
```

### Lint

```bash
make lint
```

### Build Docker Image

```bash
make docker-build IMG=k8s-autoheal-operator:dev
```

### Multi-Architecture Build

```bash
make docker-buildx IMG=<your-registry>/k8s-autoheal-operator:latest
```

## 📦 Installation Methods

### Using kubectl

```bash
kubectl apply -f https://raw.githubusercontent.com/zeldebro/k8s-autoheal-operator/main/dist/install.yaml
```

### Build the Installer Yourself

```bash
make build-installer IMG=<your-registry>/k8s-autoheal-operator:latest
# Output: dist/install.yaml
kubectl apply -f dist/install.yaml
```

## 🗑️ Uninstall

```bash
# Remove the operator
make undeploy

# Remove CRDs (if any were installed)
make uninstall
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details on:

- How to submit issues and feature requests
- How to set up a development environment
- How to submit pull requests
- Code style and conventions

## 📜 Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code.

## 🔒 Security

If you discover a security vulnerability, please see our [Security Policy](SECURITY.md) for responsible disclosure guidelines.

## 📄 License

This project is licensed under the Apache License 2.0 — see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with [Kubebuilder](https://book.kubebuilder.io/)
- Powered by [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime)
- Inspired by the Kubernetes self-healing philosophy

