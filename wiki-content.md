# TAVPBox — Complete Project Summary

> **Versi:** 0.11.2 (ZeroVer) | **Terakhir diupdate:** 2026-07-18 14:30 WIB | **Status:** Active Development

---

## 1. Project Overview

TAVPBox adalah local development environment seperti [Lando](https://lando.dev), tapi pakai [Podman](https://podman.io) (bukan Docker). Dirancang untuk developer yang ingin:

- Zero-config local dev (seperti Lando)
- RAM-efficient (Podman rootless, daemonless)
- Auto-HTTPS via Let's Encrypt wildcard cert
- LAN access untuk testing di device lain
- Lando `.lando.yml` migration

### Tech Stack

| Component | Technology |
|-----------|------------|
| Runtime | Podman (rootless, daemonless) |
| Reverse Proxy | Embedded Go proxy (HTTP :80 + HTTPS :443) |
| HTTPS | Let's Encrypt wildcard cert via lego ACME DNS-01 |
| CLI | Go (single binary, ~10MB) |
| Pre-built Image | `ghcr.io/tavp-stack/tavpbox-php:latest` |
| Domain | `*.tavp.my.id` (Cloudflare DNS) |
| Platform | Windows, macOS, Linux |

---

## 2. Architecture

```plaintext
┌─────────────────────────────────────────────────────┐
│  Host (Windows/macOS/Linux)                         │
│                                                     │
│  ┌─────────────┐  ┌──────────────────────────────┐  │
│  │  tavpbox    │  │  Embedded Go Proxy            │  │
│  │  (CLI)      │  │  - HTTP :80                   │  │
│  │             │  │  - HTTPS :443                 │  │
│  │  Commands:  │  │  - Route table (JSON)         │  │
│  │  - create   │  │  - Auto-HTTPS (lego)          │  │
│  │  - start    │  └──────────────────────────────┘  │
│  │  - stop     │                                   │
│  │  - destroy  │  ┌──────────────────────────────┐  │
│  │  - ssh      │  │  Podman Machine (WSL2)        │  │
│  │  - expose   │  │  - SSH port: 50312            │  │
│  │  - tooling  │  │  - Containers: tavp-*         │  │
│  └─────────────┘  └──────────────────────────────┘  │
└─────────────────────────────────────────────────────┘
```

### Key Files

| File | Location | Purpose |
|------|----------|---------|
| Binary | `~\AppData\Local\tavpbox\tavpbox.exe` | Installed binary |
| Routes | `~/.tavpbox/proxy/routes.json` | Proxy route mappings |
| LAN Ports | `~/.tavpbox/lan-ports.json` | Port assignments (8081-8999) |
| Certs | `~/.tavpbox/certs/` | Wildcard cert files |
| Volumes | `~/.tavpbox/volumes/<project>/` | MariaDB/PostgreSQL data |
| WSL Config | `%USERPROFILE%\.wslconfig` | WSL2 networking mode |

---

## 3. Version History (ZeroVer)

| Version | Date | Key Changes |
|---------|------|-------------|
| 0.11.2 | 2026-07-18 | phpMyAdmin world-writable fix (#7), mysqli fix (#8), adminer support, ZeroVer migration |
| 0.11.1 | 2026-07-17 | Mailpit process dies silently fix (#5, #6) |
| 0.11.0 | 2026-07-17 | WSL2 SSH fix, EnsureRunning(), startup script reliability |
| 0.10.x | 2026-07-16 | LAN access, auto-start Podman, port binding fixes |
| 0.4.x | 2026-07-15 | Pre-built images, service library (15 services, 7 recipes) |
| 0.3.x | 2026-07-14 | Auto-start proxy, startup scripts, service persistence |
| 0.2.x | 2026-07-13 | Lando migration, auto-detect webroot |
| 0.1.x | 2026-07-12 | Web panel, tooling, config management |
| 0.0.1 | 2026-07-10 | Initial release, embedded Go proxy, auto-HTTPS |

---

## 4. Session History

### Session 2026-07-18: phpMyAdmin + Adminer Fix → ZeroVer 0.11.2

**Duration:** ~6 hours (08:00 - 14:30 WIB)

**What was done:**
- **phpMyAdmin world-writable fix (#7):** Root cause = `/var/www/html` di-mount drvfs (9p) dari `C:\`, `chmod` diabaikan → `config.inc.php` selalu `0777`. Fix: pindahkan config ke `/etc` (non-drvfs, perms `0644` menempel), symlink dari webroot.
- **phpMyAdmin mysqli missing (#8):** `images/php/Containerfile` hanya install `pdo_mysql`, tidak `mysqli`. Fix: tambah `mysqli` ke `docker-php-ext-install`.
- **Adminer support:** Added dedicated nginx config on port 8081 (separate from phpMyAdmin 8080), drvfs permission fix, proxy route.
- **ZeroVer migration:** CHANGELOG/README/WIKI di-update ke `0.11.2` (major=0, patch utama).
- **Webroot fix (partial):** TAVP stack projects (Lando migration) punya `index.php` di `public/` bukan root. Fix: ubah `.tavpbox.yml` dari `webroot: .` ke `webroot: public`. Masih bermasalah (HTTP 403/404).
- **Proxy routes fix:** Rewrite `routes.json` dengan format benar, restart proxy.

**Key Commits:**
- `73b9745` fix: phpMyAdmin world-writable config.inc.php on drvfs/WSL mounts (#7)
- `7ba228a` fix: install mysqli PHP extension in php image (#8)
- `5dfb1be` feat: add proper adminer support with dedicated nginx config
- `675e505` feat: update Adminer CSS to v5.5.0 (haeckel design)
- `ca707a5` fix: expose correct ports for phpMyAdmin (8080) and adminer (8081)
- `6398e3a` docs: ZeroVer 0.11.2 migration
- `841843b` chore: remove tavpbox.exe from tracking + update .gitignore
- `0f945ba` feat: add fix-nginx.sh utility script

**Issues:**
- #7 [closed] phpMyAdmin world-writable
- #8 [closed] mysqli extension missing
- #9 [created] TAVP stack webroot issue (Lando migration)
- #4 [commented] events.post-start not auto-executed

**Release:** 0.11.2 (GitHub + Gitea)

### Previous Sessions Summary

| Version | Key Features |
|---------|-------------|
| 0.4.x | Pre-built images, service library (15 services, 7 recipes) |
| 0.3.x | Auto-start proxy, startup scripts, service persistence |
| 0.2.x | Lando migration, auto-detect webroot |
| 0.1.x | Web panel, tooling, config management |
| 0.0.x | Initial release, embedded Go proxy, auto-HTTPS |

---

## 5. Known Issues & Workarounds

### Issue #9: TAVP stack webroot issue (Lando migration)

**Status:** Open

**Problem:** TAVP stack projects (migrasi Lando) punya `index.php` di `public/` bukan root. TAVPBox generate nginx config hardcoded `root /var/www/html` → 403/404.

**Workaround:** Ubah `.tavpbox.yml` dari `webroot: .` ke `webroot: public`. Tapi masih bermasalah.

**Solution:** Update `cmd/create.go` agar auto-detect `public/index.php` dan set nginx root ke `/var/www/html/public`.

### Issue #4: events.post-start not auto-executed

**Status:** Open

**Problem:** Events in `.tavpbox.yml` under `events.post-start` are not executed after `tavpbox create` or `tavpbox rebuild`.

**Workaround:** Run commands manually via `tavpbox ssh`:
```bash
tavpbox ssh
# Then run the commands manually
```

### WSL2 SSH Port Forwarding

**Status:** Fixed in 0.11.0

**Problem:** Podman SSH socket (50312) doesn't listen on Windows even though WSL VM is running.

**Fix:** `.wslconfig` with `networkingMode=mirrored` (auto-created by tavpbox).

### Podman Machine State Confusion

**Status:** Mitigated

**Problem:** `podman machine list` shows "running" but SSH is broken.

**Workaround:** `EnsureRunning()` auto-detects and auto-restarts machine.

---

## 6. Development Setup

### Prerequisites

- Go 1.21+
- Podman Desktop
- Windows: WSL2 enabled

### Build

```bash
git clone https://git.glotama.com/tavp-stack/tavp-box.git
cd tavp-box
go build -o tavpbox.exe .
```

### Cross-compile

```bash
make cross
```

### Test

```bash
go test ./...
```

---

## 7. Container Details

### Pre-built Image: `ghcr.io/tavp-stack/tavpbox-php:latest`

| Component | Version |
|-----------|---------|
| Base | php:8.3-fpm |
| Nginx | Latest |
| MariaDB | Latest |
| Redis | Latest |
| Mailpit | v1.21.1 |
| Node.js | 20.x |
| Composer | Latest |
| Phalcon | Latest |

### Container Ports

| Port | Service |
|------|---------|
| 80 | Nginx (HTTP) |
| 9000 | PHP-FPM |
| 3306 | MariaDB |
| 6379 | Redis |
| 8025 | Mailpit Web UI |
| 1025 | Mailpit SMTP |
| 8080 | phpMyAdmin (nginx) |
| 8081 | Adminer (nginx) |

---

## 8. Active Projects

| Project | Container | Status | URL |
|---------|-----------|--------|-----|
| tavp-web-id | tavp-tavp-web-id | ⚠️ HTTP 403 | https://tavp-web-id.tavp.my.id/ |
| lula | tavp-lula | ⚠️ HTTP 404 | https://lula.tavp.my.id/ |
| test-tavp | tavp-test-tavp | ✅ HTTP 200 | https://test-tavp.tavp.my.id/ |

---

## 9. API Access (for AI/Session Continuity)

### Gitea API

```bash
# Base URL
https://git.glotama.com/api/v1/repos/tavp-stack/tavp-box

# Auth header
Authorization: token <YOUR_GITEA_TOKEN>

# List issues
GET /issues?state=all

# Get issue
GET /issues/{id}

# Create issue
POST /issues

# Update issue
PATCH /issues/{id}

# List releases
GET /releases
```

### Container Access

```bash
# SSH into container
podman exec -it tavp-<project> bash

# Run command in container
podman exec tavp-<project> <command>

# Access MariaDB
podman exec tavp-<project> mariadb -u root <database>

# Check container status
podman ps -a
```

### Podman Machine

```bash
# Check machine status
podman machine list

# Stop machine
podman machine stop

# Start machine
podman machine start
```

---

## 10. Next Steps (TODO)

### Priority 1: Fix webroot issue (Issue #9)

- Update `cmd/create.go` agar auto-detect `public/index.php`
- Set nginx root ke `/var/www/html/public` jika ditemukan
- Fix `lula` (HTTP 404) dan `tavp-web-id` (HTTP 403)

### Priority 2: Implement events.post-start (Issue #4)

- Execute commands from `.tavpbox.yml` after container starts
- This will fix Laravel apps needing nginx root change, storage permissions

### Priority 3: Windows Task Scheduler auto-start

```powershell
# Run as admin:
$action = New-ScheduledTaskAction -Execute "podman" -Argument "machine start"
$trigger = New-ScheduledTaskTrigger -AtLogOn
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "Podman Auto-Start" -RunLevel Highest
```

### Priority 4: Full restart cycle testing

1. Restart Windows
2. Run `tavpbox start`
3. Verify all projects accessible

---

## 11. Environment Setup

### Windows

```powershell
# Install Podman Desktop
# https://podman-desktop.io

# Enable WSL2
wsl --install

# Create .wslconfig
# %USERPROFILE%\.wslconfig
[wsl2]
networkingMode=mirrored
```

### macOS/Linux

```bash
# Install Podman
brew install podman  # macOS
sudo apt install podman  # Linux
```

---

## 12. Troubleshooting

### Podman not responding

```powershell
# Check if SSH port is listening
netstat -ano | findstr ":50312"

# If not listening, restart machine
podman machine stop
podman machine start
```

### Container ports not binding

```bash
# Check container status
podman ps -a

# If stopped, start it
podman start tavp-<project>
```

### Service unavailable (502/503)

```bash
# Check nginx inside container
podman exec tavp-<project> nginx -t

# Check PHP-FPM
podman exec tavp-<project> ps aux | grep php-fpm
```

### phpMyAdmin world-writable error

```bash
# Fix: symlink config to non-drvfs path
podman exec tavp-<project> bash -c 'cp /var/www/html/pma/config.inc.php /etc/pma-config.inc.php && chmod 0644 /etc/pma-config.inc.php && rm -f /var/www/html/pma/config.inc.php && ln -sf /etc/pma-config.inc.php /var/www/html/pma/config.inc.php'
```

### nginx 403 Forbidden (TAVP/Laravel projects)

```powershell
# Fix nginx config via base64
$config = @'
server {
    listen 80 default_server;
    root /var/www/html/public;
    index index.php index.html;
    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }
    location ~ \.php$ {
        fastcgi_pass 127.0.0.1:9000;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        include fastcgi_params;
    }
    location ~ /\.ht {
        deny all;
    }
}
'@
$bytes = [System.Text.Encoding]::UTF8.GetBytes($config)
$b64 = [Convert]::ToBase64String($bytes)
podman exec tavp-<project> bash -c "echo $b64 | base64 -d > /etc/nginx/sites-available/default && nginx -t && nginx -s reload"
```

---

## 13. Repository Information

### Gitea (Development)

- **URL:** https://git.glotama.com/tavp-stack/tavp-box
- **Purpose:** Primary development, issues, wiki
- **API Token:** _(disimpan lokal, TIDAK di repo)_

### GitHub (Release Mirror)

- **URL:** https://github.com/tavp-stack/tavpbox
- **Purpose:** Release binaries, public visibility
- **CI/CD:** GitHub Actions (release.yml, ci.yml)

### Documentation

- **URL:** https://docs.tavp.web.id/guide/tavpbox.html
- **Framework:** VitePress
- **Hosting:** Vercel

---

## 14. .tavpbox.yml Format

```yaml
name: my-project
recipe: tavp          # tavp, php, laravel, node, python, go
webroot: .            # or public/ for TAVP/Laravel

services:
  mariadb:
    enabled: true
  redis:
    enabled: true
  mailpit:
    enabled: true
  phpmyadmin:
    enabled: false
  adminer:
    enabled: false

env:
  APP_NAME: "My App"
  APP_ENV: local
  DB_DATABASE: mydb
  DB_USERNAME: user
  DB_PASSWORD: pass

tooling:
  php:
    cmd: php
  composer:
    cmd: composer

events:
  post-start:
    - mkdir -p storage/{logs,cache}
    - chmod -R 777 storage

ram: 512MB
cpu: 1
```

---

## 15. Commands Reference

```bash
# Lifecycle
tavpbox init              # Initialize project
tavpbox create            # Create container
tavpbox start             # Start container
tavpbox stop              # Stop container
tavpbox restart           # Restart container
tavpbox destroy           # Destroy container
tavpbox rebuild           # Destroy and recreate

# Monitoring
tavpbox list              # List all containers
tavpbox info              # Show project details
tavpbox logs              # Show container logs

# Tooling
tavpbox ssh               # SSH into container
tavpbox artisan [args]    # Run php artisan
tavpbox composer [args]   # Run composer
tavpbox npm [args]        # Run npm

# Panel
tavpbox panel             # Start web panel
tavpbox panel:stop        # Stop panel

# Proxy
tavpbox proxy:start       # Start reverse proxy
tavpbox proxy:stop        # Stop reverse proxy
tavpbox proxy:status      # Show proxy status
```

---

*Document generated by AI session on 2026-07-18. For the latest status, check Gitea issues and NEXT_STEPS.md.*
