# Architecture Decision Records (ADRs)

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## ADR-001: Why Go
- **Context**: Need a fast, statically typed language native to Kubernetes.
- **Decision**: Use Go (1.23+).
- **Reason**: Kubernetes ecosystem compatibility, fast compilation, easy static binaries, and native client-go support.

## ADR-002: Why CLI-first
- **Context**: Tool needs to serve developers and automated systems.
- **Decision**: Build the CLI as the first citizen.
- **Reason**: Enables rapid local testing, immediate CI/CD integration, and easy debugging before cluster deployment.

## ADR-003: Why deterministic rules
- **Context**: Deciding between AI-evaluated rules vs hardcoded logic.
- **Decision**: Rules must be deterministic and hardcoded.
- **Reason**: Security gates must be consistent. AI can hallucinate or drift, causing unpredictable pipeline blocks.

## ADR-004: Why one shared policy engine
- **Context**: CLI and Admission webhook process data differently.
- **Decision**: Abstract input into a single normalized structure fed to one engine.
- **Reason**: Prevents logic drift where CLI allows a manifest but Admission blocks it.

## ADR-005: Why ValidatingAdmissionWebhook
- **Context**: Need to enforce rules at the cluster level.
- **Decision**: Use standard ValidatingAdmissionWebhook.
- **Reason**: Native K8s approach. Mutating webhook is avoided to prevent hidden side-effects that cause drift between git and cluster.

## ADR-006: Why SARIF
- **Context**: Need standard output for CI platforms.
- **Decision**: Support SARIF.
- **Reason**: Native integration with GitHub Advanced Security and other SAST visualization tools.

## ADR-007: Why no AI in core decisions
- **Context**: AI is popular for security tools.
- **Decision**: AI is excluded from the *decision* path.
- **Reason**: See ADR-003. Violates the strict requirement for testability and determinism.

## ADR-008: Why security + reliability rules
- **Context**: Many tools only do security (Trivy) or only reliability (Polaris).
- **Decision**: Combine both under a "production readiness" umbrella.
- **Reason**: Missing resource limits (reliability) can cause node exhaustion (security/availability issue). They are fundamentally linked.

## ADR-009: Why Helm
- **Context**: Needs to be deployed into clusters.
- **Decision**: Provide an official Helm chart.
- **Reason**: Industry standard. Easily handles the complex TLS injection required for Webhooks.

## ADR-010: Admission failurePolicy
- **Context**: Webhook outage could brick a cluster.
- **Decision**: Provide user configuration for `Fail` vs `Ignore`, but clearly document fail-open/fail-closed semantics.
- **Reason**: Balances strict security requirements against operational stability.

## ADR-011: Policy Engine Consolidation (internal/evaluator vs internal/policy)
- **Context**: Section 6 specified both `internal/policy/` and `internal/evaluator/`.
- **Decision**: Consolidated evaluation execution directly inside `internal/policy/engine.go` and `internal/findings/findings.go` rather than splitting into a separate `evaluator` package.
- **Reason**: Avoids unnecessary package fragmentation and circular dependency risks while keeping policy normalization and rule evaluation in a unified engine package.
- **Status**: Accepted.

## ADR-012: In-Tree Profile Structs (api/ vs internal/policy)
- **Context**: Section 6 specified root `api/` directory for public types.
- **Decision**: PolicyProfile YAML models are located in `internal/policy/profile.go` for v1. Root `api/` directory reserved for future CRD definitions (`api/v1alpha1/`) when controller-runtime CRDs are introduced.
- **Reason**: Prevents premature external API exposure before profile schema stabilization.
- **Status**: Accepted.

## ADR-013: Multi-Platform Release Engineering and Cryptographic Provenance
- **Context**: Need an automated, verifiable release pipeline producing cross-platform binaries and container images without supply-chain risk.
- **Decision**: Use native Go cross-compilation with standard tooling (`go build`, `tar`, `zip`, `sha256sum`, GitHub CLI `gh release`, Docker buildx, GitHub artifact attestations) via `.github/workflows/release.yml`.
- **Reason**: Avoids heavy external binary packaging dependencies that cause toolchain drift, ensures 100% auditable release steps in GitHub Actions, and produces verifiable SHA-256 checksums and SLSA build attestations.
- **Status**: Accepted.

