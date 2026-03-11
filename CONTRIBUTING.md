# Contributing to k8s-autoheal-operator

First off, thank you for considering contributing to k8s-autoheal-operator! 🎉

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Pull Request Process](#pull-request-process)
- [Style Guide](#style-guide)

## Code of Conduct

This project adheres to the [Contributor Covenant Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior by opening an issue.

## How Can I Contribute?

### 🐛 Reporting Bugs

Before creating bug reports, please check existing issues to avoid duplicates. When creating a bug report, include:

- **Kubernetes version** (`kubectl version`)
- **Operator version/commit**
- **Steps to reproduce** the behavior
- **Expected behavior** vs **actual behavior**
- **Logs** from the operator (`kubectl logs -n k8s-autoheal-operator-system -l control-plane=controller-manager`)

### 💡 Suggesting Features

Feature requests are welcome! Please open an issue with:

- A clear description of the feature
- The motivation / use case
- Any proposed implementation details

### 🔧 Code Contributions

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for your changes
5. Ensure all tests pass
6. Submit a pull request

## Development Setup

### Prerequisites

- Go 1.24+
- Docker 17.03+
- kubectl
- A Kubernetes cluster (or [Kind](https://kind.sigs.k8s.io/) for local development)

### Getting Started

```bash
# Clone your fork
git clone https://github.com/<your-username>/k8s-autoheal-operator.git
cd k8s-autoheal-operator

# Add upstream remote
git remote add upstream https://github.com/zeldebro/k8s-autoheal-operator.git

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run the operator locally
make run

# Run linter
make lint
```

### Running E2E Tests

```bash
# Requires Kind to be installed
make test-e2e
```

## Pull Request Process

1. **Update documentation** — If your change affects user-facing behavior, update the README or relevant docs.
2. **Add tests** — All new features and bug fixes should include tests.
3. **Follow the style guide** — Run `make fmt` and `make lint` before submitting.
4. **Write clear commit messages** — Use [Conventional Commits](https://www.conventionalcommits.org/) format:
   - `feat: add new healing strategy for StatefulSets`
   - `fix: prevent duplicate restart when pod recovers`
   - `docs: update architecture diagram`
   - `test: add e2e test for OOMKilled pods`
   - `chore: update controller-runtime to v0.24`
5. **Keep PRs focused** — One feature or fix per pull request.
6. **Ensure CI passes** — All checks must be green before merging.

## Style Guide

### Go Code

- Follow standard [Go conventions](https://go.dev/doc/effective_go)
- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch issues
- Use meaningful variable and function names
- Add comments for exported functions and types
- Keep functions small and focused

### Commit Messages

We follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `ci`

### Kubernetes Resources

- Follow Kubernetes [API conventions](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md)
- Use meaningful labels and annotations
- Prefer declarative configuration

## Questions?

Feel free to open an issue for any questions about contributing. We're happy to help! 🙂

