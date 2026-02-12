package reporter

import (
	"fmt"
	"io"
	"strings"

	"github.com/stackgen-cli/devcheck/internal/models"
)

// ChecklistReporter generates a markdown fix checklist
type ChecklistReporter struct {
	writer io.Writer
}

// NewChecklistReporter creates a new ChecklistReporter
func NewChecklistReporter(w io.Writer) *ChecklistReporter {
	return &ChecklistReporter{writer: w}
}

// Report generates a fix checklist from findings
func (r *ChecklistReporter) Report(report *models.Report) error {
	fmt.Fprintln(r.writer, "# Fix Checklist")
	fmt.Fprintln(r.writer)
	fmt.Fprintf(r.writer, "**Project:** %s\n", report.Path)
	fmt.Fprintln(r.writer)

	// Separate by severity
	var blocking, warnings, info []*models.Finding
	for _, f := range report.Findings {
		switch f.Severity {
		case models.SeverityBlocking:
			blocking = append(blocking, f)
		case models.SeverityWarning:
			warnings = append(warnings, f)
		case models.SeverityInfo:
			info = append(info, f)
		}
	}

	// Blocking issues
	if len(blocking) > 0 {
		fmt.Fprintln(r.writer, "## 🚫 Must Fix (Blocking)")
		fmt.Fprintln(r.writer)
		for _, f := range blocking {
			r.writeChecklistItem(f)
		}
		fmt.Fprintln(r.writer)
	}

	// Warnings
	if len(warnings) > 0 {
		fmt.Fprintln(r.writer, "## ⚠️ Should Fix (Warnings)")
		fmt.Fprintln(r.writer)
		for _, f := range warnings {
			r.writeChecklistItem(f)
		}
		fmt.Fprintln(r.writer)
	}

	// Info
	if len(info) > 0 {
		fmt.Fprintln(r.writer, "## ℹ️ Informational")
		fmt.Fprintln(r.writer)
		for _, f := range info {
			r.writeInfoItem(f)
		}
		fmt.Fprintln(r.writer)
	}

	// Summary
	fmt.Fprintln(r.writer, "---")
	fmt.Fprintf(r.writer, "**Total:** %d blocking, %d warnings, %d info\n",
		len(blocking), len(warnings), len(info))

	return nil
}

func (r *ChecklistReporter) writeChecklistItem(f *models.Finding) {
	fmt.Fprintf(r.writer, "- [ ] **[%s]** %s\n", f.Code, f.Title)

	if len(f.Files) > 0 {
		for _, loc := range f.Files {
			if loc.Line > 0 {
				fmt.Fprintf(r.writer, "  - File: `%s:%d`\n", loc.File, loc.Line)
			} else {
				fmt.Fprintf(r.writer, "  - File: `%s`\n", loc.File)
			}
		}
	}

	if f.SuggestedFix != "" {
		fmt.Fprintf(r.writer, "  - **Fix:** %s\n", f.SuggestedFix)
	}

	fmt.Fprintln(r.writer)
}

func (r *ChecklistReporter) writeInfoItem(f *models.Finding) {
	fmt.Fprintf(r.writer, "- [%s] %s\n", f.Code, f.Title)

	if f.Details != "" {
		fmt.Fprintf(r.writer, "  - %s\n", f.Details)
	}
}

// GenerateShellScript generates a shell script with fix commands
func GenerateShellScript(report *models.Report) string {
	var sb strings.Builder

	sb.WriteString("#!/bin/bash\n")
	sb.WriteString("# Auto-generated fix script from devcheck\n")
	sb.WriteString("# Review carefully before running!\n\n")

	sb.WriteString("set -e\n\n")

	for _, f := range report.Findings {
		if f.SuggestedFix == "" {
			continue
		}

		sb.WriteString(fmt.Sprintf("# [%s] %s\n", f.Code, f.Title))

		// Try to generate actual commands from common fix patterns
		fix := f.SuggestedFix
		switch {
		case strings.HasPrefix(fix, "Add ") && strings.Contains(fix, "to .env"):
			// Parse "Add VAR=<value> to .env file"
			parts := strings.SplitN(fix, " ", 2)
			if len(parts) >= 2 {
				varPart := strings.TrimSuffix(strings.TrimPrefix(parts[1], " "), " to .env file")
				varPart = strings.TrimSuffix(varPart, " file")
				varPart = strings.Split(varPart, " to ")[0]
				sb.WriteString(fmt.Sprintf("echo '%s' >> .env\n", varPart))
			} else {
				sb.WriteString(fmt.Sprintf("# TODO: %s\n", fix))
			}
		case strings.HasPrefix(fix, "Copy "):
			// Parse "Copy X to Y"
			fix = strings.TrimPrefix(fix, "Copy ")
			parts := strings.Split(fix, " to ")
			if len(parts) == 2 {
				src := strings.TrimSpace(parts[0])
				dst := strings.Split(parts[1], " and ")[0]
				dst = strings.TrimSpace(dst)
				sb.WriteString(fmt.Sprintf("cp %s %s\n", src, dst))
			} else {
				sb.WriteString(fmt.Sprintf("# TODO: %s\n", f.SuggestedFix))
			}
		case strings.HasPrefix(fix, "Create "):
			// Parse "Create directory X"
			if strings.Contains(fix, "directory") {
				dir := strings.TrimPrefix(fix, "Create directory ")
				dir = strings.Split(dir, " or ")[0]
				sb.WriteString(fmt.Sprintf("mkdir -p %s\n", dir))
			} else {
				file := strings.TrimPrefix(fix, "Create ")
				file = strings.Split(file, " or ")[0]
				file = strings.Split(file, " in ")[0]
				sb.WriteString(fmt.Sprintf("touch %s\n", file))
			}
		case strings.HasPrefix(fix, "Install "):
			sb.WriteString(fmt.Sprintf("# TODO: %s\n", fix))
		default:
			sb.WriteString(fmt.Sprintf("# TODO: %s\n", fix))
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
