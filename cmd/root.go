package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	version   = "1.0.0"
	colorMode string
)

var rootCmd = &cobra.Command{
	Use:   "devcheck",
	Short: "Local project readiness inspector",
	Long: color.New(color.FgCyan).Sprint(`
devcheck - Local Project Readiness Inspector

`) + `Scans your repository to find what's needed to run the project locally:
missing env files, undefined variables, unmet dependencies, and more.

` + color.New(color.FgYellow).Sprint(`Read-only analysis only. No commands executed.
`),
	Version: version,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&colorMode, "color", "auto", "Color output: auto, always, never")

	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		switch colorMode {
		case "never":
			color.NoColor = true
		case "always":
			color.NoColor = false
		}
	}
}
