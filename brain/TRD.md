# Technical Requirements Document (TRD)

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Core Technologies
- **Language**: Go 1.23+
- **Kubernetes Compatibility**: 1.28-1.31 (Note: Versions marked as NOT YET TESTED until explicitly verified)

## Supported Resources
- Deployment
- StatefulSet
- DaemonSet
- Pod
- Job
- CronJob
- Service

## Interfaces
1. **CLI**: Cobra-based framework. Core commands: `scan`, `policy`, `version`.
2. **Admission Controller**: ValidatingAdmissionWebhook using AdmissionReview v1.
3. **CI/CD**: GitHub Actions.

## Engine Requirements
- **Policy Engine**: Deterministic, profile-based (e.g., baseline, strict, custom).
- **Rule Engine**: Interface-based, combining security and reliability checks.

## Delivery & Integration
- **Distribution**: Binaries, Docker images, Helm chart.
- **Output Formats**: Human-readable terminal output, JSON, SARIF.

## Performance Targets (NOT YET MEASURED)
- < 100ms for 10 resources.
- < 1s for 100 resources.
- < 10s for 1000 resources.

## Security & Reliability
- TLS for webhooks.
- RBAC with least privilege.
- No secrets stored in code.
- Testing: unit, integration, race detection, admission tests.
