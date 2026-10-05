package policy

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

// Rule is the interface all security and reliability rules implement.
type Rule interface {
	ID() string
	Title() string
	Description() string
	Category() string // "security" or "reliability"
	DefaultSeverity() findings.Severity
	Evaluate(resource *normalizer.NormalizedResource) []findings.Finding
}
