package security

import (
	"fmt"
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"strings"
)

type HostPathRule struct{}

func (r *HostPathRule) ID() string                         { return "KG-SEC-006" }
func (r *HostPathRule) Title() string                      { return "HostPath Volume Mounts" }
func (r *HostPathRule) Description() string                { return "Pods should not mount dangerous host paths" }
func (r *HostPathRule) Category() string                   { return "security" }
func (r *HostPathRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *HostPathRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil {
		return nil
	}
	var results []findings.Finding

	dangerousPaths := []string{"/", "/etc", "/var/run/docker.sock", "/var/lib/kubelet", "/sys"}

	for i, vol := range res.PodSpec.Volumes {
		if vol.HostPath != nil {
			path := vol.HostPath.Path
			isDangerous := false
			for _, dp := range dangerousPaths {
				if path == dp || strings.HasPrefix(path, dp+"/") {
					isDangerous = true
					break
				}
			}

			sev := findings.SeverityWarn
			msg := "Pod mounts host filesystem path."
			if isDangerous {
				sev = findings.SeverityBlock
				msg = fmt.Sprintf("Pod mounts dangerous host path: %s", path)
			}

			f := newFinding(r, res, "", fmt.Sprintf("spec.template.spec.volumes[%d].hostPath", i), msg, "Host paths can allow container escape or modification of host filesystem.", "Avoid using hostPath volumes, use PVs or EmptyDir instead.")
			f.Severity = sev
			results = append(results, f)
		}
	}
	return results
}
