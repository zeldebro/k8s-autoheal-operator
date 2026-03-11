# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| latest  | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in k8s-autoheal-operator, please report it responsibly.

### How to Report

**Please do NOT open a public GitHub issue for security vulnerabilities.**

Instead, please report security vulnerabilities by emailing the maintainers or by using [GitHub's private vulnerability reporting](https://github.com/zeldebro/k8s-autoheal-operator/security/advisories/new).

### What to Include

- A description of the vulnerability
- Steps to reproduce the issue
- Potential impact
- Suggested fix (if any)

### Response Timeline

- **Acknowledgment**: Within 48 hours
- **Initial assessment**: Within 1 week
- **Fix timeline**: Depends on severity, but we aim for:
  - Critical: 1-3 days
  - High: 1-2 weeks
  - Medium/Low: Next release cycle

### Scope

The following are in scope:

- The operator binary and its behavior
- Container image security
- RBAC permissions (over-privileged access)
- Kubernetes API interactions
- Dependencies with known CVEs

## Security Best Practices for Users

When deploying this operator:

1. **Use RBAC**: Deploy with least-privilege RBAC roles (provided in `config/rbac/`)
2. **Enable secure metrics**: Use `--metrics-secure=true` (default) to serve metrics over HTTPS
3. **Use leader election**: Enable `--leader-elect` in production for HA
4. **Pin image tags**: Use specific image tags, not `latest`, in production
5. **Network policies**: Apply the provided network policies from `config/network-policy/`
6. **Scan images**: Regularly scan the operator image for vulnerabilities

