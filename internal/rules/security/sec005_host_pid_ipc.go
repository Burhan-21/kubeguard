package security

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type HostPIDIPCRule struct{}

func (r *HostPIDIPCRule) ID() string { return "KG-SEC-005" }
func (r *HostPIDIPCRule) Title() string { return "Host PID/IPC Usage" }
func (r *HostPIDIPCRule) Description() string { return "Pods should not use host PID or IPC namespaces" }
func (r *HostPIDIPCRule) Category() string { return "security" }
func (r *HostPIDIPCRule) DefaultSeverity() findings.Severity { return findings.SeverityBlock }

func (r *HostPIDIPCRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	if res.PodSpec.HostPID {
		results = append(results, newFinding(r, res, "", "spec.template.spec.hostPID", "Pod shares host PID namespace.", "Allows pod to see all processes on the host.", "Remove hostPID: true."))
	}
	if res.PodSpec.HostIPC {
		results = append(results, newFinding(r, res, "", "spec.template.spec.hostIPC", "Pod shares host IPC namespace.", "Allows pod to communicate with host IPC mechanisms.", "Remove hostIPC: true."))
	}
	return results
}
