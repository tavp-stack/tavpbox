package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/certs"
	"github.com/tavp-stack/tavpbox/internal/config"
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

		// Step 2: HTTPS cert
		fmt.Println("\n[2/2] Setting up HTTPS...")
		globalCfg, _ := config.LoadGlobal()
		if globalCfg.CloudflareToken == "" {
			fmt.Println("  ⚠ Cloudflare token not set")
			fmt.Println("  > Run: tavpbox config set cloudflare_token <token>")
			fmt.Println("  > Run: tavpbox config set cloudflare_zone <zone_id>")
		} else {
			fmt.Println("  Generating wildcard cert via Let's Encrypt...")
		if _, _, err := certs.GenerateWildcardCert("tavp.my.id", globalCfg.CloudflareToken); err != nil {
			fmt.Printf("  ⚠ Cert generation failed: %v\n", err)
		} else {
			fmt.Println("  ✓ Wildcard cert generated (*.tavp.my.id)")
			fmt.Println("  Cert expires in ~90 days, auto-renew on next tavpbox setup")

			// Restart proxy to pick up new cert
			fmt.Println("  Restarting proxy...")
			restartProxy()
			fmt.Println("  ✓ Proxy restarted with new cert")
		}
		}

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
