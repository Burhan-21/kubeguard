package security

import (
	"fmt"

	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	corev1 "k8s.io/api/core/v1"
)

type RunAsRootRule struct{}

func (r *RunAsRootRule) ID() string                         { return "KG-SEC-002" }
func (r *RunAsRootRule) Title() string                      { return "Run as Root" }
func (r *RunAsRootRule) Description() string                { return "Containers should not run as root" }
func (r *RunAsRootRule) Category() string                   { return "security" }
func (r *RunAsRootRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *RunAsRootRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding

	podNonRoot := res.PodSpec.SecurityContext != nil && res.PodSpec.SecurityContext.RunAsNonRoot != nil && *res.PodSpec.SecurityContext.RunAsNonRoot
	podUserNonZero := res.PodSpec.SecurityContext != nil && res.PodSpec.SecurityContext.RunAsUser != nil && *res.PodSpec.SecurityContext.RunAsUser > 0

	checkContainer := func(cName string, sc *corev1.SecurityContext, prefix string, i int) {
		nonRoot := podNonRoot
		userNonZero := podUserNonZero
		if sc != nil {
			if sc.RunAsNonRoot != nil {
				nonRoot = *sc.RunAsNonRoot
			}
			if sc.RunAsUser != nil {
				userNonZero = *sc.RunAsUser > 0
			}
		}
		if !nonRoot && !userNonZero {
			results = append(results, newFinding(r, res, cName, fmt.Sprintf("spec.template.spec.%s[%d].securityContext", prefix, i), "Container may run as root.", "Running as root increases the impact of container breakout vulnerabilities.", "Set runAsNonRoot: true or runAsUser to a non-zero value."))
		}
	}

	for i, c := range res.PodSpec.Containers {
		checkContainer(c.Name, c.SecurityContext, "containers", i)
	}
	for i, c := range res.PodSpec.InitContainers {
		checkContainer(c.Name, c.SecurityContext, "initContainers", i)
	}
	return results
}
