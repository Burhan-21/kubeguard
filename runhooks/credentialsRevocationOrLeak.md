# Credential Revocation or Leak

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
A credential (e.g., API token, private key, Kubernetes Secret) associated with KubeGuard or its deployment infrastructure has been exposed, leaked, or compromised.

## How to detect?
* **Audit Logs**: Unusual activity originating from KubeGuard identities.
* **Secret Scanning**: GitHub Advanced Security or other CI tools flag a leaked secret in the repository.
* **Alerts**: Alerts triggered by third-party services (e.g., AWS, GCP, GitHub) notifying of a credential leak.

## How to contain?
* **Immediate Revocation**: Immediately revoke the compromised credential at the source provider. Do not wait for investigation.
* **Never Print Leaked Credentials**: Do not paste or echo the compromised credentials into chat, tickets, or logs during containment.

## How to recover?
1. **Revoke the Credential**: Go to the provider (e.g., GitHub, cloud provider, Kubernetes) and delete/revoke the token.
2. **Secret Rotation**: Generate a new credential to replace the compromised one.
3. **CI Credential Cleanup**: Update the new credential in CI/CD variables (e.g., GitHub Secrets).
4. **Kubernetes Secret Rotation**: If the credential was a Kubernetes Secret used by the KubeGuard Webhook:
   ```bash
   kubectl delete secret kubeguard-certs -n kubeguard-system
   # Generate and apply new secrets based on installation docs
   ```
5. **Git History Investigation**: If the secret was committed to the repository, use tools like BFG Repo-Cleaner or `git filter-branch` to rewrite history and remove it. Force push the cleaned branch.

## How to verify recovery?
1. Verify that systems using the *new* credential operate successfully (e.g., CI pipelines pass, Admission Webhook responds).
2. Verify that systems attempting to use the *old* credential fail with authentication errors.

## When to escalate?
* If the revoked credential had widespread access and there is evidence of malicious use.
* If you cannot successfully rotate the credential and restore KubeGuard functionality within 1 hour.

## Post-Incident Review
Conduct a blameless post-mortem to determine how the credential was leaked and implement preventative measures (e.g., pre-commit hooks for secret scanning).
