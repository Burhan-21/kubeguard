package reliability

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ResourceRequestsRule struct{}

func (r *ResourceRequestsRule) ID() string { return "KG-REL-004" }
func (r *ResourceRequestsRule) Title() string { return "Missing Resource Requests" }
func (r *ResourceRequestsRule) Description() string { return "Containers should specify resource requests" }
func (r *ResourceRequestsRule) Category() string { return "reliability" }
func (r *ResourceRequestsRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *ResourceRequestsRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	
	for i, c := range res.PodSpec.Containers {
		if c.Resources.Requests == nil || (c.Resources.Requests.Cpu().IsZero() && c.Resources.Requests.Memory().IsZero()) {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].resources.requests", i), "Container has no CPU/memory resource requests.", "Scheduler cannot make optimal placement decisions without resource requests.", "Specify CPU and memory requests."))
		}
	}
	return results
}
