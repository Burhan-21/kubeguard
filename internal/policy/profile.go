package policy

import (
	"fmt"
	"os"

	"github.com/Burhan-21/kubeguard/internal/findings"
	"gopkg.in/yaml.v3"
)

type ProfileSpec struct {
	Rules []RuleOverride `yaml:"rules"`
}

type RuleOverride struct {
	ID       string             `yaml:"id"`
	Enabled  *bool              `yaml:"enabled,omitempty"`
	Severity *findings.Severity `yaml:"severity,omitempty"`
}

type PolicyProfile struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   ProfileMeta `yaml:"metadata"`
	Spec       ProfileSpec `yaml:"spec"`
}

type ProfileMeta struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
}

func LoadProfile(path string) (*PolicyProfile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read profile %s: %w", path, err)
	}

	var profile PolicyProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse profile: %w", err)
	}

	return &profile, nil
}

func (p *PolicyProfile) BuildOverrides() (disabledRules []string, severityOverrides map[string]findings.Severity) {
	severityOverrides = make(map[string]findings.Severity)
	
	for _, r := range p.Spec.Rules {
		if r.Enabled != nil && !*r.Enabled {
			disabledRules = append(disabledRules, r.ID)
			continue
		}
		
		if r.Severity != nil {
			severityOverrides[r.ID] = *r.Severity
		}
	}
	
	return disabledRules, severityOverrides
}
