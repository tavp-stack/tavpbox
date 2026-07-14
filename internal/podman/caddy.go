package podman

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	CaddyContainer = "tavp-caddy"
	CaddyImage     = "docker.io/library/caddy:2-alpine"
)

// ensureCaddyDirs creates the Caddy config directory
func (c *Client) ensureCaddyDirs() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("home dir: %w", err)
	}
	caddyDir := filepath.Join(home, ".tavpbox", "caddy")
	os.MkdirAll(caddyDir, 0755)
	return caddyDir, nil
}

// SetupCaddy creates and starts the Caddy reverse proxy container
func (c *Client) SetupCaddy() error {
	caddyDir, err := c.ensureCaddyDirs()
	if err != nil {
		return err
	}

	// Ensure routes file exists
	routesFile := filepath.Join(caddyDir, "routes.txt")
	if _, err := os.Stat(routesFile); os.IsNotExist(err) {
		os.WriteFile(routesFile, []byte(""), 0644)
	}

	// Rebuild Caddyfile from routes
	if err := c.rebuildCaddyfile(); err != nil {
		return err
	}

	// Check if Caddy is already running
	containers, _ := c.List()
	for _, container := range containers {
		if container.Name == CaddyContainer {
			if container.Status == "running" {
				// Already running — just reload config
				c.ReloadCaddy()
				return nil
			}
			c.Remove(CaddyContainer)
		}
	}

	// Create network if not exists
	c.NetworkCreate("tavp-network")

	// Pull Caddy image
	fmt.Println("  Pulling Caddy image...")
	c.Pull(CaddyImage)

	// Start Caddy container
	fmt.Println("  Starting Caddy...")
	_, err = c.run("run", "-d",
		"--name", CaddyContainer,
		"--restart", "unless-stopped",
		"-p", "80:80",
		"-p", "443:443",
		"-v", caddyDir+"/Caddyfile:/etc/caddy/Caddyfile",
		"--network", "tavp-network",
		CaddyImage,
	)
	if err != nil {
		return fmt.Errorf("start caddy: %w", err)
	}

	return nil
}

// AddCaddyRoute adds a route for a container
func (c *Client) AddCaddyRoute(name, domain, ip string, port int) error {
	caddyDir, err := c.ensureCaddyDirs()
	if err != nil {
		return err
	}

	routesFile := filepath.Join(caddyDir, "routes.txt")

	// Read existing routes
	data, _ := os.ReadFile(routesFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Remove existing route for this domain (if any)
	var newLines []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 1 && parts[0] == domain {
			continue // skip existing route for this domain
		}
		if line != "" {
			newLines = append(newLines, line)
		}
	}

	// Add new route
	newLines = append(newLines, fmt.Sprintf("%s %s %d", domain, ip, port))

	// Write routes file
	os.WriteFile(routesFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644)

	// Rebuild Caddyfile
	return c.rebuildCaddyfile()
}

// RemoveCaddyRoute removes a route
func (c *Client) RemoveCaddyRoute(name, domain string) error {
	caddyDir, err := c.ensureCaddyDirs()
	if err != nil {
		return err
	}

	routesFile := filepath.Join(caddyDir, "routes.txt")
	data, _ := os.ReadFile(routesFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	var newLines []string
	for _, line := range lines {
		parts := strings.Fields(line)
		if len(parts) >= 1 && parts[0] == domain {
			continue
		}
		if line != "" {
			newLines = append(newLines, line)
		}
	}

	os.WriteFile(routesFile, []byte(strings.Join(newLines, "\n")+"\n"), 0644)
	return c.rebuildCaddyfile()
}

// ReloadCaddy reloads the Caddy config
func (c *Client) ReloadCaddy() error {
	_, err := c.Exec(CaddyContainer, "caddy", "reload", "--config", "/etc/caddy/Caddyfile")
	return err
}

// StopCaddy stops the Caddy container
func (c *Client) StopCaddy() error {
	return c.Stop(CaddyContainer)
}

// rebuildCaddyfile regenerates the Caddyfile from routes.txt
func (c *Client) rebuildCaddyfile() error {
	caddyDir, err := c.ensureCaddyDirs()
	if err != nil {
		return err
	}

	routesFile := filepath.Join(caddyDir, "routes.txt")
	data, _ := os.ReadFile(routesFile)
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")

	// Build Caddyfile
	var caddyfile strings.Builder

	// Global config
	caddyfile.WriteString("{\n")
	caddyfile.WriteString("\tauto_https disable_redirects\n")
	caddyfile.WriteString("}\n\n")

	// Default catch-all (if no routes)
	hasRoutes := false
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			hasRoutes = true
			break
		}
	}

	if !hasRoutes {
		caddyfile.WriteString(":80 {\n")
		caddyfile.WriteString("\trespond \"TAVPBox — No project configured\" 404\n")
		caddyfile.WriteString("}\n")
	} else {
		// Write each route — HTTP only (no TLS for local dev)
		for _, line := range lines {
			parts := strings.Fields(strings.TrimSpace(line))
			if len(parts) < 3 {
				continue
			}
			domain := parts[0]
			ip := parts[1]
			port := parts[2]

			caddyfile.WriteString(fmt.Sprintf("%s {\n", domain))
			caddyfile.WriteString(fmt.Sprintf("\treverse_proxy %s:%s\n", ip, port))
			caddyfile.WriteString("}\n\n")
		}
	}

	caddyPath := filepath.Join(caddyDir, "Caddyfile")
	return os.WriteFile(caddyPath, []byte(caddyfile.String()), 0644)
}
