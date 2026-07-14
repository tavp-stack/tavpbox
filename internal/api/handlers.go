package api

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"

	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/library"
	"github.com/tavp-stack/tavpbox/internal/podman"
)

type APIResponse struct {
	OK      bool        `json:"ok"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func jsonOK(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(APIResponse{OK: true, Data: data})
}

func jsonError(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{OK: false, Error: msg})
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	client := podman.New()
	jsonOK(w, map[string]interface{}{
		"status":  "ok",
		"podman":  client.IsAvailable(),
		"version": "dev",
	})
}

// ProjectInfo is the API representation of a project
type ProjectInfo struct {
	Name    string `json:"name"`
	Status  string `json:"status"`
	Recipe  string `json:"recipe"`
	IP      string `json:"ip"`
	Domain  string `json:"domain"`
	Webroot string `json:"webroot"`
	RAM     string `json:"ram"`
	CPU     int    `json:"cpu"`
	Image   string `json:"image"`
}

func handleListProjects(w http.ResponseWriter, r *http.Request) {
	client := podman.New()
	containers, err := client.List()
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}

	globalCfg, _ := config.LoadGlobal()
	var projects []ProjectInfo
	for _, c := range containers {
		name := client.StripPrefix(c.Name)
		status := "stopped"
		if c.Status == "running" || contains(c.Status, "Up") {
			status = "running"
		}
		ip, _ := client.GetIP(c.Name)
		domain := name + "." + globalCfg.DomainSuffix
		projects = append(projects, ProjectInfo{
			Name:   name,
			Status: status,
			IP:     ip,
			Domain: domain,
			Image:  c.Image,
		})
	}

	jsonOK(w, projects)
}

func handleGetProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		jsonError(w, 400, "project name required")
		return
	}

	client := podman.New()
	cname := client.ContainerName(name)
	ip, _ := client.GetIP(cname)
	globalCfg, _ := config.LoadGlobal()

	// Get container info
	containers, _ := client.List()
	var container *podman.ContainerInfo
	for _, c := range containers {
		if c.Name == cname {
			container = &c
			break
		}
	}

	if container == nil {
		jsonError(w, 404, "project not found")
		return
	}

	status := "stopped"
	if container.Status == "running" || contains(container.Status, "Up") {
		status = "running"
	}

	domain := name + "." + globalCfg.DomainSuffix
	jsonOK(w, ProjectInfo{
		Name:   name,
		Status: status,
		IP:     ip,
		Domain: domain,
		Image:  container.Image,
	})
}

type CreateRequest struct {
	Name     string                       `json:"name"`
	Recipe   string                       `json:"recipe"`
	Webroot  string                       `json:"webroot,omitempty"`
	Services map[string]config.ServiceConfig `json:"services,omitempty"`
	RAM      string                       `json:"ram,omitempty"`
	CPU      int                          `json:"cpu,omitempty"`
}

func handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, 400, "invalid request body")
		return
	}

	if req.Name == "" {
		jsonError(w, 400, "name is required")
		return
	}
	if req.Recipe == "" {
		req.Recipe = "tavp"
	}

	// Build services config
	services := req.Services
	if services == nil {
		services = make(map[string]config.ServiceConfig)
		recipe, ok := library.GetRecipe(req.Recipe)
		if ok {
			for _, svc := range recipe.Services {
				services[svc] = config.ServiceConfig{Enabled: true}
			}
		}
	}

	// Get default tooling
	tooling := defaultToolingForRecipe(req.Recipe)

	cfg := &config.ProjectConfig{
		Name:     req.Name,
		Recipe:   req.Recipe,
		Webroot:  req.Webroot,
		Services: services,
		Tooling:  tooling,
		RAM:      req.RAM,
		CPU:      req.CPU,
		Env: map[string]string{
			"APP_ENV": "local",
		},
	}

	// Save to user's home directory (absolute path)
	home, _ := os.UserHomeDir()
	projectDir := filepath.Join(home, req.Name)
	os.MkdirAll(projectDir, 0755)

	configPath := filepath.Join(projectDir, ".tavpbox.yml")
	if err := config.SaveProject(configPath, cfg); err != nil {
		jsonError(w, 500, "save config: "+err.Error())
		return
	}

	globalCfg, _ := config.LoadGlobal()
	domain := req.Name + "." + globalCfg.DomainSuffix

	jsonOK(w, map[string]string{
		"name":   req.Name,
		"recipe": req.Recipe,
		"status": "created",
		"path":   projectDir,
		"domain": domain,
	})
}

func handleStartProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	client := podman.New()
	cname := client.ContainerName(name)

	if err := client.Start(cname); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonOK(w, map[string]string{"status": "started"})
}

func handleStopProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	client := podman.New()
	cname := client.ContainerName(name)

	if err := client.Stop(cname); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonOK(w, map[string]string{"status": "stopped"})
}

func handleRestartProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	client := podman.New()
	cname := client.ContainerName(name)

	if err := client.Restart(cname); err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonOK(w, map[string]string{"status": "restarted"})
}

func handleDestroyProject(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	client := podman.New()
	cname := client.ContainerName(name)

	if err := client.Remove(cname); err != nil {
		jsonError(w, 500, err.Error())
		return
	}

	client.RemoveTraefikRoute(name)
	jsonOK(w, map[string]string{"status": "destroyed"})
}

func handleProjectLogs(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	client := podman.New()
	cname := client.ContainerName(name)

	output, err := client.Logs(cname, 100)
	if err != nil {
		jsonError(w, 500, err.Error())
		return
	}
	jsonOK(w, map[string]string{"logs": output})
}

type RecipeInfo struct {
	Name        string   `json:"name"`
	DisplayName string   `json:"display_name"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	Services    []string `json:"services"`
}

