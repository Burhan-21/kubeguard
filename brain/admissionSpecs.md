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
- **TLS**: Required. KubeGuard must serve over HTTPS. Provide mechanisms for cert-manager injection or self-signed CA bundle generation. Includes dynamic in-memory TLS certificate rotation via `internal/admission/CertReloader` without pod restarts.
- **Dynamic TLS Rotation Architecture**:
  - Webhook serves TLS handshakes via `tls.Config.GetCertificate` reading from lock-free `sync/atomic.Pointer[tls.Certificate]`.
  - Dual change detection:
    1. Periodic file stat/read for volume-mounted Secrets (`/etc/webhook/certs/tls.crt`, `tls.key`).
    2. Real-time Kubernetes API Secret watch via `client-go` on the exact configured namespaced Secret.
  - Initial startup: Validates initial certificate files; fails immediately if certificate is unavailable or corrupt.
  - Runtime reload: Validates key pair and x509 leaf certificate, verifies private key matches public key, and atomically swaps the active certificate pointer.
  - Failure semantics:
    - Invalid/corrupt Secret update: Retains last known-good certificate; logs actionable error; webhook continues serving without downtime.
    - Secret deletion: Retains last known-good certificate; logs warning; continues serving.
    - Watch disconnect: Reconnects automatically with exponential backoff while retaining active certificate.
    - Clean shutdown: Watchers terminate cleanly upon context cancellation without goroutine leaks.
  - Scope distinction: Server certificate rotation is fully automated and verified. CA certificate rotation and `ValidatingWebhookConfiguration.caBundle` rotation remain distinct operational procedures.
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
  - `Dynamic TLS Rotation`: Validated in-memory reload upon Secret update from Cert A to Cert B; verified zero pod restarts (`restartCount: 0`); confirmed live API server interception with rotated certificate; verified fallback retention upon invalid secret update.
  - `failurePolicy`: Verified `failurePolicy: Fail` blocks admission when webhook pods are scaled down.
- **TLS Handshake**: Successful mutual trust established using custom CA bundle and SAN-enabled TLS certificate on port 443 -> 8443.
- **RBAC**: ServiceAccount audited via `kubectl auth can-i --list` confirming strictly least-privilege read access, with namespaced `Role` restricting Secret access to `get, watch` on `kubeguard-tls` only.
