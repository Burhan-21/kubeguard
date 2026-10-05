package findings

type Severity string

const (
	SeverityPass  Severity = "PASS"
	SeverityWarn  Severity = "WARN"
	SeverityBlock Severity = "BLOCK"
)

type Finding struct {
	RuleID      string   `json:"ruleId" yaml:"ruleId"`
	Severity    Severity `json:"severity" yaml:"severity"`
	Category    string   `json:"category" yaml:"category"`
	Kind        string   `json:"kind" yaml:"kind"`
	Name        string   `json:"name" yaml:"name"`
	Namespace   string   `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Container   string   `json:"container,omitempty" yaml:"container,omitempty"`
	FieldPath   string   `json:"fieldPath,omitempty" yaml:"fieldPath,omitempty"`
	Message     string   `json:"message" yaml:"message"`
	Why         string   `json:"why,omitempty" yaml:"why,omitempty"`
	Remediation string   `json:"remediation,omitempty" yaml:"remediation,omitempty"`
	References  []string `json:"references,omitempty" yaml:"references,omitempty"`
}

type ScanResult struct {
	Findings      []Finding `json:"findings"`
	PassCount     int       `json:"passCount"`
	WarnCount     int       `json:"warnCount"`
	BlockCount    int       `json:"blockCount"`
	OverallResult Severity  `json:"overallResult"`
}

func (sr *ScanResult) Summarize() {
	sr.PassCount = 0
	sr.WarnCount = 0
	sr.BlockCount = 0
	sr.OverallResult = SeverityPass

	for _, f := range sr.Findings {
		switch f.Severity {
		case SeverityPass:
			sr.PassCount++
		case SeverityWarn:
			sr.WarnCount++
			if sr.OverallResult == SeverityPass {
				sr.OverallResult = SeverityWarn
			}
		case SeverityBlock:
			sr.BlockCount++
			sr.OverallResult = SeverityBlock
		}
	}
}

func (sr *ScanResult) ExitCode() int {
	switch sr.OverallResult {
	case SeverityWarn:
		return 1
	case SeverityBlock:
		return 2
	default:
		return 0
	}
}
