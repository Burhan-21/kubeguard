# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-10-05

### Added
- **Core Architecture**:
  - Unified, deterministic policy engine implemented in Go (`internal/policy/engine.go`).
  - Workload normalizer extracting PodSpec from `Deployment`, `StatefulSet`, `DaemonSet`, `Pod`, `Job`, and `CronJob`.
  - Streaming multi-document YAML parser supporting single files, directories, and standard input (`-`).
- **Policy Rules**:
  - 14 Security Rules: KG-SEC-001 (Privileged), KG-SEC-002 (Run as Root), KG-SEC-003 (Privilege Escalation), KG-SEC-004 (Host Network), KG-SEC-005 (Host PID/IPC), KG-SEC-006 (HostPath Mounts), KG-SEC-007 (Capabilities), KG-SEC-008 (Missing SecurityContext), KG-SEC-009 (Latest Tag), KG-SEC-010 (Missing Tag), KG-SEC-011 (Unpinned Digest), KG-SEC-012 (Plaintext Secrets), KG-SEC-013 (Automount ServiceAccount Token), KG-SEC-014 (Excessive RBAC).
  - 12 Reliability Rules: KG-REL-001 (Readiness Probe), KG-REL-002 (Liveness Probe), KG-REL-003 (Startup Probe), KG-REL-004 (Resource Requests), KG-REL-005 (Resource Limits), KG-REL-006 (Single Replica), KG-REL-007 (Deployment Strategy), KG-REL-008 (PodDisruptionBudget), KG-REL-009 (Topology / Anti-affinity), KG-REL-010 (Termination Grace Period), KG-REL-011 (NetworkPolicy), KG-REL-012 (External Service Exposure).
- **Reporters**:
  - ANSI color-coded Human terminal reporter grouped by resource.
  - Indented machine-readable JSON reporter.
  - SARIF v2.1.0 reporter for GitHub Security Code Scanning integration.
- **Admission Controller**:
  - Kubernetes ValidatingAdmissionWebhook v1 HTTPS server (`/validate`, `/healthz`, `/readyz`, `/metrics`).
  - Support for `audit`, `warn`, and `enforce` operational modes.
- **Policy Profiles**:
  - `policies/default.yaml`, `policies/development.yaml`, `policies/production.yaml`, and `policies/strict.yaml`.
- **Packaging & Delivery**:
  - Minimal multi-stage non-root Dockerfile.
  - Production-grade Helm chart with TLS, RBAC, NetworkPolicy, and ValidatingWebhookConfiguration.
  - GitHub Actions CI workflow with linting, race detection, vulnerability scanning, and SARIF upload.
  - Hardened self-scan manifest in `examples/kubeguard-self-scan/`.
