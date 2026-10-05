# KubeGuard

> **Kubernetes Deployment Security & Reliability Policy Engine**

[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.23-00ADD8?logo=go)](go.mod)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.28%E2%80%931.31-326CE5?logo=kubernetes)](https://kubernetes.io)

**KubeGuard** is an open-source security and operational reliability policy engine for Kubernetes workloads. It statically verifies whether workloads are safe, hardened, and operationally ready *before* deployment into production clusters.

It provides a single unified policy engine across:
- **Developer CLI**: Fast local feedback during development.
- **CI/CD Pipelines**: Automated pull request gating with JSON and SARIF reporting.
- **Admission Webhooks**: Enforcing runtime policy admission at the Kubernetes API server boundary.

---

## Authorship & Project Attribution

- **Author & Creator**: Shaikh Mohammed Burhan
- **GitHub**: [Burhan-21](https://github.com/Burhan-21) / [NexusForge21](https://github.com/NexusForge21)
- **License**: Apache License 2.0

*Note: KubeGuard is an independent open-source project and does not claim official affiliation with the Kubernetes project, CNCF, Google, or any commercial vendor.*

---

## The Problem KubeGuard Solves

A Kubernetes manifest can be 100% syntactically valid YAML and accepted by `kubectl apply`, yet remain fundamentally dangerous in production:

- **Security Risks**: Running as root, granting container privilege, sharing host network or PID namespaces, mounting `/var/run/docker.sock`, or embedding plain-text credentials in environment variables.
- **Reliability Deficits**: Missing readiness/liveness probes, unbounded memory and CPU consumption causing noisy neighbor starvation, single-replica deployments lacking high availability, or using `Recreate` rollout strategies leading to outages.

KubeGuard detects and halts these conditions **before deployment**.

---

## Core Architecture

```text
                 Kubernetes YAML / Manifests
                             |
                             v
                        YAML Parser (Multi-doc / Stdin / Directory)
                             |
                             v
                  Kubernetes Object Model
                             |
                             v
                     Resource Normalizer
                 (Extracts Workload PodSpec & Metadata)
                             |
                             v
                       Policy Engine
                      /             \
                     /               \
            Security Rules       Reliability Rules
             (14 Rules)             (12 Rules)
                     \               /
                      \             /
                       v           v
                        Finding Engine
                             |
                             v
                      Severity Engine
                             |
                    +--------+--------+
                    |        |        |
                   PASS     WARN    BLOCK
                    |        |        |
                    +--------+--------+
                             |
                             v
                         Reporters
                     /       |       \
                  Human    JSON     SARIF
```

---

## Policy Rules Catalog

### Security Rules (14 Rules)

| ID | Rule Title | Default Severity | Target Condition Detected |
|:---|:---|:---:|:---|
| **KG-SEC-001** | Privileged Container | `BLOCK` | Container runs with `securityContext.privileged: true` |
| **KG-SEC-002** | Root Execution Risk | `WARN` | Container runs without `runAsNonRoot: true` or with UID 0 |
| **KG-SEC-003** | Privilege Escalation | `WARN` | Container allows `allowPrivilegeEscalation: true` |
| **KG-SEC-004** | Host Network Namespace | `BLOCK` | Pod shares host network namespace (`hostNetwork: true`) |
| **KG-SEC-005** | Host PID / IPC | `BLOCK` | Pod shares host PID (`hostPID`) or IPC (`hostIPC`) namespace |
| **KG-SEC-006** | Dangerous HostPath Mounts | `BLOCK` / `WARN` | Mounts sensitive host paths (`/`, `/var/run/docker.sock`, etc.) |
| **KG-SEC-007** | Dangerous Linux Capabilities | `BLOCK` / `WARN` | Adds `SYS_ADMIN`, `ALL`, `NET_ADMIN`, or `SYS_PTRACE` |
| **KG-SEC-008** | Missing SecurityContext | `WARN` | Container lacks explicit `securityContext` restrictions |
| **KG-SEC-009** | Latest Image Tag | `WARN` | Image uses mutable `:latest` tag |
| **KG-SEC-010** | Missing Image Tag | `WARN` | Image has no tag specified (defaults to latest) |
| **KG-SEC-011** | Unpinned Image Digest | `WARN` | Image lacks immutable `@sha256:` digest pin |
| **KG-SEC-012** | Plaintext Embedded Secrets | `BLOCK` | Environment variable names suggest secrets with literal `value` |
| **KG-SEC-013** | Service Account Token Risk | `WARN` | Default automounting of ServiceAccount tokens enabled |
| **KG-SEC-014** | Excessive RBAC Permissions | `BLOCK` | Role / ClusterRole grants wildcard `*` resources and verbs |

### Reliability Rules (12 Rules)

| ID | Rule Title | Default Severity | Target Condition Detected |
|:---|:---|:---:|:---|
| **KG-REL-001** | Missing Readiness Probe | `WARN` | Workload lacks readiness probe for zero-downtime routing |
| **KG-REL-002** | Missing Liveness Probe | `WARN` | Workload lacks liveness probe for automatic self-healing |
| **KG-REL-003** | Missing Startup Probe | `WARN` | Slow-starting workloads lack startup probe |
| **KG-REL-004** | Missing Resource Requests | `WARN` | Container missing CPU or memory requests for scheduling |
| **KG-REL-005** | Missing Resource Limits | `WARN` | Container missing CPU or memory limits (OOM risk) |
| **KG-REL-006** | Single Replica Risk | `WARN` | Production deployment configured with `replicas: 1` |
| **KG-REL-007** | Deployment Strategy Risk | `WARN` | Deployment uses downtime-inducing `Recreate` strategy |
| **KG-REL-008** | Missing PodDisruptionBudget | `WARN` | Multi-replica workload has no PodDisruptionBudget |
| **KG-REL-009** | Missing Topology / Anti-Affinity | `WARN` | Workload pods have no anti-affinity or spread constraints |
| **KG-REL-010** | Termination Grace Period | `WARN` | Termination grace period set to 0 or abnormally short |
| **KG-REL-011** | Missing NetworkPolicy | `WARN` | Workload deployed without network segmentation policy |
| **KG-REL-012** | External Service Exposure | `WARN` | Service exposed via `LoadBalancer` or `NodePort` |

---

## Policy Profiles

KubeGuard supports profile-based policy customization:

- `policies/default.yaml`: Balanced profile for general workloads.
- `policies/development.yaml`: Relaxed rules (allows single replicas, relaxed image tags).
- `policies/production.yaml`: Strict operational readiness and security standards.
- `policies/strict.yaml`: Enforces `BLOCK` severity across all security and operational rules.

Example profile definition:
```yaml
apiVersion: kubeguard.dev/v1alpha1
kind: PolicyProfile
metadata:
  name: production
  description: Strict production policy profile
spec:
  rules:
    - id: KG-SEC-001
      enabled: true
      severity: BLOCK
    - id: KG-REL-001
      enabled: true
      severity: BLOCK
    - id: KG-REL-004
      enabled: true
      severity: BLOCK
```

---

## CLI Installation & Usage

### Building from Source

```bash
git clone https://github.com/Burhan-21/kubeguard.git
cd kubeguard
go build -o bin/kubeguard ./cmd/kubeguard
```

### Scan Files or Directories

```bash
# Scan a single manifest
kubeguard scan ./examples/insecure/privileged-deployment.yaml

# Scan a directory recursively with a policy profile
kubeguard scan --profile policies/production.yaml ./manifests/

# Read from stdin
kubectl kustomize ./overlays/prod | kubeguard scan -
```

### Output Formats

```bash
# Human-readable color terminal output (default)
kubeguard scan ./manifests/

# Machine-readable JSON output
kubeguard scan --output json ./manifests/ > scan-results.json

# Standard SARIF v2.1.0 output for GitHub Security tab
kubeguard scan --output sarif ./manifests/ > results.sarif
```

### Exit Codes

| Exit Code | Meaning | Condition |
|:---:|:---|:---|
| **`0`** | **PASS** | No findings, or only findings below `--fail-on` threshold |
| **`1`** | **WARN** | Warnings detected when `--fail-on=warn` |
| **`2`** | **BLOCK** | Blocking policy violations detected |
| **`3`** | **TOOL ERROR** | Manifest parse failure, profile read failure, or invalid flags |

---

## CI/CD Pipeline Integration

### GitHub Actions Workflow

```yaml
name: Security & Reliability Policy Gate

on:
  pull_request:
    branches: [ main ]

jobs:
  kubeguard-check:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'

      - name: Build KubeGuard
        run: go build -o bin/kubeguard ./cmd/kubeguard

      - name: Scan Kubernetes Manifests (SARIF)
        run: |
          ./bin/kubeguard scan --output sarif ./deploy/ > kubeguard-results.sarif
        continue-on-error: true

      - name: Upload SARIF to GitHub Security Code Scanning
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: kubeguard-results.sarif

      - name: Enforce Policy Gate
        run: |
          ./bin/kubeguard scan --profile policies/production.yaml --fail-on block ./deploy/
```

---

## Validating Admission Webhook

KubeGuard includes an HTTPS admission controller compliant with Kubernetes `AdmissionReview v1`.

### Running Locally / In Cluster

```bash
kubeguard admission \
  --port 8443 \
  --tls-cert /etc/tls/tls.crt \
  --tls-key /etc/tls/tls.key \
  --mode enforce \
  --profile /etc/kubeguard/profile.yaml
```

Supported modes:
- `audit`: Evaluates manifests, logs findings, always permits deployment.
- `warn`: Allows deployment, but returns formatted admission warnings to `kubectl`.
- `enforce`: Rejects deployments (`Allowed: false`) containing any `BLOCK` violations.

---

## Helm Chart Deployment

Deploy KubeGuard into your cluster using Helm:

```bash
helm upgrade --install kubeguard ./charts/kubeguard \
  --namespace kubeguard-system \
  --create-namespace \
  --set admission.mode=enforce \
  --set replicaCount=3
```

---

## KubeGuard Self-Scan

KubeGuard practices what it preaches. It provides a reference hardened production deployment manifest for itself in `examples/kubeguard-self-scan/`:

```bash
kubeguard scan --profile policies/production.yaml examples/kubeguard-self-scan/kubeguard-deployment.yaml
```

**Verification Status**: NOT TESTED LOCALLY (Host environment lacks installed Go compiler to build binary; verified via static manifest analysis against rule definitions).

---

## Current Engineering Status & Transparency

- **Policy Engine & Rules**: Implemented in Go source code.
- **Unit & Race Tests**: Test suites written; execution requires host Go 1.23 toolchain.
- **Self-Scan Execution**: NOT TESTED LOCALLY (Pending binary compilation).
- **Cluster Integration**: NOT TESTED on live cluster (Kind/k3d not available in local environment).
- **Benchmarks**: NOT MEASURED.
- **Static Manifest Scope**: KubeGuard scans declarative Kubernetes manifests and admission requests. It does not inspect container binary images or filesystem layers (use Trivy or Grype in conjunction).

---

## Project Repository Structure

```text
kubeguard/
├── brain/               # Engineering source of truth (18 specifications & ADRs)
├── runhooks/            # Operational failure & incident recovery playbooks
├── testPlan/            # Testing pyramid, coverage analysis, and test specifications
├── cmd/kubeguard/       # Cobra CLI (root, scan, policy, admission, version)
├── internal/
│   ├── parser/          # Multi-document YAML streaming decoder
│   ├── normalizer/      # Unified workload PodSpec extractor
│   ├── policy/          # Rule interface, engine, and YAML profile loader
│   ├── rules/
│   │   ├── security/    # 14 Security rule implementations (KG-SEC-001 - 014)
│   │   └── reliability/ # 12 Reliability rule implementations (KG-REL-001 - 012)
│   ├── findings/        # Finding data models and ScanResult summarization
│   ├── reporter/        # Terminal Human, indented JSON, and SARIF v2.1.0 reporters
│   ├── admission/       # AdmissionReview v1 HTTPS webhook handler & server
│   ├── config/          # CLI runtime configuration
│   └── version/         # Build metadata and version tags
├── policies/            # Default, Development, Production, and Strict profiles
├── examples/            # Insecure, Secure, Production, Admission, and Self-Scan examples
├── charts/kubeguard/    # Production-ready Helm chart with TLS, RBAC, and NetworkPolicy
├── testdata/            # Test manifests for unit and regression testing
├── Dockerfile           # Minimal multi-stage non-root container image
├── Makefile             # Development automation targets
└── README.md
```

---

## Relationship with AegisOps

KubeGuard is architecturally designed to complement **AegisOps**:

```text
                    AegisOps Ecosystem
                            |
            +---------------+---------------+
            |                               |
      Pre-Deployment                  Post-Deployment
            |                               |
        KubeGuard                        AegisOps
     (Static Policy &               (Runtime Observability,
    Admission Control Gate)          Drift Detection & Recovery)
            |                               |
            +---------------+---------------+
                            |
                     Kubernetes Cluster
```

- **KubeGuard answers**: *"Is this workload safe and operationally hardened enough to deploy?"*
- **AegisOps answers**: *"How is the workload behaving after deployment, and how do we remediate runtime failures?"*

---

## License

This project is licensed under the Apache License 2.0. See the [LICENSE](LICENSE) file for complete details.
