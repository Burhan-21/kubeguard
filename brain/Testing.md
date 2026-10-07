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
- **CLI Runtime Exit-Code Verification**: Marked `NOT TESTED` until the compiled binary is executed in Ubuntu CI against live YAML fixtures.
- **Rule Implementations**: 26/26 individual rules verified via Go rule-level unit tests (`security_test.go`, `reliability_test.go`).
- **End-to-End CLI Pipeline**: Marked `NOT TESTED` until executed by the CI runner against sample manifests.
