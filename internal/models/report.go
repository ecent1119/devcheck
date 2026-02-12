package models

// ReportSummary provides aggregate counts
type ReportSummary struct {
	TotalFindings int `json:"total_findings"`
	BlockingCount int `json:"blocking_count"`
	WarningCount  int `json:"warning_count"`
	InfoCount     int `json:"info_count"`
}

// Report is the complete scan result
type Report struct {
	Path      string        `json:"path"`
	Artifacts *Artifacts    `json:"artifacts"`
	Findings  []*Finding    `json:"findings"`
	Summary   ReportSummary `json:"summary"`
}

// CalculateSummary computes summary counts from findings
func (r *Report) CalculateSummary() {
	r.Summary = ReportSummary{}
	for _, f := range r.Findings {
		r.Summary.TotalFindings++
		switch f.Severity {
		case SeverityBlocking:
			r.Summary.BlockingCount++
		case SeverityWarning:
			r.Summary.WarningCount++
		case SeverityInfo:
			r.Summary.InfoCount++
		}
	}
}

// HasBlocking checks if there are any blocking findings
func (r *Report) HasBlocking() bool {
	return r.Summary.BlockingCount > 0
}

// FilterBySeverity returns findings at or above the given severity
func (r *Report) FilterBySeverity(minSeverity Severity) []*Finding {
	minLevel := SeverityLevel(minSeverity)
	var result []*Finding
	for _, f := range r.Findings {
		if SeverityLevel(f.Severity) >= minLevel {
			result = append(result, f)
		}
	}
	return result
}
