# Error Handling Protocol

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

Never expose stack traces to normal users. All errors must be wrapped, contextualized, and categorized.

## Error Categories

| Category | Description | Exit Code | Fail Behavior (Admission) |
|---|---|---|---|
| `PARSE_ERROR` | Malformed YAML/JSON. | 3 | Fail-closed (Reject) |
| `UNSUPPORTED_RESOURCE` | Resource kind not recognized/supported. | 0 (Warn/Skip) | Fail-open (Allow/Skip) |
| `INVALID_POLICY` | Corrupted or invalid profile config. | 3 | Fail-closed (Reject) |
| `POLICY_ENGINE_ERROR` | Internal logic failure during evaluation. | 3 | Configurable via `failurePolicy` |
| `CONFIGURATION_ERROR` | Invalid CLI flags or startup parameters. | 3 | N/A |
| `ADMISSION_ERROR` | Malformed AdmissionReview request. | HTTP 400 | Fail-closed |
| `TLS_ERROR` | Certificate/handshake failure. | 3 | Fail-closed |
| `INTERNAL_ERROR` | Unexpected panic or nil pointer. | 3 | Configurable via `failurePolicy` |

## Principles
1. **User-Facing Messages**: Must be actionable. Explain *what* went wrong and *how* to fix it.
2. **Logging Level**: Internal errors log at `ERROR`. Unsupported resources log at `WARN` or `DEBUG`.
3. **Retry Behavior**: KubeGuard CLI does not retry (stateless). Admission webhook leaves retries to the Kubernetes API Server.
4. **Stack Traces**: Only emitted in debug/trace modes or in developer logs.
