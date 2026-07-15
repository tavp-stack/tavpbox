# TAVPBox

> **Lando Dockerless** — Local development environment seperti [Lando](https://lando.dev), tapi pakai [Podman](https://podman.io) (bukan Docker).

```
┌─────────────────────────────────────────────────────────┐
│  TAVPBox                                                │
├──────────────────┬──────────────────────────────────────┤
│ Runtime          │ Podman (rootless, daemonless)        │
│ Reverse Proxy    │ Embedded Go proxy (HTTP + HTTPS)     │
│ HTTPS            │ Let's Encrypt wildcard cert          │
│ RAM / container  │ ~50-80MB                             │
│ 20 project       │ ~1.5GB (vs Docker ~3.2GB)           │
│ Auto domain      │ *.tavp.my.id                         │
│ Config file      │ .tavpbox.yml / .lando.yml            │
│ Platform         │ Windows, macOS, Linux                │
│ CLI language     │ Go (single binary)                   │
│ Web UI           │ Built-in panel (Tailwind + Alpine)   │
│ Lando migration  │ Full .lando.yml support              │
│ License          │ MIT                                  │
└──────────────────┴──────────────────────────────────────┘
```

---

## Install

### Option 1: Download Binary

Download from [GitHub Releases](https://github.com/tavp-stack/tavpbox/releases):

| Platform | File |
|----------|------|
| Windows | `tavpbox-windows-amd64.exe` |
| macOS (Intel) | `tavpbox-darwin-amd64` |
| macOS (M1/M2) | `tavpbox-darwin-arm64` |
| Linux (x64) | `tavpbox-linux-amd64` |
| Linux (ARM) | `tavpbox-linux-arm64` |

Add to PATH:
```powershell
# Windows
move tavpbox.exe C:\Users\<you>\AppData\Local\tavpbox\

# macOS/Linux
sudo mv tavpbox /usr/local/bin/
```

### Option 2: Build from Source

```bash
git clone https://github.com/tavp-stack/tavpbox.git
cd tavpbox
go build -o tavpbox .
```

### Option 3: Go Install

```bash
go install github.com/tavp-stack/tavpbox@latest
```

---

## Quick Start

```powershell
# 1. Init project
cd ~/projects/my-app
tavpbox init

# 2. Create container
tavpbox create

# 3. Open browser
# https://my-app.tavp.my.id
```

### Migrasi dari Lando

```powershell
cd ~/lando-project
tavpbox create
# https://project.tavp.my.id → jalan!
```

---

## Semua Commands

### Lifecycle

| Command | Description |
|---------|-------------|
| `tavpbox init` | Initialize project (creates `.tavpbox.yml`) |
| `tavpbox create` | Create and start container |
| `tavpbox start` | Start container |
| `tavpbox stop` | Stop container |
| `tavpbox restart` | Restart container |
| `tavpbox destroy` | Destroy container permanently |
| `tavpbox rebuild` | Destroy and recreate container |

### Monitoring

| Command | Description |
|---------|-------------|
| `tavpbox list` | List all containers |
| `tavpbox info` | Show project details (URLs, DB creds) |
| `tavpbox logs` | Show container logs |

### Tooling

| Command | Description |
|---------|-------------|
| `tavpbox tooling` | List available tooling commands |
| `tavpbox artisan [args]` | Run `php artisan` in container |
| `tavpbox composer [args]` | Run `composer` in container |
| `tavpbox npm [args]` | Run `npm` in container |
| `tavpbox ssh [cmd]` | SSH into container or run command |

### Panel

| Command | Description |
|---------|-------------|
| `tavpbox panel` | Start web panel at http://localhost:8080 |
| `tavpbox panel -p 3000` | Start on custom port |
| `tavpbox panel:stop` | Stop panel |

### Proxy

| Command | Description |
|---------|-------------|
| `tavpbox proxy:start` | Start reverse proxy |
| `tavpbox proxy:stop` | Stop reverse proxy |
| `tavpbox proxy:status` | Show proxy status + routes |

### Config

| Command | Description |
|---------|-------------|
| `tavpbox config list` | List all configuration |
| `tavpbox config set <key> <value>` | Set config value |
| `tavpbox config get <key>` | Get config value |

| `tavpbox version` | Show version |

---

## Config: `.tavpbox.yml`

```yaml
name: my-project
recipe: tavp
webroot: public
services:
  mariadb:
    enabled: true
  redis:
    enabled: true
  mailpit:
    enabled: true
env:
  APP_NAME: "My Project"
  APP_ENV: local
  DB_DATABASE: my_database
  DB_USERNAME: my_user
  DB_PASSWORD: my_password
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer
  npm:
    cmd: npm
  test:
    cmd: php artisan test
ram: 512MB
cpu: 1
```

---

## Lando Migration

TAVPBox mendukung penuh `.lando.yml`. Cukup jalankan `tavpbox create` di folder yang ada `.lando.yml`.

### Yang di-support:
- ✅ `recipe` (lamp, laravel, dll)
- ✅ `services` (mariadb, mysql, redis, mailpit, phpmyadmin, dll)
- ✅ `tooling` (artisan, composer, npm, mysql, dll)
- ✅ `proxy` (*.lndo.site → *.tavp.my.id)
- ✅ `events.post-start` (build/run commands)
- ✅ `services.*.overrides.environment` (env vars)
- ✅ `services.*.creds` (DB credentials)

### Contoh migrasi:

```powershell
# Project Lando
cd ~/kos-kosan.id
cat .lando.yml
# name: koskosan
# recipe: lamp
# services:
#   appserver: { type: php:8.4 }
#   database: { type: mysql:8.0, creds: { user: koskosan } }
#   redis: { type: redis:7 }
#   mailpit: { type: mailpit }
# proxy:
#   appserver: [koskosan.lndo.site]

# Migrasi ke TAVPBox
tavpbox info
# Recipe:    laravel
# Services:  mariadb, redis, mailpit
# Domain:    http://koskosan.tavp.my.id
# Database:  koskosan/koskosan/koskosan

tavpbox create
# https://koskosan.tavp.my.id → jalan!
```

---

## Recipes

| Recipe | Description | Image | Default Services |
|--------|-------------|-------|------------------|
| `tavp` | TAVP Stack (PHP 8.3 + Nginx + Node 20) | ubuntu:24.04 | mariadb, redis, mailpit |
| `laravel` | Laravel | ubuntu:24.04 | mariadb, redis, mailpit |
| `php` | Generic PHP | ubuntu:24.04 | mariadb, redis |
| `node` | Node.js | node:20-alpine | redis |
| `go` | Go | golang:1.22-alpine | — |
| `python` | Python | python:3.12-slim | redis |
| `blank` | Empty container | ubuntu:24.04 | — |

## Services

| Service | Category | Description |
|---------|----------|-------------|
| mariadb | database | MySQL-compatible RDBMS |
| mysql | database | MySQL |
| postgres | database | PostgreSQL |
| mongodb | database | NoSQL document DB |
| redis | cache | In-memory cache |
| memcached | cache | Distributed cache |
| mailpit | mail | Email testing (SMTP + web UI) |
| mailhog | mail | Email testing |
| phpmyadmin | admin | Database admin UI |
| adminer | admin | Lightweight DB manager |
| elasticsearch | search | Search engine |
| rabbitmq | queue | Message broker |
| beanstalkd | queue | Work queue |
| apache | webserver | Apache HTTP server |
| varnish | cache | HTTP reverse proxy cache |

---

## HTTPS

HTTPS otomatis. TAVPBox sudah include wildcard cert `*.tavp.my.id` yang valid. Developer gak perlu setup apa-apa.

```powershell
tavpbox create
# https://myproject.tavp.my.id → langsung jalan
```

Cert wildcard di-embed di binary. Browser auto-trust. Expired ~90 hari, admin release binary baru.

---

## Web Panel

```powershell
tavpbox panel
# Opens http://localhost:8080
```

Features:
- Dashboard (all projects with status)
- Create project wizard
- Project detail (logs, URLs, DB credentials)
- Start/Stop/Restart/Destroy actions
- Recipe & service browser

---

## Architecture

```
tavpbox (Go binary)
├── CLI (cobra)
│   ├── init, create, start, stop, restart, destroy, rebuild
│   ├── ssh, list, info, logs
│   ├── tooling (dynamic subcommands)
│   ├── panel (web UI)
│   ├── proxy (reverse proxy management)
│   └── config (configuration)
├── Podman client (exec wrapper)
├── Embedded Go proxy
│   ├── HTTP :80
│   ├── HTTPS :443
│   └── Dynamic routes (routes.json)
├── Wildcard cert (*.tavp.my.id) embedded
├── Service library (15 services)
├── Recipe library (7 recipes)
├── Lando parser (.lando.yml)
├── Plugin engine (~/.tavpbox/plugins/)
└── API server (REST + embedded panel)
```

---

## Multi-Platform

| Platform | Architecture | Binary |
|----------|-------------|--------|
| Windows | amd64 | `tavpbox-windows-amd64.exe` |
| macOS | amd64 | `tavpbox-darwin-amd64` |
| macOS | arm64 (M1/M2) | `tavpbox-darwin-arm64` |
| Linux | amd64 | `tavpbox-linux-amd64` |
| Linux | arm64 | `tavpbox-linux-arm64` |

Cross-compile:
```bash
make cross
# Output: dist/tavpbox-{os}-{arch}
```

---

## Development

```bash
# Build
go build -o tavpbox .

# Cross-compile
make cross

# Run
./tavpbox version

# Test
go test ./...
```

---

## Troubleshooting

### Podman not found
Install Podman Desktop: https://podman-desktop.io

### Container already exists
```powershell
tavpbox destroy
tavpbox create
```

### Port already in use
```powershell
tavpbox proxy:stop
tavpbox proxy:start
```

---

## License

MIT

---

## Links

- **GitHub**: https://github.com/tavp-stack/tavpbox
- **Gitea**: https://git.glotama.com/tavp-stack/tavp-box
- **Issues**: https://github.com/tavp-stack/tavpbox/issues
- **Docs**: https://docs.tavp.web.id/guide/tavpbox.html
