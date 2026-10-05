# Detailed Test Plan & Test Case Catalog

**Author**: Shaikh Mohammed Burhan (GitHub: [Burhan-21](https://github.com/Burhan-21) / [NexusForge21](https://github.com/NexusForge21))  
**License**: Apache 2.0

This test plan defines explicit test cases, inputs, expected findings, and assertion criteria for all KubeGuard components.

---

## Part 1: Security Rules (TC-KG-001 to TC-KG-014)

| Test ID | Rule ID | Target Condition | Test Input Manifest | Expected Severity | Expected Exit Code |
|:---|:---|:---|:---|:---:|:---:|
| **TC-KG-001** | `KG-SEC-001` | Privileged container | Container with `securityContext.privileged: true` | `BLOCK` | 2 |
| **TC-KG-002** | `KG-SEC-002` | Root execution risk | Container without `runAsNonRoot: true` and `runAsUser: 0` | `WARN` | 1 |
| **TC-KG-003** | `KG-SEC-003` | Privilege escalation | Container with `allowPrivilegeEscalation: true` or omitted | `WARN` | 1 |
| **TC-KG-004** | `KG-SEC-004` | Host network namespace | Pod with `hostNetwork: true` | `BLOCK` | 2 |
| **TC-KG-005** | `KG-SEC-005` | Host PID / IPC | Pod with `hostPID: true` or `hostIPC: true` | `BLOCK` | 2 |
| **TC-KG-006** | `KG-SEC-006` | Dangerous hostPath mounts | Volume mount of `/var/run/docker.sock` or `/` | `BLOCK` | 2 |
| **TC-KG-007** | `KG-SEC-007` | Dangerous capabilities | Capabilities add `SYS_ADMIN` or `ALL` | `BLOCK` | 2 |
| **TC-KG-008** | `KG-SEC-008` | Missing securityContext | Container with nil `securityContext` | `WARN` | 1 |
| **TC-KG-009** | `KG-SEC-009` | Mutable latest image tag | Image `nginx:latest` | `WARN` | 1 |
| **TC-KG-010** | `KG-SEC-010` | Missing image tag | Image `redis` without explicit tag | `WARN` | 1 |
| **TC-KG-011** | `KG-SEC-011` | Unpinned image digest | Image without `@sha256:` digest | `WARN` | 1 |
| **TC-KG-012** | `KG-SEC-012` | Plaintext embedded secret | Env var `DB_PASSWORD` with literal `value` | `BLOCK` | 2 |
| **TC-KG-013** | `KG-SEC-013` | Automount ServiceAccount token | Pod with `automountServiceAccountToken: true` or nil | `WARN` | 1 |
| **TC-KG-014** | `KG-SEC-014` | Excessive RBAC permissions | ClusterRole with wildcard `*` resources and verbs | `BLOCK` | 2 |

---

## Part 2: Reliability Rules (TC-KG-015 to TC-KG-026)

| Test ID | Rule ID | Target Condition | Test Input Manifest | Expected Severity | Expected Exit Code |
|:---|:---|:---|:---|:---:|:---:|
| **TC-KG-015** | `KG-REL-001` | Missing readiness probe | Container with nil `readinessProbe` | `WARN` | 1 |
| **TC-KG-016** | `KG-REL-002` | Missing liveness probe | Container with nil `livenessProbe` | `WARN` | 1 |
| **TC-KG-017** | `KG-REL-003` | Missing startup probe | Workload with liveness probe but no startup probe | `WARN` | 1 |
| **TC-KG-018** | `KG-REL-004` | Missing resource requests | Container without CPU/memory `requests` | `WARN` | 1 |
| **TC-KG-019** | `KG-REL-005` | Missing resource limits | Container without CPU/memory `limits` | `WARN` | 1 |
| **TC-KG-020** | `KG-REL-006` | Single replica production risk | Deployment configured with `replicas: 1` | `WARN` | 1 |
| **TC-KG-021** | `KG-REL-007` | Deployment strategy risk | Deployment using `Recreate` strategy | `WARN` | 1 |
| **TC-KG-022** | `KG-REL-008` | Missing PodDisruptionBudget | Multi-replica workload without PDB | `WARN` | 1 |
| **TC-KG-023** | `KG-REL-009` | Missing topology/anti-affinity | Multi-replica workload without pod anti-affinity | `WARN` | 1 |
| **TC-KG-024** | `KG-REL-010` | Short termination grace period | Pod with `terminationGracePeriodSeconds: 0` | `WARN` | 1 |
| **TC-KG-025** | `KG-REL-011` | Missing NetworkPolicy | Workload deployed without network segmentation | `WARN` | 1 |
| **TC-KG-026** | `KG-REL-012` | External service exposure | Service with `type: LoadBalancer` or `NodePort` | `WARN` | 1 |

---

## Part 3: Parser & Input Handling (TC-KG-027 to TC-KG-034)

- **TC-KG-027**: Single document YAML parsing (`internal/parser/parser_test.go:TestParseReaderSingle`)
- **TC-KG-028**: Multi-document stream YAML parsing separated by `---` (`TestParseReaderMultiDocument`)
- **TC-KG-029**: Empty documents and comment-only streams (`TestParseReaderEmpty`)
- **TC-KG-030**: Malformed and unparseable YAML syntax handling (`TestParseReaderInvalid`) -> Graceful tool error (exit code 3)
- **TC-KG-031**: Directory traversal parsing only `.yaml`/`.yml` (`TestParseDirectory`)
- **TC-KG-032**: Standard input (`-`) streaming parsing
- **TC-KG-033**: UTF-8 and BOM-encoded YAML stream ingestion
- **TC-KG-034**: Oversized YAML document resource exhaustion bounds

---

## Part 4: Normalizer Workload Extraction (TC-KG-035 to TC-KG-040)

- **TC-KG-035**: Extract `PodSpec` and `replicas` from `Deployment` (`apps/v1`)
- **TC-KG-036**: Extract `PodSpec` from standalone `Pod` (`v1`)
- **TC-KG-037**: Extract `PodSpec` from `StatefulSet` (`apps/v1`)
- **TC-KG-038**: Extract `PodSpec` from `DaemonSet` (`apps/v1`)
- **TC-KG-039**: Extract `PodSpec` from `Job` and `CronJob` (`batch/v1`)
- **TC-KG-040**: Graceful non-workload normalization for `Service`, `Role`, `ClusterRole`, `NetworkPolicy`

---

## Part 5: Policy Engine & Overrides (TC-KG-041 to TC-KG-046)

- **TC-KG-041**: Empty rule set produces 0 findings (`TestEngineNoRules`)
- **TC-KG-042**: Rule evaluation and finding aggregation (`TestEngineWithRules`)
- **TC-KG-043**: Profile severity override modifies finding severity (`TestSeverityOverrides`)
- **TC-KG-044**: Profile rule disabling suppresses evaluation (`TestEngineDisabledRules`)
- **TC-KG-045**: Multi-resource evaluation aggregation (`TestEvaluateAll`)
- **TC-KG-046**: Exit code resolution: 0 for PASS, 1 for WARN, 2 for BLOCK, 3 for tool error

---

## Part 6: Reporting Verification (TC-KG-047 to TC-KG-050)

- **TC-KG-047**: Terminal Human reporter renders resource headers, rule IDs, messages, and color highlights
- **TC-KG-048**: JSON reporter emits valid JSON conforming to `findings.ScanResult`
- **TC-KG-049**: SARIF v2.1.0 reporter matches Oasis SARIF schema with correct rule drivers and error/warning level mappings
- **TC-KG-050**: Format dispatch via `--output` flag

---

## Part 7: Admission Controller (TC-KG-051 to TC-KG-055)

- **TC-KG-051**: `enforce` mode denies admission (`Allowed: false`) for manifests with `BLOCK` findings (`TestAdmissionHandlerEnforceBlock`)
- **TC-KG-052**: `audit` mode permits admission (`Allowed: true`) even with `BLOCK` findings (`TestAdmissionHandlerAuditMode`)
- **TC-KG-053**: `warn` mode permits admission and populates admission warnings array (`TestAdmissionHandlerWarnMode`)
- **TC-KG-054**: Compliant workload admitted without warnings (`TestAdmissionHandlerCleanWorkload`)
- **TC-KG-055**: Malformed AdmissionReview request handles nil gracefully (`TestAdmissionHandlerNilObject`)
