# Test Strategy

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## Testing Pyramid
KubeGuard follows a standard testing pyramid to ensure reliability and performance.

1. **Unit Testing**
   * Testing individual functions and methods (e.g., specific rule evaluations, string manipulation).
   * Fast, deterministic, no external dependencies.
2. **Component Testing**
   * Testing larger components in isolation (e.g., Policy Engine evaluating a full normalized struct).
3. **Integration Testing**
   * Testing the interaction between components and external systems.
   * **Kubernetes Integration**: Use `sigs.k8s.io/controller-runtime/pkg/envtest` to run a local API server and etcd for testing the Admission Webhook locally without a full cluster.
4. **End-to-End (E2E) Testing**
   * Complete flow from CLI execution to final output, or testing the Helm chart installation.

## Key Testing Principles

* **Deterministic Testing**: Tests must not rely on external network calls or time-based logic that can cause flakiness. Mock external dependencies (like container registries).
* **Race Detection**: All Go tests must be run with the `-race` flag (`go test -race ./...`) in CI to prevent concurrent map writes and other race conditions.
* **Fuzz Testing**: Apply Go's native fuzzing (`go test -fuzz`) to the YAML Parser and Normalizer to identify panics or memory issues caused by malformed input.
* **Security Testing**:
   * Test malicious YAML (e.g., YAML bombs, deep recursion).
   * Test oversized input to ensure proper limits are enforced.
* **Negative Testing**: Intentionally provide invalid configurations to ensure error handling is robust (e.g., invalid CLI flags, invalid policy files).
* **Failure Testing**: Simulate failures (e.g., engine panics) to verify safe failure behavior (Fail-Closed).
