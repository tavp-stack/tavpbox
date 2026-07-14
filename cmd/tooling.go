package cmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

// defaultTooling returns default tooling commands based on recipe
func defaultTooling(recipe string) map[string]config.ToolingConfig {
	switch recipe {
	case "tavp", "laravel":
		return map[string]config.ToolingConfig{
			"artisan":  {Cmd: "php artisan"},
			"composer": {Cmd: "composer"},
			"npm":      {Cmd: "npm"},
			"npx":      {Cmd: "npx"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php artisan test"},
		}
	case "php":
		return map[string]config.ToolingConfig{
			"composer": {Cmd: "composer"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php vendor/bin/phpunit"},
		}
	case "node":
		return map[string]config.ToolingConfig{
			"npm":  {Cmd: "npm"},
			"npx":  {Cmd: "npx"},
			"yarn": {Cmd: "yarn"},
			"pnpm": {Cmd: "pnpm"},
			"node": {Cmd: "node"},
		}
	case "go":
		return map[string]config.ToolingConfig{
			"go": {Cmd: "go"},
		}
	case "python":
		return map[string]config.ToolingConfig{
			"python": {Cmd: "python3"},
			"pip":    {Cmd: "pip3"},
			"pytest": {Cmd: "python3 -m pytest"},
		}
	default:
		return nil
	}
}

// GetTooling returns the merged tooling config (defaults + user overrides)
func GetTooling(cfg *config.ProjectConfig) map[string]config.ToolingConfig {
	tooling := defaultTooling(cfg.Recipe)
	if tooling == nil {
		tooling = make(map[string]config.ToolingConfig)
	}
	// User overrides from .tavpbox.yml
	for name, tc := range cfg.Tooling {
		tooling[name] = tc
	}
	return tooling
}

// RegisterToolingCommands dynamically adds tooling subcommands to root
func RegisterToolingCommands() {
	_, cfg, err := config.FindProject()
	if err != nil {
		return
	}

	tooling := GetTooling(cfg)
	if len(tooling) == 0 {
		return
	}

	for name, tc := range tooling {
		cmdName := name
		cmdDef := tc
		rootCmd.AddCommand(&cobra.Command{
			Use:                cmdName + " [args...]",
			Short:              fmt.Sprintf("Run: %s", cmdDef.Cmd),
			DisableFlagParsing: true,
			RunE: func(cmd *cobra.Command, args []string) error {
				return runTooling(cfg, cmdDef, args)
			},
		})
	}
}

// runTooling executes a tooling command inside the container
func runTooling(cfg *config.ProjectConfig, tc config.ToolingConfig, args []string) error {
	client := podman.New()
	cname := client.ContainerName(cfg.Name)

	fullCmd := tc.Cmd
	if len(args) > 0 {
		fullCmd = tc.Cmd + " " + strings.Join(args, " ")
	}

	// Split into command parts for Exec
	parts := strings.Fields(fullCmd)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	output, err := client.Exec(cname, parts...)
	if len(output) > 0 {
		fmt.Println(output)
	}
	return err
}

var toolingListCmd = &cobra.Command{
	Use:     "tooling",
	Aliases: []string{"tools"},
	Short:   "List available tooling commands",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return fmt.Errorf(".tavpbox.yml not found. Run: tavpbox init")
		}

		tooling := GetTooling(cfg)
		if len(tooling) == 0 {
			fmt.Println("No tooling commands available.")
			return nil
		}

		// Sort for consistent output
		names := make([]string, 0, len(tooling))
		for name := range tooling {
			names = append(names, name)
		}
		sort.Strings(names)

		fmt.Printf("Tooling for '%s' (%s):\n\n", cfg.Name, cfg.Recipe)
		for _, name := range names {
			tc := tooling[name]
			fmt.Printf("  tavpbox %-15s → %s\n", name, tc.Cmd)
		}
		fmt.Println("\nUsage: tavpbox <command> [args...]")
		fmt.Println("  Example: tavpbox artisan migrate")
		return nil
	},
}
