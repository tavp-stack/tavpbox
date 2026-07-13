package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Loader struct {
	pluginDir string
}

func NewLoader(pluginDir string) *Loader {
	return &Loader{pluginDir: pluginDir}
}

func (l *Loader) LoadAll() (*Registry, error) {
	registry := NewRegistry()

	// Load stack plugins
	stackDir := filepath.Join(l.pluginDir, "stacks")
	if err := registry.LoadFromDir(stackDir); err != nil {
		// Non-fatal: directory might not exist
	}

	// Load service plugins
	serviceDir := filepath.Join(l.pluginDir, "services")
	if err := registry.LoadFromDir(serviceDir); err != nil {
		// Non-fatal: directory might not exist
	}

	return registry, nil
}

func (l *Loader) LoadFromFile(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin Plugin
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, fmt.Errorf("invalid plugin YAML: %w", err)
	}

	if plugin.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	return &plugin, nil
}

func (l *Loader) InstallPlugin(path string) error {
	plugin, err := l.LoadFromFile(path)
	if err != nil {
		return err
	}

	var targetDir string
	switch plugin.Category {
	case PluginTypeStack:
		targetDir = filepath.Join(l.pluginDir, "stacks")
	case PluginTypeService:
		targetDir = filepath.Join(l.pluginDir, "services")
	default:
		return fmt.Errorf("unknown plugin category: %s", plugin.Category)
	}

	os.MkdirAll(targetDir, 0755)

	// Copy plugin file
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	targetPath := filepath.Join(targetDir, plugin.Name+".yml")
	return os.WriteFile(targetPath, data, 0644)
}

func (l *Loader) RemovePlugin(name string) error {
	// Try stacks
	stackPath := filepath.Join(l.pluginDir, "stacks", name+".yml")
	if err := os.Remove(stackPath); err == nil {
		return nil
	}

	// Try services
	servicePath := filepath.Join(l.pluginDir, "services", name+".yml")
	if err := os.Remove(servicePath); err == nil {
		return nil
	}

	return fmt.Errorf("plugin '%s' not found", name)
}
