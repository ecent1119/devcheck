package checker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stackgen-cli/devcheck/internal/config"
	"github.com/stackgen-cli/devcheck/internal/models"
	"github.com/stackgen-cli/devcheck/internal/tools"
	"gopkg.in/yaml.v3"
)

// Options configures the checker behavior
type Options struct {
	EnableSourceScanning bool
	Config               *config.Config
	CheckToolVersions    bool
}

// Check runs all checks against the detected artifacts
func Check(basePath string, artifacts *models.Artifacts) []*models.Finding {
	return CheckWithOptions(basePath, artifacts, Options{})
}

// CheckWithOptions runs all checks with configurable options
func CheckWithOptions(basePath string, artifacts *models.Artifacts, opts Options) []*models.Finding {
	var findings []*models.Finding

	// Check env vars in compose files
	findings = append(findings, checkComposeEnvRefs(basePath, artifacts)...)

	// Check env example vs env
	findings = append(findings, checkEnvExample(basePath, artifacts)...)

	// Check compose depends_on
	findings = append(findings, checkComposeDependsOn(basePath, artifacts)...)

	// Check build contexts (Dockerfile existence)
	findings = append(findings, checkBuildContexts(basePath, artifacts)...)

	// Add info findings
	findings = append(findings, addLanguageInfo(artifacts)...)

	// Add run hints from README
	findings = append(findings, checkReadmeHints(basePath, artifacts)...)

	// Source code env scanning (if enabled)
	if opts.EnableSourceScanning {
		findings = append(findings, checkSourceCodeEnvRefs(basePath, artifacts)...)
	}

	// Tool version checks (if enabled)
	if opts.CheckToolVersions && opts.Config != nil && opts.Config.ToolVersions != nil {
		findings = append(findings, checkToolVersions(opts.Config.ToolVersions)...)
	}

	// Custom rules from config
	if opts.Config != nil {
		findings = append(findings, checkCustomRules(basePath, artifacts, opts.Config)...)
		findings = append(findings, checkRequiredEnvVars(basePath, artifacts, opts.Config)...)
	}

	// Filter out ignored codes if config provided
	if opts.Config != nil {
		findings = filterIgnoredFindings(findings, opts.Config)
	}

	return findings
}

// checkComposeEnvRefs checks for ${VAR} references in compose files
func checkComposeEnvRefs(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	// Collect defined env vars from all env files
	definedVars := make(map[string]bool)
	for _, envFile := range artifacts.EnvFiles {
		if envFile.Found {
			vars := parseEnvFile(filepath.Join(basePath, envFile.Path))
			for k := range vars {
				definedVars[k] = true
			}
		}
	}

	// Parse compose files for ${VAR} references
	varRefRegex := regexp.MustCompile(`\$\{([^}:]+)(?::-[^}]*)?\}`)

	for _, composeFile := range artifacts.ComposeFiles {
		if !composeFile.Found {
			continue
		}

		content, err := os.ReadFile(filepath.Join(basePath, composeFile.Path))
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(strings.NewReader(string(content)))
		lineNum := 0
		for scanner.Scan() {
			lineNum++
			line := scanner.Text()
			matches := varRefRegex.FindAllStringSubmatch(line, -1)
			for _, match := range matches {
				if len(match) > 1 {
					varName := match[1]
					if !definedVars[varName] && !isStandardVar(varName) {
						finding := models.NewFinding(
							"ENV001",
							models.SeverityBlocking,
							fmt.Sprintf("${%s} referenced but not defined", varName),
						).WithDetails(fmt.Sprintf("Variable ${%s} is used in %s but is not defined in any .env file", varName, composeFile.Path)).
							WithFile(composeFile.Path, lineNum).
							WithFix(fmt.Sprintf("Add %s=<value> to .env file", varName))

						findings = append(findings, finding)
					}
				}
			}
		}
	}

	return findings
}

