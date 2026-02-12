package models

// Severity represents the impact level of a finding
type Severity string

const (
	SeverityBlocking Severity = "blocking"
	SeverityWarning  Severity = "warning"
	SeverityInfo     Severity = "info"
)

// SourceLocation represents a location in a file
type SourceLocation struct {
	File   string `json:"file"`
	Line   int    `json:"line,omitempty"`
	Column int    `json:"column,omitempty"`
}

// Finding represents a single finding from the scan
type Finding struct {
	Code         string           `json:"code"`
	Severity     Severity         `json:"severity"`
	Title        string           `json:"title"`
	Details      string           `json:"details,omitempty"`
	Files        []SourceLocation `json:"files,omitempty"`
	SuggestedFix string           `json:"suggested_fix,omitempty"`
}

// NewFinding creates a new finding
func NewFinding(code string, severity Severity, title string) *Finding {
	return &Finding{
		Code:     code,
		Severity: severity,
		Title:    title,
	}
}

// WithDetails adds details to the finding
func (f *Finding) WithDetails(details string) *Finding {
	f.Details = details
	return f
}

// WithFile adds a file location to the finding
func (f *Finding) WithFile(file string, line int) *Finding {
	f.Files = append(f.Files, SourceLocation{File: file, Line: line})
	return f
}

// WithFix adds a suggested fix to the finding
func (f *Finding) WithFix(fix string) *Finding {
	f.SuggestedFix = fix
	return f
}

// SeverityLevel returns a numeric level for severity comparison
func SeverityLevel(s Severity) int {
	switch s {
	case SeverityBlocking:
		return 3
	case SeverityWarning:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}
