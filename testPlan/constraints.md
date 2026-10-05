# Testing Constraints

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

The testing and development of KubeGuard are bound by the following constraints:

1. **No Production Cluster Required**: Development and testing must not rely on a live, external Kubernetes cluster.
2. **Local-First**: All unit, component, and integration tests must run fully locally. Admission webhook integration testing must use `envtest`.
3. **No Paid Infrastructure**: The test suite must run on standard developer machines and free tiers of CI tools (e.g., GitHub Actions) without requiring paid cloud resources or paid scanning tools.
4. **No Secrets in Repo**: The repository must not contain real credentials, API keys, or certificates. Any required secrets for testing must be generated dynamically or mocked.
5. **No Destructive Experiments**: Tests must not perform destructive actions on any environment outside of the local sandbox or test runner execution context.
6. **No Fabricated Data**: Benchmark data, user statistics, stars, or adoption metrics must not be fabricated in any documentation or test outputs. Data should be realistic but explicitly marked as synthetic if used for testing.
7. **No External Affiliation Claims**: Testing outputs and documentation must not claim affiliation with Kubernetes, CNCF, Google, or other entities.
