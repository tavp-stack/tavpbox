package service

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/tavp-stack/tavpbox/internal/lxd"
)

//go:embed scripts/*.sh
var embeddedScripts embed.FS

type Manager struct {
	client *lxd.Client
}

func NewManager(client *lxd.Client) *Manager {
	return &Manager{client: client}
}

func (m *Manager) Install(containerName, serviceName string) error {
	script, err := m.getServiceScript(serviceName)
	if err != nil {
		return err
	}

	// Push script to container
	tmpFile := filepath.Join(os.TempDir(), fmt.Sprintf("tavpbox-svc-%s.sh", serviceName))
	os.WriteFile(tmpFile, []byte(script), 0755)
	m.client.Push(containerName, tmpFile, "/tmp/svc-install.sh")
	os.Remove(tmpFile)

	// Execute service installer
	_, err = m.client.ExecNoTTY(containerName, "bash", "-c", "chmod +x /tmp/svc-install.sh && bash /tmp/svc-install.sh")
	if err != nil {
		return fmt.Errorf("service install failed: %w", err)
	}

	return nil
}

func (m *Manager) InstallAll(containerName string, services []string) error {
	for _, svc := range services {
		if err := m.Install(containerName, svc); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) getServiceScript(name string) (string, error) {
	// Try embedded first
	scriptPath := fmt.Sprintf("scripts/%s.sh", name)
	data, err := embeddedScripts.ReadFile(scriptPath)
	if err == nil {
		return string(data), nil
	}

	// Try user's custom services
	home, _ := os.UserHomeDir()
	userPath := filepath.Join(home, ".tavpbox", "plugins", "services", name+".sh")
	data, err = os.ReadFile(userPath)
	if err == nil {
		return string(data), nil
	}

	return "", fmt.Errorf("service '%s' not found", name)
}
