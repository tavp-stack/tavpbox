package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/proxy"
)

var proxyPort int

var proxyStartCmd = &cobra.Command{
	Use:   "proxy:start",
	Short: "Start the TAVPBox reverse proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if already running
		if isProxyRunning() {
			fmt.Println("Proxy is already running")
			return nil
		}

		p := proxy.New(proxyPort)
		saveProxyPID()
		return p.Start()
	},
}

var proxyStopCmd = &cobra.Command{
	Use:   "proxy:stop",
	Short: "Stop the TAVPBox reverse proxy",
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		pidFile := filepath.Join(home, ".tavpbox", "proxy.pid")
		data, err := os.ReadFile(pidFile)
		if err != nil {
			fmt.Println("Proxy is not running")
			return nil
		}
		pid, _ := strconv.Atoi(string(data))
		if pid == 0 {
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
			fmt.Printf("Proxy stopped (PID %d)\n", pid)
		}
		os.Remove(pidFile)
		return nil
	},
}

var proxyStatusCmd = &cobra.Command{
	Use:   "proxy:status",
	Short: "Show proxy status",
	Run: func(cmd *cobra.Command, args []string) {
		if isProxyRunning() {
			home, _ := os.UserHomeDir()
			pidFile := filepath.Join(home, ".tavpbox", "proxy.pid")
			data, _ := os.ReadFile(pidFile)
			fmt.Printf("Proxy is running (PID %s) on port %d\n", string(data), proxyPort)

			p := proxy.New(proxyPort)
			routes := p.Routes()
			if len(routes) > 0 {
				fmt.Println("\nRoutes:")
				for _, r := range routes {
					fmt.Printf("  %s → %s:%d\n", r.Domain, r.IP, r.Port)
				}
			}
		} else {
			fmt.Println("Proxy is not running")
		}
	},
}

func isProxyRunning() bool {
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".tavpbox", "proxy.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}
	pid, _ := strconv.Atoi(string(data))
	if pid == 0 {
		return false
	}
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	// On Windows, FindProcess always succeeds. Try to check if alive.
	return p.Signal(os.Signal(nil)) == nil
}

func saveProxyPID() {
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".tavpbox", "proxy.pid")
	os.MkdirAll(filepath.Dir(pidFile), 0755)
	os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
}

func ensureProxyRunning() {
	if isProxyRunning() {
		return
	}

	execPath, _ := os.Executable()
	cmd := exec.Command(execPath, "proxy:start", "-p", "80")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Detach process (platform-specific)
	detachProcess(cmd)

	if err := cmd.Start(); err != nil {
		fmt.Printf("  ⚠ Could not start proxy: %v\n", err)
		return
	}

	// Give it a moment to start
	time.Sleep(500 * time.Millisecond)
	fmt.Println("  Proxy started in background (port 80)")
}

func restartProxy() {
	// Stop existing proxy
	home, _ := os.UserHomeDir()
	pidFile := filepath.Join(home, ".tavpbox", "proxy.pid")
	if data, err := os.ReadFile(pidFile); err == nil {
		pid, _ := strconv.Atoi(string(data))
		if pid > 0 {
			if p, err := os.FindProcess(pid); err == nil {
				p.Kill()
			}
		}
		os.Remove(pidFile)
	}
	time.Sleep(500 * time.Millisecond)

	// Start new proxy
	ensureProxyRunning()
}

func init() {
	proxyStartCmd.Flags().IntVarP(&proxyPort, "port", "p", 80, "Port to listen on")
	proxyStopCmd.Flags().IntVarP(&proxyPort, "port", "p", 80, "Port")
	proxyStatusCmd.Flags().IntVarP(&proxyPort, "port", "p", 80, "Port")

	rootCmd.AddCommand(proxyStartCmd, proxyStopCmd, proxyStatusCmd)
}