// checkEnvExample compares .env.example with .env
func checkEnvExample(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	// Check if .env.example exists but .env doesn't
	hasExample := artifacts.HasEnvExample()
	hasEnv := artifacts.HasEnv()

	if hasExample && !hasEnv {
		var examplePath string
		for _, e := range artifacts.EnvExamples {
			if e.Found {
				examplePath = e.Path
				break
			}
		}
		findings = append(findings, models.NewFinding(
			"ENV003",
			models.SeverityWarning,
			".env.example exists but .env is missing",
		).WithDetails(fmt.Sprintf("%s exists but no .env file found", examplePath)).
			WithFile(examplePath, 0).
			WithFix("Copy .env.example to .env and fill in values"))
	}

	// Compare keys in .env.example vs .env
	if hasExample && hasEnv {
		var exampleVars, envVars map[string]string
		var examplePath, envPath string

		for _, e := range artifacts.EnvExamples {
			if e.Found {
				examplePath = e.Path
				exampleVars = parseEnvFile(filepath.Join(basePath, e.Path))
				break
			}
		}

		for _, e := range artifacts.EnvFiles {
			if e.Found && (e.Path == ".env" || e.Path == ".env.local") {
				envPath = e.Path
				envVars = parseEnvFile(filepath.Join(basePath, e.Path))
				break
			}
		}

		if exampleVars != nil && envVars != nil {
			for key := range exampleVars {
				if _, ok := envVars[key]; !ok {
					findings = append(findings, models.NewFinding(
						"ENV002",
						models.SeverityWarning,
						fmt.Sprintf("%s has %s but %s does not", examplePath, key, envPath),
					).WithDetails(fmt.Sprintf("Variable %s is defined in %s but missing from %s", key, examplePath, envPath)).
						WithFix(fmt.Sprintf("Add %s=<value> to %s", key, envPath)))
				}
			}
		}
	}

	return findings
}

// checkComposeDependsOn validates depends_on references
func checkComposeDependsOn(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	for _, composeFile := range artifacts.ComposeFiles {
		if !composeFile.Found {
			continue
		}

		content, err := os.ReadFile(filepath.Join(basePath, composeFile.Path))
		if err != nil {
			continue
		}

		var compose struct {
			Services map[string]struct {
				DependsOn yaml.Node `yaml:"depends_on"`
			} `yaml:"services"`
		}

		if err := yaml.Unmarshal(content, &compose); err != nil {
			continue
		}

		// Collect all service names
		serviceNames := make(map[string]bool)
		for name := range compose.Services {
			serviceNames[name] = true
		}

		// Check depends_on references
		for svcName, svc := range compose.Services {
			deps := extractDependsOn(&svc.DependsOn)
			for _, dep := range deps {
				if !serviceNames[dep] {
					findings = append(findings, models.NewFinding(
						"CMP001",
						models.SeverityBlocking,
						fmt.Sprintf("Service %s depends on unknown service %s", svcName, dep),
					).WithDetails(fmt.Sprintf("depends_on references %s which is not defined in %s", dep, composeFile.Path)).
						WithFile(composeFile.Path, 0).
						WithFix(fmt.Sprintf("Add service %s to %s or remove from depends_on", dep, composeFile.Path)))
				}
			}
		}
	}

	return findings
}

// addLanguageInfo adds informational findings about detected languages
func addLanguageInfo(artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	if artifacts.DetectedLang != "" {
		details := fmt.Sprintf("Detected %s project", artifacts.DetectedLang)
		if artifacts.PackageManager != "" {
			details += fmt.Sprintf(" with %s", artifacts.PackageManager)
		}

		findings = append(findings, models.NewFinding(
			"LANG001",
			models.SeverityInfo,
			details,
		))
	}

	return findings
}

