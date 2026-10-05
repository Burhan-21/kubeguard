package reliability

import (
	"github.com/Burhan-21/kubeguard/internal/policy"
)

func AllReliabilityRules() []policy.Rule {
	return []policy.Rule{
		&ReadinessProbeRule{},
		&LivenessProbeRule{},
		&StartupProbeRule{},
		&ResourceRequestsRule{},
		&ResourceLimitsRule{},
		&SingleReplicaRule{},
		&DeploymentStrategyRule{},
		&PDBRule{},
		&TopologyRule{},
		&TerminationRule{},
		&NetworkPolicyRule{},
		&ServiceExposureRule{},
	}
}
