package podman

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	TraefikContainer = "tavp-traefik"
	TraefikImage     = "docker.io/library/traefik:v3.0"
)

// ensureTraefikDirs creates the Traefik config directories if they don't exist
func (c *Client) ensureTraefikDirs() (string, string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", "", fmt.Errorf("home dir: %w", err)
	}
	configDir := filepath.Join(home, ".tavpbox", "traefik")
	dynamicDir := filepath.Join(configDir, "dynamic")
	letsencryptDir := filepath.Join(configDir, "letsencrypt")

	os.MkdirAll(configDir, 0755)
	os.MkdirAll(dynamicDir, 0755)
	os.MkdirAll(letsencryptDir, 0755)

	return configDir, dynamicDir, nil
}

// SetupTraefik creates and starts the Traefik reverse proxy container
func (c *Client) SetupTraefik() error {
	// Ensure directories exist first (even if Traefik is already running)
	configDir, dynamicDir, err := c.ensureTraefikDirs()
	if err != nil {
		return err
	}

	// Always write/update traefik.yml
	// Local dev — no ACME, Traefik uses self-signed cert automatically
	traefikConfig := `api:
  dashboard: true
  insecure: true

providers:
  file:
    directory: /etc/traefik/dynamic
    watch: true

entryPoints:
  web:
    address: ":80"
  websecure:
    address: ":443"
`
	os.WriteFile(filepath.Join(configDir, "traefik.yml"), []byte(traefikConfig), 0644)

	// Remove stale routes.yml that breaks Traefik v3
	os.Remove(filepath.Join(dynamicDir, "routes.yml"))

	// Check if Traefik is already running
	containers, _ := c.List()
	for _, container := range containers {
		if container.Name == TraefikContainer {
			if container.Status == "running" {
				// Restart to pick up config changes
				c.Stop(TraefikContainer)
				c.Remove(TraefikContainer)
				break
			}
			c.Remove(TraefikContainer)
		}
	}

	// Create network if not exists
	c.NetworkCreate("tavp-network")

	// Pull Traefik image
	fmt.Println("  Pulling Traefik image...")
	c.Pull(TraefikImage)

	// Start Traefik container
	fmt.Println("  Starting Traefik...")
	_, err = c.run("run", "-d",
		"--name", TraefikContainer,
		"--restart", "unless-stopped",
		"-p", "80:80",
		"-p", "443:443",
		"-p", "8090:8080",
		"-v", configDir+"/traefik.yml:/etc/traefik/traefik.yml:ro",
		"-v", dynamicDir+":/etc/traefik/dynamic",
		"--network", "tavp-network",
		TraefikImage,
	)
	if err != nil {
		return fmt.Errorf("start traefik: %w", err)
	}

	return nil
}

// AddTraefikRoute adds a route for a container to Traefik
func (c *Client) AddTraefikRoute(name, domain, ip string, port int) error {
	_, dynamicDir, err := c.ensureTraefikDirs()
	if err != nil {
		return fmt.Errorf("traefik dirs: %w", err)
	}

	routeConfig := fmt.Sprintf("http:\n  routers:\n    %s-http:\n      rule: \"Host(`%s`)\"\n      service: %s\n      entryPoints:\n        - web\n    %s-https:\n      rule: \"Host(`%s`)\"\n      service: %s\n      entryPoints:\n        - websecure\n      tls: {}\n  services:\n    %s:\n      loadBalancer:\n        servers:\n          - url: \"http://%s:%d\"\n",
		name, domain, name, name, domain, name, name, ip, port)

	return os.WriteFile(filepath.Join(dynamicDir, name+".yml"), []byte(routeConfig), 0644)
}

// RemoveTraefikRoute removes a route from Traefik
func (c *Client) RemoveTraefikRoute(name string) error {
	home, _ := os.UserHomeDir()
	routeFile := filepath.Join(home, ".tavpbox", "traefik", "dynamic", name+".yml")
	return os.Remove(routeFile)
}

// StopTraefik stops the Traefik container
func (c *Client) StopTraefik() error {
	return c.Stop(TraefikContainer)
}
