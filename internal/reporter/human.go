package reporter

import (
	"fmt"
	"io"
	"os"

	"github.com/Burhan-21/kubeguard/internal/findings"
)

func ReportHuman(result *findings.ScanResult, w io.Writer) error {
	useColor := false
	if os.Getenv("NO_COLOR") == "" {
		// Assuming it isatty for simplicity in this implementation
		useColor = true
	}

	colorRed := "\033[31m"
	colorYellow := "\033[33m"
	colorGreen := "\033[32m"
	colorReset := "\033[0m"

	fmt.Fprintln(w, "KubeGuard Deployment Readiness")
	fmt.Fprintln(w, "")

	grouped := make(map[string][]findings.Finding)
	for _, f := range result.Findings {
		key := fmt.Sprintf("%s/%s", f.Kind, f.Name)
		grouped[key] = append(grouped[key], f)
	}

	blockCount := 0
	warnCount := 0
	passCount := 0

	for key, fList := range grouped {
		fmt.Fprintf(w, "%s\n\n", key)
		for _, f := range fList {
			sevLabel := string(f.Severity)
			if useColor {
				switch f.Severity {
				case findings.SeverityBlock:
					sevLabel = fmt.Sprintf("%s%-5s%s", colorRed, f.Severity, colorReset)
					blockCount++
				case findings.SeverityWarn:
					sevLabel = fmt.Sprintf("%s%-5s%s", colorYellow, f.Severity, colorReset)
					warnCount++
				case findings.SeverityPass:
					sevLabel = fmt.Sprintf("%s%-5s%s", colorGreen, f.Severity, colorReset)
					passCount++
				}
			} else {
				sevLabel = fmt.Sprintf("%-5s", f.Severity)
				switch f.Severity {
				case findings.SeverityBlock: blockCount++
				case findings.SeverityWarn: warnCount++
				case findings.SeverityPass: passCount++
				}
			}
			fmt.Fprintf(w, "  %s  %s  %s\n", sevLabel, f.RuleID, f.Message)
		}
		fmt.Fprintln(w, "")
	}

	overallResult := "PASS"
	resultColor := colorGreen
	if blockCount > 0 {
		overallResult = "BLOCK"
		resultColor = colorRed
	} else if warnCount > 0 {
		overallResult = "WARN"
		resultColor = colorYellow
	}

	resLabel := overallResult
	if useColor {
		resLabel = fmt.Sprintf("%s%s%s", resultColor, overallResult, colorReset)
	}

	fmt.Fprintf(w, "Result: %s (%d findings: %d BLOCK, %d WARN)\n", resLabel, blockCount+warnCount, blockCount, warnCount)
	return nil
}
