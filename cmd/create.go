package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"
	"github.com/tavp-stack/tavpbox/internal/config"
	"github.com/tavp-stack/tavpbox/internal/lxd"
	"github.com/tavp-stack/tavpbox/internal/network"
	"github.com/tavp-stack/tavpbox/internal/service"
	"github.com/tavp-stack/tavpbox/internal/stack"
	"gopkg.in/yaml.v3"
)

var createCmd = &cobra.Command{
	Use:   "create [tavpbox.yml]",
	Short: "Create a new development box",
	RunE:  runCreate,
}

func runCreate(cmd *cobra.Command, args []string) error {
	var projectCfg *config.ProjectConfig

	if len(args) > 0 {
		data, err := os.ReadFile(args[0])
		if err != nil {
			return err
		}
		projectCfg = &config.ProjectConfig{}
		if err := yaml.Unmarshal(data, projectCfg); err != nil {
			return fmt.Errorf("invalid YAML: %w", err)
		}
	} else {
		// Try to find .tavpbox.yml in current directory
		_, cfg, err := config.FindProject()
		if err == nil {
			projectCfg = cfg
		} else {
			// Run TUI wizard
			projectCfg, err = runCreateTUI()
			if err != nil {
				return err
			}
		}
	}

	return createBox(projectCfg)
}

