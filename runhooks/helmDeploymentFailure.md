# Helm Deployment Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
The deployment of KubeGuard via Helm failed (e.g., failed install, failed upgrade, webhook deployment failure, RBAC failure, resource scheduling failure).

## How to detect?
* `helm install` or `helm upgrade` command fails.
* Pods in the `kubeguard-system` namespace are in CrashLoopBackOff or Pending state.

## How to contain?
* If an upgrade fails, avoid making further manual changes to the resources. Use Helm to manage the state.

## How to recover?
1. Check helm status: `helm status kubeguard -n kubeguard-system`
2. Check helm history: `helm history kubeguard -n kubeguard-system`
3. Inspect pod failures: `kubectl describe pod -l app=kubeguard -n kubeguard-system`
4. **Rollback**: If an upgrade failed and the cluster is in a bad state, rollback to the previous successful release:
   ```bash
   helm rollback kubeguard <revision-number> -n kubeguard-system
   ```
5. Fix the underlying issue (e.g., missing RBAC permissions, insufficient nodes for scheduling) and retry the deployment.

## How to verify recovery?
Verify that all KubeGuard pods are Running and Ready, and that the `helm status` shows deployed successfully.

## When to escalate?
* If Helm becomes stuck in a "pending-upgrade" or "pending-install" state that cannot be resolved with a rollback.
