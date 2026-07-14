package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage TAVPBox configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a configuration value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		value := args[1]

		globalCfg, _ := config.LoadGlobal()

		switch key {
		case "cloudflare_token":
			globalCfg.CloudflareToken = value
		case "cloudflare_zone":
			globalCfg.CloudflareZone = value
		case "domain_suffix":
			globalCfg.DomainSuffix = value
		default:
			return fmt.Errorf("unknown config key: %s", key)
		}

		if err := config.SaveGlobal(globalCfg); err != nil {
			return err
		}
		fmt.Printf("Set %s = %s\n", key, value)
		return nil
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration value",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		key := args[0]
		globalCfg, _ := config.LoadGlobal()

		switch key {
		case "cloudflare_token":
			if globalCfg.CloudflareToken != "" {
				fmt.Println(globalCfg.CloudflareToken[:8] + "...")
			} else {
				fmt.Println("(not set)")
			}
		case "cloudflare_zone":
			fmt.Println(globalCfg.CloudflareZone)
		case "domain_suffix":
			fmt.Println(globalCfg.DomainSuffix)
		default:
			fmt.Printf("Unknown key: %s\n", key)
		}
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configuration",
	Run: func(cmd *cobra.Command, args []string) {
		globalCfg, _ := config.LoadGlobal()
		home, _ := os.UserHomeDir()
		configPath := filepath.Join(home, ".tavpbox", "config.yml")

		fmt.Printf("Config file: %s\n\n", configPath)
		fmt.Printf("  domain_suffix:     %s\n", globalCfg.DomainSuffix)
		fmt.Printf("  default_ram:       %s\n", globalCfg.DefaultRAM)
		fmt.Printf("  default_cpu:       %d\n", globalCfg.DefaultCPU)
		fmt.Printf("  default_image:     %s\n", globalCfg.DefaultImage)
		fmt.Printf("  panel_port:        %d\n", globalCfg.PanelPort)
		if globalCfg.CloudflareToken != "" {
			fmt.Printf("  cloudflare_token:  %s...\n", globalCfg.CloudflareToken[:8])
		} else {
			fmt.Printf("  cloudflare_token:  (not set)\n")
		}
		fmt.Printf("  cloudflare_zone:   %s\n", globalCfg.CloudflareZone)
	},
}

func init() {
	configCmd.AddCommand(configSetCmd, configGetCmd, configListCmd)
	rootCmd.AddCommand(configCmd)
}
