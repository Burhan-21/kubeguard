package security

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type HostNetworkRule struct{}

func (r *HostNetworkRule) ID() string    { return "KG-SEC-004" }
func (r *HostNetworkRule) Title() string { return "Host Network Usage" }
func (r *HostNetworkRule) Description() string {
	return "Pods should not use the host network namespace"
}
func (r *HostNetworkRule) Category() string                   { return "security" }
func (r *HostNetworkRule) DefaultSeverity() findings.Severity { return findings.SeverityBlock }

func (r *HostNetworkRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding
	if res.PodSpec.HostNetwork {
		results = append(results, newFinding(r, res, "", "spec.template.spec.hostNetwork", "Pod uses host network namespace.", "Allows pod to access host's network interfaces and potentially snoop traffic.", "Remove hostNetwork unless explicitly required for network plugins."))
	}
	return results
}
