package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"github.com/stackgen-cli/devcheck/internal/config"
)

var initConfigCmd = &cobra.Command{
	Use:   "init-config",
	Short: "Create a .devcheck.yaml configuration file",
	Long: `Create a .devcheck.yaml configuration file with example settings.

The configuration file allows you to:
- Define custom validation rules
- Specify minimum tool versions (docker, docker-compose, etc.)
- List required environment variables
- Ignore specific finding codes
- Map build contexts to Dockerfiles`,
	RunE: runInitConfig,
}

var initConfigForce bool

func init() {
	initConfigCmd.Flags().BoolVarP(&initConfigForce, "force", "f", false, "Overwrite existing config file")
	rootCmd.AddCommand(initConfigCmd)
}

func runInitConfig(cmd *cobra.Command, args []string) error {
	configPath := ".devcheck.yaml"

	// Check if file already exists
	if _, err := os.Stat(configPath); err == nil && !initConfigForce {
		return fmt.Errorf("%s already exists (use --force to overwrite)", configPath)
	}

	// Write example config
	if err := os.WriteFile(configPath, []byte(config.ExampleConfig()), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	color.Green("✅ Created %s", configPath)
	fmt.Println("\nEdit this file to customize devcheck behavior:")
	fmt.Println("  - Add custom validation rules")
	fmt.Println("  - Specify minimum tool versions")
	fmt.Println("  - Define required environment variables")
	fmt.Println("  - Ignore specific finding codes")

	return nil
}
