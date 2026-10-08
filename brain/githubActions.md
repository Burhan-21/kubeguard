# GitHub Actions & CI/CD

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Pipeline Stages
Every commit to `main` and all Pull Requests must pass the following strict pipeline:

1. **Checkout**: Source code retrieval.
2. **Go Setup**: Initialize Go environment (target version defined in TRD).
3. **Dependencies**: `go mod tidy` and verify `go.sum`.
4. **Linting & Formatting**: `gofmt`, `golangci-lint`.
5. **Static Analysis**: `go vet`.
6. **Unit Tests**: Fast, isolated tests (`go test -short`).
7. **Race Detector**: `go test -race`.
8. **Integration Tests**: File-based parsing and policy engine tests.
9. **Admission Tests**: Using `envtest` or `kind` to spin up a transient API server and test webhook behavior.
10. **Security Scan**: `govulncheck` / Trivy SAST for KubeGuard codebase.
11. **Build**: Compile binary for target architectures (Linux, Darwin, Windows; amd64, arm64).
12. **Container Build**: Build Docker image.
13. **Container Scan**: Scan the KubeGuard image for OS-level vulnerabilities.
14. **SARIF Output**: Upload SARIF logs to GitHub Code Scanning (for KubeGuard itself).

## Failure Conditions
CI must **fail on any critical failures**: 
- Broken builds
- Failing tests
- Race conditions detected
- High/Critical CVEs in code or dependencies

No merging bypassing CI checks.

## Release Workflow (.github/workflows/release.yml)
Additive workflow triggered on push of version tags (`v*`) and manual `workflow_dispatch`.

1. **Target Platforms**: `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, `windows/amd64`.
2. **Deterministic Metadata**: Static CGO-free build with `-s -w` and version/commit/build-date injected via `-ldflags`.
3. **Artifact Packaging**: Compressed `.tar.gz` for Linux/macOS and `.zip` for Windows.
4. **Supply-Chain Security**:
   - SHA-256 checksums written to `checksums.txt` and verified via `sha256sum -c`.
   - Build provenance attestation using `actions/attest-build-provenance@v2`.
5. **GitHub Release**: Attached binaries, archives, and checksums published to official GitHub Release.
6. **Container Publishing**: Multi-arch image built and published to GitHub Container Registry (`ghcr.io/burhan-21/kubeguard:<tag>`), scanned via Trivy for zero critical/high CVEs.
