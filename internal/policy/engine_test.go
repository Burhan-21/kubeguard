package policy

import (
	"testing"

	"github.com/Burhan-21/kubeguard/internal/findings"
	"github.com/Burhan-21/kubeguard/internal/normalizer"
)

type mockRule struct {
	id        string
	severity  findings.Severity
	violation bool
}

func (m *mockRule) ID() string                         { return m.id }
func (m *mockRule) Title() string                      { return "Mock Rule " + m.id }
func (m *mockRule) Description() string                { return "Mock description" }
func (m *mockRule) Category() string                   { return "security" }
func (m *mockRule) DefaultSeverity() findings.Severity { return m.severity }
func (m *mockRule) Evaluate(res *normalizer.NormalizedResource) []findings.Finding {
	if m.violation {
		return []findings.Finding{
			{
				RuleID:   m.id,
				Severity: m.severity,
				Category: "security",
				Kind:     res.Kind,
				Name:     res.Name,
				Message:  "Violation in " + m.id,
			},
		}
	}
	return nil
}

func TestEngineNoRules(t *testing.T) {
	engine := NewEngine(nil)
	res := &normalizer.NormalizedResource{Kind: "Pod", Name: "test"}
	result := engine.Evaluate(res)
	if len(result) != 0 {
		t.Errorf("expected 0 findings, got %d", len(result))
	}
}

func TestEngineWithRules(t *testing.T) {
	r1 := &mockRule{id: "R1", severity: findings.SeverityBlock, violation: true}
	r2 := &mockRule{id: "R2", severity: findings.SeverityWarn, violation: false}

	engine := NewEngine([]Rule{r1, r2})
	res := &normalizer.NormalizedResource{Kind: "Deployment", Name: "my-app"}
	findingsList := engine.Evaluate(res)

	if len(findingsList) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findingsList))
	}
	if findingsList[0].RuleID != "R1" {
		t.Errorf("expected finding for R1, got %s", findingsList[0].RuleID)
	}
	if findingsList[0].Severity != findings.SeverityBlock {
		t.Errorf("expected severity BLOCK, got %s", findingsList[0].Severity)
	}
}

func TestSeverityOverrides(t *testing.T) {
	r1 := &mockRule{id: "R1", severity: findings.SeverityWarn, violation: true}
	engine := NewEngine([]Rule{r1})

	engine.SetOverrides(map[string]findings.Severity{
		"R1": findings.SeverityBlock,
	})

	res := &normalizer.NormalizedResource{Kind: "Deployment", Name: "my-app"}
	findingsList := engine.Evaluate(res)

	if len(findingsList) != 1 {
		t.Fatalf("expected 1 finding, got %d", len(findingsList))
	}
	if findingsList[0].Severity != findings.SeverityBlock {
		t.Errorf("expected overridden severity BLOCK, got %s", findingsList[0].Severity)
	}
}

func TestEngineDisabledRules(t *testing.T) {
	r1 := &mockRule{id: "R1", severity: findings.SeverityBlock, violation: true}
	r2 := &mockRule{id: "R2", severity: findings.SeverityWarn, violation: true}
	engine := NewEngine([]Rule{r1, r2})

	engine.SetDisabled([]string{"R1"})

	res := &normalizer.NormalizedResource{Kind: "Pod", Name: "pod1"}
	findingsList := engine.Evaluate(res)

	if len(findingsList) != 1 {
		t.Fatalf("expected 1 finding (R2 only), got %d", len(findingsList))
	}
	if findingsList[0].RuleID != "R2" {
		t.Errorf("expected finding for R2, got %s", findingsList[0].RuleID)
	}
}

func TestEvaluateAll(t *testing.T) {
	r1 := &mockRule{id: "R1", severity: findings.SeverityBlock, violation: true}
	r2 := &mockRule{id: "R2", severity: findings.SeverityWarn, violation: true}
	engine := NewEngine([]Rule{r1, r2})

	res1 := &normalizer.NormalizedResource{Kind: "Deployment", Name: "app1"}
	res2 := &normalizer.NormalizedResource{Kind: "Pod", Name: "app2"}

	scanResult := engine.EvaluateAll([]*normalizer.NormalizedResource{res1, res2})

	if len(scanResult.Findings) != 4 {
		t.Errorf("expected 4 total findings across 2 resources, got %d", len(scanResult.Findings))
	}
	if scanResult.BlockCount != 2 {
		t.Errorf("expected 2 block findings, got %d", scanResult.BlockCount)
	}
	if scanResult.WarnCount != 2 {
		t.Errorf("expected 2 warn findings, got %d", scanResult.WarnCount)
	}
	if scanResult.OverallResult != findings.SeverityBlock {
		t.Errorf("expected overall BLOCK, got %s", scanResult.OverallResult)
	}
	if scanResult.ExitCode() != 2 {
		t.Errorf("expected exit code 2, got %d", scanResult.ExitCode())
	}
}
