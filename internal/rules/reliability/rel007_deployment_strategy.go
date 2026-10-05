package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	appsv1 "k8s.io/api/apps/v1"
)

type DeploymentStrategyRule struct{}

func (r *DeploymentStrategyRule) ID() string { return "KG-REL-007" }
func (r *DeploymentStrategyRule) Title() string { return "Deployment Strategy" }
func (r *DeploymentStrategyRule) Description() string { return "Deployments should use RollingUpdate strategy" }
func (r *DeploymentStrategyRule) Category() string { return "reliability" }
func (r *DeploymentStrategyRule) DefaultSeverity() findings.Severity { return findings.SeverityWarn }

func (r *DeploymentStrategyRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if res.Kind != "Deployment" { return nil }
	var results []findings.Finding
	
	if res.Strategy == nil || res.Strategy.Type == appsv1.RecreateDeploymentStrategyType {
		results = append(results, newFinding(r, res, "", "spec.strategy", "Deployment uses Recreate strategy which causes downtime.", "Recreate strategy kills all pods before creating new ones.", "Use RollingUpdate strategy for zero-downtime deployments."))
	}
	return results
}
