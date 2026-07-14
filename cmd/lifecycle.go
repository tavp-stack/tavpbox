package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start the project container",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return err
		}
		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		fmt.Printf("Starting '%s'...\n", cfg.Name)
		if err := client.Start(cname); err != nil {
			return fmt.Errorf("start: %w", err)
		}

		globalCfg, _ := config.LoadGlobal()
		domain := cfg.Name + "." + globalCfg.DomainSuffix
		fmt.Printf("✓ %s started\n", cfg.Name)
		fmt.Printf("  URL: http://%s\n", domain)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the project container",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return err
		}
		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		fmt.Printf("Stopping '%s'...\n", cfg.Name)
		if err := client.Stop(cname); err != nil {
			return fmt.Errorf("stop: %w", err)
		}
		fmt.Printf("✓ %s stopped — RAM freed\n", cfg.Name)
		return nil
	},
}

var restartCmd = &cobra.Command{
	Use:   "restart",
	Short: "Restart the project container",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return err
		}
		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		fmt.Printf("Restarting '%s'...\n", cfg.Name)
		if err := client.Restart(cname); err != nil {
			return fmt.Errorf("restart: %w", err)
		}

		globalCfg, _ := config.LoadGlobal()
		domain := cfg.Name + "." + globalCfg.DomainSuffix
		fmt.Printf("✓ %s restarted\n", cfg.Name)
		fmt.Printf("  URL: http://%s\n", domain)
		return nil
	},
}
