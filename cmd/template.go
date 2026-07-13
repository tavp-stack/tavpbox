package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"gopkg.in/yaml.v3"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage starter templates",
}

var templateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List available templates",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Available Templates:")
		fmt.Println("────────────────────────────────────────────")
		fmt.Println("  tavp-starter     TAVP Stack starter (PHP + Phalcon + Nginx)")
		fmt.Println("  laravel-starter  Laravel starter (PHP + Nginx + MySQL)")
		fmt.Println("  node-starter     Node.js starter (Express + Nginx)")
		fmt.Println("  python-starter   Python starter (Flask + Nginx)")
		fmt.Println("  blank            Empty container")
		fmt.Println("")
		fmt.Println("Usage: tavpbox create --template=<name>")
		return nil
	},
}

var templates = map[string]string{
	"tavp-starter":    "https://github.com/tavp-stack/tavp-starter.git",
	"laravel-starter": "https://github.com/tavp-stack/laravel-starter.git",
	"node-starter":    "https://github.com/tavp-stack/node-starter.git",
	"python-starter":  "https://github.com/tavp-stack/python-starter.git",
}

func cloneTemplate(templateName, targetDir string) error {
	repoURL, ok := templates[templateName]
	if !ok {
		return fmt.Errorf("template '%s' not found", templateName)
	}

	cmd := exec.Command("git", "clone", "--depth=1", repoURL, targetDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func createFromTemplate(templateName, projectName string) error {
	tmpDir := filepath.Join(os.TempDir(), "tavpbox-template-"+projectName)
	defer os.RemoveAll(tmpDir)

	fmt.Printf("Cloning template '%s'...\n", templateName)
	if err := cloneTemplate(templateName, tmpDir); err != nil {
		return fmt.Errorf("failed to clone template: %w", err)
	}

	// Copy template to current directory
	targetDir := filepath.Join(".", projectName)
	if err := copyDir(tmpDir, targetDir); err != nil {
		return fmt.Errorf("failed to copy template: %w", err)
	}

	// Update .tavpbox.yml with project name
	configPath := filepath.Join(targetDir, ".tavpbox.yml")
	if _, err := os.Stat(configPath); err == nil {
		data, err := os.ReadFile(configPath)
		if err == nil {
			var cfg config.ProjectConfig
			if err := yaml.Unmarshal(data, &cfg); err == nil {
				cfg.Name = projectName
				newData, _ := yaml.Marshal(cfg)
				os.WriteFile(configPath, newData, 0644)
			}
		}
	}

	fmt.Printf("✓ Template '%s' created at %s\n", templateName, targetDir)
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		targetPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(targetPath, data, info.Mode())
	})
}

func init() {
	templateCmd.AddCommand(templateListCmd)
	rootCmd.AddCommand(templateCmd)
}
