package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
	"github.com/tavp-stack/tavpbox/internal/proxy"
)

var destroyCmd = &cobra.Command{
	Use:     "destroy",
	Aliases: []string{"delete"},
	Short:   "Destroy the project container permanently",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return err
		}
		client := podman.New()
		cname := client.ContainerName(cfg.Name)
		globalCfg, _ := config.LoadGlobal()
		domain := cfg.Name + "." + globalCfg.DomainSuffix

		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("Destroy '%s' permanently? [y/N]: ", cfg.Name)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))
		if answer != "y" && answer != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}

		fmt.Printf("Destroying '%s'...\n", cfg.Name)
		if err := client.Remove(cname); err != nil {
			return fmt.Errorf("destroy: %w", err)
		}

		// Remove proxy route
		p := proxy.New(80)
		p.RemoveRoute(domain)
		p.RemoveRoute("mailpit." + domain)
		p.RemoveRoute("adminer." + domain)

		// Release LAN port
		lanMgr := proxy.NewLanPortManager()
		lanMgr.Release(cfg.Name)

		fmt.Printf("✓ %s destroyed\n", cfg.Name)
		return nil
	},
}

var rebuildCmd = &cobra.Command{
	Use:   "rebuild",
	Short: "Rebuild the project container (data preserved)",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return err
		}
		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		fmt.Printf("Rebuilding '%s'...\n", cfg.Name)
		client.Remove(cname)
		return createCmd.RunE(cmd, args)
	},
}
