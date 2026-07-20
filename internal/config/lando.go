package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// LandoConfig represents a .lando.yml file
type LandoConfig struct {
	Name      string                   `yaml:"name"`
	Recipe    string                   `yaml:"recipe"`
	Config    LandoRecipeConfig        `yaml:"config"`
	Excludes  []string                 `yaml:"excludes"`
	Services  map[string]LandoService  `yaml:"services"`
	Proxy     map[string][]string      `yaml:"proxy"`
	Tooling   map[string]LandoTooling  `yaml:"tooling"`
	Events    map[string][]LandoEvent  `yaml:"events"`
}

// LandoEvent represents a single event entry like "- appserver: composer install"
type LandoEvent struct {
	Service string
	Command string
}

func (e *LandoEvent) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind == yaml.MappingNode && len(value.Content) == 2 {
		e.Service = value.Content[0].Value
		e.Command = value.Content[1].Value
		return nil
	}
	// Fallback: treat as plain string
	e.Command = value.Value
	return nil
}

type LandoRecipeConfig struct {
	Webroot    string `yaml:"webroot"`
	PHP        string `yaml:"php"`
	Database   string `yaml:"database"`
	Xdebug     bool   `yaml:"xdebug"`
}

type LandoService struct {
	Type       string                 `yaml:"type"`
	Webroot    string                 `yaml:"webroot"`
	Xdebug     bool                   `yaml:"xdebug"`
	Portforward interface{}          `yaml:"portforward"`
	Creds      map[string]string      `yaml:"creds"`
	Build      []string               `yaml:"build"`
	Run        []string               `yaml:"run"`
	MailFrom   []string               `yaml:"mailFrom"`
	Overrides  LandoOverrides         `yaml:"overrides"`
}

type LandoOverrides struct {
	Environment map[string]string `yaml:"environment"`
	Tmpfs       []string          `yaml:"tmpfs"`
	Ports       []string          `yaml:"ports"`
}

type LandoTooling struct {
	Service     string `yaml:"service"`
	Cmd         string `yaml:"cmd"`
	User        string `yaml:"user"`
	Description string `yaml:"description"`
}

