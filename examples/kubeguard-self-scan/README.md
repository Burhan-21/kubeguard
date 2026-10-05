# KubeGuard Self-Scan

This directory contains the production-grade deployment manifest for **KubeGuard itself**, used to verify that KubeGuard passes its own policy engine and enforces the standards it recommends.

## Running Self-Scan

```bash
kubeguard scan --profile production examples/kubeguard-self-scan/kubeguard-deployment.yaml
```

## Security & Reliability Profile

The self-scan manifest demonstrates complete compliance with KubeGuard policies:

- **Security Hardening**:
  - `runAsNonRoot: true` with non-root UID 10001
  - `allowPrivilegeEscalation: false`
  - `readOnlyRootFilesystem: true`
  - `capabilities.drop: ["ALL"]`
  - Pinned image tag with immutable SHA-256 digest
  - No host namespaces (`hostNetwork`, `hostPID`, `hostIPC` disabled)
  - No `hostPath` volumes (TLS secrets mounted via read-only projected volume)

- **Operational Reliability**:
  - 3 replicas for high availability
  - `RollingUpdate` strategy with zero downtime (`maxUnavailable: 0`)
  - Pod anti-affinity across node topology
  - Explicit CPU & memory resource requests and limits
  - Liveness probe, readiness probe, and startup probe configured
  - Graceful termination period of 30 seconds
