package reliability

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type StartupProbeRule struct{}

func (r *StartupProbeRule) ID() string    { return "KG-REL-003" }
func (r *StartupProbeRule) Title() string { return "Missing Startup Probe" }
func (r *StartupProbeRule) Description() string {
	return "Containers with slow initialization should have a startup probe"
}
func (r *StartupProbeRule) Category() string                   { return "reliability" }
func (r *StartupProbeRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *StartupProbeRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.StartupProbe == nil && c.LivenessProbe != nil {
			delay := c.LivenessProbe.InitialDelaySeconds
			if delay >= 30 || delay == 0 {
				results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].startupProbe", i), "Container may benefit from a startup probe.", "Liveness probe might kill container during slow startup.", "Add a startup probe for containers with slow initialization."))
			}
		}
	}
	return results
}
