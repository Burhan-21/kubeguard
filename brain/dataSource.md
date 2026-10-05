# Data Sources

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Primary Data Sources (Runtime & Configuration)
- **Kubernetes YAML Files**: Local or CI/CD static manifests.
- **K8s API Objects**: Fetched dynamically if running against a live cluster.
- **AdmissionReview**: JSON payload provided by the K8s API server during webhook invocation.
- **Policy YAML**: Custom user-defined profiles and rule toggles.
- **KubeGuard Config**: Application configuration parameters.

## Reference Documentation (Knowledge Base)
- Kubernetes Official Documentation
- Pod Security Standards (PSS)
- OWASP Kubernetes Top 10
- CIS Kubernetes Benchmarks
- NIST Application Container Security Guide

## Important Distinctions
- **No External Network Calls**: The KubeGuard scanner must operate entirely offline or intra-cluster. It must **not** phone home, make external network calls to fetch remote schemas, or dynamically query the internet during a scan.
- **Configuration vs. Runtime**: Reference configurations are baked into the binary or supplied via local config files; runtime data is strictly what is passed via CLI args, stdin, or Webhook payloads.
