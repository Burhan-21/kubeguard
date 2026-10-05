package reliability

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type ReadinessProbeRule struct{}

func (r *ReadinessProbeRule) ID() string { return "KG-REL-001" }
func (r *ReadinessProbeRule) Title() string { return "Missing Readiness Probe" }
func (r *ReadinessProbeRule) Description() string { return "Containers should have a readiness probe" }
func (r *ReadinessProbeRule) Category() string { return "reliability" }
func (r *ReadinessProbeRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *ReadinessProbeRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if c.ReadinessProbe == nil {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].readinessProbe", i), "Container has no readiness probe.", "Traffic might be sent to pods that are not ready.", "Define a readiness probe."))
		}
	}
	return results
}
