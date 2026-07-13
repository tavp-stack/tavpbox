package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/lxd"
)

var startCmd = &cobra.Command{
	Use:   "start <name>",
	Short: "Start a box",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		fmt.Printf("Starting %s...\n", name)
		if err := client.Start(containerName); err != nil {
			return err
		}

		globalCfg, _ := config.LoadGlobal()
		domain := name + "." + globalCfg.DomainSuffix
		fmt.Printf("✓ %s started — http://%s\n", name, domain)
		return nil
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop <name>",
	Short: "Stop a box (RAM freed immediately)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		fmt.Printf("Stopping %s...\n", name)
		if err := client.Stop(containerName); err != nil {
			return err
		}
		fmt.Printf("✓ %s stopped — RAM freed\n", name)
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all boxes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := lxd.New()
		globalCfg, _ := config.LoadGlobal()

		containers, err := client.List()
		if err != nil {
			return err
		}

		fmt.Printf("%-20s %-10s %-18s %-15s\n", "NAME", "STATUS", "IP", "URL")
		fmt.Println("────────────────────────────────────────────────────────────────────")

		for _, c := range containers {
			name := client.StripPrefix(c.Name)
			status := c.Status
			if status == "RUNNING" {
				status = "✓ running"
			} else {
				status = "○ stopped"
			}
			url := name + "." + globalCfg.DomainSuffix
			fmt.Printf("%-20s %-10s %-18s %-15s\n", name, status, c.IP, url)
		}
		return nil
	},
}

var destroyCmd = &cobra.Command{
	Use:   "destroy <name>",
	Short: "Destroy a box permanently",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		fmt.Printf("⚠ This will permanently delete '%s'. Continue? [y/N] ", name)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("Cancelled.")
			return nil
		}

		fmt.Printf("Destroying %s...\n", name)
		if err := client.Delete(containerName); err != nil {
			return err
		}

		fmt.Printf("✓ %s destroyed\n", name)
		return nil
	},
}

var rebuildCmd = &cobra.Command{
	Use:   "rebuild <name>",
	Short: "Recreate box (data in mapped folders preserved)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		_, projectCfg, err := config.FindProject()
		if err != nil {
			return fmt.Errorf("project config not found: %w", err)
		}

		client := lxd.New()
		containerName := client.ContainerName(name)
		client.Delete(containerName)

		fmt.Printf("Rebuilding %s...\n", name)
		return createBox(projectCfg)
	},
}

var sshCmd = &cobra.Command{
	Use:   "ssh",
	Short: "Enter a box terminal or run a command",
	FParseErrWhitelist: cobra.FParseErrWhitelist{
		UnknownFlags: true,
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 0 && args[0] == "--" {
			args = args[1:]
		}
		if len(args) < 1 {
			return cmd.Help()
		}

		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		if len(args) > 1 {
			command := strings.Join(args[1:], " ")
			output, err := client.ExecNoTTY(containerName, "bash", "-c", command)
			if len(output) > 0 {
				fmt.Print(output)
			}
			if err != nil {
				return fmt.Errorf("exit status 1")
			}
			return nil
		}

		return client.Exec(containerName, true, "/bin/bash")
	},
}

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show box info",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		ip, _ := client.GetIP(containerName)
		_, projectCfg, _ := config.FindProject()

		fmt.Printf("Name: %s\n", name)
		if projectCfg != nil {
			fmt.Printf("Stack: %s\n", projectCfg.Stack)
			fmt.Printf("Services: %s\n", strings.Join(projectCfg.Services, ", "))
			fmt.Printf("Webroot: %s\n", projectCfg.Webroot)
		}
		fmt.Printf("IP: %s\n", ip)
		return nil
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show system status",
	RunE: func(cmd *cobra.Command, args []string) error {
		client := lxd.New()

		containers, err := client.List()
		if err != nil {
			return err
		}

		fmt.Printf("%-20s %-10s %-18s %-10s %-10s\n", "BOX", "STATUS", "IP", "RAM", "CPU")
		fmt.Println("────────────────────────────────────────────────────────────────────────")

		for _, c := range containers {
			name := client.StripPrefix(c.Name)
			status := c.Status
			if status == "RUNNING" {
				status = "✓ running"
			} else {
				status = "○ stopped"
			}

			// Get resource usage
			ram := "-"
			cpu := "-"
			if c.Status == "RUNNING" {
				info, err := client.ExecNoTTY(c.Name, "cat", "/proc/meminfo")
				if err == nil {
					// Parse MemAvailable
					for _, line := range strings.Split(info, "\n") {
						if strings.HasPrefix(line, "MemAvailable:") {
							parts := strings.Fields(line)
							if len(parts) >= 2 {
								ram = parts[1] + " kB"
							}
						}
					}
				}
			}

			fmt.Printf("%-20s %-10s %-18s %-10s %-10s\n", name, status, c.IP, ram, cpu)
		}

		fmt.Printf("\nTotal: %d boxes\n", len(containers))
		return nil
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <name>",
	Short: "Display logs for a box",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		client := lxd.New()
		containerName := client.ContainerName(name)

		// Get nginx access log
		fmt.Println("=== Nginx Access Log ===")
		output, err := client.ExecNoTTY(containerName, "tail", "-n", "50", "/var/log/nginx/access.log")
		if err == nil && len(output) > 0 {
			fmt.Print(output)
		}

		// Get PHP error log
		fmt.Println("\n=== PHP Error Log ===")
		output, err = client.ExecNoTTY(containerName, "tail", "-n", "50", "/var/log/php8.3-fpm.log")
		if err == nil && len(output) > 0 {
			fmt.Print(output)
		}

		// Get system log
		fmt.Println("\n=== System Log ===")
		output, err = client.ExecNoTTY(containerName, "journalctl", "-n", "50", "--no-pager")
		if err == nil && len(output) > 0 {
			fmt.Print(output)
		}

		return nil
	},
}

var execCmd = &cobra.Command{
	Use:   "exec <name> <command>",
	Short: "Execute a command in a box",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		command := strings.Join(args[1:], " ")
		client := lxd.New()
		containerName := client.ContainerName(name)

		output, err := client.ExecNoTTY(containerName, "bash", "-c", command)
		if len(output) > 0 {
			fmt.Print(output)
		}
		return err
	},
}

var snapshotCmd = &cobra.Command{
	Use:   "snapshot <name> [snapshot-name]",
	Short: "Create a snapshot",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		snapName := "snap-0"
		if len(args) > 1 {
			snapName = args[1]
		}

		client := lxd.New()
		containerName := client.ContainerName(name)

		fmt.Printf("Creating snapshot '%s' for %s...\n", snapName, name)
		if err := client.Snapshot(containerName, snapName); err != nil {
			return err
		}
		fmt.Printf("✓ Snapshot created: %s/%s\n", name, snapName)
		return nil
	},
}