## ADR-014: Dynamic TLS Certificate Rotation via In-Memory Reloader and Dual Watchers
- **Context**: Admission webhook previously required pod restarts to pick up renewed TLS certificates. Operational downtime or handshake failures occurred during certificate expiration or rotation.
- **Decision**: Implemented `internal/admission/CertReloader` using standard library `tls.Config.GetCertificate` backed by lock-free `sync/atomic.Pointer[tls.Certificate]`. Supported dual watch mechanisms: periodic file stat/read for mounted Secret volumes and Kubernetes Secret API watching via `client-go`. RBAC permissions strictly scoped to namespaced `Role` granting `get, watch` on the specific TLS secret name.
- **Reason**: Zero lock contention during high-throughput TLS handshakes, instant sub-second rotation upon Kubernetes Secret modification, automatic fallback to file-based rotation when running outside cluster, and retention of the last known-good certificate if corrupted/invalid certificate material is encountered. CA bundle and CA rotation remain decoupled and explicitly distinct from server certificate rotation.
- **Status**: Accepted.

## ADR-015: Multi-Version Kubernetes Compatibility Matrix Testing via KinD
- **Context**: KubeGuard claimed compatibility with Kubernetes v1.28–v1.31 in documentation, but only v1.31 was verified on a live cluster in CI. A production-ready policy engine must prove compatibility across supported Kubernetes minor versions against actual API server and admission webhook behaviors.
- **Decision**: Introduce a parameterized CI matrix in GitHub Actions running the complete admission integration test suite across four pinned KinD node images (`v1.28.15`, `v1.29.12`, `v1.30.8`, `v1.31.4`) with SHA-256 digest pinning. Each matrix entry executes the identical 12-scenario admission test suite, including TLS handshakes, ALLOW/DENY, warning events, RBAC restrictions, failurePolicy behavior, and dynamic TLS rotation without pod restart.
- **Reason**: Reuses existing end-to-end integration logic without duplication, ensures isolated parallel testing on clean ephemeral clusters, pins immutable node image digests for deterministic reproducibility, and converts unverified compatibility claims into empirical, automated evidence.
- **Status**: Accepted.

## ADR-016: Stateless Active-Active Multi-Replica Webhook HA and In-Tree Observability
- **Context**: The admission webhook previously ran as a single pod (`replicaCount: 1`) without a PodDisruptionBudget, rolling update strategy guarantees, or operational metrics. Pod restarts or disruptions risked blocking cluster admissions when `failurePolicy: Fail` was configured.
- **Decision**: Implemented active-active multi-replica HA without leader election:
  1. Defaulted `replicaCount: 2` with `RollingUpdate` strategy (`maxSurge: 1`, `maxUnavailable: 0`) and `PodDisruptionBudget` (`minAvailable: 1`).
  2. Maintained pure statelessness: each replica loads identical policy profile ConfigMaps and independently watches the TLS Secret via `client-go` and mounted volumes.
  3. Added lightweight, dependency-free Prometheus metrics (`/metrics`) using atomic counters and gauges tracking request decisions, duration, policy errors, certificate reload status, and certificate expiration.
- **Reason**: Validating admission webhooks do not write mutable state to clusters, rendering leader election unnecessary overhead. Running active-active replicas behind the Kubernetes ClusterIP Service provides true zero-downtime admission availability, seamless rolling restarts, and resilient survival of individual pod terminations.
- **Status**: Accepted.

## ADR-017: Prometheus Latency Histogram and Deterministic Admission Load Testing
- **Context**: Operational observability requires standard Prometheus histogram distribution tracking for admission request latencies ($p_{50}, p_{90}, p_{99}$). Furthermore, admission throughput, concurrency ceilings, and replica scaling characteristics must be empirically validated under synthetic and burst workloads without introducing external heavyweight dependencies.
- **Decision**:
  1. Implemented standard Prometheus Histogram (`kubeguard_admission_request_duration_seconds`) with standard upper bounds (`le`: 0.001, 0.005, 0.010, 0.025, 0.050, 0.100, 0.250, 0.500, 1.0, `+Inf`) using thread-safe, lock-free `atomic.Uint64` bucket counters, sum, and count. Retained backward-compatible cumulative duration counter `kubeguard_admission_request_duration_seconds_total`.
  2. Built a deterministic in-tree load generation tool (`test/load/` and `kubeguard load-test` CLI) executing warm-up, sustained, burst, and recovery phases across compliant, denied, and mixed AdmissionReview payloads.
  3. Structured automated verification into CI: lightweight integration verification in `.github/workflows/ci.yml` (verifying live KinD admission endpoints and histogram population) and a dedicated stress workflow `.github/workflows/load-test.yml` comparing single-replica vs multi-replica scaling performance.
- **Reason**: Atomic array counters avoid mutex contention in high-concurrency admission handlers. Standard histogram formatting enables native Grafana visualization and Prometheus quantile calculation (`histogram_quantile`). In-tree load generation eliminates dependence on external load tools (e.g. k6, locust) and ensures 100% reproducible benchmark testing across environments.
- **Status**: Accepted.





