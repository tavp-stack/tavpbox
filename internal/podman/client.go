package podman

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

const ContainerPrefix = "tavp-"

type Client struct{}

func New() *Client { return &Client{} }

type ContainerInfo struct {
	Name   string
	Status string
	IP     string
	Image  string
	Ports  string
}

// Podman binary path
func (c *Client) bin() string {
	if runtime.GOOS == "windows" {
		// On Windows, podman.exe should be in PATH after Podman Desktop install
		return "podman"
	}
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman"
	}
	return "/usr/bin/podman"
}

// Run a podman command and return output
func (c *Client) run(args ...string) (string, error) {
	bin := c.bin()
	cmd := exec.Command(bin, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return stdout.String(), fmt.Errorf("podman %s: %s", strings.Join(args, " "), strings.TrimSpace(stderr.String()))
	}
	return strings.TrimSpace(stdout.String()), nil
}

// ContainerName returns the prefixed container name
func (c *Client) ContainerName(name string) string {
	return ContainerPrefix + name
}

// StripPrefix removes the container prefix from a name
func (c *Client) StripPrefix(name string) string {
	return strings.TrimPrefix(name, ContainerPrefix)
}

// IsAvailable checks if Podman is installed and working
func (c *Client) IsAvailable() bool {
	_, err := c.run("version")
	return err == nil
}

// Create creates a new container
func (c *Client) Create(name, image string, ports []string, env map[string]string, volumes []string, labels map[string]string) error {
	args := []string{"run", "-d", "--name", name}

	// Add port mappings
	for _, port := range ports {
		args = append(args, "-p", port)
	}

	// Add environment variables
	for key, value := range env {
		args = append(args, "-e", fmt.Sprintf("%s=%s", key, value))
	}

	// Add volume mounts
	for _, volume := range volumes {
		args = append(args, "-v", volume)
	}

	// Add labels for Traefik
	for key, value := range labels {
		args = append(args, "-l", fmt.Sprintf("%s=%s", key, value))
	}

	// Add restart policy
	args = append(args, "--restart", "unless-stopped")

	// Add network
	args = append(args, "--network", "tavp-network")

	// Add image
	args = append(args, image)

	// Use default CMD from image (startup script in Containerfile)
	_, err := c.run(args...)
	return err
}

// Start starts a container
func (c *Client) Start(name string) error {
	_, err := c.run("start", name)
	return err
}

// Stop stops a container
func (c *Client) Stop(name string) error {
	_, err := c.run("stop", name)
	return err
}

// Restart restarts a container
func (c *Client) Restart(name string) error {
	_, err := c.run("restart", name)
	return err
}

// Remove removes a container
func (c *Client) Remove(name string) error {
	_, err := c.run("rm", "-f", name)
	return err
}

// Exec executes a command in a running container
func (c *Client) Exec(name string, cmdArgs ...string) (string, error) {
	args := append([]string{"exec", name}, cmdArgs...)
	return c.run(args...)
}

// ExecInteractive executes a command interactively
func (c *Client) ExecInteractive(name string, cmdArgs ...string) error {
	bin := c.bin()
	args := append([]string{"exec", "-it", name}, cmdArgs...)
	cmd := exec.Command(bin, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// List returns all containers with the tavp- prefix
func (c *Client) List() ([]ContainerInfo, error) {
	output, err := c.run("ps", "-a", "--filter", fmt.Sprintf("name=%s", ContainerPrefix),
		"--format", "{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}")
	if err != nil {
		return nil, err
	}

	var containers []ContainerInfo
	for _, line := range strings.Split(output, "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 4)
		if len(parts) >= 2 {
			info := ContainerInfo{
				Name:   parts[0],
				Status: parts[1],
			}
			if len(parts) >= 3 {
				info.Ports = parts[2]
			}
			if len(parts) >= 4 {
				info.Image = parts[3]
			}
			containers = append(containers, info)
		}
	}
	return containers, nil
}

// GetIP returns the IP address of a container
func (c *Client) GetIP(name string) (string, error) {
	// Try Podman format first (network-specific)
	output, err := c.run("inspect", "-f", "{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}", name)
	if err == nil && output != "" && output != "no value" {
		return output, nil
	}

	// Fallback: try Docker format
	output, err = c.run("inspect", "-f", "{{.NetworkSettings.IPAddress}}", name)
	if err == nil && output != "" && output != "no value" {
		return output, nil
	}

	// Fallback: get from network inspect
	netOutput, err := c.run("network", "inspect", "tavp-network")
	if err == nil {
		// Parse the network output to find the container IP
		lines := strings.Split(netOutput, "\n")
		for i, line := range lines {
			if strings.Contains(line, name) {
				// Look for IP in nearby lines
				for j := i; j < len(lines) && j < i+10; j++ {
					if strings.Contains(lines[j], "ipnet") {
						// Extract IP from "ipnet": "10.89.0.5/24"
						parts := strings.Split(lines[j], "\"")
						for _, p := range parts {
							if strings.Contains(p, "/") {
								ip := strings.Split(p, "/")[0]
								if ip != "" {
									return ip, nil
								}
							}
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("no IP found for %s", name)
}

// Inspect returns detailed container info
func (c *Client) Inspect(name string) (string, error) {
	return c.run("inspect", name)
}

// GetHostPort returns the host port mapped to a container port
func (c *Client) GetHostPort(name, containerPort string) int {
	output, err := c.run("port", name, containerPort)
	if err != nil || output == "" {
		return 0
	}
	// Output format: "0.0.0.0:32768" or "[::]:32768"
	parts := strings.Split(output, ":")
	if len(parts) >= 2 {
		portStr := strings.TrimSpace(parts[len(parts)-1])
		port := 0
		for _, ch := range portStr {
			if ch >= '0' && ch <= '9' {
				port = port*10 + int(ch-'0')
			}
		}
		return port
	}
	return 0
}

// Logs returns container logs
func (c *Client) Logs(name string, tail int) (string, error) {
	args := []string{"logs"}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	args = append(args, name)
	return c.run(args...)
}

// Pull pulls an image
func (c *Client) Pull(image string) error {
	_, err := c.run("pull", image)
	return err
}

// NetworkCreate creates a podman network
func (c *Client) NetworkCreate(name string) error {
	_, err := c.run("network", "create", name)
	return err
}

// NetworkConnect connects a container to a network
func (c *Client) NetworkConnect(network, container string) error {
	_, err := c.run("network", "connect", network, container)
	return err
}

// VolumeCreate creates a podman volume
func (c *Client) VolumeCreate(name string) error {
	_, err := c.run("volume", "create", name)
	return err
}

// Commit creates an image from a container
func (c *Client) Commit(container, image string) (string, error) {
	return c.run("commit", container, image)
}

// Push pushes an image to a registry
func (c *Client) Push(image string) (string, error) {
	return c.run("push", image)
}

// ListImages lists all local images
func (c *Client) ListImages() (string, error) {
	return c.run("images", "--format", "table {{.Repository}}\t{{.Tag}}\t{{.Size}}")
}
