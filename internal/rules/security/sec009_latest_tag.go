package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"strings"
)

type LatestTagRule struct{}

func (r *LatestTagRule) ID() string                         { return "KG-SEC-009" }
func (r *LatestTagRule) Title() string                      { return "Use of :latest Tag" }
func (r *LatestTagRule) Description() string                { return "Containers should not use the latest image tag" }
func (r *LatestTagRule) Category() string                   { return "security" }
func (r *LatestTagRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *LatestTagRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if strings.HasSuffix(c.Image, ":latest") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].image", i), "Container uses :latest image tag.", "The latest tag is mutable and can lead to unexpected behaviors.", "Pin to a specific version tag."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if strings.HasSuffix(c.Image, ":latest") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].image", i), "Container uses :latest image tag.", "The latest tag is mutable and can lead to unexpected behaviors.", "Pin to a specific version tag."))
		}
	}
	return results
}
