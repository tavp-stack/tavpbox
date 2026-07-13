package stack

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

func (m *Manager) Install(containerName, stackName string, env map[string]string) error {
	script, err := m.getStackScript(stackName)
	if err != nil {
		return err
	}

	// Replace env variables in script
	for key, value := range env {
		script = replaceVar(script, key, value)
	}

	// Push script to container
	tmpFile := filepath.Join(os.TempDir(), "tavpbox-stack-install.sh")
	os.WriteFile(tmpFile, []byte(script), 0755)
	m.client.Push(containerName, tmpFile, "/tmp/stack-install.sh")
	os.Remove(tmpFile)

	// Execute stack installer
	_, err = m.client.ExecNoTTY(containerName, "bash", "-c", "chmod +x /tmp/stack-install.sh && bash /tmp/stack-install.sh")
	if err != nil {
		return fmt.Errorf("stack install failed: %w", err)
	}

	return nil
}

func (m *Manager) getStackScript(name string) (string, error) {
	// Try embedded first
	scriptPath := fmt.Sprintf("scripts/%s.sh", name)
	data, err := embeddedScripts.ReadFile(scriptPath)
	if err == nil {
		return string(data), nil
	}

	// Try user's custom stacks
	home, _ := os.UserHomeDir()
	userPath := filepath.Join(home, ".tavpbox", "plugins", "stacks", name+".sh")
	data, err = os.ReadFile(userPath)
	if err == nil {
		return string(data), nil
	}

	return "", fmt.Errorf("stack '%s' not found", name)
}

func replaceVar(script, key, value string) string {
	return replaceAll(script, fmt.Sprintf("${%s}", key), value)
}

func replaceAll(s, old, new string) string {
	result := ""
	for {
		idx := findIndex(s, old)
		if idx == -1 {
			result += s
			break
		}
		result += s[:idx] + new
		s = s[idx+len(old):]
	}
	return result
}

func findIndex(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