func createBox(projectCfg *config.ProjectConfig) error {
	client := lxd.New()
	containerName := client.ContainerName(projectCfg.Name)

	globalCfg, _ := config.LoadGlobal()

	distro := projectCfg.Distro
	if distro == "" {
		distro = globalCfg.DefaultDistro
	}

	ram := projectCfg.RAM
	if ram == "" {
		ram = globalCfg.DefaultRAM
	}
	cpu := projectCfg.CPU
	if cpu == 0 {
		cpu = globalCfg.DefaultCPU
	}

	fmt.Printf("Creating box '%s'...\n", projectCfg.Name)

	fmt.Printf("  [1/5] Creating container (%s, RAM: %s, CPU: %d)...\n", distro, ram, cpu)
	if err := client.Create(containerName, distro, ram, cpu); err != nil {
		return err
	}

	if projectCfg.Webroot != "" {
		absPath, _ := filepath.Abs(projectCfg.Webroot)
		fmt.Printf("  [2/5] Mapping %s → /var/www/html...\n", absPath)
		os.MkdirAll(absPath, 0755)
		if err := client.MapHostDir(containerName, absPath, "/var/www/html"); err != nil {
			return fmt.Errorf("map host dir: %w", err)
		}
	}

	fmt.Printf("  [3/5] Installing stack '%s'...\n", projectCfg.Stack)
	stackMgr := stack.NewManager(client)
	if err := stackMgr.Install(containerName, projectCfg.Stack, nil); err != nil {
		fmt.Printf("  ⚠ Stack install warning: %v\n", err)
		fmt.Printf("  → You can install manually: tavpbox ssh %s\n", projectCfg.Name)
	}

	fmt.Printf("  [4/5] Installing services: %s...\n", strings.Join(projectCfg.Services, ", "))
	svcMgr := service.NewManager(client)
	if err := svcMgr.InstallAll(containerName, projectCfg.Services); err != nil {
		fmt.Printf("  ⚠ Service install warning: %v\n", err)
	}

	// Inject environment variables
	if len(projectCfg.Env) > 0 {
		fmt.Printf("  [4.5/5] Injecting environment variables...\n")
		envScript := "#!/bin/bash\n"
		for key, value := range projectCfg.Env {
			envScript += fmt.Sprintf("echo 'export %s=\"%s\"' >> /etc/environment\n", key, value)
		}
		tmpEnvFile := "/tmp/tavpbox-env.sh"
		os.WriteFile(tmpEnvFile, []byte(envScript), 0755)
		client.Push(containerName, tmpEnvFile, "/tmp/env-setup.sh")
		os.Remove(tmpEnvFile)
		client.ExecNoTTY(containerName, "bash", "-c", "chmod +x /tmp/env-setup.sh && bash /tmp/env-setup.sh")
	}

	fmt.Printf("  [5/5] Configuring networking...\n")
	domain := projectCfg.Name + "." + globalCfg.DomainSuffix
	projectCfg.Domain = domain
	config.SaveProject(".", projectCfg)

	// Setup auto-domain via dnsmasq (non-fatal)
	ip, _ := client.GetIP(containerName)
	if ip != "" {
		if err := network.AddDnsmasqEntry(projectCfg.Name, ip); err != nil {
			fmt.Printf("  ⚠ DNS setup warning: %v\n", err)
		}
		if err := network.AddCaddyRoute(domain, ip, 80); err != nil {
			fmt.Printf("  ⚠ Caddy setup warning: %v\n", err)
		}

		// Setup mailpit subdomain if mailpit is installed
		for _, svc := range projectCfg.Services {
			if svc == "mailpit" {
				mailDomain := "mail." + domain
				network.AddCaddyRoute(mailDomain, ip, 8025)
			}
		}
	}

	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║  ✓  Box '%s' created successfully!                         ║
╠══════════════════════════════════════════════════════════════╣
║                                                              ║
║  App  : http://%s
║                                                              ║
║  Commands:                                                   ║
║    tavpbox ssh %s            Enter box               ║
║    tavpbox start %s          Start box               ║
║    tavpbox stop %s           Stop box                ║
║    tavpbox destroy %s        Delete box              ║
╚══════════════════════════════════════════════════════════════╝
`, projectCfg.Name, domain, projectCfg.Name, projectCfg.Name, projectCfg.Name, projectCfg.Name)

	return nil
}

func installStack(client *lxd.Client, container, stack string) {
	scripts := map[string]string{
		"tavp":    tavpStackScript,
		"laravel": laravelStackScript,
		"node":    nodeStackScript,
		"python":  pythonStackScript,
		"blank":   "#!/bin/bash\necho 'Blank stack'",
	}

	script, ok := scripts[stack]
	if !ok {
		script = scripts["blank"]
	}

	tmpFile := "/tmp/tavpbox-stack.sh"
	os.WriteFile(tmpFile, []byte(script), 0755)
	client.Push(container, tmpFile, "/tmp/stack-install.sh")
	os.Remove(tmpFile)

	client.ExecNoTTY(container, "bash", "-c", "chmod +x /tmp/stack-install.sh && bash /tmp/stack-install.sh")
}

func installService(client *lxd.Client, container, service string) {
	scripts := map[string]string{
		"mariadb":    mariadbScript,
		"redis":      redisScript,
		"postgres":   postgresScript,
		"mailpit":    mailpitScript,
		"phpmyadmin": phpmyadminScript,
	}

	script, ok := scripts[service]
	if !ok {
		return
	}

	tmpFile := "/tmp/tavpbox-svc.sh"
	os.WriteFile(tmpFile, []byte(script), 0755)
	client.Push(container, tmpFile, "/tmp/svc-install.sh")
	os.Remove(tmpFile)

	client.ExecNoTTY(container, "bash", "-c", "chmod +x /tmp/svc-install.sh && bash /tmp/svc-install.sh")
}

// ─── TUI ───

type createModel struct {
	step          int
	name          string
	selectedStack string
	services      []string
	selectedSvc   map[string]bool
	stacks        []string
	cursor        int
	result        *config.ProjectConfig
}

func runCreateTUI() (*config.ProjectConfig, error) {
	m := createModel{
		step:        1,
		selectedSvc: make(map[string]bool),
		services:    []string{"mariadb", "redis", "postgres", "mailpit", "phpmyadmin"},
		stacks:      []string{"tavp", "laravel", "node", "python", "blank"},
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	result, err := p.Run()
	if err != nil {
		return nil, err
	}

	model := result.(createModel)
	if model.result == nil {
		return nil, fmt.Errorf("cancelled")
	}
	return model.result, nil
}

func (m createModel) Init() tea.Cmd { return nil }

func (m createModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			maxItems := len(m.stacks)
			if m.step == 4 {
				maxItems = len(m.services)
			}
			if m.cursor < maxItems-1 {
				m.cursor++
			}

		case " ":
			if m.step == 4 {
				svc := m.services[m.cursor]
				m.selectedSvc[svc] = !m.selectedSvc[svc]
			}

		case "enter":
			switch m.step {
			case 1:
				m.step = 2
				m.cursor = 0
				return m, nil
			case 2:
				m.selectedStack = m.stacks[m.cursor]
				m.step = 3
				m.cursor = 0
				return m, nil
			case 3:
				m.step = 4
				m.cursor = 0
				return m, nil
			case 4:
				var svcs []string
				for s, sel := range m.selectedSvc {
					if sel {
						svcs = append(svcs, s)
					}
				}
				m.result = &config.ProjectConfig{
					Name:     m.name,
					Stack:    m.selectedStack,
					Services: svcs,
					Webroot:  ".",
				}
				return m, tea.Quit
			}
		}
	}
	return m, nil
}

func (m createModel) View() string {
	s := titleStyle.Render(fmt.Sprintf("⚡ Create Box (step %d/4)", m.step))
	s += "\n\n"

	switch m.step {
	case 1:
		s += "Box name:\n\n"
		s += "  [Type name and press Enter]\n"
		s += "  (becomes: <name>.tavp.local)\n"

	case 2:
		s += "Select stack:\n\n"
		for i, st := range m.stacks {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			s += fmt.Sprintf("  %s %s\n", cursor, st)
		}
		s += "\n  ↑↓ navigate · enter select"

	case 3:
		s += "Select services (space to toggle):\n\n"
		for i, svc := range m.services {
			cursor := " "
			if i == m.cursor {
				cursor = ">"
			}
			check := " "
			if m.selectedSvc[svc] {
				check = "✓"
			}
			s += fmt.Sprintf("  %s [%s] %s\n", cursor, check, svc)
		}
		s += "\n  ↑↓ navigate · space toggle · enter confirm"

	case 4:
		s += "Confirm:\n\n"
		s += fmt.Sprintf("  Name: %s\n", m.name)
		s += fmt.Sprintf("  Stack: %s\n", m.selectedStack)
		s += fmt.Sprintf("  Services: %s\n", strings.Join(m.services, ", "))
		s += "\n  enter to create"
	}

	return docStyle.Render(s)
}

// ─── Stack Scripts ───

var tavpStackScript = `#!/bin/bash
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y software-properties-common
add-apt-repository -y ppa:ondrej/php
apt-get update
apt-get install -y php8.3-fpm php8.3-cli php8.3-curl php8.3-mbstring php8.3-xml php8.3-zip php8.3-bcmath php8.3-intl php8.3-gd php8.3-sqlite3 php8.3-pgsql php8.3-mysql php8.3-opcache php8.3-redis nginx composer curl git
cat > /etc/nginx/nginx.conf <<'NGINX'
worker_processes auto;
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 100M;
    server {
        listen 80 default_server;
        root /var/www/html/public;
        index index.php index.html;
        location / { try_files $uri $uri/ /index.php?$query_string; }
        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include fastcgi_params;
        }
    }
}
NGINX
systemctl enable php8.3-fpm nginx
systemctl start php8.3-fpm nginx
echo "TAVP stack installed!"
`

var laravelStackScript = `#!/bin/bash
set -e
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y software-properties-common
add-apt-repository -y ppa:ondrej/php
apt-get update
apt-get install -y php8.3-fpm php8.3-cli php8.3-curl php8.3-mbstring php8.3-xml php8.3-zip php8.3-bcmath php8.3-intl php8.3-gd php8.3-sqlite3 php8.3-pgsql php8.3-mysql php8.3-opcache php8.3-redis nginx composer curl git
cat > /etc/nginx/nginx.conf <<'NGINX'
worker_processes auto;
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    sendfile on;
    keepalive_timeout 65;
    client_max_body_size 100M;
    server {
        listen 80 default_server;
        root /var/www/html/public;
        index index.php;
        location / { try_files $uri $uri/ /index.php?$query_string; }
        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            include fastcgi_params;
        }
    }
}
NGINX
systemctl enable php8.3-fpm nginx
systemctl start php8.3-fpm nginx
echo "Laravel stack installed!"
`

var nodeStackScript = `#!/bin/bash
set -e
curl -fsSL https://deb.nodesource.com/setup_20.x | bash -
apt-get install -y nodejs nginx
npm install -g yarn pnpm
cat > /etc/nginx/nginx.conf <<'NGINX'
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    server {
        listen 80;
        root /var/www/html;
        index index.html;
        location / {
            proxy_pass http://127.0.0.1:3000;
            proxy_http_version 1.1;
            proxy_set_header Upgrade $http_upgrade;
            proxy_set_header Connection 'upgrade';
            proxy_set_header Host $host;
        }
    }
}
NGINX
systemctl enable nginx
systemctl start nginx
echo "Node.js stack installed!"
`

var pythonStackScript = `#!/bin/bash
set -e
apt-get update
apt-get install -y python3 python3-pip python3-venv nginx
cat > /etc/nginx/nginx.conf <<'NGINX'
events { worker_connections 1024; }
http {
    include /etc/nginx/mime.types;
    server {
        listen 80;
        root /var/www/html;
        index index.html;
        location / {
            proxy_pass http://127.0.0.1:8000;
            proxy_set_header Host $host;
        }
    }
}
NGINX
systemctl enable nginx
systemctl start nginx
echo "Python stack installed!"
`

// ─── Service Scripts ───

var mariadbScript = `#!/bin/bash
set -e
apt-get update
apt-get install -y mariadb-server mariadb-client
mysql_install_db --user=root --datadir=/var/lib/mysql 2>/dev/null || true
service mariadb start 2>/dev/null || systemctl start mariadb 2>/dev/null || true
mysql -u root -e "CREATE DATABASE IF NOT EXISTS app; CREATE USER IF NOT EXISTS 'app'@'localhost' IDENTIFIED BY 'app'; GRANT ALL ON app.* TO 'app'@'localhost'; FLUSH PRIVILEGES;" 2>/dev/null || true
echo "MariaDB installed!"
`

var redisScript = `#!/bin/bash
set -e
apt-get update
apt-get install -y redis-server
systemctl start redis-server 2>/dev/null || service redis-server start 2>/dev/null || true
echo "Redis installed!"
`

var postgresScript = `#!/bin/bash
set -e
apt-get update
apt-get install -y postgresql postgresql-client
service postgresql start 2>/dev/null || systemctl start postgresql 2>/dev/null || true
su - postgres -c "psql -c \"CREATE USER app WITH PASSWORD 'app' CREATEDB;\"" 2>/dev/null || true
su - postgres -c "psql -c \"CREATE DATABASE app OWNER app;\"" 2>/dev/null || true
echo "PostgreSQL installed!"
`

var mailpitScript = `#!/bin/bash
set -e
curl -sL https://github.com/axllent/mailpit/releases/latest/download/mailpit_linux_amd64.tar.gz | tar xz -C /usr/local/bin/
cat > /etc/systemd/system/mailpit.service <<EOF
[Unit]
Description=Mailpit
After=network.target
[Service]
ExecStart=/usr/local/bin/mailpit --listen 0.0.0.0:8025 --smtp 0.0.0.0:1025
Restart=always
[Install]
WantedBy=multi-user.target
EOF
systemctl daemon-reload
systemctl enable mailpit
systemctl start mailpit
echo "Mailpit installed!"
`

var phpmyadminScript = `#!/bin/bash
set -e
apt-get update
apt-get install -y phpmyadmin
ln -sf /usr/share/phpmyadmin /var/www/html/pma 2>/dev/null || true
echo "phpMyAdmin installed!"
`