// checkReadmeHints scans README for run instructions
func checkReadmeHints(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	if artifacts.Readme == nil || !artifacts.Readme.Found {
		return findings
	}

	content, err := os.ReadFile(filepath.Join(basePath, artifacts.Readme.Path))
	if err != nil {
		return findings
	}

	text := strings.ToLower(string(content))

	// Look for common run commands
	patterns := []struct {
		pattern string
		hint    string
	}{
		{"docker compose up", "docker compose up"},
		{"docker-compose up", "docker-compose up"},
		{"pnpm install", "pnpm install"},
		{"pnpm dev", "pnpm dev"},
		{"npm install", "npm install"},
		{"npm run dev", "npm run dev"},
		{"yarn install", "yarn install"},
		{"yarn dev", "yarn dev"},
		{"go run", "go run"},
		{"make run", "make run"},
		{"make dev", "make dev"},
	}

	for _, p := range patterns {
		if strings.Contains(text, p.pattern) {
			findings = append(findings, models.NewFinding(
				"HINT001",
				models.SeverityInfo,
				fmt.Sprintf("Likely entrypoint: %s (from README)", p.hint),
			))
			break // Only report first match
		}
	}

	return findings
}

// parseEnvFile reads an env file and returns key-value pairs
func parseEnvFile(path string) map[string]string {
	result := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return result
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes
			value = strings.Trim(value, `"'`)
			result[key] = value
		}
	}

	return result
}

// extractDependsOn extracts dependency names from depends_on node
func extractDependsOn(node *yaml.Node) []string {
	var deps []string

	if node == nil || node.Kind == 0 {
		return deps
	}

	// List form
	if node.Kind == yaml.SequenceNode {
		for _, item := range node.Content {
			if item.Kind == yaml.ScalarNode {
				deps = append(deps, item.Value)
			}
		}
		return deps
	}

	// Map form
	if node.Kind == yaml.MappingNode {
		for i := 0; i < len(node.Content); i += 2 {
			deps = append(deps, node.Content[i].Value)
		}
	}

	return deps
}

// isStandardVar checks if a variable is a standard system variable
func isStandardVar(name string) bool {
	standard := map[string]bool{
		"HOME":       true,
		"USER":       true,
		"PATH":       true,
		"PWD":        true,
		"SHELL":      true,
		"TERM":       true,
		"HOSTNAME":   true,
		"UID":        true,
		"GID":        true,
	}
	return standard[name]
}

// checkSourceCodeEnvRefs scans source code for environment variable usage
func checkSourceCodeEnvRefs(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	// Collect defined env vars
	definedVars := make(map[string]bool)
	for _, envFile := range artifacts.EnvFiles {
		if envFile.Found {
			vars := parseEnvFile(filepath.Join(basePath, envFile.Path))
			for k := range vars {
				definedVars[k] = true
			}
		}
	}

	// Patterns to detect env var usage in source code
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`process\.env\.([A-Za-z_][A-Za-z0-9_]*)`),                         // Node.js
		regexp.MustCompile(`os\.Getenv\s*\(\s*"([A-Za-z_][A-Za-z0-9_]*)"\s*\)`),              // Go
		regexp.MustCompile(`os\.environ\s*\[\s*['"]([A-Za-z_][A-Za-z0-9_]*)['"]\s*\]`),       // Python dict
		regexp.MustCompile(`os\.getenv\s*\(\s*['"]([A-Za-z_][A-Za-z0-9_]*)['"]`),             // Python getenv
		regexp.MustCompile(`System\.getenv\s*\(\s*"([A-Za-z_][A-Za-z0-9_]*)"\s*\)`),          // Java
		regexp.MustCompile(`Environment\.GetEnvironmentVariable\s*\(\s*"([A-Za-z_][A-Za-z0-9_]*)"\s*\)`), // C#
		regexp.MustCompile(`env::var\s*\(\s*"([A-Za-z_][A-Za-z0-9_]*)"\s*\)`),                // Rust
	}

	// File extensions to scan
	extensions := map[string]bool{
		".go":    true,
		".js":    true,
		".ts":    true,
		".jsx":   true,
		".tsx":   true,
		".py":    true,
		".java":  true,
		".cs":    true,
		".rs":    true,
	}

	// Track found undefined vars to avoid duplicates
	foundUndefined := make(map[string]bool)

	// Walk source files
	filepath.Walk(basePath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			// Skip common non-source directories
			if info != nil && info.IsDir() {
				name := info.Name()
				if name == "node_modules" || name == "vendor" || name == ".git" || name == "__pycache__" || name == "target" || name == "bin" || name == "obj" {
					return filepath.SkipDir
				}
			}
			return nil
		}

		ext := filepath.Ext(path)
		if !extensions[ext] {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		relPath, _ := filepath.Rel(basePath, path)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			for _, pattern := range patterns {
				matches := pattern.FindAllStringSubmatch(line, -1)
				for _, match := range matches {
					if len(match) >= 2 {
						varName := match[1]
						if !definedVars[varName] && !isStandardVar(varName) && !foundUndefined[varName] {
							foundUndefined[varName] = true
							findings = append(findings, models.NewFinding(
								"SRC001",
								models.SeverityWarning,
								fmt.Sprintf("Environment variable '%s' used in source but not defined", varName),
							).WithDetails(fmt.Sprintf("Variable %s is accessed in source code but not found in any .env file", varName)).
								WithFile(relPath, lineNum+1).
								WithFix(fmt.Sprintf("Add %s=<value> to .env file", varName)))
						}
					}
				}
			}
		}

		return nil
	})

	return findings
}

