package security

import (
	"fmt"
	"strings"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type EmbeddedSecretRule struct{}

func (r *EmbeddedSecretRule) ID() string { return "KG-SEC-012" }
func (r *EmbeddedSecretRule) Title() string { return "Embedded Secrets" }
func (r *EmbeddedSecretRule) Description() string { return "Secrets should not be embedded as plain environment variables" }
func (r *EmbeddedSecretRule) Category() string { return "security" }
func (r *EmbeddedSecretRule) DefaultSeverity() findings.Severity { return findings.SeverityBlock }

func (r *EmbeddedSecretRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	
	secretNames := []string{"PASSWORD", "SECRET", "TOKEN", "API_KEY"}
	
	for i, c := range res.PodSpec.Containers {
		for j, env := range c.Env {
			if env.Value != "" {
				upperName := strings.ToUpper(env.Name)
				for _, sn := range secretNames {
					if strings.Contains(upperName, sn) {
						results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].env[%d]", i, j), "Potential secret embedded as plain environment variable.", "Plaintext secrets in env vars can be easily exposed.", "Use Kubernetes Secrets with valueFrom.secretKeyRef."))
						break
					}
				}
			}
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		for j, env := range c.Env {
			if env.Value != "" {
				upperName := strings.ToUpper(env.Name)
				for _, sn := range secretNames {
					if strings.Contains(upperName, sn) {
						results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].env[%d]", i, j), "Potential secret embedded as plain environment variable.", "Plaintext secrets in env vars can be easily exposed.", "Use Kubernetes Secrets with valueFrom.secretKeyRef."))
						break
					}
				}
			}
		}
	}
	return results
}
