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
