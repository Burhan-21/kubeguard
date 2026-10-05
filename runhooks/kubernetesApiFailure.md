# Kubernetes API Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
KubeGuard or its deployment tools (like Helm) cannot communicate with the Kubernetes API server due to unavailability, timeouts, auth failures, or throttling.

## How to detect?
* `kubectl` commands time out or return 5xx errors.
* KubeGuard pods log errors like "context deadline exceeded" or "Unauthorized".
* API server metrics show high error rates or latency.

## How to contain?
* Stop automated CI/CD deployments targeting the cluster to avoid exacerbating the load.

## How to recover?
1. **Unavailability/Timeout**: Wait for the control plane to recover. If managed (EKS/GKE/AKS), check provider status. If self-managed, check API server logs. Use exponential backoff for retries.
2. **Auth Failure**: Verify the kubeconfig or service account tokens used by KubeGuard or CI are valid and have not expired.
3. **Throttling**: If KubeGuard is being throttled, reduce the frequency of API calls or check for retry loops in controllers.

## How to verify recovery?
Run `kubectl get nodes` or `kubectl get namespaces` successfully.

## When to escalate?
* If the API server remains unavailable for an extended period.