func handleListRecipes(w http.ResponseWriter, r *http.Request) {
	var recipes []RecipeInfo
	for name, recipe := range library.RecipeLibrary {
		recipes = append(recipes, RecipeInfo{
			Name:        name,
			DisplayName: recipe.DisplayName,
			Description: recipe.Description,
			Image:       recipe.Image,
			Services:    recipe.Services,
		})
	}
	jsonOK(w, recipes)
}

type ServiceInfo struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

func handleListServices(w http.ResponseWriter, r *http.Request) {
	var services []ServiceInfo
	for name, svc := range library.ServiceLibrary {
		services = append(services, ServiceInfo{
			Name:        name,
			DisplayName: svc.DisplayName,
			Description: svc.Description,
			Category:    svc.Category,
		})
	}
	jsonOK(w, services)
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstr(s, substr))
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// defaultToolingForRecipe returns default tooling for a recipe (API helper)
func defaultToolingForRecipe(recipe string) map[string]config.ToolingConfig {
	switch recipe {
	case "tavp", "laravel":
		return map[string]config.ToolingConfig{
			"artisan":  {Cmd: "php artisan"},
			"composer": {Cmd: "composer"},
			"npm":      {Cmd: "npm"},
			"npx":      {Cmd: "npx"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php artisan test"},
		}
	case "php":
		return map[string]config.ToolingConfig{
			"composer": {Cmd: "composer"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php vendor/bin/phpunit"},
		}
	case "node":
		return map[string]config.ToolingConfig{
			"npm":  {Cmd: "npm"},
			"npx":  {Cmd: "npx"},
			"yarn": {Cmd: "yarn"},
			"pnpm": {Cmd: "pnpm"},
			"node": {Cmd: "node"},
		}
	case "go":
		return map[string]config.ToolingConfig{
			"go": {Cmd: "go"},
		}
	case "python":
		return map[string]config.ToolingConfig{
			"python": {Cmd: "python3"},
			"pip":    {Cmd: "pip3"},
			"pytest": {Cmd: "python3 -m pytest"},
		}
	default:
		return nil
	}
}
