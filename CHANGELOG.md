# Changelog

## 0.11.2

### Fix (Critical)
- phpMyAdmin: Wrong permissions on configuration file (world writable) on WSL/drvfs mounts (#7)
- phpMyAdmin: mysqli extension missing in PHP 8.3 image (#8)
- Adminer: added dedicated nginx config on port 8081 with drvfs permission fix

### Fix (Details)
- `cmd/create.go`: Move phpMyAdmin `config.inc.php` to `/etc` (non-drvfs) + symlink to fix world-writable check
- `images/php/Containerfile`: Added `mysqli` to `docker-php-ext-install` (phpMyAdmin needs it)
- `cmd/create.go`: Adminer now has proper nginx config on port 8081 (separate from phpMyAdmin 8080)
- `cmd/create.go`: Apply same drvfs permission fix (symlink to `/etc`) for Adminer

## 0.11.1

### Fix (Critical)
- Mailpit process dies silently after container restart (#5, #6)
- Added health check in startup loop (every 10s)
- Auto-restart mailpit if process dies
- Applied to both start.sh and buildStartupScript()

## 0.11.0

### Fix (Critical)
- WSL2 SSH port forwarding: Created `.wslconfig` with `networkingMode=mirrored`
- Root cause: Podman SSH socket (50312) listened inside WSL but wasn't forwarded to Windows
- `podman machine list` showed "running" but `podman ps` failed

### Feature
- `EnsureRunning()`: Auto-detects and auto-fixes Podman SSH (stop + start machine)
- Startup script: Added delays between services (MariaDB 2s, PHP-FPM 1s)
- Startup script: Nginx retry loop (3 attempts) on container start
- Pre-built image rebuilt with fixed startup script

### Fix (LAN Access)
- Container ports bind to `0.0.0.0` (not `127.0.0.1`) for LAN access
- Fixed ports 8081-8999 per project with `lan-ports.json`
- `tavpbox expose` command shows LAN URLs

### Fix (Various)
- `isProxyRunning()` uses port 80 check (not PID-based)
- `restartProxy()` kills by port (netstat + taskkill on Windows)
- `configureAdminServices()` auto-configure nginx for phpmyadmin/adminer

## 0.10.0-0.10.14

### Feature
- LAN access: `tavpbox expose` command with fixed ports 8081-8999
- Auto-start Podman machine on `tavpbox start`
- Multiple EnsureRunning iterations (SSH port check, auto-restart, etc.)

### Fix
- Caddy and Traefik code fully removed from codebase
- Port binding: `0.0.0.0` prefix for all interfaces
- Proxy detection: port-based instead of PID-based

## 0.4.2

### Feature
- Pre-built images dengan semua service (MariaDB, Redis, Mailpit)
- installService() skip jika service udah ada di image
- Start script pakai php-fpm8.2 langsung
- Gak ada lagi apt-get timeout

## 0.4.1

### Feature
- Pakai pre-built images (ghcr.io/tavp-stack/tavpbox-*)
- Skip apt-get jika packages udah ada di image
- Recipe install instant jika pakai pre-built image
- Fallback ke apt-get jika image gak pre-built

## 0.4.0

### Feature
- Pre-built image system (Containerfiles for PHP, Node, Go, Python)
- tavpbox image build — build custom image dari container
- tavpbox image push — push ke registry
- tavpbox image pull — pull dari registry
- tavpbox image list — list local images
- Makefile targets untuk build/push official images
- Official images: ghcr.io/tavp-stack/tavpbox-{php,node,go,python}

## 0.3.4

### Optimization
- DEBIAN_FRONTEND=noninteractive untuk semua apt-get commands
- Recipe install lebih cepat (kurangi interactive prompts)
- Service install dioptimasi (mariadb, mysql, postgres, redis)

## 0.3.3

### Fix
- Auto-start proxy on all commands (PersistentPreRun)
- Proxy otomatis jalan saat `tavpbox create`, `start`, `restart`, `info`, dll
- Developer gak perlu start proxy manual

## 0.3.2

### Fix
- Auto-start services on container restart
- Startup script (/usr/local/bin/tavpbox-startup.sh) created after install
- Services persist across host restart (nginx, php-fpm, mariadb, redis, mailpit)
- Container gak perlu re-install services setelah host restart

## 0.3.1

### Fix
- Service install tanpa systemctl (langsung start process)
- MariaDB: mysqld --user=mysql
- MySQL: mysqld --user=mysql
- PostgreSQL: pg_ctlcluster
- Redis: redis-server --daemonize yes
- Semua service work tanpa systemd

## 0.3.0

### Feature
- Wildcard cert (*.tavp.my.id) embedded di binary
- Auto-extract ke ~/.tavpbox/certs/ saat HTTPS request pertama
- Developer gak perlu run `tavpbox setup`
- Admin: `tavpbox setup` + `make cert` + `make release` untuk renew

## 0.2.2

### Fix
- setup: restart proxy setelah generate cert (cert baru langsung dipake)
- Proxy auto-reload routes.json kalau file berubah

## 0.2.1

### Fix
- Phalcon + Node.js diinstall otomatis untuk tavp recipe
- Nginx fastcgi_pass pakai unix socket (bukan TCP 9000)
- Sync recipes.go dan create.go

## 0.2.0

### Optimization
- Recipe install 2x faster (--no-install-recommends, removed php-pear php8.3-dev gcc make)
- Proxy auto-reload routes (watch routes.json every 2s)
- Proxy auto-start before adding routes
- Webroot auto-detect from current directory

## 0.1.0

### Architecture Change
- LXC/LXD → Podman (rootless, daemonless)
- Traefik/Caddy → Embedded Go proxy (zero dependency, ~10MB RAM)
- Self-signed cert → Let's Encrypt wildcard cert
- Full Lando migration support

### New Features
- Embedded Go reverse proxy (HTTP :80 + HTTPS :443)
- Wildcard cert `*.tavp.my.id` via Let's Encrypt (ACME DNS-01)
- Dynamic tooling commands (artisan, composer, npm, etc.)
- Web panel (`tavpbox panel`) with Tailwind + Alpine.js
- Full Lando migration (services, tooling, env, proxy, build/run)
- Auto-route update on rebuild
- Config management (`tavpbox config set/get/list`)
- Multi-platform (Windows, macOS, Linux)

### Commands Added
- `tavpbox tooling` — List tooling commands
- `tavpbox panel` — Start web panel
- `tavpbox panel:stop` — Stop panel
- `tavpbox proxy:start` — Start reverse proxy
- `tavpbox proxy:stop` — Stop reverse proxy
- `tavpbox proxy:status` — Show proxy status
- `tavpbox config set/get/list` — Configuration management
- `tavpbox setup` — Install dependencies + generate cert

### Files Changed
- `internal/config/lando.go` — Lando YAML parser + converter
- `internal/proxy/proxy.go` — Embedded Go reverse proxy
- `internal/certs/certs.go` — Let's Encrypt ACME via lego
- `internal/podman/client.go` — Podman wrapper
- `cmd/create.go` — Container creation + recipe install
- `cmd/tooling.go` — Dynamic tooling commands
- `cmd/panel.go` — Web panel server
- `cmd/proxy.go` — Proxy management
- `cmd/config.go` — Configuration management
- `cmd/setup.go` — Dependencies + cert setup
- `internal/api/` — REST API + embedded panel

## 0.0.1

### Initial Release
- LXC container management
- TUI wizard for init and create
- Multi-stack support (TAVP, Laravel, Node.js, Python, Blank)
- Service plugins (MariaDB, Redis, PostgreSQL, Mailpit, phpMyAdmin)
- Auto-domain (*.tavp.local)
- Plugin system (YAML-based)
- Custom tooling commands
- Image management
- Snapshot system
- Cross-platform (Linux, macOS, Windows/WSL2)
