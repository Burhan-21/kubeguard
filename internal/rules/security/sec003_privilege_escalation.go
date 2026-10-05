package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type PrivilegeEscalationRule struct{}

func (r *PrivilegeEscalationRule) ID() string { return "KG-SEC-003" }
func (r *PrivilegeEscalationRule) Title() string { return "Privilege Escalation Allowed" }
func (r *PrivilegeEscalationRule) Description() string { return "Containers should not allow privilege escalation" }
func (r *PrivilegeEscalationRule) Category() string { return "security" }
func (r *PrivilegeEscalationRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *PrivilegeEscalationRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.SecurityContext == nil || c.SecurityContext.AllowPrivilegeEscalation == nil || *c.SecurityContext.AllowPrivilegeEscalation {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].securityContext.allowPrivilegeEscalation", i), "Container allows privilege escalation.", "Processes could gain more privileges than their parent.", "Set allowPrivilegeEscalation: false."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if c.SecurityContext == nil || c.SecurityContext.AllowPrivilegeEscalation == nil || *c.SecurityContext.AllowPrivilegeEscalation {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].securityContext.allowPrivilegeEscalation", i), "Container allows privilege escalation.", "Processes could gain more privileges than their parent.", "Set allowPrivilegeEscalation: false."))
		}
	}
	return results
}
