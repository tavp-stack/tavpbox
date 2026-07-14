# TAVPBox

> **Lando Dockerless** — Local development environment seperti [Lando](https://lando.dev), tapi pakai [Podman](https://podman.io) (bukan Docker).

```
┌─────────────────────────────────────────────────────────┐
│  TAVPBox                                                │
├──────────────────┬──────────────────────────────────────┤
│ Runtime          │ Podman (rootless, daemonless)        │
│ Reverse Proxy    │ Traefik (auto-HTTPS)                 │
│ RAM / container  │ ~50-80MB                             │
│ 20 project       │ ~1.5GB (vs Docker ~3.2GB)           │
│ Auto domain      │ *.tavp.my.id                         │
│ Config file      │ .tavpbox.yml                         │
│ Platform         │ Windows, macOS, Linux                │
│ CLI language     │ Go (single binary)                   │
│ Web UI           │ Built-in panel (Tailwind + Alpine)   │
│ License          │ MIT                                  │
└──────────────────┴──────────────────────────────────────┘
```

---

## Install

### Prerequisites

Install [Podman Desktop](https://podman-desktop.io) first, then:

```powershell
# Windows
podman machine init
podman machine start

# macOS
podman machine init
podman machine start

# Linux
sudo apt install podman   # or: sudo dnf install podman
```

### Install TAVPBox

**Option 1: Download binary**

Download from [Releases](https://github.com/tavp-stack/tavpbox/releases) and add to PATH.

**Option 2: Build from source**

```bash
git clone https://github.com/tavp-stack/tavpbox.git
cd tavpbox
go build -o tavpbox .
# Move to PATH:
# Windows: move tavpbox.exe C:\Users\<you>\AppData\Local\tavpbox\
# macOS/Linux: sudo mv tavpbox /usr/local/bin/
```

**Option 3: Go install**

```bash
go install github.com/tavp-stack/tavpbox@latest
```

---

## Quick Start

```bash
# 1. Init project
cd ~/projects/my-app
tavpbox init

# 2. Create container (installs nginx, PHP, services)
tavpbox create

# 3. Open in browser
# http://my-app.tavp.my.id

# 4. SSH into container
tavpbox ssh

# 5. Run tooling commands
tavpbox artisan migrate
tavpbox composer install
tavpbox npm run dev
```

---

## Commands

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

### Setup

| Command | Description |
|---------|-------------|
| `tavpbox setup` | Install dependencies (Podman) |
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

## Tooling

Tooling commands run inside the container. Define them in `.tavpbox.yml`:

```yaml
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer
  npm:
    cmd: npm
  test:
    cmd: php artisan test
```

Then use them directly:

```bash
tavpbox artisan migrate
tavpbox composer install
tavpbox npm run dev
tavpbox test
```

Default tooling is auto-detected from recipe:
- **tavp/laravel**: artisan, composer, npm, npx, php, test
- **php**: composer, php, test
- **node**: npm, npx, yarn, pnpm, node
- **go**: go
- **python**: python, pip, pytest

---

## Web Panel

```bash
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
├── Podman client (exec wrapper)
├── Traefik (reverse proxy container)
│   ├── *.tavp.my.id → container:80
│   └── Auto-HTTPS via ACME
├── Service library (15 services)
├── Recipe library (7 recipes)
├── Plugin engine (~/.tavpbox/plugins/)
├── API server (REST + embedded panel)
└── Tooling engine (dynamic subcommands)
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
```

---

## License

MIT

---

## Links

- **GitHub**: https://github.com/tavp-stack/tavpbox
- **Issues**: https://github.com/tavp-stack/tavpbox/issues
