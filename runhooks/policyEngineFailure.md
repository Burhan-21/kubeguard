# Policy Engine Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
The KubeGuard Policy Engine failed during evaluation of a Kubernetes manifest. This means the engine crashed, panicked, or encountered a fatal error before it could determine if the manifest was compliant.

## How to detect?
* **CLI Behavior**: The CLI exits with a specific status code (Exit 3) indicating an evaluation error, rather than rule violations.
* **CI Behavior**: The CI pipeline fails unexpectedly without producing a SARIF report, or logs show engine panics.
* **Admission Behavior**: The Kubernetes API server logs show webhook failures, or KubeGuard pod logs show errors/crashes during `AdmissionReview` processing.

## How to contain?
* **Logging**: Capture the stack trace or error message from the KubeGuard output or pod logs.
* **Admission Webhook**: The containment behavior depends on the `failurePolicy` of the `ValidatingWebhookConfiguration`.

## How to recover?
1. Identify the input YAML that caused the failure.
2. Run the KubeGuard CLI locally against the problematic YAML with debug logging enabled to reproduce the issue.
3. Determine if the issue is malformed input that bypasses validation, or a bug in the Policy Engine logic.
4. Apply a patch to fix the engine bug, or update the parser to properly handle/reject the malformed input.

### Fail-Open vs Fail-Closed Decision
In the Admission Webhook context, KubeGuard's `failurePolicy` dictates behavior:
* **Fail (Recommended)**: Denies the request. Secure, but can block cluster operations if KubeGuard is crashing.
* **Ignore**: Allows the request. Insecure, but maintains cluster availability.
* **Decision**: Always prefer Fail for security, but be prepared to temporarily switch to Ignore or bypass the webhook (e.g., by adding a namespace selector exclusion) if a critical production deployment is blocked by an engine bug.

## How to verify recovery?
1. Execute the KubeGuard CLI against the previously failing YAML; it should now exit cleanly (either passing, or failing with a correct validation error, but not crashing).
2. If in cluster, deploy the patched KubeGuard and verify the webhook processes the resource successfully.

## When to escalate?
* If the engine failure is causing a production outage (due to `failurePolicy: Fail`) and cannot be quickly patched or bypassed.
* If the failure indicates a broader systemic issue with the Go runtime or core dependencies.
