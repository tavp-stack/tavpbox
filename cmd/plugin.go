package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/lxd"
	"gopkg.in/yaml.v3"
)

var pluginCmd = &cobra.Command{
	Use:   "plugin",
	Short: "Manage plugins",
}

var pluginListCmd = &cobra.Command{
	Use:   "list",
	Short: "List installed plugins",
	RunE: func(cmd *cobra.Command, args []string) error {
		home := config.HomeDir()

		// List stack plugins
		fmt.Println("Stacks:")
		stackDir := filepath.Join(home, "plugins", "stacks")
		entries, err := os.ReadDir(stackDir)
		if err == nil {
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".yml") {
					name := strings.TrimSuffix(entry.Name(), ".yml")
					fmt.Printf("  - %s\n", name)
				}
			}
		}

		// List service plugins
		fmt.Println("\nServices:")
		serviceDir := filepath.Join(home, "plugins", "services")
		entries, err = os.ReadDir(serviceDir)
		if err == nil {
			for _, entry := range entries {
				if strings.HasSuffix(entry.Name(), ".yml") {
					name := strings.TrimSuffix(entry.Name(), ".yml")
					fmt.Printf("  - %s\n", name)
				}
			}
		}

		return nil
	},
}

var pluginInstallCmd = &cobra.Command{
	Use:   "install <path>",
	Short: "Install a plugin from a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path := args[0]

		// Read plugin file
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("cannot read file: %w", err)
		}

		// Parse plugin
		var plugin struct {
			Name     string `yaml:"name"`
			Category string `yaml:"category"`
		}
		if err := yaml.Unmarshal(data, &plugin); err != nil {
			return fmt.Errorf("invalid plugin YAML: %w", err)
		}

		if plugin.Name == "" {
			return fmt.Errorf("plugin name is required")
		}

		// Determine target directory
		home := config.HomeDir()
		var targetDir string
		switch plugin.Category {
		case "stack":
			targetDir = filepath.Join(home, "plugins", "stacks")
		case "service":
			targetDir = filepath.Join(home, "plugins", "services")
		default:
			return fmt.Errorf("unknown plugin category: %s", plugin.Category)
		}

		os.MkdirAll(targetDir, 0755)

		// Copy plugin file
		targetPath := filepath.Join(targetDir, plugin.Name+".yml")
		if err := os.WriteFile(targetPath, data, 0644); err != nil {
			return fmt.Errorf("failed to install plugin: %w", err)
		}

		fmt.Printf("✓ Plugin '%s' installed\n", plugin.Name)
		return nil
	},
}

var pluginRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a plugin",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		home := config.HomeDir()

		// Try stacks
		stackPath := filepath.Join(home, "plugins", "stacks", name+".yml")
		if err := os.Remove(stackPath); err == nil {
			fmt.Printf("✓ Plugin '%s' removed\n", name)
			return nil
		}

		// Try services
		servicePath := filepath.Join(home, "plugins", "services", name+".yml")
		if err := os.Remove(servicePath); err == nil {
			fmt.Printf("✓ Plugin '%s' removed\n", name)
			return nil
		}

		return fmt.Errorf("plugin '%s' not found", name)
	},
}

// Tooling system - dynamic commands from .tavpbox.yml
func registerToolingCommands() {
	_, projectCfg, err := config.FindProject()
	if err != nil {
		return
	}

	if projectCfg.Tooling == nil {
		return
	}

	for name, tooling := range projectCfg.Tooling {
		cmdName := name
		cmdStr := tooling.Cmd

		toolingCmd := &cobra.Command{
			Use:   cmdName,
			Short: fmt.Sprintf("Run '%s' in container", cmdStr),
			FParseErrWhitelist: cobra.FParseErrWhitelist{
				UnknownFlags: true,
			},
			RunE: func(cmd *cobra.Command, args []string) error {
				client := lxd.New()
				containerName := client.ContainerName(projectCfg.Name)

				command := cmdStr
				if len(args) > 0 {
					command = cmdStr + " " + strings.Join(args, " ")
				}

				output, err := client.ExecNoTTY(containerName, "bash", "-c", command)
				if len(output) > 0 {
					fmt.Print(output)
				}
				return err
			},
		}

		rootCmd.AddCommand(toolingCmd)
	}
}

func init() {
	pluginCmd.AddCommand(pluginListCmd)
	pluginCmd.AddCommand(pluginInstallCmd)
	pluginCmd.AddCommand(pluginRemoveCmd)
	rootCmd.AddCommand(pluginCmd)

	// Register tooling commands from .tavpbox.yml
	registerToolingCommands()
}
