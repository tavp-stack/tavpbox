# TAVPBox — Complete Project Summary

> **Versi:** 0.11.2 | **Terakhir diupdate:** 2026-07-18 | **Status:** Stable

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

```
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

## 3. Session History

### Session 2026-07-18: phpMyAdmin + Adminer Fix → ZeroVer 0.11.2

**Duration:** ~3 hours (08:00 - 11:30 WIB)

**What was done:**
- **phpMyAdmin world-writable fix (#7):** Root cause = `/var/www/html` di-mount drvfs (9p) dari `C:\`, `chmod` diabaikan → `config.inc.php` selalu `0777`. Fix: pindahkan config ke `/etc` (non-drvfs, perms `0644` menempel), symlink dari webroot.
- **phpMyAdmin mysqli missing (#8):** `images/php/Containerfile` hanya install `pdo_mysql`, tidak `mysqli`. Fix: tambah `mysqli` ke `docker-php-ext-install`.
- **Adminer support:** Added dedicated nginx config on port 8081 (separate from phpMyAdmin 8080), drvfs permission fix, proxy route.
- **ZeroVer migration:** CHANGELOG/README/WIKI di-update ke `0.11.2` (major=0, patch utama).

**Key Commits:**
- `73b9745` fix: phpMyAdmin world-writable config.inc.php on drvfs/WSL mounts (#7)
- `7ba228a` fix: install mysqli PHP extension in php image (#8)
- `5dfb1be` feat: add proper adminer support with dedicated nginx config (#8 follow-up)
- `...` docs: ZeroVer 0.11.2 changelog + README + WIKI

**Issues:**
- #7 [open→fix] phpMyAdmin world-writable (fixed, commented)
- #8 [open→fix] mysqli extension missing (fixed, commented)
- #4 [open] events.post-start not auto-executed

**Version:** 0.11.2 (ZeroVer: major=0, patch incremented)

### Session 2026-07-17: Podman SSH Fix + v1.11.0 Release

**Duration:** ~2 hours (18:00 - 20:00 WIB)

**What was done:**
- Investigated Podman SSH socket (50312) not listening despite `podman machine list` showing "running"
- **Root cause found:** WSL2 localhost forwarding broken — SSH listened inside WSL but wasn't forwarded to Windows
- **Fix:** Created `.wslconfig` with `networkingMode=mirrored` in `%USERPROFILE%`
- Implemented `EnsureRunning()` auto-fix: detect → stop → start → wait for SSH
- Fixed startup script: MariaDB sleep 2, PHP-FPM sleep 1, nginx retry 3x
- Rebuilt pre-built image `ghcr.io/tavp-stack/tavpbox-php:latest`
- Recreated lula container, verified HTTP 200

**Key Commits:**
- `54a4f5a` feat: auto-fix Podman SSH
- `dd24eaf` fix: comprehensive Podman + nginx fixes
- `a30932f` docs: add v1.11.0 changelog
- `59c598a` docs: add Status Terkini section
- `3062698` docs: add NEXT_STEPS.md
- `d061256` docs: add SESSION_LOG.md

**Release:** v1.11.0 (GitHub + Gitea)

**Issues:**
- #1 [closed] Port binding fix
- #2 [closed] Post-start events + port binding
- #3 [closed] HTTP→HTTPS + Service unavailable
- #4 [created] events.post-start not auto-executed

### Previous Sessions Summary

| Version | Key Features |
|---------|-------------|
| v1.4.x | Pre-built images, service library (15 services, 7 recipes) |
| v1.3.x | Auto-start proxy, startup scripts, service persistence |
| v1.2.x | Lando `.lando.yml` migration, auto-detect webroot |
| v1.1.x | Web panel, tooling, config management |
| v1.0.x | Initial release, embedded Go proxy, auto-HTTPS |

---

## 4. Commands Reference

### Core Commands

```bash
# Initialize from .lando.yml
tavpbox init

# Create container
tavpbox create

# Start (auto-fixes Podman if needed)
tavpbox start

# Stop
tavpbox stop

# Restart
tavpbox restart

# Destroy
tavpbox destroy

# SSH into container
tavpbox ssh

# Show LAN URLs
tavpbox expose
```

### Proxy Commands

```bash
tavpbox proxy:start
tavpbox proxy:stop
tavpbox proxy:status
```

### Tooling Commands

```bash
tavpbox artisan <cmd>      # Laravel artisan
tavpbox composer <cmd>     # Composer
tavpbox npm <cmd>          # NPM
tavpbox node <cmd>         # Node.js
tavpbox php <cmd>          # PHP
```

### Image Commands

```bash
tavpbox image build
tavpbox image push
tavpbox image pull
tavpbox image list
```

### Config Commands

```bash
tavpbox config set <key> <value>
tavpbox config get <key>
tavpbox config list
```

---

## 5. Container Details

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

### Host Port Mapping

| Project | HTTP | Mailpit |
|---------|------|---------|
| lula | 8081 | 34789 |
| tavp-web-id | 41509 | 38439 |
| test-tavp | 33165 | 37863 |

---

## 6. .tavpbox.yml Format

```yaml
name: my-project
recipe: tavp          # tavp, php, laravel, node, python, go
webroot: .            # or public/

