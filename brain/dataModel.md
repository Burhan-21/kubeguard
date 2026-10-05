# Data Model

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## Internal Data Structures (Go Types)

### Resource & Evaluation
- `Resource`: Represents raw parsed input.
- `NormalizedResource`: Standardized internal representation for the Policy Engine.
- `ScanResult`: Aggregated outcome of a full scan.
- `EvaluationResult`: Outcome of evaluating a single rule on a single resource.

### Policies & Rules
- `Policy`: A collection of rules.
- `PolicyProfile`: A defined baseline (e.g., 'strict', 'baseline').
- `Rule`: Interface for evaluating a specific check.
- `RuleMetadata`: Information about the rule (ID, name, description, reference).

### Finding
```go
type Finding struct {
    RuleID      string   `json:"ruleId"`
    Severity    Severity `json:"severity"`
    Category    string   `json:"category"`
    Kind        string   `json:"kind"`
    Name        string   `json:"name"`
    Namespace   string   `json:"namespace,omitempty"`
    Container   string   `json:"container,omitempty"`
    FieldPath   string   `json:"fieldPath,omitempty"`
    Message     string   `json:"message"`
    Why         string   `json:"why,omitempty"`
    Remediation string   `json:"remediation,omitempty"`
    References  []string `json:"references,omitempty"`
}
```

### FindingSeverity
Enum representing strictness: `PASS`, `WARN`, `BLOCK`.

## JSON Output Example
```json
{
  "scanId": "12345-abcde",
  "timestamp": "2026-10-05T10:00:00Z",
  "summary": {
    "totalScanned": 1,
    "passed": 0,
    "warnings": 1,
    "blocks": 1
  },
  "findings": [
    {
      "ruleId": "KG-SEC-001",
      "severity": "BLOCK",
      "category": "Security",
      "resource": "Deployment/myapp",
      "namespace": "default",
      "kind": "Deployment",
      "name": "myapp",
      "container": "main-app",
      "fieldPath": "spec.template.spec.containers[0].securityContext.runAsNonRoot",
      "message": "Container may run as root.",
      "why": "Running as root allows attackers to exploit container runtime vulnerabilities.",
      "remediation": "Set runAsNonRoot: true in the container's securityContext.",
      "references": ["https://kubernetes.io/docs/concepts/security/pod-security-standards/"]
    }
  ]
}
```
