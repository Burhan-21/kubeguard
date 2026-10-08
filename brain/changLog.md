# Changelog

**Author**: Shaikh Mohammed Burhan (GitHub: [Burhan-21](https://github.com/Burhan-21) / [NexusForge21](https://github.com/NexusForge21))

All notable changes to KubeGuard will be documented in this file.

## [0.1.0] - 2026-10-05

### Added
- **Core Engine & Architecture**:
  - Deterministic evaluation engine (`internal/policy/engine.go`) shared across CLI, CI, and Admission Webhook.
  - Multi-document YAML streaming parser (`internal/parser/parser.go`).
  - Workload normalizer (`internal/normalizer/normalizer.go`) supporting standard workloads and non-workload objects.
- **Rule Implementations**:
  - Full suite of 14 security rules (`KG-SEC-001` through `KG-SEC-014`).
  - Full suite of 12 reliability rules (`KG-REL-001` through `KG-REL-012`).
  - Positive and negative unit test coverage for every rule.
- **Reporting System**:
  - Human terminal reporter, JSON reporter, and SARIF v2.1.0 reporter.
- **Kubernetes Admission Controller**:
  - `AdmissionReview v1` webhook handler (`internal/admission/handler.go`) supporting `audit`, `warn`, and `enforce` modes.
  - TLS HTTP server (`internal/admission/server.go`) with health and ready probes.
- **Packaging & Delivery**:
  - Production Helm chart (`charts/kubeguard/`) with least-privilege RBAC and NetworkPolicy.
  - Minimal non-root Dockerfile based on distroless debian12.
  - GitHub Actions CI workflow (`.github/workflows/ci.yml`).
  - KubeGuard self-scan validation (`examples/kubeguard-self-scan/`).
  - Automated cross-platform release pipeline (`.github/workflows/release.yml`) for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`.
  - SHA-256 cryptographic checksums (`checksums.txt`) and GitHub build provenance attestations.
  - Container image repository published on GitHub Container Registry (`ghcr.io/burhan-21/kubeguard:v0.1.0`).
