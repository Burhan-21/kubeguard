package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type SingleReplicaRule struct{}

func (r *SingleReplicaRule) ID() string    { return "KG-REL-006" }
func (r *SingleReplicaRule) Title() string { return "Single Replica" }
func (r *SingleReplicaRule) Description() string {
	return "Workloads should have multiple replicas for high availability"
}
func (r *SingleReplicaRule) Category() string                   { return "reliability" }
func (r *SingleReplicaRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *SingleReplicaRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.Kind != "Deployment" && res.Kind != "StatefulSet" {
		return nil
	}
	var results []findings.Finding

	if res.Replicas != nil && *res.Replicas == 1 {
		results = append(results, newFinding(r, res, "", "spec.replicas", "Workload has only a single replica.", "Single replicas provide no high availability if the pod or node fails.", "Consider multiple replicas for high availability."))
	}
	return results
}
