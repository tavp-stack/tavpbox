package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a project in the current folder",
	RunE: func(cmd *cobra.Command, args []string) error {
		cwd, _ := os.Getwd()

		// Check if .tavpbox.yml exists
		_, _, err := config.FindProject()
		if err == nil {
			fmt.Println("Project already initialized.")
			return nil
		}

		// Check for .lando.yml — auto-convert if found
		landoPath := filepath.Join(cwd, ".lando.yml")
		if _, err := os.Stat(landoPath); err == nil {
			fmt.Println("✓ Found .lando.yml — auto-converting...")
			lando, err := config.ParseLando(landoPath)
			if err != nil {
				return fmt.Errorf("parse .lando.yml: %w", err)
			}
			cfg := config.ConvertLandoToTavpbox(lando)

			// Force webroot from .lando.yml if present
			if lando.Config.Webroot != "" {
				cfg.Webroot = lando.Config.Webroot
			}

			if err := config.SaveProject(".tavpbox.yml", cfg); err != nil {
				return fmt.Errorf("save config: %w", err)
			}
			globalCfg, _ := config.LoadGlobal()
			domain := cfg.Name + "." + globalCfg.DomainSuffix
			fmt.Printf("\n✓ Migrated from Lando!\n")
			fmt.Printf("  File:     .tavpbox.yml (from .lando.yml)\n")
			fmt.Printf("  Recipe:   %s\n", cfg.Recipe)
			fmt.Printf("  Webroot:  %s\n", cfg.Webroot)
			fmt.Printf("  Services: ")
			first := true
			for svc, sCfg := range cfg.Services {
				if sCfg.Enabled {
					if !first { fmt.Print(", ") }
					fmt.Print(svc)
					first = false
				}
			}
			fmt.Println()
			fmt.Printf("  Domain:   http://%s\n", domain)
			fmt.Printf("\n  Next: tavpbox create\n")
			return nil
		}

		reader := bufio.NewReader(os.Stdin)

		// Check Podman
		client := podman.New()
		if !client.IsAvailable() {
			return fmt.Errorf("podman not found. Install Podman Desktop from https://podman-desktop.io")
		}
		fmt.Println("✓ Podman ready")

		// Project name
		defaultName := filepath.Base(cwd)
		fmt.Printf("\nProject name [%s]: ", defaultName)
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			name = defaultName
		}

		// Recipe
		recipes := []string{"tavp", "laravel", "php", "node", "go", "python", "blank"}
		fmt.Println("\nRecipe:")
		for i, r := range recipes {
			mark := " "
			if r == "tavp" {
				mark = ">"
			}
			fmt.Printf("  %s [%d] %s\n", mark, i+1, r)
		}
		fmt.Print("Select [1]: ")
		recipeInput, _ := reader.ReadString('\n')
		recipeInput = strings.TrimSpace(recipeInput)
		recipe := "tavp"
		if recipeInput != "" {
			if idx := atoi(recipeInput); idx >= 1 && idx <= len(recipes) {
				recipe = recipes[idx-1]
			} else {
				recipe = recipeInput
			}
		}

		// Services
		allServices := []string{"mariadb", "mysql", "postgres", "mongodb", "redis", "memcached", "mailpit", "mailhog", "phpmyadmin", "adminer", "elasticsearch", "rabbitmq"}
		defaultServices := "mariadb redis mailpit"
		if recipe == "laravel" {
			defaultServices = "mariadb redis mailpit"
		}
		fmt.Printf("\nServices (comma/space separated):\n")
		fmt.Printf("  Available: %s\n", strings.Join(allServices, ", "))
		fmt.Printf("  Default [%s]: ", defaultServices)
		svcInput, _ := reader.ReadString('\n')
		svcInput = strings.TrimSpace(svcInput)
		if svcInput == "" {
			svcInput = defaultServices
		}
		svcInput = strings.ReplaceAll(svcInput, ",", " ")
		services := strings.Fields(svcInput)

		// Webroot
		fmt.Print("\nWebroot [public]: ")
		webrootInput, _ := reader.ReadString('\n')
		webrootInput = strings.TrimSpace(webrootInput)
		if webrootInput == "" {
			webrootInput = "public"
		}

		// RAM
		fmt.Print("RAM limit [512MB]: ")
		ramInput, _ := reader.ReadString('\n')
		ramInput = strings.TrimSpace(ramInput)
		if ramInput == "" {
			ramInput = "512MB"
		}

		// CPU
		fmt.Print("CPU cores [1]: ")
		cpuInput, _ := reader.ReadString('\n')
		cpuInput = strings.TrimSpace(cpuInput)
		cpu := 1
		if cpuInput != "" {
			cpu = atoi(cpuInput)
		}

		// Build services config
		servicesCfg := make(map[string]config.ServiceConfig)
		for _, svc := range services {
			servicesCfg[svc] = config.ServiceConfig{Enabled: true}
		}

		// Build config
		cfg := &config.ProjectConfig{
			Name:     name,
			Recipe:   recipe,
			Webroot:  webrootInput,
			Services: servicesCfg,
			Tooling:  defaultTooling(recipe),
			RAM:      ramInput,
			CPU:      cpu,
			Env: map[string]string{
				"APP_ENV": "local",
			},
		}

		// Save config
		if err := config.SaveProject(".tavpbox.yml", cfg); err != nil {
			return fmt.Errorf("save config: %w", err)
		}

		globalCfg, _ := config.LoadGlobal()
		domain := name + "." + globalCfg.DomainSuffix
		fmt.Printf("\n✓ Project '%s' initialized!\n", name)
		fmt.Printf("  File: .tavpbox.yml\n")
		fmt.Printf("  URL:  http://%s\n", domain)
		fmt.Printf("\n  Next: tavpbox create\n")
		return nil
	},
}

func atoi(s string) int {
	n := 0
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0')
		} else {
			return 0
		}
	}
	return n
}
