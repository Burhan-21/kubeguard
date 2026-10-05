package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type NetworkPolicyRule struct{}

func (r *NetworkPolicyRule) ID() string { return "KG-REL-011" }
func (r *NetworkPolicyRule) Title() string { return "Network Policy Recommended" }
func (r *NetworkPolicyRule) Description() string { return "Workloads should be restricted by NetworkPolicies" }
func (r *NetworkPolicyRule) Category() string { return "reliability" }
func (r *NetworkPolicyRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *NetworkPolicyRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.Kind == "Deployment" || res.Kind == "StatefulSet" || res.Kind == "DaemonSet" {
		return []findings.Finding{
			newFinding(r, res, "", "", "Consider restricting network access with a NetworkPolicy.", "Pods are open to all traffic by default.", "Implement a default deny NetworkPolicy and explicitly allow required traffic."),
		}
	}
	return nil
}
