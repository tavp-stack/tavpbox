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
		if webroot == "" || webroot == "." {
			// Check .lando.yml for webroot override
			wd, _ := os.Getwd()
			landoPath := filepath.Join(wd, ".lando.yml")
			if lando, err := config.ParseLando(landoPath); err == nil && lando.Config.Webroot != "" {
				webroot = lando.Config.Webroot
			}
			if webroot == "" {
				webroot = "."
			}
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

		// Create startup script for auto-restart
		fmt.Printf("  Creating startup script...\n")
		startupScript := buildStartupScript(cfg)
		client.Exec(cname, "bash", "-c", fmt.Sprintf("cat > /usr/local/bin/tavpbox-startup.sh << 'STARTEOF'\n%s\nSTARTEOF\nchmod +x /usr/local/bin/tavpbox-startup.sh", startupScript))

		// Run startup script in background (not blocking)
		client.Exec(cname, "bash", "-c", "nohup /usr/local/bin/tavpbox-startup.sh > /var/log/tavpbox-startup.log 2>&1 &")

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
		if cfg.Services["phpmyadmin"].Enabled {
			// phpMyAdmin accessible via /pma/ path on main app
			p.AddRoute("phpmyadmin."+domain, "127.0.0.1", hostPort)
		}
		if cfg.Services["phpmyadmin"].Enabled {
			// phpMyAdmin runs on port 80 inside container via nginx
			p.AddRoute("phpmyadmin."+domain, "127.0.0.1", hostPort)
		}

		// Ensure HTTPS cert exists
		certs.GetWildcardCert("tavp.my.id")

		// Proxy auto-detects route changes via file watcher — no restart needed

		fmt.Printf("\n✓ Box '%s' created and running!\n", cfg.Name)
		fmt.Printf("  Direct:  http://localhost:%d\n", hostPort)
		fmt.Printf("  HTTP:    http://%s\n", domain)
		fmt.Printf("  HTTPS:   https://%s\n", domain)
		if cfg.Services["mailpit"].Enabled || cfg.Services["mailhog"].Enabled {
			fmt.Printf("  Mailpit: http://mailpit.%s\n", domain)
		}
		if cfg.Services["adminer"].Enabled {
			fmt.Printf("  Adminer: http://adminer.%s\n", domain)
		}
		if cfg.Services["phpmyadmin"].Enabled {
			fmt.Printf("  phpMyAdmin: http://phpmyadmin.%s\n", domain)
		}
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
		return "ghcr.io/tavp-stack/tavpbox-php:latest"
	case "node":
		return "ghcr.io/tavp-stack/tavpbox-node:latest"
	case "go":
		return "ghcr.io/tavp-stack/tavpbox-go:latest"
	case "python":
		return "ghcr.io/tavp-stack/tavpbox-python:latest"
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
		return installPHPServer(client, cname, cfg)
	case "laravel":
		return installLaravel(client, cname, cfg)
	case "node":
		return installNode(client, cname, cfg)
	case "go":
		return installGo(client, cname, cfg)
	case "python":
		return installPython(client, cname, cfg)
	default:
		return nil
	}
}

