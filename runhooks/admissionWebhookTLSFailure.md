# Admission Webhook TLS Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
The Kubernetes API server cannot communicate securely with the KubeGuard Admission Webhook due to a TLS or certificate-related issue.

## How to detect?
* `kubectl apply` commands fail with TLS errors (e.g., "x509: certificate signed by unknown authority", "x509: certificate has expired or is not yet valid").
* KubeGuard pod logs may show TLS handshake errors.

## How to contain?
TLS failures typically block the webhook entirely. Containment involves diagnosing the specific TLS issue immediately.

## How to recover?
1. **Identify the Issue**:
   * **Expired Cert**: The webhook certificate has passed its expiration date.
   * **Invalid Cert / Wrong CA**: The `caBundle` in the `ValidatingWebhookConfiguration` does not match the CA that signed the webhook's certificate.
   * **Secret Missing**: The Kubernetes Secret containing the TLS keypair is deleted or unavailable.
2. **Rotate Certificates**:
   * If using `cert-manager`, check the `Certificate` resource status and force a renewal if necessary.
   * If using manual certificates, generate a new CA and keypair, update the Secret in the `kubeguard-system` namespace, and patch the `caBundle` in the `ValidatingWebhookConfiguration`.
3. **Restart Pods**: After updating the TLS Secret, restart the KubeGuard webhook pods so they load the new certificates.
   ```bash
   kubectl rollout restart deployment kubeguard -n kubeguard-system
   ```

## How to verify recovery?
Execute an `AdmissionReview` by applying a resource and verifying the TLS error is gone and the webhook evaluates the resource.
```bash
kubectl apply -f test/fixtures/safe-deployment.yaml
```

## When to escalate?
* If you cannot successfully rotate the certificates or update the `caBundle` after 30 minutes.
* If `cert-manager` is failing cluster-wide and requires infrastructure team intervention.