services:
  mariadb:
    enabled: true
  redis:
    enabled: true
  mailpit:
    enabled: true
  phpmyadmin:
    enabled: false

env:
  APP_NAME: "My App"
  APP_ENV: local
  APP_DEBUG: "true"
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

## 7. Known Issues & Workarounds

### Issue #4: events.post-start not auto-executed

**Status:** Open

**Problem:** Events in `.tavpbox.yml` under `events.post-start` are not executed after `tavpbox create` or `tavpbox rebuild`.

**Workaround:** Run commands manually via `tavpbox ssh`:
```bash
tavpbox ssh
# Then run the commands manually
```

### WSL2 SSH Port Forwarding

**Status:** Fixed in v1.11.0

**Problem:** Podman SSH socket (50312) doesn't listen on Windows even though WSL VM is running.

**Fix:** `.wslconfig` with `networkingMode=mirrored` (auto-created by tavpbox).

### Podman Machine State Confusion

**Status:** Mitigated

**Problem:** `podman machine list` shows "running" but SSH is broken.

**Workaround:** `EnsureRunning()` auto-detects and auto-restarts machine.

---

## 8. Development Setup

### Prerequisites

- Go 1.21+
- Podman Desktop
- Windows: WSL2 enabled

### Build

```bash
git clone https://github.com/tavp-stack/tavpbox.git
cd tavpbox
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

## 9. Repository Information

### GitHub (Release mirror)

- **URL:** https://github.com/tavp-stack/tavpbox
- **Purpose:** Release binaries, public visibility
- **CI/CD:** GitHub Actions (release.yml, ci.yml)

### Gitea (Development)

- **URL:** https://git.glotama.com/tavp-stack/tavp-box
- **Purpose:** Primary development, issues, wiki
- **API Token:** _(disimpan lokal, TIDAK di repo)_

### Documentation

- **URL:** https://docs.tavp.web.id/guide/tavpbox.html
- **Framework:** VitePress
- **Hosting:** Vercel

---

## 10. API Access (for AI/Session Continuity)

### Gitea API

```bash
# Base URL
https://git.glotama.com/api/v1/repos/tavp-stack/tavp-box

# Auth header
Authorization: token <GITEA_TOKEN>

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

# Get repo info
GET /
```

### Container Access

```bash
# SSH into container
podman exec -it tavp-lula bash

# Run command in container
podman exec tavp-lula <command>

# Access MariaDB
podman exec tavp-lula mariadb -u root luta_app

# Check container status
podman ps -a

# Check port mappings
podman port tavp-lula
```

### Podman Machine

```bash
# Check machine status
podman machine list

# Stop machine
podman machine stop

# Start machine
podman machine start

# Restart (fixes most issues)
podman machine stop && podman machine start
```

---

## 11. Troubleshooting

### Podman not responding

```powershell
# Check if SSH port is listening
netstat -ano | findstr ":50312"

# If not listening, restart machine
podman machine stop
podman machine start

# If still not working, check .wslconfig
cat $env:USERPROFILE\.wslconfig
# Should have: networkingMode=mirrored
```

### Container ports not binding

```bash
# Check container status
podman ps -a

# If stopped, start it
podman start tavp-lula

# If running but ports not bound
podman restart tavp-lula
```

### Service unavailable (502/503)

```bash
# Check nginx inside container
podman exec tavp-lula nginx -t

# Check PHP-FPM
podman exec tavp-lula ps aux | grep php-fpm

# Restart services
podman exec tavp-lula nginx -s reload
```

### Post-start events not executed

```bash
# Manual execution
tavpbox ssh

# Inside container:
mkdir -p storage/{logs,cache,compiled/volt}
chmod -R 777 storage
sed -i 's|root /var/www/html;|root /var/www/html/public;|' /etc/nginx/sites-available/default
nginx -s reload
```

---

## 12. Next Steps

### Priority 1: events.post-start executor (Issue #4)

- Implement automatic execution of `events.post-start` after container starts
- This will fix Laravel apps needing nginx root change, storage permissions, DB users

### Priority 2: Windows Task Scheduler auto-start

```powershell
# Run as admin:
$action = New-ScheduledTaskAction -Execute "podman" -Argument "machine start"
$trigger = New-ScheduledTaskTrigger -AtLogOn
Register-ScheduledTask -Action $action -Trigger $trigger -TaskName "Podman Auto-Start" -RunLevel Highest
```

### Priority 3: Full restart cycle testing

1. Restart Windows
2. Run `tavpbox start`
3. Verify all projects accessible

---

## 13. Token/Cost Tracking

| Metric | Value |
|--------|-------|
| Session tokens | ~232,600 |
| Cache hit rate | ~60% |
| Cost per session | ~$0.26 (Rp5,117) |
| USD/IDR rate | Rp22,000/$1 |

---

*Document generated by AI session on 2026-07-17. For the latest status, check Gitea issues and NEXT_STEPS.md.*
