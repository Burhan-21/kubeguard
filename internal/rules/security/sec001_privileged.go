package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type PrivilegedRule struct{}

func (r *PrivilegedRule) ID() string { return "KG-SEC-001" }
func (r *PrivilegedRule) Title() string { return "Privileged Container" }
func (r *PrivilegedRule) Description() string { return "Containers must not run as privileged" }
func (r *PrivilegedRule) Category() string { return "security" }
func (r *PrivilegedRule) DefaultSeverity() findings.Severity { return findings.SeverityBlock }

func (r *PrivilegedRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].securityContext.privileged", i), "Container is configured as privileged.", "Privileged containers have access to all host devices.", "Remove privileged mode unless explicitly required."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if c.SecurityContext != nil && c.SecurityContext.Privileged != nil && *c.SecurityContext.Privileged {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].securityContext.privileged", i), "Container is configured as privileged.", "Privileged containers have access to all host devices.", "Remove privileged mode unless explicitly required."))
		}
	}
	return results
}
