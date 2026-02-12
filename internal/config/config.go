// Package config handles .devcheck.yaml configuration files
package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents a .devcheck.yaml configuration
type Config struct {
	// CustomRules allows users to define custom variable validation rules
	CustomRules []CustomRule `yaml:"custom_rules,omitempty"`

	// ToolVersions specifies minimum required tool versions
	ToolVersions *ToolVersions `yaml:"tool_versions,omitempty"`

	// IgnorePatterns are file patterns to ignore during scanning
	IgnorePatterns []string `yaml:"ignore_patterns,omitempty"`

	// IgnoreCodes are finding codes to ignore (e.g., "ENV001")
	IgnoreCodes []string `yaml:"ignore_codes,omitempty"`

	// RequiredEnvVars is a list of env vars that must be defined
	RequiredEnvVars []string `yaml:"required_env_vars,omitempty"`

	// BuildContexts maps service names to expected Dockerfile paths
	BuildContexts map[string]string `yaml:"build_contexts,omitempty"`
}

// CustomRule defines a custom validation rule
type CustomRule struct {
	ID          string `yaml:"id"`
	Pattern     string `yaml:"pattern"`      // Variable name pattern (regex)
	Required    bool   `yaml:"required"`     // Whether matching vars must be defined
	Description string `yaml:"description"`  // Human-readable description
	Severity    string `yaml:"severity"`     // blocking, warning, info
}

// ToolVersions specifies minimum tool versions
type ToolVersions struct {
	Docker        string `yaml:"docker,omitempty"`
	DockerCompose string `yaml:"docker_compose,omitempty"`
	Go            string `yaml:"go,omitempty"`
	Node          string `yaml:"node,omitempty"`
	Python        string `yaml:"python,omitempty"`
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		CustomRules:    []CustomRule{},
		IgnorePatterns: []string{},
		IgnoreCodes:    []string{},
	}
}

// Load attempts to load a config from the given path
// Returns default config if file doesn't exist
func Load(basePath string) (*Config, error) {
	configPaths := []string{
		filepath.Join(basePath, ".devcheck.yaml"),
		filepath.Join(basePath, ".devcheck.yml"),
		filepath.Join(basePath, "devcheck.yaml"),
		filepath.Join(basePath, "devcheck.yml"),
	}

	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			return loadFromFile(path)
		}
	}

	// No config file found, return default
	return DefaultConfig(), nil
}

// loadFromFile loads configuration from a specific file
func loadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// LoadFromFile loads configuration from a specific file path
func LoadFromFile(path string) (*Config, error) {
	return loadFromFile(path)
}

// ShouldIgnoreCode checks if a finding code should be ignored
func (c *Config) ShouldIgnoreCode(code string) bool {
	for _, ignore := range c.IgnoreCodes {
		if ignore == code {
			return true
		}
	}
	return false
}

// ExampleConfig returns an example configuration string
func ExampleConfig() string {
	return `# .devcheck.yaml - devcheck configuration file
#
# Define custom rules for environment variable validation
custom_rules:
  - id: "DB_REQUIRED"
    pattern: "^DATABASE_"
    required: true
    description: "Database configuration variables must be defined"
    severity: blocking

# Minimum tool versions
tool_versions:
  docker: "20.10.0"
  docker_compose: "2.0.0"

# Files to ignore during scanning
ignore_patterns:
  - "*.backup"
  - "deprecated/"

# Finding codes to ignore
ignore_codes:
  - "HINT001"

# Environment variables that must always be defined
required_env_vars:
  - "NODE_ENV"
  - "DATABASE_URL"

# Map service names to expected Dockerfile paths
# devcheck will verify these exist
build_contexts:
  api: "./api"
  web: "./frontend"
`
}
