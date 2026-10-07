package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"strings"
)

type UnpinnedDigestRule struct{}

func (r *UnpinnedDigestRule) ID() string                         { return "KG-SEC-011" }
func (r *UnpinnedDigestRule) Title() string                      { return "Unpinned Image Digest" }
func (r *UnpinnedDigestRule) Description() string                { return "Containers should pin images by digest" }
func (r *UnpinnedDigestRule) Category() string                   { return "security" }
func (r *UnpinnedDigestRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *UnpinnedDigestRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding
	for i, c := range res.PodSpec.Containers {
		if !strings.Contains(c.Image, "@sha256:") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.containers[%d].image", i), "Container image is not pinned by digest.", "Tags can be modified; digests guarantee image immutability.", "Consider using image digest for immutable deployments."))
		}
	}
	for i, c := range res.PodSpec.InitContainers {
		if !strings.Contains(c.Image, "@sha256:") {
			results = append(results, newFinding(r, res, c.Name, fmt.Sprintf("spec.template.spec.initContainers[%d].image", i), "Container image is not pinned by digest.", "Tags can be modified; digests guarantee image immutability.", "Consider using image digest for immutable deployments."))
		}
	}
	return results
}