func installPHPServer(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	webroot := "/var/www/html"
	if cfg.Webroot != "" && cfg.Webroot != "." {
		webroot = "/var/www/html/" + cfg.Webroot
	}

	// Check if packages are already installed (pre-built image)
	_, err := client.Exec(cname, "bash", "-c", "command -v nginx && command -v php-fpm")
	if err == nil {
		// Already installed, just configure and start
		// Install Phalcon if missing
		client.Exec(cname, "bash", "-c", "pecl install phalcon 2>/dev/null && echo 'extension=phalcon.so' > /usr/local/etc/php/conf.d/phalcon.ini || true")
		// Configure nginx with correct webroot
		_, err = client.Exec(cname, "bash", "-c", fmt.Sprintf(`
mkdir -p /run/php

# Create storage symlinks for Laravel/TAVP
mkdir -p /tmp/storage/framework/views /tmp/storage/framework/cache /tmp/storage/framework/sessions /tmp/bootstrap-cache
rm -rf /var/www/html/storage /var/www/html/bootstrap/cache 2>/dev/null || true
ln -sf /tmp/storage /var/www/html/storage
ln -sf /tmp/bootstrap-cache /var/www/html/bootstrap/cache

cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root %s;
    index index.php index.html;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX
php-fpm &
nginx 2>/dev/null || true
`, webroot))
		return err
	}

	// Not pre-built, install from scratch
	_, err = client.Exec(cname, "bash", "-c", `
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql php8.3-gd composer curl wget git unzip
apt-get install -y -qq --no-install-recommends php-pear php8.3-dev gcc make
pecl channel-update pecl.php.net 2>/dev/null
pecl install phalcon 2>/dev/null || true
echo "extension=phalcon.so" > /etc/php/8.3/mods-available/phalcon.ini
phpenmod phalcon 2>/dev/null || true
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y -qq --no-install-recommends nodejs
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

func installLaravel(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	webroot := "/var/www/html"
	if cfg.Webroot != "" && cfg.Webroot != "." {
		webroot = "/var/www/html/" + cfg.Webroot
	}

	// Check if packages are already installed (pre-built image)
	_, err := client.Exec(cname, "bash", "-c", "command -v nginx && command -v php-fpm")
	if err == nil {
		_, err = client.Exec(cname, "bash", "-c", fmt.Sprintf(`
mkdir -p /run/php
cat > /etc/nginx/sites-available/default <<'NGINX'
server {
    listen 80 default_server;
    root %s;
    index index.php;
    location / { try_files $uri $uri/ /index.php?$query_string; }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht { deny all; }
}
NGINX
php-fpm &
nginx 2>/dev/null || true
`, webroot))
		return err
	}

	// Not pre-built, install from scratch
	_, err = client.Exec(cname, "bash", "-c", `
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends nginx php8.3-fpm php8.3-cli php8.3-mbstring php8.3-xml php8.3-curl php8.3-zip php8.3-bcmath php8.3-intl php8.3-mysql php8.3-gd composer curl wget git unzip
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

func installNode(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	// Check if packages are already installed (pre-built image)
	_, err := client.Exec(cname, "bash", "-c", "command -v nginx && command -v node")
	if err == nil {
		_, err = client.Exec(cname, "bash", "-c", `
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
nginx 2>/dev/null || true
`)
		return err
	}

	// Not pre-built, install from scratch
	_, err = client.Exec(cname, "bash", "-c", `
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends nginx
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

func installGo(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	_, err := client.Exec(cname, "bash", "-c", `apt-get update -qq && apt-get install -y -qq nginx curl`)
	return err
}

func installPython(client *podman.Client, cname string, cfg *config.ProjectConfig) error {
	_, err := client.Exec(cname, "bash", "-c", `
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq --no-install-recommends nginx python3 python3-pip python3-venv curl

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
	// Check if service is already installed (pre-built image)
	checkCmd := map[string]string{
		"mariadb": "command -v mysqld",
		"mysql":   "command -v mysqld",
		"redis":   "command -v redis-server",
		"mailpit": "test -f /usr/local/bin/mailpit",
	}
	if cmd, ok := checkCmd[svcName]; ok {
		if _, err := client.Exec(cname, "bash", "-c", cmd); err == nil {
			return nil // already installed
		}
	}

	scripts := map[string]string{
		"mariadb": `export DEBIAN_FRONTEND=noninteractive
apt-get install -y -qq mariadb-server mariadb-client 2>/dev/null
mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
mysqld --user=mysql --datadir=/var/lib/mysql &
sleep 3
mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true`,
		"mysql": `export DEBIAN_FRONTEND=noninteractive
apt-get install -y -qq mysql-server mysql-client 2>/dev/null
mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
mysqld --user=mysql &
sleep 3`,
		"postgres": `export DEBIAN_FRONTEND=noninteractive
apt-get install -y -qq postgresql postgresql-client 2>/dev/null
su - postgres -c "pg_ctlcluster $(pg_lsclusters -h | head -1 | awk '{print $1, $2}') start" 2>/dev/null || true
su - postgres -c "psql -c \"CREATE USER app WITH PASSWORD 'app' CREATEDB;\"" 2>/dev/null || true
su - postgres -c "psql -c \"CREATE DATABASE app OWNER app;\"" 2>/dev/null || true`,
		"redis": `export DEBIAN_FRONTEND=noninteractive
apt-get install -y -qq redis-server 2>/dev/null
redis-server --daemonize yes 2>/dev/null || true`,
		"mailpit": `curl -sL https://github.com/axllent/mailpit/releases/latest/download/mailpit_linux_amd64.tar.gz | tar xz -C /usr/local/bin/
nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &`,
		"adminer": `mkdir -p /var/www/html/adminer
curl -sL https://www.adminer.org/latest.php -o /var/www/html/adminer/index.php
curl -sL https://www.adminer.org/download/v5.4.4/designs/haeckel/adminer.css -o /var/www/html/adminer/adminer.css
chmod 644 /var/www/html/adminer/index.php /var/www/html/adminer/adminer.css`,
		"phpmyadmin": `export DEBIAN_FRONTEND=noninteractive
apt-get install -y -qq phpmyadmin 2>/dev/null
# Symlink ke webroot yang benar (bisa /var/www/html/public/pma untuk Laravel)
mkdir -p /var/www/html/public
ln -sf /usr/share/phpmyadmin /var/www/html/public/pma 2>/dev/null || true
ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true`,
	}

	if script, ok := scripts[svcName]; ok {
		_, err := client.Exec(cname, "bash", "-c", script)
		return err
	}
	return nil
}

// buildStartupScript creates a script that starts all installed services
func buildStartupScript(cfg *config.ProjectConfig) string {
	script := `#!/bin/bash
# TAVPBox auto-start services

# Start MariaDB/MySQL if installed
if command -v mysqld &> /dev/null; then
    mkdir -p /run/mysqld && chown mysql:mysql /run/mysqld
    mysqld --user=mysql --datadir=/var/lib/mysql &
fi

# Start Redis if installed
if command -v redis-server &> /dev/null; then
    redis-server --daemonize yes 2>/dev/null || true
fi

# Start Mailpit if installed
if [ -f /usr/local/bin/mailpit ]; then
    nohup /usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025 > /var/log/mailpit.log 2>&1 &
fi

# Keep container alive
while true; do sleep 3600; done
`
	return script
}
