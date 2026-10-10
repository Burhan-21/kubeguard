# Security Requirements

**Author**: Shaikh Mohammed Burhan (GitHub: NexusForge21)

This document outlines the security posture of KubeGuard itself, and the boundaries it enforces.

## 1. Application Security
- **Least Privilege**: The admission controller RBAC must only request permissions necessary to read policies and handle admission reviews. For dynamic TLS certificate rotation, the controller is strictly granted namespaced read-only permissions (`get`, `watch`) on its designated TLS secret (`resourceNames: ["kubeguard-tls"]`). It must never be granted cluster-admin, wildcard write permissions, or Secret access across namespaces.
- **Webhook TLS & Dynamic Rotation**: All admission controller traffic must be encrypted via strictly validated TLS certificates. Dynamic certificate reload safely validates new x509 key pairs in memory before atomically updating active certificates. If corrupt or invalid Secret data is supplied, the webhook retains the last known-good certificate and emits an actionable error without restarting.
- **Data Confidentiality & Zero Secret Logging**: Private keys, certificate PEM bytes, and raw Kubernetes Secret data must NEVER be logged or leaked into error messages, audit events, or application logs. Only non-sensitive operational metadata (CommonName, SerialNumber, Expiration timestamp) may be logged.
- **No Secrets in Code**: Hardcoded credentials are strictly forbidden. Use service accounts, secrets, or environment variables.

## 2. Threat Modeling & Protections
- **YAML Parser Attacks**: Implement limits on parse depth and file size to prevent YAML bombs (Billion Laughs attack) and resource exhaustion.
- **Oversized Input**: Enforce a strict max payload size on the admission webhook to prevent memory exhaustion (OOM).
- **Policy Tampering**: In cluster environments, ensure ConfigMaps/Secrets containing KubeGuard policies are protected via Kubernetes RBAC.
- **Malicious Manifests**: The engine must safely sandbox string evaluation and not execute arbitrary code embedded in manifests.

## 3. Detection Capabilities
- **Sensitive Data Detection**: Rules to prevent hardcoded credentials or unencrypted secrets in manifests.
- **Container Security**: Enforcement of Pod Security Standards (PSS), e.g., dropping capabilities, non-root users, read-only root filesystems.
- **Supply-Chain Security**: Capability to warn on deprecated, unmaintained, or insecure image registries. (Note: Deep image scanning is out-of-scope; KubeGuard checks manifest integrity).

## 4. Operational Security
- **Admission Webhook Compromise**: If KubeGuard is compromised, fail-closed vs fail-open (`failurePolicy: Fail` vs `Ignore`) determines blast radius. Defaults should prioritize cluster stability.
- **Audit Logging**: Webhook ALLOW/DENY decisions must be loggable for security incident investigations.
- **Dependency Vulnerabilities**: CI/CD must strictly include `govulncheck` and standard SAST tooling to prevent supply chain compromise of the scanner itself.
