# API & Interface Contracts

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

KubeGuard exposes functionality exclusively through specific, stable interfaces. **There is no REST API** unless explicitly implemented in future phases.

## 1. CLI Interface
The primary human and CI interface.

### Commands
- `kubeguard scan <path/to/yaml>`
- `kubeguard admission --port <port> --tls-cert <cert> --tls-key <key>`
- `kubeguard policy list`
- `kubeguard version`

### Example Usage
```bash
kubeguard scan ./k8s-manifests \
  --profile strict \
  --output sarif \
  --fail-on BLOCK
```

### Exit Codes
- `0`: PASS (No findings, or only findings below `--fail-on` threshold)
- `1`: WARN (Findings reached WARN threshold)
- `2`: BLOCK (Findings reached BLOCK threshold, critical security/reliability issues)
- `3`: TOOL ERROR (Invalid configuration, parser failure, system error)

## 2. Admission API Contract
The interface for the Kubernetes ValidatingAdmissionWebhook.

### Request
Expects standard Kubernetes `AdmissionReview` v1 payload.
```json
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "request": {
    "uid": "12345",
    "kind": {"group":"apps","version":"v1","kind":"Deployment"},
    "resource": {"group":"apps","version":"v1","resource":"deployments"},
    "namespace": "default",
    "operation": "CREATE",
    "object": { ... }
  }
}
```

### Response
Returns an `AdmissionReview` v1 response.
```json
{
  "apiVersion": "admission.k8s.io/v1",
  "kind": "AdmissionReview",
  "response": {
    "uid": "12345",
    "allowed": false,
    "status": {
      "code": 403,
      "message": "KubeGuard: DENIED - Container may run as root [KG-SEC-001]"
    }
  }
}
```
