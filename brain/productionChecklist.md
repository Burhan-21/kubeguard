# Production Release Checklist

**Author**: Shaikh Mohammed Burhan (GitHub: [Burhan-21](https://github.com/Burhan-21) / [NexusForge21](https://github.com/NexusForge21))

Every release must pass this checklist before deployment or tagging.

## Pre-Release Gate Verification Status

| Checklist Item | Status | Evidence / Notes |
|:---|:---:|:---|
| **Build passes** | `NOT TESTED` | Go compiler not installed on host environment. Code written for Go 1.23. |
| **Unit tests pass** | `NOT TESTED` | Unit test files written across parser, normalizer, engine, rules, reporters; runtime execution pending Go installation. |
| **Race detector passes** | `NOT TESTED` | `go test -race` requires Go toolchain with CGo/compiler on host or CI runner. |
| **Code security scan passes** | `NOT TESTED` | `govulncheck` / SAST requires Go binary or CI runner execution. |
| **Container scan passes** | `NOT TESTED` | Docker daemon not running in local environment. |
| **No secrets committed** | `PASS` | Repository audit confirmed zero private keys, API tokens, or credentials committed. |
| **CLI table output verified** | `NOT TESTED` | Implemented in `internal/reporter/human.go`; live binary execution pending. |
| **JSON output verified** | `NOT TESTED` | Implemented in `internal/reporter/json.go`; live binary execution pending. |
| **SARIF output verified** | `NOT TESTED` | Implemented in `internal/reporter/sarif.go`; schema validated against Oasis SARIF 2.1.0 standard. |
| **Admission controller verified** | `NOT TESTED` | `AdmissionReview v1` handler implemented; live Kubernetes API server / kind cluster not available. |
| **TLS configuration verified** | `NOT TESTED` | HTTPS listener implemented in `server.go`; live certificate generation/handshake pending cluster deployment. |
| **RBAC reviewed** | `PASS` | Minimal `ClusterRole` and `ClusterRoleBinding` created in Helm chart, strictly restricted to admission review. |
| **Helm chart validated** | `NOT TESTED` | Helm CLI not installed on host; template files created in `charts/kubeguard/`. |
| **Documentation matches code** | `PASS` | Brain docs, README, and API contracts audited and synchronized with Go source code. |
| **CHANGELOG updated** | `PASS` | `CHANGELOG.md` and `brain/changLog.md` synchronized for release v0.1.0. |
| **Version updated** | `PASS` | Version `0.1.0` set in `internal/version/version.go`, `Chart.yaml`, and documentation. |
| **Release artifacts verified** | `NOT TESTED` | Compilation of release tarballs pending host or CI build. |
