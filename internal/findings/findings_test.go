package findings

import (
	"testing"
)

func TestScanResultSummarizeAndExitCode(t *testing.T) {
	// Empty findings -> PASS, exit 0
	srEmpty := &ScanResult{}
	srEmpty.Summarize()
	if srEmpty.OverallResult != SeverityPass || srEmpty.ExitCode() != 0 {
		t.Errorf("empty findings expected PASS / 0, got %s / %d", srEmpty.OverallResult, srEmpty.ExitCode())
	}

	// Pass findings -> PASS, exit 0
	srPass := &ScanResult{
		Findings: []Finding{
			{RuleID: "TEST-001", Severity: SeverityPass},
		},
	}
	srPass.Summarize()
	if srPass.OverallResult != SeverityPass || srPass.ExitCode() != 0 || srPass.PassCount != 1 {
		t.Errorf("pass findings expected PASS / 0, got %s / %d (count %d)", srPass.OverallResult, srPass.ExitCode(), srPass.PassCount)
	}

	// Warn findings -> WARN, exit 1
	srWarn := &ScanResult{
		Findings: []Finding{
			{RuleID: "TEST-001", Severity: SeverityPass},
			{RuleID: "TEST-002", Severity: SeverityWarn},
		},
	}
	srWarn.Summarize()
	if srWarn.OverallResult != SeverityWarn || srWarn.ExitCode() != 1 || srWarn.WarnCount != 1 {
		t.Errorf("warn findings expected WARN / 1, got %s / %d", srWarn.OverallResult, srWarn.ExitCode())
	}

	// Block findings -> BLOCK, exit 2
	srBlock := &ScanResult{
		Findings: []Finding{
			{RuleID: "TEST-001", Severity: SeverityWarn},
			{RuleID: "TEST-002", Severity: SeverityBlock},
		},
	}
	srBlock.Summarize()
	if srBlock.OverallResult != SeverityBlock || srBlock.ExitCode() != 2 || srBlock.BlockCount != 1 {
		t.Errorf("block findings expected BLOCK / 2, got %s / %d", srBlock.OverallResult, srBlock.ExitCode())
	}
}
