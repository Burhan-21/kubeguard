package security

import (
	"fmt"
	"strings"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type MissingTagRule struct{}

func (r *MissingTagRule) ID() string { return "KG-SEC-010" }
func (r *MissingTagRule) Title() string { return "Missing Image Tag" }
func (r *MissingTagRule) Description() string { return "Containers should specify an image tag" }
func (r *MissingTagRule) Category() string { return "security" }
func (r *MissingTagRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *MissingTagRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if !strings.Contains(c.Image, ":") || strings.HasSuffix(c.Image, ":") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].image", i), "Container image has no tag specified.", "Images without tags default to latest which is mutable.", "Specify an explicit image tag."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if !strings.Contains(c.Image, ":") || strings.HasSuffix(c.Image, ":") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].image", i), "Container image has no tag specified.", "Images without tags default to latest which is mutable.", "Specify an explicit image tag."))
		}
	}
	return results
}
