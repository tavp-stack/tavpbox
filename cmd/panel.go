package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/api"
	"github.com/tavp-stack/tavpbox/internal/config"
)

var panelPort int
var panelOpen bool

var panelCmd = &cobra.Command{
	Use:   "panel",
	Short: "Start the TAVPBox web panel",
	RunE: func(cmd *cobra.Command, args []string) error {
		globalCfg, _ := config.LoadGlobal()
		if panelPort == 0 {
			panelPort = globalCfg.PanelPort
			if panelPort == 0 {
				panelPort = 8080
			}
		}

		url := fmt.Sprintf("http://localhost:%d", panelPort)
		fmt.Printf("Starting TAVPBox Panel on port %d...\n", panelPort)
		fmt.Printf("Dashboard: %s\n", url)
		fmt.Println("Press Ctrl+C to stop")

		if panelOpen {
			openBrowser(url)
		}

		// Save PID for panel:stop
		savePID()

		return api.Start(panelPort)
	},
}

var panelStopCmd = &cobra.Command{
	Use:   "panel:stop",
	Short: "Stop the TAVPBox panel (if running in background)",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		pidFile := filepath.Join(home, ".tavpbox", "panel.pid")
		data, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Println("No panel PID file found")
			return nil
		}
		pidStr := string(data)
		pid, err := strconv.Atoi(pidStr)
		if err != nil || pid == 0 {
			fmt.Println("Invalid PID")
			return nil
		}

		p, err := os.FindProcess(pid)
		if err != nil {
			fmt.Printf("Process %d not found\n", pid)
			os.Remove(pidFile)
			return nil
		}

		if err := p.Kill(); err != nil {
			fmt.Printf("Could not stop process %d: %v\n", pid, err)
		} else {
			fmt.Printf("Panel stopped (PID %d)\n", pid)
		}
		os.Remove(pidFile)
		return nil
	},
}

func savePID() {
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".tavpbox", "panel.pid")
	os.MkdirAll(filepath.Dir(pidFile), 0755)
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	cmd.Start()
}

func init() {
	panelCmd.Flags().IntVarP(&panelPort, "port", "p", 0, "Port to listen on (default: 8080)")
	panelCmd.Flags().BoolVar(&panelOpen, "open", true, "Open browser automatically")
	rootCmd.AddCommand(panelCmd)
	rootCmd.AddCommand(panelStopCmd)
}
