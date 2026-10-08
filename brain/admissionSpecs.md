# Admission Controller Specifications

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Architecture Flow
`kubectl apply` → `K8s API Server` → `ValidatingWebhookConfig` → `KubeGuard Service` → `PolicyEngine` → `AdmissionResponse (ALLOW/DENY)`

## Operating Modes
- **Audit**: Log findings, but return `ALLOW`.
- **Warn**: Return `ALLOW`, but populate the `warnings` field in the AdmissionResponse to show warnings directly in the user's terminal.
- **Enforce**: Return `DENY` for any finding that hits the configured BLOCK threshold.

## Validating Webhook Configuration
- **API Version**: `admission.k8s.io/v1`
- **TLS**: Required. KubeGuard must serve over HTTPS. Provide mechanisms for cert-manager injection or self-signed CA bundle generation.
- **failurePolicy**: Configurable (`Fail` or `Ignore`). Default recommended is `Ignore` for initial rollouts, moving to `Fail` for strict environments.
- **timeoutSeconds**: Must be strict (e.g., `5` seconds) to prevent APIServer stalls.
- **Namespace Selectors**: Support `namespaceSelector` to easily exempt system namespaces (e.g., `kube-system`).

## Deployment Considerations
- **Resource Limits**: Strict CPU/Memory requests and limits must be defined in the Helm chart.
- **securityContext**: Run as non-root, read-only root filesystem, drop ALL capabilities.
- **NetworkPolicy**: Limit ingress to only the API Server, limit egress (deny all external egress).
- **Rollout Strategy**: High availability (minReplicas > 1, PodDisruptionBudget) to prevent single point of failure blocking cluster deployments.

## Rule Consistency
The Admission controller utilizes the **exact same policy engine** as the CLI. A manifest that passes `kubeguard scan` locally must identically pass the admission webhook.

## Live Cluster Verification Status
- **Environment**: KinD Ephemeral Kubernetes Cluster v1.31 in GitHub Actions (Run `#37806478180`, Job `113412441486`).
- **Interception Tests**:
  - `ALLOW`: Compliant deployment admitted by API server.
  - `DENY`: Insecure deployment violating `KG-SEC-001` rejected with clear rule violation reason.
  - `WARN`: Admitted with user-facing warnings.
  - `Consistency`: Validated identical policy decision between CLI and webhook.
  - `Malformed Review`: Handled via HTTP 400 Bad Request without process crash.
  - `failurePolicy`: Verified `failurePolicy: Fail` blocks admission when webhook pods are scaled down.
- **TLS Handshake**: Successful mutual trust established using custom CA bundle and SAN-enabled TLS certificate on port 443 -> 8443.
- **RBAC**: ServiceAccount audited via `kubectl auth can-i --list` confirming strictly least-privilege read access.
