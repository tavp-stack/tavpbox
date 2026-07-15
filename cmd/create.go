package cmd

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/certs"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/podman"
	"github.com/tavp-stack/tavpbox/internal/proxy"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and start a project container",
	RunE: func(cmd *cobra.Command, args []string) error {
		_, cfg, err := config.FindProject()
		if err != nil {
			return fmt.Errorf(".tavpbox.yml not found. Run: tavpbox init")
		}

		globalCfg, _ := config.LoadGlobal()
		client := podman.New()
		cname := client.ContainerName(cfg.Name)

		fmt.Printf("Creating box '%s' (%s recipe)...\n", cfg.Name, cfg.Recipe)

		image := getImage(cfg, globalCfg)
		ports := getPorts(cfg)

		env := make(map[string]string)
		env["APP_ENV"] = "local"
		for k, v := range cfg.Env {
			env[k] = v
		}

		webroot := cfg.Webroot
		if webroot == "" {
			webroot = "."
		}
		absWebroot, _ := filepath.Abs(webroot)
		volumes := []string{
			fmt.Sprintf("%s:/var/www/html", absWebroot),
		}

		domain := cfg.Name + "." + globalCfg.DomainSuffix

		// 1. Pull image
		fmt.Printf("  [1/4] Pulling image %s...\n", image)
		if err := client.Pull(image); err != nil {
			return fmt.Errorf("pull image: %w", err)
		}

		// 2. Create container
		fmt.Printf("  [2/4] Creating container...\n")
		client.Remove(cname)
		if err := client.Create(cname, image, ports, env, volumes, map[string]string{}); err != nil {
			return fmt.Errorf("create container: %w", err)
		}

		// 3. Start container
		fmt.Printf("  [3/4] Starting container...\n")
		if err := client.Start(cname); err != nil {
			return fmt.Errorf("start container: %w", err)
		}
		time.Sleep(2 * time.Second)

		// 4. Install recipe + services
		fmt.Printf("  [4/4] Installing %s recipe...\n", cfg.Recipe)
		if err := installRecipe(client, cname, cfg); err != nil {
			fmt.Printf("  ⚠ Recipe install warning: %v\n", err)
		}

		for svcName, svcCfg := range cfg.Services {
			if !svcCfg.Enabled {
				continue
			}
			fmt.Printf("  Installing %s...\n", svcName)
			if err := installService(client, cname, svcName); err != nil {
				fmt.Printf("  ⚠ %s install warning: %v\n", svcName, err)
			}
		}

		// Execute Lando build/run commands if present
		if buildCmds, ok := cfg.Env["LANDO_BUILD_CMDS"]; ok && buildCmds != "" {
			fmt.Printf("  Running build commands...\n")
			if _, err := client.Exec(cname, "bash", "-c", buildCmds); err != nil {
				fmt.Printf("  ⚠ Build commands warning: %v\n", err)
			}
		}

		// Get container IP and host port
		ip, _ := client.GetIP(cname)
		hostPort := client.GetHostPort(cname, "80")

		// Ensure proxy is running before adding routes
		ensureProxyRunning()

		// Add proxy route for domain access
		p := proxy.New(80)
		p.AddRoute(domain, "127.0.0.1", hostPort)

		// Add routes for services
		if cfg.Services["mailpit"].Enabled || cfg.Services["mailhog"].Enabled {
			mailpitPort := client.GetHostPort(cname, "8025")
			if mailpitPort > 0 {
				p.AddRoute("mailpit."+domain, "127.0.0.1", mailpitPort)
			}
		}
		if cfg.Services["adminer"].Enabled {
			adminerPort := client.GetHostPort(cname, "8080")
			if adminerPort > 0 {
				p.AddRoute("adminer."+domain, "127.0.0.1", adminerPort)
			}
		}

		// Ensure HTTPS cert exists
		certs.GetWildcardCert("tavp.my.id")

		// Restart proxy to pick up new routes
		restartProxy()

		fmt.Printf("\n✓ Box '%s' created and running!\n", cfg.Name)
		fmt.Printf("  Direct:  http://localhost:%d\n", hostPort)
		fmt.Printf("  HTTP:    http://%s\n", domain)
		fmt.Printf("  HTTPS:   https://%s\n", domain)
		if ip != "" {
			fmt.Printf("  IP:      %s\n", ip)
		}
		fmt.Printf("  SSH:     tavpbox ssh\n")

		return nil
	},
}

