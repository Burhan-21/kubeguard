package reporter

import (
	"encoding/json"
	"io"

	"github.com/Burhan-21/kubeguard/internal/findings"
)

type SarifLog struct {
	Schema  string     `json:"$schema"`
	Version string     `json:"version"`
	Runs    []SarifRun `json:"runs"`
}

type SarifRun struct {
	Tool    SarifTool     `json:"tool"`
	Results []SarifResult `json:"results"`
}

type SarifTool struct {
	Driver SarifDriver `json:"driver"`
}

type SarifDriver struct {
	Name    string      `json:"name"`
	Version string      `json:"version"`
	Rules   []SarifRule `json:"rules,omitempty"`
}

type SarifRule struct {
	ID               string       `json:"id"`
	ShortDescription SarifMessage `json:"shortDescription"`
}

type SarifResult struct {
	RuleID  string       `json:"ruleId"`
	Level   string       `json:"level"`
	Message SarifMessage `json:"message"`
}

type SarifMessage struct {
	Text string `json:"text"`
}

func ReportSARIF(result *findings.ScanResult, w io.Writer) error {
	run := SarifRun{
		Tool: SarifTool{
			Driver: SarifDriver{
				Name:    "KubeGuard",
				Version: "dev", // Ideally passed in
			},
		},
		Results: []SarifResult{},
	}

	for _, f := range result.Findings {
		level := "note"
		switch f.Severity {
		case findings.SeverityBlock:
			level = "error"
		case findings.SeverityWarn:
			level = "warning"
		}

		res := SarifResult{
			RuleID: f.RuleID,
			Level:  level,
			Message: SarifMessage{
				Text: f.Message,
			},
		}
		run.Results = append(run.Results, res)
	}

	log := SarifLog{
		Schema:  "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/sarif-2.1/schema/sarif-schema-2.1.0.json",
		Version: "2.1.0",
		Runs:    []SarifRun{run},
	}

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(log)
}
