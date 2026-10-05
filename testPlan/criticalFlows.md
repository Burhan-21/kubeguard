# Critical Flows

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

The following 6 critical flows represent the core functionality of KubeGuard and must be thoroughly tested.

1. **Safe manifest (Happy Path)**
   * **Flow**: YAML input → Parser → Normalizer → Rule Engine → PASS
   * **Outcome**: KubeGuard confirms the manifest is compliant and allows it.

2. **Security violation (Enforcement Path)**
   * **Flow**: YAML input with a privileged container → Parser → Normalizer → Rule Engine identifies `privileged=true` → Policy Engine triggers rule `KG-SEC-001` → BLOCK
   * **Outcome**: KubeGuard successfully detects the violation, assigns the correct severity, and exits/denies appropriately.

3. **CI Pipeline Integration**
   * **Flow**: Pull Request created → CI Workflow triggers KubeGuard CLI → Vulnerability detected → Output formatted to SARIF → SARIF uploaded to GitHub Advanced Security.
   * **Outcome**: Findings are correctly displayed in the GitHub PR UI.

4. **Admission Control (In-Cluster Enforcement)**
   * **Flow**: User runs `kubectl apply` → Kubernetes API Server intercepts → Sends `AdmissionReview` to KubeGuard Webhook → Webhook parses request → Policy Engine evaluates → Rule violation found → Webhook returns `AdmissionResponse` with `allowed: false`.
   * **Outcome**: The resource creation is blocked by the cluster.

5. **Policy Failure (Safe Failure Behavior)**
   * **Flow**: Invalid policy configuration provided (e.g., syntax error) → Engine attempts evaluation → Evaluation fails.
   * **Outcome**: System handles failure securely. In CI, it exits with an error code (Exit 3). In the Webhook (if `failurePolicy: Fail`), it denies the request rather than failing open.

6. **Webhook Recovery (TLS Handling)**
   * **Flow**: Webhook certificate expires → API server rejects connection (TLS failure) → Detection alerts triggered → Operator rotates certs → Verification testing.
   * **Outcome**: Secure communication is restored and the webhook resumes processing requests.
