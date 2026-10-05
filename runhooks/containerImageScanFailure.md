# Container Image Scan Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
KubeGuard's container image scanning integration failed (e.g., scanner unavailable, image pull failure, registry failure, vuln DB unavailable).

## How to detect?
* CI pipeline fails during the image scanning step.
* KubeGuard logs indicate failure to reach the scanner API or registry.

## How to contain?
* **Distinction**: Crucially, distinguish a 'scan failed' event from a 'scan completed and image is safe' event.
* **Never** treat an unavailable scanner as a successful scan. The pipeline or admission request MUST fail if the scan cannot be completed.

## How to recover?
1. **Scanner Unavailable / Vuln DB Down**: Check the status of the external scanning service (e.g., Trivy, Clair). Wait and retry.
2. **Image Pull/Registry Failure**: Ensure registry credentials are correct and the registry is online. Check for rate limits (e.g., Docker Hub).

## How to verify recovery?
Manually trigger the image scan in CI or locally and ensure it completes successfully and returns a valid vulnerability report.

## When to escalate?
* If the external scanning service or registry experiences a prolonged outage.
