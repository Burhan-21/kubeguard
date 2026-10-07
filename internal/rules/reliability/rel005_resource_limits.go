package reliability

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ResourceLimitsRule struct{}

func (r *ResourceLimitsRule) ID() string                         { return "KG-REL-005" }
func (r *ResourceLimitsRule) Title() string                      { return "Missing Resource Limits" }
func (r *ResourceLimitsRule) Description() string                { return "Containers should specify resource limits" }
func (r *ResourceLimitsRule) Category() string                   { return "reliability" }
func (r *ResourceLimitsRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *ResourceLimitsRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding

	for i, c := range res.PodSpec.Containers {
		if c.Resources.Limits == nil || (c.Resources.Limits.Cpu().IsZero() && c.Resources.Limits.Memory().IsZero()) {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].resources.limits", i), "Container has no CPU/memory resource limits.", "Containers without limits can consume all node resources.", "Specify CPU and memory limits."))
		}
	}
	return results
}
