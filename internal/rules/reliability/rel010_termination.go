package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type TerminationRule struct{}

func (r *TerminationRule) ID() string { return "KG-REL-010" }
func (r *TerminationRule) Title() string { return "Short Termination Grace Period" }
func (r *TerminationRule) Description() string { return "Pods should have a sufficient termination grace period" }
func (r *TerminationRule) Category() string { return "reliability" }
func (r *TerminationRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *TerminationRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.PodSpec == nil { return nil }
	var results []findings.Finding
	
	if res.PodSpec.TerminationGracePeriodSeconds != nil && *res.PodSpec.TerminationGracePeriodSeconds < 5 {
		results = append(results, newFinding(r, res, "", "spec.template.spec.terminationGracePeriodSeconds", "Termination grace period is very short.", "Containers might be killed before cleanly shutting down.", "Set a reasonable terminationGracePeriodSeconds (default 30s)."))
	}
	return results
}
