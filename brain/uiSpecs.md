# UI Specifications

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

## v1 Interfaces
For version 1, KubeGuard will have **NO Web UI**. 
The primary interfaces are:
1. Command Line Interface (CLI)
2. CI/CD integration (via CLI exit codes)
3. Admission Controller Webhook (Kubernetes API)
4. Programmatic outputs (SARIF, JSON)

## Future Scope (Optional Dashboard)
Do **not** build a dashboard for v1. If a Web UI is developed in future versions, it must include:
- **Scan Overview**: High-level metrics on pass/fail rates.
- **Policy Results**: Detailed breakdown of applied policies.
- **Security/Reliability Findings**: Filterable lists of vulnerabilities and misconfigurations.
- **Admission Events**: Audit log of webhook ALLOW/DENY decisions.
- **Rule Statistics**: Which rules trigger most often.
- **Historical Scans**: Timeline of changes in compliance over time.
