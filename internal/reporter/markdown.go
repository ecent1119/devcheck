package reporter

import (
	"fmt"
	"io"

	"github.com/stackgen-cli/devcheck/internal/models"
)

// MarkdownReporter outputs findings as Markdown
type MarkdownReporter struct {
	writer io.Writer
}

// NewMarkdownReporter creates a new MarkdownReporter
func NewMarkdownReporter(w io.Writer) *MarkdownReporter {
	return &MarkdownReporter{writer: w}
}

// Report outputs the report as Markdown
func (r *MarkdownReporter) Report(report *models.Report) error {
	// Header
	fmt.Fprintf(r.writer, "# devcheck Report\n\n")
	fmt.Fprintf(r.writer, "**Path:** `%s`\n\n", report.Path)

	// Summary
	blocking := 0
	warnings := 0
	info := 0
	for _, f := range report.Findings {
		switch f.Severity {
		case models.SeverityBlocking:
			blocking++
		case models.SeverityWarning:
			warnings++
		case models.SeverityInfo:
			info++
		}
	}

	fmt.Fprintf(r.writer, "## Summary\n\n")
	fmt.Fprintf(r.writer, "| Severity | Count |\n")
	fmt.Fprintf(r.writer, "|----------|-------|\n")
	fmt.Fprintf(r.writer, "| 🔴 Blocking | %d |\n", blocking)
	fmt.Fprintf(r.writer, "| 🟡 Warning | %d |\n", warnings)
	fmt.Fprintf(r.writer, "| 🔵 Info | %d |\n\n", info)

	// Blocking issues
	if blocking > 0 {
		fmt.Fprintf(r.writer, "## 🔴 Blocking Issues\n\n")
		for _, f := range report.Findings {
			if f.Severity == models.SeverityBlocking {
				r.printFinding(f)
			}
		}
	}

	// Warnings
	if warnings > 0 {
		fmt.Fprintf(r.writer, "## 🟡 Warnings\n\n")
		for _, f := range report.Findings {
			if f.Severity == models.SeverityWarning {
				r.printFinding(f)
			}
		}
	}

	// Info
	if info > 0 {
		fmt.Fprintf(r.writer, "## 🔵 Info\n\n")
		for _, f := range report.Findings {
			if f.Severity == models.SeverityInfo {
				r.printFinding(f)
			}
		}
	}

	// Verdict
	fmt.Fprintf(r.writer, "---\n\n")
	if blocking > 0 {
		fmt.Fprintf(r.writer, "**❌ Project has blocking issues that must be resolved**\n")
	} else if warnings > 0 {
		fmt.Fprintf(r.writer, "**⚠️ Project has warnings to review**\n")
	} else {
		fmt.Fprintf(r.writer, "**✅ Project looks ready to run**\n")
	}

	return nil
}

func (r *MarkdownReporter) printFinding(f *models.Finding) {
	fmt.Fprintf(r.writer, "### `%s` %s\n\n", f.Code, f.Title)

	for _, loc := range f.Files {
		if loc.Line > 0 {
			fmt.Fprintf(r.writer, "- **Location:** `%s:%d`\n", loc.File, loc.Line)
		} else {
			fmt.Fprintf(r.writer, "- **File:** `%s`\n", loc.File)
		}
	}

	if f.Details != "" {
		fmt.Fprintf(r.writer, "- **Details:** %s\n", f.Details)
	}

	if f.SuggestedFix != "" {
		fmt.Fprintf(r.writer, "- **Fix:** %s\n", f.SuggestedFix)
	}

	fmt.Fprintln(r.writer)
}
