package config

import (
	"fmt"
	"os"
	"path/filepath"
)

func ValidateProject(cfg *ProjectConfig) error {
	if cfg.Name == "" {
		return fmt.Errorf("name is required")
	}

	if cfg.Stack == "" {
		cfg.Stack = "tavp"
	}

	if cfg.Webroot == "" {
		cfg.Webroot = "."
	}

	if cfg.RAM == "" {
		cfg.RAM = "512MB"
	}

	if cfg.CPU == 0 {
		cfg.CPU = 1
	}

	if cfg.Distro == "" {
		cfg.Distro = "ubuntu/24.04"
	}

	// Validate webroot exists
	if cfg.Webroot != "." {
		absPath, err := filepath.Abs(cfg.Webroot)
		if err != nil {
			return fmt.Errorf("invalid webroot path: %w", err)
		}
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return fmt.Errorf("webroot path does not exist: %s", absPath)
		}
	}

	// Validate services
	validServices := map[string]bool{
		"mariadb":        true,
		"mysql":          true,
		"postgresql":     true,
		"postgres":       true,
		"redis":          true,
		"memcached":      true,
		"mailpit":        true,
		"elasticsearch":  true,
		"meilisearch":    true,
		"minio":          true,
		"phpmyadmin":     true,
		"adminer":        true,
		"rabbitmq":       true,
	}

	for _, svc := range cfg.Services {
		if !validServices[svc] {
			return fmt.Errorf("unknown service: %s", svc)
		}
	}

	// Validate stack
	validStacks := map[string]bool{
		"tavp":     true,
		"laravel":  true,
		"symfony":  true,
		"wordpress": true,
		"node":     true,
		"python":   true,
		"go":       true,
		"ruby":     true,
		"blank":    true,
	}

	if !validStacks[cfg.Stack] {
		return fmt.Errorf("unknown stack: %s", cfg.Stack)
	}

	return nil
}
