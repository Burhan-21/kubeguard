package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	corev1 "k8s.io/api/core/v1"
)

type CapabilitiesRule struct{}

func (r *CapabilitiesRule) ID() string { return "KG-SEC-007" }
func (r *CapabilitiesRule) Title() string { return "Dangerous Capabilities" }
func (r *CapabilitiesRule) Description() string { return "Containers should not add dangerous capabilities" }
func (r *CapabilitiesRule) Category() string { return "security" }
func (r *CapabilitiesRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *CapabilitiesRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	
	checkCaps := func(cName string, sc *corev1.SecurityContext, prefix string, i int) {
		if sc == nil || sc.Capabilities == nil { return }
		for _, cap := range sc.Capabilities.Add {
			c := string(cap)
			if c == "SYS_ADMIN" || c == "ALL" {
				f := newFinding(r, res, cName, fmt.Sprintf("spec.template.spec.%s[%d].securityContext.capabilities", prefix, i), "Container adds dangerous capability.", "Capability grants excessive privileges.", "Drop unnecessary capabilities.")
				f.Severity = findings.SeverityBlock
				results = append(results, f)
			} else if c == "NET_ADMIN" || c == "SYS_PTRACE" || c == "NET_RAW" {
				results = append(results, newFinding(r, res, cName, fmt.Sprintf("spec.template.spec.%s[%d].securityContext.capabilities", prefix, i), "Container adds dangerous capability.", "Capability could be misused.", "Drop unnecessary capabilities."))
			}
		}
	}
	
	for i, c := range res.PodSpec.Containers {
		checkCaps(c.Name, c.SecurityContext, "containers", i)
	}
	for i, c := range res.PodSpec.InitContainers {
		checkCaps(c.Name, c.SecurityContext, "initContainers", i)
	}
	return results
}