// ParseLando parses a .lando.yml file and returns a LandoConfig
func ParseLando(path string) (*LandoConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	cfg := &LandoConfig{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// ConvertLandoToTavpbox converts a LandoConfig to a ProjectConfig
func ConvertLandoToTavpbox(lando *LandoConfig) *ProjectConfig {
	cfg := &ProjectConfig{
		Name:     lando.Name,
		Webroot:  lando.Config.Webroot,
		Services: make(map[string]ServiceConfig),
		Tooling:  make(map[string]ToolingConfig),
		Env:      make(map[string]string),
	}

	// Detect recipe from services
	cfg.Recipe = detectRecipe(lando)

	// Map services
	mapServices(lando, cfg)

	// Map tooling
	mapTooling(lando, cfg)

	// Map environment variables
	mapEnvironment(lando, cfg)

	// Set defaults
	if cfg.Webroot == "" {
		cfg.Webroot = "public"
	}
	if cfg.TZ == "" {
		cfg.TZ = "Asia/Jakarta"
	}
	if cfg.RAM == "" {
		cfg.RAM = "512MB"
	}
	if cfg.CPU == 0 {
		cfg.CPU = 1
	}

	return cfg
}

func detectRecipe(lando *LandoConfig) string {
	// Check recipe field
	switch lando.Recipe {
	case "lamp", "lemp":
		return "laravel" // Most Lando PHP projects are Laravel
	case "laravel":
		return "laravel"
	case "mean", "MERN":
		return "node"
	case "python":
		return "python"
	case "go":
		return "go"
	}

	// Detect from services
	for _, svc := range lando.Services {
		if strings.Contains(svc.Type, "php") {
			return "laravel"
		}
		if strings.Contains(svc.Type, "node") {
			return "node"
		}
		if strings.Contains(svc.Type, "python") {
			return "python"
		}
	}

	// Default
	if lando.Config.PHP != "" {
		return "laravel"
	}

	return "tavp"
}

func mapServices(lando *LandoConfig, cfg *ProjectConfig) {
	for _, svc := range lando.Services {
		switch {
		// Database services
		case strings.Contains(svc.Type, "mariadb") || strings.Contains(svc.Type, "mysql"):
			cfg.Services["mariadb"] = ServiceConfig{Enabled: true}
			// Map credentials
			if svc.Creds != nil {
				if user, ok := svc.Creds["user"]; ok {
					cfg.Env["DB_USERNAME"] = user
				}
				if pass, ok := svc.Creds["password"]; ok {
					cfg.Env["DB_PASSWORD"] = pass
				}
				if db, ok := svc.Creds["database"]; ok {
					cfg.Env["DB_DATABASE"] = db
				}
			}

		case strings.Contains(svc.Type, "postgres"):
			cfg.Services["postgres"] = ServiceConfig{Enabled: true}

		case strings.Contains(svc.Type, "mongodb"):
			cfg.Services["mongodb"] = ServiceConfig{Enabled: true}

		// Cache services
		case strings.Contains(svc.Type, "redis"):
			cfg.Services["redis"] = ServiceConfig{Enabled: true}

		case strings.Contains(svc.Type, "memcached"):
			cfg.Services["memcached"] = ServiceConfig{Enabled: true}

		// Mail services
		case strings.Contains(svc.Type, "mailpit"):
			cfg.Services["mailpit"] = ServiceConfig{Enabled: true}

		case strings.Contains(svc.Type, "mailhog"):
			cfg.Services["mailhog"] = ServiceConfig{Enabled: true}

		// Admin services
		case strings.Contains(svc.Type, "phpmyadmin") || strings.Contains(svc.Type, "pma"):
			cfg.Services["phpmyadmin"] = ServiceConfig{Enabled: true}

		case strings.Contains(svc.Type, "adminer"):
			cfg.Services["adminer"] = ServiceConfig{Enabled: true}

		// Search services
		case strings.Contains(svc.Type, "elasticsearch") || strings.Contains(svc.Type, "elastic"):
			cfg.Services["elasticsearch"] = ServiceConfig{Enabled: true}

		// Queue services
		case strings.Contains(svc.Type, "rabbitmq"):
			cfg.Services["rabbitmq"] = ServiceConfig{Enabled: true}

		case strings.Contains(svc.Type, "beanstalkd"):
			cfg.Services["beanstalkd"] = ServiceConfig{Enabled: true}
		}

		// Store build/run commands for later execution
		if len(svc.Build) > 0 || len(svc.Run) > 0 {
			// These will be stored as tooling or env
			if cfg.Env["LANDO_BUILD_CMDS"] == "" {
				cmds := append(svc.Build, svc.Run...)
				cfg.Env["LANDO_BUILD_CMDS"] = strings.Join(cmds, " && ")
			}
		}

		// Map environment variables from overrides
		if svc.Overrides.Environment != nil {
			for k, v := range svc.Overrides.Environment {
				cfg.Env[k] = v
			}
		}
	}

	// Set database defaults
	if _, ok := cfg.Services["mariadb"]; ok {
		if _, ok := cfg.Env["DB_HOST"]; !ok {
			cfg.Env["DB_HOST"] = "localhost"
		}
		if _, ok := cfg.Env["DB_PORT"]; !ok {
			cfg.Env["DB_PORT"] = "3306"
		}
		if _, ok := cfg.Env["DB_DATABASE"]; !ok {
			cfg.Env["DB_DATABASE"] = "app"
		}
		if _, ok := cfg.Env["DB_USERNAME"]; !ok {
			cfg.Env["DB_USERNAME"] = "app"
		}
		if _, ok := cfg.Env["DB_PASSWORD"]; !ok {
			cfg.Env["DB_PASSWORD"] = "app"
		}
	}

	// Set mail defaults
	if _, ok := cfg.Services["mailpit"]; ok {
		if _, ok := cfg.Env["MAIL_HOST"]; !ok {
			cfg.Env["MAIL_HOST"] = "localhost"
		}
		if _, ok := cfg.Env["MAIL_PORT"]; !ok {
			cfg.Env["MAIL_PORT"] = "1025"
		}
	}

	// Set Redis defaults
	if _, ok := cfg.Services["redis"]; ok {
		if _, ok := cfg.Env["REDIS_HOST"]; !ok {
			cfg.Env["REDIS_HOST"] = "localhost"
		}
		if _, ok := cfg.Env["REDIS_PORT"]; !ok {
			cfg.Env["REDIS_PORT"] = "6379"
		}
	}
}

func mapTooling(lando *LandoConfig, cfg *ProjectConfig) {
	for _, tool := range lando.Tooling {
		cfg.Tooling[tool.Cmd] = ToolingConfig{
			Cmd: tool.Cmd,
		}
	}

	// Add default tooling based on recipe if not already set
	defaults := defaultTooling(cfg.Recipe)
	for name, tool := range defaults {
		if _, ok := cfg.Tooling[name]; !ok {
			cfg.Tooling[name] = tool
		}
	}
}

func defaultTooling(recipe string) map[string]ToolingConfig {
	switch recipe {
	case "tavp", "laravel":
		return map[string]ToolingConfig{
			"artisan":  {Cmd: "php artisan"},
			"composer": {Cmd: "composer"},
			"npm":      {Cmd: "npm"},
			"npx":      {Cmd: "npx"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php artisan test"},
		}
	case "php":
		return map[string]ToolingConfig{
			"composer": {Cmd: "composer"},
			"php":      {Cmd: "php"},
			"test":     {Cmd: "php vendor/bin/phpunit"},
		}
	case "node":
		return map[string]ToolingConfig{
			"npm":  {Cmd: "npm"},
			"npx":  {Cmd: "npx"},
			"yarn": {Cmd: "yarn"},
			"pnpm": {Cmd: "pnpm"},
			"node": {Cmd: "node"},
		}
	case "go":
		return map[string]ToolingConfig{
			"go": {Cmd: "go"},
		}
	case "python":
		return map[string]ToolingConfig{
			"python": {Cmd: "python3"},
			"pip":    {Cmd: "pip3"},
			"pytest": {Cmd: "python3 -m pytest"},
		}
	default:
		return nil
	}
}

func mapEnvironment(lando *LandoConfig, cfg *ProjectConfig) {
	// Set APP_ENV if not set
	if _, ok := cfg.Env["APP_ENV"]; !ok {
		cfg.Env["APP_ENV"] = "local"
	}

	// Set APP_URL based on proxy
	for _, domains := range lando.Proxy {
		if len(domains) > 0 {
			domain := domains[0]
			// Convert lndo.site to tavp.my.id
			domain = strings.Replace(domain, ".lndo.site", ".tavp.my.id", 1)
			cfg.Env["APP_URL"] = "http://" + domain
			break
		}
	}
}

// GetLandoProxyDomains returns the proxy domains converted to tavp.my.id
func GetLandoProxyDomains(lando *LandoConfig) map[string]string {
	domains := make(map[string]string)
	for svcName, landoDomains := range lando.Proxy {
		for _, domain := range landoDomains {
			// Convert lndo.site to tavp.my.id
			newDomain := strings.Replace(domain, ".lndo.site", ".tavp.my.id", 1)
			domains[svcName] = newDomain
		}
	}
	return domains
}

// GetLandoBuildCommands returns the build commands for a service
func GetLandoBuildCommands(lando *LandoConfig, serviceName string) []string {
	if svc, ok := lando.Services[serviceName]; ok {
		return svc.Build
	}
	return nil
}

// GetLandoRunCommands returns the run commands for a service
func GetLandoRunCommands(lando *LandoConfig, serviceName string) []string {
	if svc, ok := lando.Services[serviceName]; ok {
		return svc.Run
	}
	return nil
}

// GetLandoPostStartCommands returns the post-start commands
func GetLandoPostStartCommands(lando *LandoConfig) []string {
	if events, ok := lando.Events["post-start"]; ok {
		var cmds []string
		for _, ev := range events {
			cmds = append(cmds, ev.Command)
		}
		return cmds
	}
	return nil
}

// FindLandoProject finds a .lando.yml file and converts it
func FindLandoProject() (string, *ProjectConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}

	landoPath := filepath.Join(wd, ".lando.yml")
	if _, err := os.Stat(landoPath); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("no .lando.yml found")
	}

	lando, err := ParseLando(landoPath)
	if err != nil {
		return "", nil, fmt.Errorf("parse .lando.yml: %w", err)
	}

	cfg := ConvertLandoToTavpbox(lando)
	return landoPath, cfg, nil
}