// checkBuildContexts validates that Dockerfiles exist in build contexts
func checkBuildContexts(basePath string, artifacts *models.Artifacts) []*models.Finding {
	var findings []*models.Finding

	for _, composeFile := range artifacts.ComposeFiles {
		if !composeFile.Found {
			continue
		}

		content, err := os.ReadFile(filepath.Join(basePath, composeFile.Path))
		if err != nil {
			continue
		}

		var compose struct {
			Services map[string]struct {
				Build interface{} `yaml:"build"`
			} `yaml:"services"`
		}

		if err := yaml.Unmarshal(content, &compose); err != nil {
			continue
		}

		for svcName, svc := range compose.Services {
			if svc.Build == nil {
				continue
			}

			var context string
			var dockerfile string = "Dockerfile"

			switch build := svc.Build.(type) {
			case string:
				context = build
			case map[string]interface{}:
				if c, ok := build["context"].(string); ok {
					context = c
				}
				if df, ok := build["dockerfile"].(string); ok {
					dockerfile = df
				}
			}

			if context == "" {
				continue
			}

			// Check if Dockerfile exists in context
			dockerfilePath := filepath.Join(basePath, context, dockerfile)
			if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
				findings = append(findings, models.NewFinding(
					"BUILD001",
					models.SeverityBlocking,
					fmt.Sprintf("Dockerfile not found for service %s", svcName),
				).WithDetails(fmt.Sprintf("Service %s expects %s at %s but it doesn't exist", svcName, dockerfile, filepath.Join(context, dockerfile))).
					WithFile(composeFile.Path, 0).
					WithFix(fmt.Sprintf("Create %s in %s or update build.context", dockerfile, context)))
			}

			// Check if context directory exists
			contextPath := filepath.Join(basePath, context)
			if _, err := os.Stat(contextPath); os.IsNotExist(err) {
				findings = append(findings, models.NewFinding(
					"BUILD002",
					models.SeverityBlocking,
					fmt.Sprintf("Build context directory not found for service %s", svcName),
				).WithDetails(fmt.Sprintf("Service %s references build context %s which doesn't exist", svcName, context)).
					WithFile(composeFile.Path, 0).
					WithFix(fmt.Sprintf("Create directory %s or update build.context", context)))
			}
		}
	}

	return findings
}

