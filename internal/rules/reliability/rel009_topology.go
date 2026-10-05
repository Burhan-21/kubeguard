package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type TopologyRule struct{}

func (r *TopologyRule) ID() string { return "KG-REL-009" }
func (r *TopologyRule) Title() string { return "Missing Topology Spread" }
func (r *TopologyRule) Description() string { return "Workloads should use topology spread or anti-affinity" }
func (r *TopologyRule) Category() string { return "reliability" }
func (r *TopologyRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *TopologyRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if (res.Kind == "Deployment" || res.Kind == "StatefulSet") && res.Replicas != nil && *res.Replicas > 1 && res.PodSpec != nil {
		hasAffinity := res.PodSpec.Affinity != nil && res.PodSpec.Affinity.PodAntiAffinity != nil
		hasTopology := len(res.PodSpec.TopologySpreadConstraints) > 0
		
		if !hasAffinity && !hasTopology {
			return []findings.Finding{
				newFinding(r, res, "", "spec.template.spec.topologySpreadConstraints", "No topology spread or anti-affinity configured.", "All replicas might be scheduled on the same node.", "Add pod anti-affinity or topology spread constraints for resilience."),
			}
		}
	}
	return nil
}
