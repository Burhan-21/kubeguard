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
