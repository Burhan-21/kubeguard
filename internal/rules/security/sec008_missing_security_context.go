package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type MissingSecurityContextRule struct{}

func (r *MissingSecurityContextRule) ID() string    { return "KG-SEC-008" }
func (r *MissingSecurityContextRule) Title() string { return "Missing Security Context" }
func (r *MissingSecurityContextRule) Description() string {
	return "Containers should define a security context"
}
func (r *MissingSecurityContextRule) Category() string { return "security" }
func (r *MissingSecurityContextRule) DefaultSeverity() findings.Severity {
	return findings.SeverityWarn
}

func (r *MissingSecurityContextRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.SecurityContext == nil {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].securityContext", i), "Container has no security context defined.", "Security contexts limit container privileges.", "Define a securityContext with appropriate restrictions."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if c.SecurityContext == nil {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].securityContext", i), "Container has no security context defined.", "Security contexts limit container privileges.", "Define a securityContext with appropriate restrictions."))
		}
	}
	return results
}
