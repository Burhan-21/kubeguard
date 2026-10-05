package reporter

import (
	"encoding/json"
	"io"

	"github.com/Burhan-21/kubeguard/internal/findings"
)

func ReportJSON(result *findings.ScanResult, w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
