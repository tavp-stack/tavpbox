package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/podman"
	"github.com/tavp-stack/tavpbox/internal/proxy"
)

var exposeCmd = &cobra.Command{
	Use:   "expose",
	Short: "Show LAN access URLs for all projects",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := podman.New()
		containers, err := client.List()
		if err != nil {
			return fmt.Errorf("list containers: %w", err)
		}

		if len(containers) == 0 {
			fmt.Println("No containers found. Run: tavpbox create")
			return nil
		}

		lanMgr := proxy.NewLanPortManager()
		hostIP := proxy.GetHostIP()

		fmt.Println()
		fmt.Println("🌐 LAN Access")
		fmt.Println()
		fmt.Println("┌─────────────────────┬──────────────────────────────────────┬───────┐")
		fmt.Println("│ Project             │ URL                                  │ Port  │")
		fmt.Println("├─────────────────────┼──────────────────────────────────────┼───────┤")

		var domains []string
		for _, c := range containers {
			name := client.StripPrefix(c.Name)

			// Check if has assigned LAN port
			lanPort := lanMgr.Get(name)

			// Fallback: use current host port from container
			if lanPort == 0 {
				lanPort = client.GetHostPort(c.Name, "80")
			}

			if lanPort == 0 {
				continue
			}

			url := fmt.Sprintf("http://%s:%d", hostIP, lanPort)
			padding := strings.Repeat(" ", 38-len(url))
			portStr := fmt.Sprintf("%d", lanPort)
			portPadding := strings.Repeat(" ", 5-len(portStr))
			namePadding := strings.Repeat(" ", 20-len(name))

			fmt.Printf("│ %s%s │ %s%s │ %s%s │\n", name, namePadding, url, padding, portStr, portPadding)
			domains = append(domains, name+".tavp.my.id")
		}

		fmt.Println("└─────────────────────┴──────────────────────────────────────┴───────┘")
		fmt.Println()

		// Hosts entry
		if len(domains) > 0 {
			hostsLine := fmt.Sprintf("%s %s", hostIP, strings.Join(domains, " "))
			fmt.Println("📋 Tambahkan ke hosts device lain (opsional):")
			fmt.Printf("   %s\n", hostsLine)
		}
		fmt.Println()
		fmt.Println("💡 Cara pakai:")
		fmt.Printf("   1. Buka browser di device lain\n")
		fmt.Printf("   2. Ketik: http://%s:<port>\n", hostIP)
		fmt.Println("   3. (Opsional) Tambah hosts entry untuk domain access")
		fmt.Println()

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exposeCmd)
}
