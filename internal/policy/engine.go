package policy

import (
	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

// Engine evaluates normalized resources against a set of rules.
type Engine struct {
	rules     []Rule
	overrides map[string]findings.Severity // ruleID -> severity override
	disabled  map[string]bool              // ruleID -> is disabled
}

func NewEngine(rules []Rule) *Engine {
	return &Engine{
		rules:     rules,
		overrides: make(map[string]findings.Severity),
		disabled:  make(map[string]bool),
	}
}

func (e *Engine) SetOverrides(overrides map[string]findings.Severity) {
	e.overrides = overrides
}

func (e *Engine) SetDisabled(disabled []string) {
	e.disabled = make(map[string]bool)
	for _, id := range disabled {
		e.disabled[id] = true
	}
}

func (e *Engine) Evaluate(resource *normalizer.NormalizedResource) []findings.Finding {
	var allFindings []findings.Finding

	for _, rule := range e.rules {
		if e.disabled[rule.ID()] {
			continue
		}
		ruleFindings := rule.Evaluate(resource)
		
		for i := range ruleFindings {
			if override, exists := e.overrides[rule.ID()]; exists {
				ruleFindings[i].Severity = override
			}
		}
		
		allFindings = append(allFindings, ruleFindings...)
	}

	return allFindings
}

func (e *Engine) EvaluateAll(resources []*normalizer.NormalizedResource) *findings.ScanResult {
	result := &findings.ScanResult{}
	
	for _, res := range resources {
		resFindings := e.Evaluate(res)
		result.Findings = append(result.Findings, resFindings...)
	}
	
	result.Summarize()
	return result
}