// checkToolVersions checks if required tools are installed with correct versions
func checkToolVersions(versions *config.ToolVersions) []*models.Finding {
	var findings []*models.Finding

	requirements := make(map[string]string)
	if versions.Docker != "" {
		requirements["docker"] = versions.Docker
	}
	if versions.DockerCompose != "" {
		requirements["docker-compose"] = versions.DockerCompose
	}
	if versions.Go != "" {
		requirements["go"] = versions.Go
	}
	if versions.Node != "" {
		requirements["node"] = versions.Node
	}
	if versions.Python != "" {
		requirements["python"] = versions.Python
	}

	checks := tools.CheckVersions(requirements)

	for _, check := range checks {
		if !check.Available {
			findings = append(findings, models.NewFinding(
				"TOOL001",
				models.SeverityBlocking,
				fmt.Sprintf("Required tool '%s' not found", check.Tool),
			).WithDetails(fmt.Sprintf("Tool %s is required but not installed or not in PATH", check.Tool)).
				WithFix(fmt.Sprintf("Install %s version %s or higher", check.Tool, check.Required)))
		} else if !check.Satisfied {
			findings = append(findings, models.NewFinding(
				"TOOL002",
				models.SeverityWarning,
				fmt.Sprintf("Tool '%s' version too old: %s < %s", check.Tool, check.Current, check.Required),
			).WithDetails(fmt.Sprintf("Tool %s version %s is installed but minimum %s is required", check.Tool, check.Current, check.Required)).
				WithFix(fmt.Sprintf("Upgrade %s to version %s or higher", check.Tool, check.Required)))
		}
	}

	return findings
}

// checkCustomRules applies custom rules from config
func checkCustomRules(basePath string, artifacts *models.Artifacts, cfg *config.Config) []*models.Finding {
	var findings []*models.Finding

	if len(cfg.CustomRules) == 0 {
		return findings
	}

	// Collect all defined vars
	definedVars := make(map[string]bool)
	for _, envFile := range artifacts.EnvFiles {
		if envFile.Found {
			vars := parseEnvFile(filepath.Join(basePath, envFile.Path))
			for k := range vars {
				definedVars[k] = true
			}
		}
	}

	for _, rule := range cfg.CustomRules {
		if !rule.Required {
			continue
		}

		pattern, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		// Check if any matching variable is defined
		found := false
		for name := range definedVars {
			if pattern.MatchString(name) {
				found = true
				break
			}
		}

		if !found {
			severity := models.SeverityWarning
			if rule.Severity == "blocking" {
				severity = models.SeverityBlocking
			} else if rule.Severity == "info" {
				severity = models.SeverityInfo
			}

			findings = append(findings, models.NewFinding(
				"CUSTOM-"+rule.ID,
				severity,
				fmt.Sprintf("Custom rule '%s' not satisfied", rule.ID),
			).WithDetails(rule.Description).
				WithFix(fmt.Sprintf("Define a variable matching pattern: %s", rule.Pattern)))
		}
	}

	return findings
}

// checkRequiredEnvVars checks that required env vars from config are defined
func checkRequiredEnvVars(basePath string, artifacts *models.Artifacts, cfg *config.Config) []*models.Finding {
	var findings []*models.Finding

	if len(cfg.RequiredEnvVars) == 0 {
		return findings
	}

	// Collect all defined vars
	definedVars := make(map[string]bool)
	for _, envFile := range artifacts.EnvFiles {
		if envFile.Found {
			vars := parseEnvFile(filepath.Join(basePath, envFile.Path))
			for k := range vars {
				definedVars[k] = true
			}
		}
	}

	for _, required := range cfg.RequiredEnvVars {
		if !definedVars[required] {
			findings = append(findings, models.NewFinding(
				"REQ001",
				models.SeverityBlocking,
				fmt.Sprintf("Required variable '%s' not defined", required),
			).WithDetails(fmt.Sprintf("Variable %s is configured as required in .devcheck.yaml but is not defined", required)).
				WithFix(fmt.Sprintf("Add %s=<value> to .env file", required)))
		}
	}

	return findings
}

// filterIgnoredFindings removes findings with codes in the ignore list
func filterIgnoredFindings(findings []*models.Finding, cfg *config.Config) []*models.Finding {
	if len(cfg.IgnoreCodes) == 0 {
		return findings
	}

	var filtered []*models.Finding
	for _, f := range findings {
		if !cfg.ShouldIgnoreCode(f.Code) {
			filtered = append(filtered, f)
		}
	}
	return filtered
}
