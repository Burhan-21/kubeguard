package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type PDBRule struct{}

func (r *PDBRule) ID() string { return "KG-REL-008" }
func (r *PDBRule) Title() string { return "Missing PodDisruptionBudget" }
func (r *PDBRule) Description() string { return "Workloads with multiple replicas should have a PodDisruptionBudget" }
func (r *PDBRule) Category() string { return "reliability" }
func (r *PDBRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *PDBRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if (res.Kind == "Deployment" || res.Kind == "StatefulSet") && res.Replicas != nil && *res.Replicas > 1 {
		return []findings.Finding{
			newFinding(r, res, "", "", "Consider creating a PodDisruptionBudget for this workload.", "PDBs prevent disruptions during voluntary node drains.", "Create a PDB with appropriate minAvailable/maxUnavailable."),
		}
	}
	return nil
}
