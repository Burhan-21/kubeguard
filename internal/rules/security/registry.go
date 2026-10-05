package security

import (
	"github.com/Burhan-21/kubeguard/internal/policy"
)

func AllSecurityRules() []policy.Rule {
	return []policy.Rule{
		&PrivilegedRule{},
		&RunAsRootRule{},
		&PrivilegeEscalationRule{},
		&HostNetworkRule{},
		&HostPIDIPCRule{},
		&HostPathRule{},
		&CapabilitiesRule{},
		&MissingSecurityContextRule{},
		&LatestTagRule{},
		&MissingTagRule{},
		&UnpinnedDigestRule{},
		&EmbeddedSecretRule{},
		&ServiceAccountTokenRule{},
		&ExcessiveRBACRule{},
	}
}
