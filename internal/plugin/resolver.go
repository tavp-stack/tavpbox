package plugin

import (
	"fmt"
	"strings"
)

type Resolver struct {
	registry *Registry
}

func NewResolver(registry *Registry) *Resolver {
	return &Resolver{registry: registry}
}

func (r *Resolver) ResolveComponents(pluginName string, userSelections map[string]string) (map[string]string, error) {
	plugin := r.registry.Get(pluginName)
	if plugin == nil {
		return nil, fmt.Errorf("plugin '%s' not found", pluginName)
	}

	resolved := make(map[string]string)

	// Start with preset defaults
	for key, value := range plugin.Preset {
		resolved[key] = value
	}

	// Override with user selections
	for key, value := range userSelections {
		if component, ok := plugin.Components[key]; ok {
			// Validate version
			if !r.isValidVersion(component, value) {
				return nil, fmt.Errorf("invalid version '%s' for component '%s'. Available: %s",
					value, key, strings.Join(component.Versions, ", "))
			}
			resolved[key] = value
		}
	}

	return resolved, nil
}

func (r *Resolver) isValidVersion(component Component, version string) bool {
	for _, v := range component.Versions {
		if v == version {
			return true
		}
	}
	return false
}

func (r *Resolver) GetPluginWithResolvedVersions(pluginName string, userSelections map[string]string) (*Plugin, map[string]string, error) {
	plugin := r.registry.Get(pluginName)
	if plugin == nil {
		return nil, nil, fmt.Errorf("plugin '%s' not found", pluginName)
	}

	resolved, err := r.ResolveComponents(pluginName, userSelections)
	if err != nil {
		return nil, nil, err
	}

	return plugin, resolved, nil
}
