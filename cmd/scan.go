package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stackgen-cli/devcheck/internal/checker"
	"github.com/stackgen-cli/devcheck/internal/config"
	"github.com/stackgen-cli/devcheck/internal/detector"
	"github.com/stackgen-cli/devcheck/internal/models"
	"github.com/stackgen-cli/devcheck/internal/profiles"
	"github.com/stackgen-cli/devcheck/internal/reporter"
)

var (
	formatFlag        string
	composeFile       string
	envFiles          []string
	strictMode        bool
	noColor           bool
	profileName       string
	checkToolVersions bool
	configFile        string
	generateFixList   string
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan a project for local dev readiness",
	Long: `Scan a project directory for local development readiness issues.

Available profiles:
  default  Standard development checks
  strict   All checks enabled, fail on any issue
  ci       CI mode - blocking and warnings only
  minimal  Only blocking issues
  full     Full analysis including source code scanning

Configuration:
  Create a .devcheck.yaml file to customize rules, required variables,
  tool versions, and ignored checks. See --init-config for an example.

Examples:
  devcheck scan
  devcheck scan /path/to/project
  devcheck scan --format json
  devcheck scan --strict
  devcheck scan --profile ci
  devcheck scan --check-tools
  devcheck scan --fix-list fixes.md`,
	Args: cobra.MaximumNArgs(1),
	Run:  runScan,
}

func init() {
	scanCmd.Flags().StringVarP(&formatFlag, "format", "f", "text", "Output format: text, json, markdown, checklist")
	scanCmd.Flags().StringVar(&composeFile, "compose", "", "Specify compose file path")
	scanCmd.Flags().StringSliceVar(&envFiles, "env", nil, "Specify env file(s)")
	scanCmd.Flags().BoolVar(&strictMode, "strict", false, "Exit 1 if blocking findings exist")
	scanCmd.Flags().BoolVar(&noColor, "no-color", false, "Disable color output")
	scanCmd.Flags().StringVarP(&profileName, "profile", "p", "default", fmt.Sprintf("Check profile (%s)", strings.Join(profiles.List(), ", ")))
	scanCmd.Flags().BoolVar(&checkToolVersions, "check-tools", false, "Check tool versions (docker, docker-compose, etc.)")
	scanCmd.Flags().StringVar(&configFile, "config", "", "Custom config file path")
	scanCmd.Flags().StringVar(&generateFixList, "fix-list", "", "Generate fix checklist to file (markdown)")

	rootCmd.AddCommand(scanCmd)
}

func runScan(cmd *cobra.Command, args []string) {
	// Get profile
	profile := profiles.Get(profileName)
	if profile == nil {
		color.Red("Unknown profile: %s (available: %s)", profileName, strings.Join(profiles.List(), ", "))
		os.Exit(2)
	}

	// Determine scan path
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}

	// Resolve to absolute path
	absPath, err := filepath.Abs(scanPath)
	if err != nil {
		color.Red("Error resolving path: %v", err)
		os.Exit(2)
	}

	// Check path exists
	if _, err := os.Stat(absPath); err != nil {
		color.Red("Path not found: %s", absPath)
		os.Exit(2)
	}

	// Load config
	var cfg *config.Config
	if configFile != "" {
		// Load from specified file
		cfg, err = config.LoadFromFile(configFile)
		if err != nil {
			color.Red("Error loading config: %v", err)
			os.Exit(2)
		}
	} else {
		// Try to load from project directory
		cfg, err = config.Load(absPath)
		if err != nil {
			color.Yellow("Warning: could not load config: %v", err)
			cfg = config.DefaultConfig()
		}
	}

	// Detect artifacts
	artifacts := detector.Detect(absPath, composeFile, envFiles)

	// Run checks with profile options
	opts := checker.Options{
		EnableSourceScanning: profile.EnableSourceScanning,
		Config:               cfg,
		CheckToolVersions:    checkToolVersions,
	}
	findings := checker.CheckWithOptions(absPath, artifacts, opts)

	// Filter findings based on profile
	findings = profile.FilterFindings(findings)

	// Create report
	report := &models.Report{
		Path:      absPath,
		Artifacts: artifacts,
		Findings:  findings,
	}

	// Calculate summary
	report.CalculateSummary()

	// Generate fix list if requested
	if generateFixList != "" {
		f, err := os.Create(generateFixList)
		if err != nil {
			color.Red("Error creating fix list: %v", err)
			os.Exit(2)
		}
		defer f.Close()

		r := reporter.NewChecklistReporter(f)
		if err := r.Report(report); err != nil {
			color.Red("Error generating fix list: %v", err)
			os.Exit(2)
		}
		color.Green("Fix checklist written to %s", generateFixList)
	}

	// Output based on format
	switch formatFlag {
	case "json":
		r := reporter.NewJSONReporter(os.Stdout, true)
		if err := r.Report(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating JSON: %v\n", err)
			os.Exit(2)
		}
	case "markdown":
		r := reporter.NewMarkdownReporter(os.Stdout)
		if err := r.Report(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating markdown: %v\n", err)
			os.Exit(2)
		}
	case "checklist":
		r := reporter.NewChecklistReporter(os.Stdout)
		if err := r.Report(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating checklist: %v\n", err)
			os.Exit(2)
		}
	default:
		r := reporter.NewTextReporter(os.Stdout, noColor)
		if err := r.Report(report); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating text: %v\n", err)
			os.Exit(2)
		}
	}

	// Exit code handling
	if strictMode && report.Summary.BlockingCount > 0 {
		os.Exit(1)
	}
}
