package reporter

import (
	"io"

	"github.com/Burhan-21/kubeguard/internal/findings"
)

func Report(format string, result *findings.ScanResult, w io.Writer) error {
	switch format {
	case "json":
		return ReportJSON(result, w)
	case "sarif":
		return ReportSARIF(result, w)
	default:
		return ReportHuman(result, w)
	}
}
