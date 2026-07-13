package lxd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const ContainerPrefix = "tb-"

type Client struct{}

func New() *Client {
	return &Client{}
}

type ContainerInfo struct {
	Name    string
	Status  string
	IP      string
	Image   string
	Created string
}

func (c *Client) lxcBin() string {
	if runtime.GOOS == "windows" {
		return "wsl"
	}
	if _, err := exec.LookPath("lxc"); err == nil {
		return "lxc"
	}
	return "/snap/bin/lxc"
}

func (c *Client) lxcArgs(args []string) []string {
	if runtime.GOOS == "windows" {
		var quoted []string
		for _, a := range args {
			quoted = append(quoted, fmt.Sprintf("'%s'", strings.ReplaceAll(a, "'", "'\\''")))
		}
		script := fmt.Sprintf(`#!/bin/bash
export PATH=$PATH:/snap/bin
# Start LXD daemon if not running
if ! pgrep -x lxd > /dev/null 2>&1; then
  nohup lxd --group lxd > /dev/null 2>&1 &
  sleep 3
fi
lxc %s
`, strings.Join(quoted, " "))
		tmpFile := filepath.Join(os.TempDir(), "tavpbox-lxc.sh")
		os.WriteFile(tmpFile, []byte(strings.ReplaceAll(script, "\r\n", "\n")), 0755)
		wslPath := strings.ReplaceAll(tmpFile, "\\", "/")
		wslPath = strings.Replace(wslPath, "C:", "/mnt/c", 1)
		wslPath = strings.Replace(wslPath, "c:", "/mnt/c", 1)
		return []string{"-d", "Ubuntu", "-u", "root", "--", "bash", wslPath}
	}
	return args
}

func (c *Client) run(args ...string) error {
	bin := c.lxcBin()
	finalArgs := c.lxcArgs(args)
	cmd := exec.Command(bin, finalArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("lxc %s: %s", strings.Join(args, " "), string(output))
	}
	return nil
}

func (c *Client) runOutput(args ...string) (string, error) {
	bin := c.lxcBin()
	finalArgs := c.lxcArgs(args)
	cmd := exec.Command(bin, finalArgs...)
	var stdout strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = nil
	err := cmd.Run()
	return stdout.String(), err
}

func (c *Client) runTTY(args ...string) error {
	bin := c.lxcBin()
	if runtime.GOOS == "windows" {
		fullCmd := fmt.Sprintf("export PATH=$PATH:/snap/bin && lxc %s", strings.Join(args, " "))
		cmd := exec.Command(bin, "-d", "Ubuntu", "--", "bash", "-c", fullCmd)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (c *Client) ContainerName(name string) string {
	return ContainerPrefix + name
}

func (c *Client) StripPrefix(name string) string {
	return strings.TrimPrefix(name, ContainerPrefix)
}

func (c *Client) IsAvailable() bool {
	if runtime.GOOS == "windows" {
		wslCmd := exec.Command("wsl", "--status")
		if wslCmd.Run() != nil {
			return false
		}
		_, err := c.runOutput("version")
		return err == nil
	}
	bin := c.lxcBin()
	cmd := exec.Command(bin, "version")
	return cmd.Run() == nil
}

func (c *Client) Create(name, image, ram string, cpu int) error {
	imageMap := map[string]string{
		"ubuntu/24.04": "ubuntu:24.04",
		"ubuntu/22.04": "ubuntu:22.04",
		"alpine/3.20":  "images:alpine/3.20",
		"debian/12":    "images:debian/12",
		"fedora/40":    "images:fedora/40",
		"archlinux":    "images:archlinux",
	}
	lxcImage, ok := imageMap[image]
	if !ok {
		lxcImage = "ubuntu:24.04"
	}

	args := []string{"launch", lxcImage, name}
	if ram != "" {
		args = append(args, "-c", fmt.Sprintf("limits.memory=%s", ram))
	}
	if cpu > 0 {
		args = append(args, "-c", fmt.Sprintf("limits.cpu=%d", cpu))
	}
	if err := c.run(args...); err != nil {
		return err
	}
	return c.WaitReady(name, 30*time.Second)
}

func (c *Client) Start(name string) error {
	return c.run("start", name)
}

func (c *Client) Stop(name string) error {
	return c.run("stop", name, "--force")
}

func (c *Client) Delete(name string) error {
	c.run("stop", name, "--force")
	time.Sleep(1 * time.Second)
	return c.run("delete", name, "--force")
}

func (c *Client) Exec(name string, interactive bool, command ...string) error {
	args := []string{"exec", name}
	if interactive {
		args = append(args, "--mode", "interactive")
	}
	args = append(args, "--")
	args = append(args, command...)
	return c.runTTY(args...)
}

func (c *Client) ExecNoTTY(name string, command ...string) (string, error) {
	args := []string{"exec", name, "--"}
	args = append(args, command...)
	return c.runOutput(args...)
}

func (c *Client) List() ([]ContainerInfo, error) {
	output, err := c.runOutput("list", "--format", "csv", "-c", "ns4")
	if err != nil {
		return nil, err
	}

	var containers []ContainerInfo
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ",", 3)
		if len(parts) >= 2 {
			ip := ""
			if len(parts) >= 3 {
				ip = strings.TrimSpace(parts[2])
				ip = strings.Split(ip, " ")[0]
			}
			containers = append(containers, ContainerInfo{
				Name:   strings.TrimSpace(parts[0]),
				Status: strings.TrimSpace(parts[1]),
				IP:     ip,
			})
		}
	}
	return containers, nil
}

func (c *Client) GetIP(name string) (string, error) {
	output, err := c.runOutput("list", name, "--format", "csv", "-c", "4")
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(output, ",") {
		line = strings.TrimSpace(line)
		if strings.Contains(line, ".") {
			return strings.Split(line, " ")[0], nil
		}
	}
	return "", fmt.Errorf("no IP found for %s", name)
}

func (c *Client) MapHostDir(name, hostPath, containerPath string) error {
	wslPath := hostPath
	if runtime.GOOS == "windows" {
		wslPath = strings.ReplaceAll(hostPath, "\\", "/")
		if len(wslPath) >= 2 && wslPath[1] == ':' {
			drive := strings.ToLower(string(wslPath[0]))
			wslPath = "/mnt/" + drive + wslPath[2:]
		}
	}
	return c.run("config", "device", "add", name, "webroot", "disk",
		"source="+wslPath, "path="+containerPath)
}

func (c *Client) Push(name, hostPath, containerPath string) error {
	return c.run("file", "push", hostPath, fmt.Sprintf("%s/%s", name, containerPath))
}

func (c *Client) Pull(name, containerPath, hostPath string) error {
	return c.run("file", "pull", fmt.Sprintf("%s/%s", name, containerPath), hostPath)
}

func (c *Client) Snapshot(name, snapshotName string) error {
	return c.run("snapshot", name, snapshotName)
}

func (c *Client) Restore(name, snapshotName string) error {
	return c.run("restore", name, snapshotName)
}

func (c *Client) WaitReady(name string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		output, _ := c.runOutput("list", name, "--format", "csv", "-c", "n,s")
		if strings.Contains(output, "RUNNING") {
			time.Sleep(2 * time.Second)
			return nil
		}
		time.Sleep(500 * time.Millisecond)
	}
	return fmt.Errorf("container %s not ready within %s", name, timeout)
}
