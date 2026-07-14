package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// ServicePlugin represents a service plugin loaded from YAML
type ServicePlugin struct {
	Name        string   `yaml:"name"`
	DisplayName string   `yaml:"display_name"`
	Description string   `yaml:"description"`
	Category    string   `yaml:"category"`
	Ports       []int    `yaml:"ports,omitempty"`
	Script      string   `yaml:"script,omitempty"`
}

// ServicesDir returns the path to the plugins/services directory
func ServicesDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".tavpbox", "plugins", "services")
}

// LoadPlugins loads all service plugins from ~/.tavpbox/plugins/services/
func LoadPlugins() (map[string]ServicePlugin, error) {
	dir := ServicesDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		// No plugins directory, return empty
		return make(map[string]ServicePlugin), nil
	}

	plugins := make(map[string]ServicePlugin)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		ymlPath := filepath.Join(dir, entry.Name(), "service.yml")
		data, err := os.ReadFile(ymlPath)
		if err != nil {
			continue
		}
		var p ServicePlugin
		if err := yaml.Unmarshal(data, &p); err != nil {
			continue
		}
		p.Script = filepath.Join(dir, entry.Name(), "install.sh")
		plugins[p.Name] = p
	}
	return plugins, nil
}

// GetPluginScript returns the install script path for a plugin
func GetPluginScript(name string) string {
	dir := ServicesDir()
	scriptPath := filepath.Join(dir, name, "install.sh")
	if _, err := os.Stat(scriptPath); err == nil {
		return scriptPath
	}
	return ""
}

// ListPlugins returns all available plugins
func ListPlugins() ([]ServicePlugin, error) {
	plugins, err := LoadPlugins()
	if err != nil {
		return nil, err
	}
	var list []ServicePlugin
	for _, p := range plugins {
		list = append(list, p)
	}
	return list, nil
}

// CreatePluginTemplate creates a template plugin directory
func CreatePluginTemplate(name string) error {
	dir := ServicesDir()
	pluginDir := filepath.Join(dir, name)
	os.MkdirAll(pluginDir, 0755)

	// Create service.yml
	serviceYml := fmt.Sprintf(`name: %s
display_name: "%s"
description: "%s service"
category: custom
ports:
  - 8080
`, name, name, name)
	os.WriteFile(filepath.Join(pluginDir, "service.yml"), []byte(serviceYml), 0644)

	// Create install.sh
	installSh := fmt.Sprintf(`#!/bin/bash
set -e
# TODO: Add installation commands for %s
echo "Installing %s..."
`, name, name)
	os.WriteFile(filepath.Join(pluginDir, "install.sh"), []byte(installSh), 0755)

	return nil
}
