package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigDir  = ".tavpbox"
	ConfigFile = "config.yml"
	BoxesDir   = "boxes"
)

type GlobalConfig struct {
	DomainSuffix  string        `yaml:"domain_suffix"`
	DefaultRAM    string        `yaml:"default_ram"`
	DefaultCPU    int           `yaml:"default_cpu"`
	DefaultDistro string        `yaml:"default_distro"`
	Network       NetworkConfig `yaml:"network"`
}

type NetworkConfig struct {
	Bridge string `yaml:"bridge"`
	Subnet string `yaml:"subnet"`
	DNS    string `yaml:"dns"`
}

func HomeDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ConfigDir)
}

func LoadGlobal() (*GlobalConfig, error) {
	path := filepath.Join(HomeDir(), ConfigFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultGlobal(), nil
	}
	var cfg GlobalConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return DefaultGlobal(), nil
	}
	return &cfg, nil
}

func SaveGlobal(cfg *GlobalConfig) error {
	os.MkdirAll(HomeDir(), 0755)
	path := filepath.Join(HomeDir(), ConfigFile)
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func DefaultGlobal() *GlobalConfig {
	return &GlobalConfig{
		DomainSuffix:  "tavp.local",
		DefaultRAM:    "512MB",
		DefaultCPU:    1,
		DefaultDistro: "ubuntu/24.04",
		Network: NetworkConfig{
			Bridge: "lxdbr0",
			Subnet: "10.0.3.0/24",
			DNS:    "10.0.3.1",
		},
	}
}
