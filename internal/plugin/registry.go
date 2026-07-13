package plugin

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type PluginType string

const (
	PluginTypeStack   PluginType = "stack"
	PluginTypeService PluginType = "service"
)

type Plugin struct {
	Name        string            `yaml:"name"`
	DisplayName string            `yaml:"display_name"`
	Description string            `yaml:"description"`
	Version     string            `yaml:"version"`
	Author      string            `yaml:"author"`
	Category    PluginType        `yaml:"category"`
	Icon        string            `yaml:"icon"`
	Components  map[string]Component `yaml:"components"`
	Preset      map[string]string `yaml:"preset"`
	Provision   string            `yaml:"provision"`
	Defaults    map[string]string `yaml:"defaults,omitempty"`
	Ports       []int             `yaml:"ports,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
}

type Component struct {
	Label       string   `yaml:"label"`
	Type        string   `yaml:"type"`
	Versions    []string `yaml:"versions"`
	Default     string   `yaml:"default"`
	Optional    bool     `yaml:"optional,omitempty"`
	Required    bool     `yaml:"required,omitempty"`
	Description string   `yaml:"description,omitempty"`
}

type Registry struct {
	plugins map[string]*Plugin
}

func NewRegistry() *Registry {
	return &Registry{
		plugins: make(map[string]*Plugin),
	}
}

func (r *Registry) LoadFromDir(dir string) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".yml" && filepath.Ext(entry.Name()) != ".yaml" {
			continue
		}

		path := filepath.Join(dir, entry.Name())
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var plugin Plugin
		if err := yaml.Unmarshal(data, &plugin); err != nil {
			continue
		}

		if plugin.Name != "" {
			r.plugins[plugin.Name] = &plugin
		}
	}

	return nil
}

func (r *Registry) Get(name string) *Plugin {
	return r.plugins[name]
}

func (r *Registry) List() []*Plugin {
	var result []*Plugin
	for _, p := range r.plugins {
		result = append(result, p)
	}
	return result
}

func (r *Registry) ListByType(pluginType PluginType) []*Plugin {
	var result []*Plugin
	for _, p := range r.plugins {
		if p.Category == pluginType {
			result = append(result, p)
		}
	}
	return result
}

func LoadPlugin(path string) (*Plugin, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var plugin Plugin
	if err := yaml.Unmarshal(data, &plugin); err != nil {
		return nil, err
	}

	if plugin.Name == "" {
		return nil, fmt.Errorf("plugin name is required")
	}

	return &plugin, nil
}
