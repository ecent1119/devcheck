package reporter

import (
	"encoding/json"
	"io"

	"github.com/stackgen-cli/devcheck/internal/models"
)

// JSONReporter outputs findings as JSON
type JSONReporter struct {
	writer io.Writer
	pretty bool
}

// NewJSONReporter creates a new JSONReporter
func NewJSONReporter(w io.Writer, pretty bool) *JSONReporter {
	return &JSONReporter{writer: w, pretty: pretty}
}

// Report outputs the report as JSON
func (r *JSONReporter) Report(report *models.Report) error {
	var encoder *json.Encoder
	encoder = json.NewEncoder(r.writer)
	if r.pretty {
		encoder.SetIndent("", "  ")
	}
	return encoder.Encode(report)
}
