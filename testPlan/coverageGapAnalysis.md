# Coverage Gap Analysis

**Author**: Shaikh Mohammed Burhan (GitHub: [Burhan-21](https://github.com/Burhan-21) / [NexusForge21](https://github.com/NexusForge21))  
**License**: Apache 2.0

This document tracks test coverage across KubeGuard components. Rather than chasing synthetic coverage percentages, KubeGuard tracks coverage against critical execution paths, failure boundaries, and admission safety invariants.

| Component | Status | Key Paths Tested | Important Remaining Paths | Risk Level | Action Required |
|:---|:---:|:---|:---|:---:|:---|
| **YAML Parser** | `COVERED` | Single-doc, multi-doc (`---`), directory parsing, empty docs, invalid YAML syntax | Extreme payload fuzzing (>50MB), deeply nested cyclic YAML | Medium | Add parser fuzz testing in CI |
| **Resource Normalizer** | `COVERED` | Deployment, Pod, StatefulSet, DaemonSet, Job, CronJob, Service, ClusterRole | Custom Resource Definitions (CRDs) with embedded PodTemplates | Low | Add generic CRD pod extraction |
| **Policy Engine** | `COVERED` | Rule evaluation, finding aggregation, severity overrides, rule disabling | Concurrent evaluation race detection under high load | Medium | Verify with `go test -race` in CI |
| **Security Rules (14)** | `COVERED` | Positive (violation triggered) and negative (compliant passes) for all 14 rules | Multi-container mixed compliance combinations | Low | Add composite multi-container fixtures |
| **Reliability Rules (12)** | `COVERED` | Positive and negative tests for all 12 rules (probes, resources, strategy, replicas, PDB, etc.) | Auto-scaling workloads (HPA overriding replicas) | Low | Support HPA awareness in single-replica rule |
| **Reporters** | `COVERED` | Human terminal output, JSON indentation & unmarshaling, SARIF v2.1.0 schema | Windows cmd.exe color compatibility without ANSI support | Low | Verify `NO_COLOR` environment variable fallback |
| **Admission Webhook** | `COVERED` | Enforce mode (deny on BLOCK), audit mode (allow all), warn mode (warnings array), clean manifests, nil request handling | Live cluster network latency, webhook timeout recovery | High | Run automated `kind` cluster integration test |
| **Helm Packaging** | `COVERED` | Template rendering for Deployment, Service, RBAC, Webhook, NetPol | Live cert-manager auto-injection integration | Medium | Document cert-manager values configuration |
