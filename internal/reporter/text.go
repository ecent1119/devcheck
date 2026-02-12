package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/fatih/color"
	"github.com/stackgen-cli/devcheck/internal/models"
)

// TextReporter outputs findings as colored terminal text
type TextReporter struct {
	writer io.Writer
	noColor bool
}

// NewTextReporter creates a new TextReporter
func NewTextReporter(w io.Writer, noColor bool) *TextReporter {
	if noColor {
		color.NoColor = true
	}
	return &TextReporter{writer: w, noColor: noColor}
}

// Report outputs the report as colored text
func (r *TextReporter) Report(report *models.Report) error {
	// Header
	fmt.Fprintf(r.writer, "devcheck scan: %s\n", report.Path)
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))
	fmt.Fprintln(r.writer)

	// Summary by severity
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

	// Print summary line
	redBold := color.New(color.FgRed, color.Bold)
	yellowBold := color.New(color.FgYellow, color.Bold)
	cyanBold := color.New(color.FgCyan)
	greenBold := color.New(color.FgGreen, color.Bold)

	if blocking > 0 {
		redBold.Fprintf(r.writer, "BLOCKING: %d  ", blocking)
	}
	if warnings > 0 {
		yellowBold.Fprintf(r.writer, "WARNINGS: %d  ", warnings)
	}
	if info > 0 {
		cyanBold.Fprintf(r.writer, "INFO: %d", info)
	}
	fmt.Fprintln(r.writer)
	fmt.Fprintln(r.writer)

	// Print blocking issues first
	if blocking > 0 {
		redBold.Fprintln(r.writer, "BLOCKING ISSUES")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		for _, f := range report.Findings {
			if f.Severity == models.SeverityBlocking {
				r.printFinding(f, redBold)
			}
		}
		fmt.Fprintln(r.writer)
	}

	// Print warnings
	if warnings > 0 {
		yellowBold.Fprintln(r.writer, "WARNINGS")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		for _, f := range report.Findings {
			if f.Severity == models.SeverityWarning {
				r.printFinding(f, yellowBold)
			}
		}
		fmt.Fprintln(r.writer)
	}

	// Print info
	if info > 0 {
		cyanBold.Fprintln(r.writer, "INFO")
		fmt.Fprintln(r.writer, strings.Repeat("-", 40))
		for _, f := range report.Findings {
			if f.Severity == models.SeverityInfo {
				r.printFinding(f, cyanBold)
			}
		}
		fmt.Fprintln(r.writer)
	}

	// Final verdict
	fmt.Fprintln(r.writer, strings.Repeat("=", 60))
	if blocking > 0 {
		redBold.Fprintln(r.writer, "✗ Project has blocking issues that must be resolved")
	} else if warnings > 0 {
		yellowBold.Fprintln(r.writer, "⚠ Project has warnings to review")
	} else {
		greenBold.Fprintln(r.writer, "✓ Project looks ready to run")
	}

	return nil
}

func (r *TextReporter) printFinding(f *models.Finding, c *color.Color) {
	c.Fprintf(r.writer, "[%s] ", f.Code)
	fmt.Fprintln(r.writer, f.Title)

	for _, loc := range f.Files {
		if loc.Line > 0 {
			fmt.Fprintf(r.writer, "    at %s:%d\n", loc.File, loc.Line)
		} else {
			fmt.Fprintf(r.writer, "    in %s\n", loc.File)
		}
	}

	if f.Details != "" {
		fmt.Fprintf(r.writer, "    %s\n", f.Details)
	}

	if f.SuggestedFix != "" {
		color.New(color.FgGreen).Fprintf(r.writer, "    → Fix: %s\n", f.SuggestedFix)
	}
	fmt.Fprintln(r.writer)
}
