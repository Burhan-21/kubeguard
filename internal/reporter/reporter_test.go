package reporter

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/Burhan-21/kubeguard/internal/findings"
)

func sampleScanResult() *findings.ScanResult {
	sr := &findings.ScanResult{
		Findings: []findings.Finding{
			{
				RuleID:   "KG-SEC-001",
				Severity: findings.SeverityBlock,
				Category: "security",
				Kind:     "Deployment",
				Name:     "payment-service",
				Message:  "Container is configured as privileged.",
			},
			{
				RuleID:   "KG-REL-001",
				Severity: findings.SeverityWarn,
				Category: "reliability",
				Kind:     "Deployment",
				Name:     "payment-service",
				Message:  "Readiness probe is missing.",
			},
		},
	}
	sr.Summarize()
	return sr
}

func TestReportHuman(t *testing.T) {
	sr := sampleScanResult()
	var buf bytes.Buffer
	if err := ReportHuman(sr, &buf); err != nil {
		t.Fatalf("ReportHuman failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Deployment/payment-service") {
		t.Errorf("output missing resource header: %s", out)
	}
	if !strings.Contains(out, "KG-SEC-001") {
		t.Errorf("output missing rule ID KG-SEC-001: %s", out)
	}
	if !strings.Contains(out, "Result: BLOCK") {
		t.Errorf("output missing overall Result BLOCK: %s", out)
	}
}

func TestReportJSON(t *testing.T) {
	sr := sampleScanResult()
	var buf bytes.Buffer
	if err := ReportJSON(sr, &buf); err != nil {
		t.Fatalf("ReportJSON failed: %v", err)
	}

	var parsed findings.ScanResult
	if err := json.Unmarshal(buf.Bytes(), &parsed); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if parsed.BlockCount != 1 || parsed.WarnCount != 1 {
		t.Errorf("expected 1 block and 1 warn, got block=%d, warn=%d", parsed.BlockCount, parsed.WarnCount)
	}
	if parsed.OverallResult != findings.SeverityBlock {
		t.Errorf("expected overall BLOCK, got %s", parsed.OverallResult)
	}
}

func TestReportSARIF(t *testing.T) {
	sr := sampleScanResult()
	var buf bytes.Buffer
	if err := ReportSARIF(sr, &buf); err != nil {
		t.Fatalf("ReportSARIF failed: %v", err)
	}

	var sarif SarifLog
	if err := json.Unmarshal(buf.Bytes(), &sarif); err != nil {
		t.Fatalf("invalid SARIF JSON output: %v", err)
	}

	if sarif.Version != "2.1.0" {
		t.Errorf("expected SARIF version 2.1.0, got %s", sarif.Version)
	}
	if len(sarif.Runs) == 0 || len(sarif.Runs[0].Results) != 2 {
		t.Fatalf("expected 2 SARIF results, got %d", len(sarif.Runs[0].Results))
	}
	if sarif.Runs[0].Results[0].Level != "error" {
		t.Errorf("expected block finding to map to SARIF level error, got %s", sarif.Runs[0].Results[0].Level)
	}
	if sarif.Runs[0].Results[1].Level != "warning" {
		t.Errorf("expected warn finding to map to SARIF level warning, got %s", sarif.Runs[0].Results[1].Level)
	}
}

func TestReportDispatch(t *testing.T) {
	sr := sampleScanResult()
	var buf bytes.Buffer

	if err := Report("json", sr, &buf); err != nil {
		t.Fatalf("Report with json failed: %v", err)
	}
	buf.Reset()
	if err := Report("sarif", sr, &buf); err != nil {
		t.Fatalf("Report with sarif failed: %v", err)
	}
	buf.Reset()
	if err := Report("human", sr, &buf); err != nil {
		t.Fatalf("Report with human failed: %v", err)
	}
}
