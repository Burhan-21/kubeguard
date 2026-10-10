# Testing Requirements

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Testing Pyramid

1. **Unit Tests**: Isolated testing of Go functions, structures, and utility methods.
2. **Component (Rule/Parser) Tests**: Testing the YAML parser against various valid/invalid schemas. Testing individual rules in isolation.
3. **Integration Tests**: Running the CLI against directories of YAML files. Validating JSON/SARIF output correctness. Testing policy engine concurrency.
4. **E2E / Admission Tests**: Deploying KubeGuard into a `kind` cluster and sending real `kubectl apply` commands to verify webhook interception.

## Rule Testing Strictness
Every security and reliability rule **MUST** include:
- **Positive Test Case**: Manifest explicitly violating the rule (should trigger finding).
- **Negative Test Case**: Manifest correctly configured (should NOT trigger finding).
- **Edge Cases**: Missing fields, null values, malformed data structures.

## Additional Testing
- **Race Detection**: `go test -race` must run in CI.
- **Concurrency**: Engine must be tested with thousands of concurrent resource evaluations to ensure thread safety.
- **Security Testing**: Verify that parser resource limits correctly block YAML bombs.
- **Helm Tests**: Validate template rendering against multiple values variations.
- **Container Tests**: Ensure the built Docker image runs as non-root and responds correctly to basic commands.

## Verification Gate Distinction
- **Unit-Level Exit-Code Contract**: Verified via `internal/findings/findings_test.go` (`0 = PASS`, `1 = WARN`, `2 = BLOCK`, `3 = TOOL ERROR`).
- **CLI Runtime Exit-Code Verification**: Verified (`PASS`) via Ubuntu CI execution of `./bin/kubeguard scan` against fixtures with assertions on exit codes 0, 1, 2, and 3.
- **Rule Implementations**: 26/26 individual rules verified via Go rule-level unit tests under race detector (`security_test.go`, `reliability_test.go`).
- **End-to-End CLI Pipeline**: Verified (`PASS`) via Ubuntu CI runner executing single, multi-doc, directory, and self-scan with JSON and SARIF validation.
- **Live Kubernetes Integration**: Verified (`PASS`) via Ubuntu CI ephemeral KinD cluster running live Kubernetes API server with ValidatingWebhookConfiguration, TLS communication, ALLOW, DENY (KG-SEC-001), WARN, malformed request handling (HTTP 400), failurePolicy: Fail verification, and live dynamic TLS certificate rotation.
- **Dynamic TLS Certificate Rotation**: Verified (`PASS`) via comprehensive test suite in `internal/admission/reloader_test.go` and live KinD integration:
  - `Test A (Initial certificate)`: Verified startup succeeds with valid cert, TLS handshakes succeed, and startup fails immediately if initial certificate is corrupt or missing.
  - `Test B (Certificate rotation)`: Verified transition from Cert A to Cert B serves Cert B on subsequent TLS connections without process restart.
  - `Test C (Invalid rotation)`: Verified corrupt PEM and mismatched private keys are safely rejected, retaining last known-good certificate while server remains healthy.
  - `Test D (Rapid updates)`: Verified rapid succession of 10 certificate updates without deadlocks, serving final valid certificate.
  - `Test E (Shutdown)`: Verified file and secret watchers terminate cleanly upon context cancellation without goroutine leaks.
  - `Test F (Race safety)`: Verified concurrent TLS client requests against active certificate rotation under `go test -race` with zero data races.
  - `Live KinD Cluster Rotation`: Verified on ephemeral KinD cluster: Secret updated in `kubeguard-system`, webhook served new certificate over live TLS connection, container `restartCount: 0` (zero restarts), and live Kubernetes API server admitted workloads with rotated certificate.
- **Performance Benchmarks**: Verified (`PASS`) via automated benchmark suite (`test/benchmark/benchmark_test.go`) evaluating 10, 100, and 1000 resource parsing/normalization/policy evaluation and admission latency.

