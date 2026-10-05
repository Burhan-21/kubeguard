# Product Requirements Document (PRD)

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Problem Statement
Kubernetes manifests can be syntactically valid but operationally unsafe or incomplete. Existing tools either focus purely on security vulnerabilities or lack developer experience and explainability, leading to misconfigurations reaching production.

## Target Users
- Developers
- DevOps Engineers
- Site Reliability Engineers (SREs)
- Platform Engineers
- Security Engineers
- Students
- CI/CD Maintainers

## User Journeys
1. **Developer (Local):** Write YAML → `kubeguard scan` → Fix misconfigurations → PASS.
2. **CI/CD Maintainer (Pipeline):** Open PR → GitHub Action runs KubeGuard → PASS/WARN/BLOCK based on severity.
3. **Cluster Admin (Admission):** Apply YAML to cluster → KubeGuard Validating Webhook → ALLOW/DENY based on cluster policies.

## Market Context & Differentiation
KubeGuard is **NOT** a replacement for Kyverno, OPA, Kubescape, Trivy, or Polaris. 
**Differentiation:**
- Focuses heavily on the combination of Security + Reliability + Deployment Readiness.
- High emphasis on Explainability (the "Why") and Developer Experience (DX).
- Pre-deployment focus, perfectly paired with AegisOps (post-deployment).

## AegisOps Relationship
- **KubeGuard**: Pre-deployment validation, CI/CD scanning, and admission control.
- **AegisOps**: Post-deployment runtime security and observability.
