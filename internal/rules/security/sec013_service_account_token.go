package security

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ServiceAccountTokenRule struct{}

func (r *ServiceAccountTokenRule) ID() string { return "KG-SEC-013" }
func (r *ServiceAccountTokenRule) Title() string { return "Service Account Token Automount" }
func (r *ServiceAccountTokenRule) Description() string { return "Service account tokens should not be automounted unless required" }
func (r *ServiceAccountTokenRule) Category() string { return "security" }
func (r *ServiceAccountTokenRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *ServiceAccountTokenRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	if res.PodSpec.AutomountServiceAccountToken == nil || *res.PodSpec.AutomountServiceAccountToken {
		results = append(results, newFinding(r, res, "", "spec.template.spec.automountServiceAccountToken", "Service account token is automounted.", "Automounted tokens can be used by an attacker if the container is compromised.", "Set automountServiceAccountToken: false unless required."))
	}
	return results
}
