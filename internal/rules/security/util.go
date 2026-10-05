package security

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
	"github.com/Burhan-21/kubeguard/internal/policy"
)

func newFinding(rule policy.Rule, res *normalizer.NormalizedResource, container, fieldPath, message, why, remediation string) findings.Finding {
	return findings.Finding{
		RuleID:      rule.ID(),
		Severity:    rule.DefaultSeverity(),
		Category:    rule.Category(),
		Kind:        res.Kind,
		Name:        res.Name,
		Namespace:   res.Namespace,
		Container:   container,
		FieldPath:   fieldPath,
		Message:     message,
		Why:         why,
		Remediation: remediation,
	}
}
