package cmd

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/lxd"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup TAVPBox dependencies (WSL, Ubuntu, LXD)",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TAVPBox Setup")
		fmt.Println("=============")
		fmt.Println("")

		// Check WSL
		if runtime.GOOS == "windows" {
			fmt.Println("[1/3] Checking WSL2...")
			wslCheck := exec.Command("wsl", "--status")
			if wslCheck.Run() != nil {
				fmt.Println("  ! WSL2 not found")
				fmt.Println("  > Please install WSL2:")
				fmt.Println("    1. Open PowerShell as Administrator")
				fmt.Println("    2. Run: wsl --install")
				fmt.Println("    3. Restart computer")
				return fmt.Errorf("WSL not available")
			}
			fmt.Println("  + WSL2 is available")

			// Check Ubuntu
			fmt.Println("[2/3] Checking Ubuntu WSL...")
			ubuntuCheck := exec.Command("wsl", "-d", "Ubuntu", "--", "echo", "ok")
			if ubuntuCheck.Run() != nil {
				fmt.Println("  ! Ubuntu not found")
				fmt.Println("  > Please install Ubuntu:")
				fmt.Println("    1. Open PowerShell as Administrator")
				fmt.Println("    2. Run: wsl --install -d Ubuntu")
				fmt.Println("    3. Set username/password when prompted")
				return fmt.Errorf("Ubuntu not available")
			}
			fmt.Println("  + Ubuntu is available")
		}

		// Check LXD
		fmt.Println("[3/3] Checking LXD...")
		client := lxd.New()
		if client.IsAvailable() {
			fmt.Println("  + LXD is available")
		} else {
			fmt.Println("  ! LXD not found")
			fmt.Println("  > Installing LXD...")
			fmt.Println("  > This may take a few minutes...")
			
			// Try to install LXD
			if runtime.GOOS == "windows" {
				// Install LXD in WSL
				installCmd := exec.Command("wsl", "-d", "Ubuntu", "-u", "root", "--", "bash", "-c",
					"export PATH=$PATH:/snap/bin && snap install lxd && lxd init --auto")
				if installCmd.Run() != nil {
					fmt.Println("  ! LXD installation failed")
					fmt.Println("  > Please install manually:")
					fmt.Println("    1. Open Ubuntu WSL")
					fmt.Println("    2. Run: sudo snap install lxd")
					fmt.Println("    3. Run: sudo lxd init --auto")
					return fmt.Errorf("LXD installation failed")
				}
			} else {
				// Install LXD natively
				installCmd := exec.Command("sudo", "snap", "install", "lxd")
				if installCmd.Run() != nil {
					fmt.Println("  ! LXD installation failed")
					return fmt.Errorf("LXD installation failed")
				}
			}
			fmt.Println("  + LXD installed")
		}

		fmt.Println("")
		fmt.Println("Setup complete! Run: tavpbox init")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