func getImage(cfg *config.ProjectConfig, globalCfg *config.GlobalConfig) string {
	if cfg.Image != "" {
		return cfg.Image
	}
	switch cfg.Recipe {
	case "tavp", "php", "laravel":
		return "docker.io/library/ubuntu:24.04"
	case "node":
		return "docker.io/library/node:20-alpine"
	case "go":
		return "docker.io/library/golang:1.22-alpine"
	case "python":
		return "docker.io/library/python:3.12-slim"
	default:
		if globalCfg.DefaultImage != "" {
			return globalCfg.DefaultImage
		}
		return "docker.io/library/ubuntu:24.04"
	}
}

func getPorts(cfg *config.ProjectConfig) []string {
	ports := []string{"80"} // Always map container port 80 to host

	for svcName, svcCfg := range cfg.Services {
		if !svcCfg.Enabled {
			continue
		}
		switch svcName {
		case "mailpit":
			ports = append(ports, "8025", "1025")
		case "phpmyadmin", "adminer":
			ports = append(ports, "8080")
		}
	}

	return ports
}

func installRecipe(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	switch cfg.Recipe {
	case "tavp", "php":
		return installPHPServer(client, cname)
	case "laravel":
		return installLaravel(client, cname)
	case "node":
		return installNode(client, cname)
	case "go":
		return installGo(client, cname)
	case "python":
		return installPython(client, cname)
	default:
		return nil
	}
}

func installPHPServer(client *podman.Client, cname string) error {
	_, err := client.Exec(cname, "bash", "-c", `
apt-get update -qq && apt-get install -y -qq --no-install-recommends \
  nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml \
  php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql \
  php8.3-pgsql php8.3-redis php8.3-sqlite3 php8.3-gd \
  composer curl wget git unzip

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html;
    index index.php index.html;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX

service php8.3-fpm start 2>/dev/null; service nginx start 2>/dev/null
`)
	return err
}

func installLaravel(client *podman.Client, cname string) error {
	_, err := client.Exec(cname, "bash", "-c", `
apt-get update -qq && apt-get install -y -qq --no-install-recommends \
  nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml \
  php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql \
  php8.3-pgsql php8.3-redis php8.3-sqlite3 php8.3-gd \
  composer curl wget git unzip

curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq --no-install-recommends nodejs

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root /var/www/html/public;
    index index.php;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass unix:/run/php/php8.3-fpm.sock;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX

service php8.3-fpm start 2>/dev/null; service nginx start 2>/dev/null
`)
	return err
}

func installNode(client *podman.Client, cname string) error {
	_, err := client.Exec(cname, "bash", "-c", `
apt-get update -qq && apt-get install -y -qq nginx
npm install -g yarn pnpm

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80;
    root /var/www/html;
    location / {
        proxy_pass http://127.0.0.1:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
    }
}
NGINX

systemctl start nginx 2>/dev/null || service nginx start
`)
	return err
}

func installGo(client *podman.Client, cname string) error {
	_, err := client.Exec(cname, "bash", "-c", `apt-get update -qq && apt-get install -y -qq nginx curl`)
	return err
}

func installPython(client *podman.Client, cname string) error {
	_, err := client.Exec(cname, "bash", "-c", `
apt-get update -qq && apt-get install -y -qq nginx python3 python3-pip python3-venv curl

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80;
    root /var/www/html;
    location / {
        proxy_pass http://127.0.0.1:8000;
        proxy_set_header Host $host;
    }
}
NGINX

systemctl start nginx 2>/dev/null || service nginx start
`)
	return err
}

func installService(client *podman.Client, cname, svcName string) error {
	scripts := map[string]string{
		"mariadb": `apt-get install -y -qq mariadb-server mariadb-client
service mariadb start 2>/dev/null || systemctl start mariadb 2>/dev/null || true
mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true`,
		"mysql": `apt-get install -y -qq mysql-server mysql-client
service mysql start 2>/dev/null || systemctl start mysql 2>/dev/null || true`,
		"postgres": `apt-get install -y -qq postgresql postgresql-client
service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true`,
		"redis": `apt-get install -y -qq redis-server
service redis-server start 2>/dev/null || systemctl start redis-server 2>/dev/null || true`,
		"mailpit": `curl -sL https://github.com/axllent/mailpit/releases/latest/download/mailpit_linux_amd64.tar.gz | tar xz -C /usr/local/bin/
nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &`,
		"adminer": `mkdir -p /var/www/html/adminer
curl -sL https://www.adminer.org/latest.php -o /var/www/html/adminer/index.php
curl -sL https://www.adminer.org/download/v5.4.4/designs/haeckel/adminer.css -o /var/www/html/adminer/adminer.css
chmod 644 /var/www/html/adminer/index.php /var/www/html/adminer/adminer.css`,
	}

	if script, ok := scripts[svcName]; ok {
		_, err := client.Exec(cname, "bash", "-c", script)
		return err
	}
	return nil
}
