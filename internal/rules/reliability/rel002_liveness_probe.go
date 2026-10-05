package reliability

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type LivenessProbeRule struct{}

func (r *LivenessProbeRule) ID() string { return "KG-REL-002" }
func (r *LivenessProbeRule) Title() string { return "Missing Liveness Probe" }
func (r *LivenessProbeRule) Description() string { return "Containers should have a liveness probe" }
func (r *LivenessProbeRule) Category() string { return "reliability" }
func (r *LivenessProbeRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *LivenessProbeRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.LivenessProbe == nil {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].livenessProbe", i), "Container has no liveness probe.", "Pod will not be restarted if it deadlocks.", "Define a liveness probe."))
		}
	}
	return results
}
