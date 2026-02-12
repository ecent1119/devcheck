// Package profiles provides preset configurations for devcheck
package profiles

import "github.com/stackgen-cli/devcheck/internal/models"

// Profile represents a configuration profile
type Profile struct {
	Name        string
	Description string
	// MinSeverity filters findings to this severity or higher
	MinSeverity models.Severity
	// EnabledChecks specifies which check codes to enable (empty = all)
	EnabledChecks []string
	// DisabledChecks specifies which check codes to disable
	DisabledChecks []string
	// EnableSourceScanning enables source code env var scanning
	EnableSourceScanning bool
	// IncludeInfo includes info-level findings in output
	IncludeInfo bool
}

// BuiltinProfiles contains all available preset profiles
var BuiltinProfiles = map[string]*Profile{
	"default": {
		Name:                 "default",
		Description:          "Standard development checks",
		MinSeverity:          models.SeverityInfo,
		EnableSourceScanning: false,
		IncludeInfo:          true,
	},
	"strict": {
		Name:                 "strict",
		Description:          "Strict mode - all checks enabled, fail on any issue",
		MinSeverity:          models.SeverityInfo,
		EnableSourceScanning: true,
		IncludeInfo:          true,
	},
	"ci": {
		Name:                 "ci",
		Description:          "CI mode - blocking and warnings only, no info",
		MinSeverity:          models.SeverityWarning,
		EnableSourceScanning: false,
		IncludeInfo:          false,
	},
	"minimal": {
		Name:                 "minimal",
		Description:          "Minimal mode - only blocking issues",
		MinSeverity:          models.SeverityBlocking,
		EnableSourceScanning: false,
		IncludeInfo:          false,
	},
	"full": {
		Name:                 "full",
		Description:          "Full analysis including source code scanning",
		MinSeverity:          models.SeverityInfo,
		EnableSourceScanning: true,
		IncludeInfo:          true,
	},
}

// Get returns a profile by name, or nil if not found
func Get(name string) *Profile {
	return BuiltinProfiles[name]
}

// List returns all available profile names
func List() []string {
	names := make([]string, 0, len(BuiltinProfiles))
	for name := range BuiltinProfiles {
		names = append(names, name)
	}
	return names
}

// FilterFindings filters findings based on profile settings
func (p *Profile) FilterFindings(findings []*models.Finding) []*models.Finding {
	var filtered []*models.Finding

	for _, f := range findings {
		// Check severity threshold
		if models.SeverityLevel(f.Severity) < models.SeverityLevel(p.MinSeverity) {
			continue
		}

		// Check if info findings should be included
		if f.Severity == models.SeverityInfo && !p.IncludeInfo {
			continue
		}

		// Check enabled/disabled checks
		if len(p.EnabledChecks) > 0 && !contains(p.EnabledChecks, f.Code) {
			continue
		}
		if contains(p.DisabledChecks, f.Code) {
			continue
		}

		filtered = append(filtered, f)
	}

	return filtered
}

func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
