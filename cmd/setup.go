package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup TAVPBox dependencies",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TAVPBox Setup")
		fmt.Println("=============")
		fmt.Println("")

		client := podman.New()

		// Step 1: Podman
		fmt.Println("[1/3] Checking Podman...")
		if client.IsAvailable() {
			fmt.Println("  ✓ Podman ready")
		} else {
			fmt.Println("  ! Podman not found")
			fmt.Println("  > Installing Podman...")
			if err := installPodman(); err != nil {
				fmt.Println("  ! Auto-install failed")
				fmt.Println("  > Please install manually:")
				fmt.Println("    1. Download Podman Desktop: https://podman-desktop.io")
				fmt.Println("    2. Install and run Podman Desktop")
				fmt.Println("    3. Run 'podman machine init' then 'podman machine start'")
				return fmt.Errorf("podman not available")
			}
			fmt.Println("  ✓ Podman installed")
		}

		// Step 2: HTTPS cert (disabled — HTTP only)
		fmt.Println("\n[2/2] HTTPS disabled (HTTP only)")
		fmt.Println("  All local dev uses plain HTTP")

		fmt.Println("\n✓ Setup complete! Run: tavpbox init")
		return nil
	},
}

func installPodman() error {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("winget", "install", "-e", "--id", "RedHat.PodmanDesktop")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err == nil {
			return nil
		}
		cmd = exec.Command("choco", "install", "podman-desktop", "-y")
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()

	case "darwin":
		cmd := exec.Command("brew", "install", "podman")
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()

	case "linux":
		cmd := exec.Command("sudo", "apt-get", "install", "-y", "podman")
		cmd.Stdout = nil
		cmd.Stderr = nil
		if err := cmd.Run(); err == nil {
			return nil
		}
		cmd = exec.Command("sudo", "dnf", "install", "-y", "podman")
		cmd.Stdout = nil
		cmd.Stderr = nil
		return cmd.Run()

	default:
		return fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
