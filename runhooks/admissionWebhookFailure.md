# Admission Webhook Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
The KubeGuard Validating Admission Webhook is failing to process `AdmissionReview` requests from the Kubernetes API server.

## How to detect?
* Users receive errors when attempting to apply resources via `kubectl` (e.g., "Internal error occurred: failed calling webhook").
* KubeGuard webhook pod logs show exceptions, timeouts, or malformed responses.
* Kubernetes API server metrics indicate high latency or errors for the KubeGuard webhook.

## How to contain?
The safety trade-off is governed by the `failurePolicy` in the `ValidatingWebhookConfiguration`:
* **Fail**: The API server rejects the request. Safe, but impacts availability. KubeGuard recommends this setting for strict security.
* **Ignore**: The API server admits the request. Available, but compromises security.
* To contain a critical availability incident, you may temporarily bypass the webhook by labeling the affected namespace so it is excluded by the webhook's `namespaceSelector`.

## How to recover?
1. **Check Pod Health**: Ensure the KubeGuard webhook pods are running and passing readiness probes.
   ```bash
   kubectl get pods -n kubeguard-system
   ```
2. **Check Logs**: Inspect pod logs for exceptions or internal errors.
   ```bash
   kubectl logs -l app=kubeguard -n kubeguard-system
   ```
3. **Timeout Issues**: If the webhook is timing out, ensure the KubeGuard pods have sufficient CPU/Memory resources, and check network connectivity between the API server and the pods.
4. **Malformed Response**: If the webhook is returning malformed JSON, check the KubeGuard version and ensure it matches the expected API version of the cluster.
5. Restart the webhook pods if they appear deadlocked or in a bad state.

## How to verify recovery?
Apply a test resource that should be allowed, and one that should be denied, and verify the correct behavior.
```bash
kubectl apply -f test/fixtures/safe-deployment.yaml
kubectl apply -f test/fixtures/privileged-pod.yaml
```

## When to escalate?
* If the webhook continues to fail after pod restarts and resource adjustments.
* If the issue is a suspected bug in the KubeGuard webhook server implementation.
