# GitHub Actions Failure

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)
**License**: Apache 2.0

## What failed?
A GitHub Actions CI workflow for KubeGuard failed (e.g., test failure, security scan failure, build failure, Docker failure, SARIF upload failure, dependency failure).

## How to detect?
* GitHub PR shows a red X.
* Email or slack notifications from GitHub Actions.

## How to contain?
* Do not bypass branch protections to merge failing code.

## How to recover?
1. **Analyze Logs**: Review the detailed logs in the GitHub Actions UI.
2. **Define Retry vs Defect**:
   * If the failure is a transient network error (e.g., pulling a dependency), retry the job.
   * If the failure is a deterministic test failure or build error, it is an actual defect. Fix the code and push a new commit.
3. **SARIF Upload Failure**: Ensure the SARIF output format from KubeGuard is strictly valid according to the SARIF schema.

## How to verify recovery?
The PR checks pass and all workflows show a green checkmark.

## When to escalate?
* If GitHub Actions itself is experiencing an outage.
