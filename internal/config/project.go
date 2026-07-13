package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type ProjectConfig struct {
	Name        string            `yaml:"name"`
	Description string            `yaml:"description,omitempty"`
	Stack       string            `yaml:"stack"`
	Services    []string          `yaml:"services,omitempty"`
	Webroot     string            `yaml:"webroot,omitempty"`
	RAM         string            `yaml:"ram,omitempty"`
	CPU         int               `yaml:"cpu,omitempty"`
	Distro      string            `yaml:"distro,omitempty"`
	Env         map[string]string `yaml:"env,omitempty"`
	Domain      string            `yaml:"domain,omitempty"`
	Tooling     map[string]Tooling `yaml:"tooling,omitempty"`
}

type Tooling struct {
	Cmd string `yaml:"cmd"`
}

func LoadProject(dir string) (*ProjectConfig, error) {
	path := filepath.Join(dir, ".tavpbox.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg ProjectConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveProject(dir string, cfg *ProjectConfig) error {
	path := filepath.Join(dir, ".tavpbox.yml")
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func FindProject() (string, *ProjectConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	for {
		cfg, err := LoadProject(dir)
		if err == nil {
			return dir, cfg, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", nil, os.ErrNotExist
}
