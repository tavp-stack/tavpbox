# TAVPBox

> **LXC-based dev environment** — seperti [Lando](https://lando.dev), tapi lebih irit RAM karena pakai LXC bukan Docker.

```
┌─────────────────────────────────────────────────────────┐
│  TAVPBox — Ringkasan                                     │
├──────────────────┬──────────────────────────────────────┤
│ Runtime          │ LXC/LXD (system container)           │
│ RAM / container  │ ~30-50MB                             │
│ 20 project       │ ~700MB (vs Docker ~3.2GB)           │
│ Phalcon          │ Compile sekali, persist forever      │
│ Auto domain      │ *.tavp.local                         │
│ Config file      │ .tavpbox.yml                         │
│ Platform         │ Linux, macOS, Windows (WSL2)        │
└──────────────────┴──────────────────────────────────────┘
```

---

## Install

### Windows (PowerShell as Administrator)

```powershell
# Download dan install
iex (irm 'https://get.tavp.dev/setup-tavpbox.ps1' -UseB)
```

Atau manual:
```powershell
# 1. Enable WSL2
wsl --install --no-distribution

# 2. Install Ubuntu
wsl --install -d Ubuntu

# 3. Install LXD di dalam WSL
wsl -d Ubuntu -- sudo snap install lxd
wsl -d Ubuntu -- sudo lxd init --auto

# 4. Download binary
Invoke-WebRequest -Uri "https://github.com/tavp-stack/tavpbox/releases/latest/download/tavpbox-windows-amd64.exe" -OutFile tavpbox.exe

# 5. Jalankan
.\tavpbox.exe init
```

### macOS

```bash
# Install via Lima
curl -fsSL https://get.tavp.dev/setup-tavpbox.sh | bash
```

### Linux

```bash
# Install langsung
sudo curl -fsSL https://get.tavp.dev/setup-tavpbox.sh | bash
```

---

## Quick Start (5 Menit)

### 1. Inisialisasi

```bash
tavpbox init
```

TUI wizard akan muncul:
- Pilih distro (Ubuntu, Alpine, Debian, dll)
- Set domain suffix (*.tavp.local)
- Set default RAM per box

### 2. Buat Project

```bash
cd ~/projects/my-app
tavpbox create
```

TUI wizard akan muncul:
- Masukkan nama box
- Pilih stack (TAVP, Laravel, Node, Python, Blank)
- Pilih services (MariaDB, Redis, Mailpit, dll)

### 3. Akses Project

```bash
# Buka di browser
open http://my-app.tavp.local

# SSH ke container
tavpbox ssh my-app

# Jalankan command di container
tavpbox exec my-app php artisan migrate
```

---

## Semua Commands

### Lifecycle

| Command | Description |
|---------|-------------|
| `tavpbox init` | Setup pertama kali (TUI wizard) |
| `tavpbox create` | Buat box baru (TUI wizard atau dari file) |
| `tavpbox start <name>` | Start box |
| `tavpbox stop <name>` | Stop box (RAM langsung bebas) |
| `tavpbox restart <name>` | Restart box |
| `tavpbox destroy <name>` | Hapus box permanen |
| `tavpbox rebuild <name>` | Rebuild box (data di folder tetap ada) |

### Monitoring

| Command | Description |
|---------|-------------|
| `tavpbox list` | Lihat semua box |
| `tavpbox status` | Lihat status system + resource usage |
| `tavpbox info <name>` | Detail box (IP, stack, services) |
| `tavpbox logs <name>` | Lihat logs (nginx, PHP, system) |
| `tavpbox snapshot <name>` | Buat snapshot |

### Exec & SSH

| Command | Description |
|---------|-------------|
| `tavpbox ssh <name>` | Masuk terminal box |
| `tavpbox ssh <name> <cmd>` | Jalankan command di box |
| `tavpbox exec <name> <cmd>` | Jalankan command di box |

### Plugin & Images

| Command | Description |
|---------|-------------|
| `tavpbox plugin list` | Lihat plugin terinstall |
| `tavpbox plugin install <file>` | Install plugin dari YAML |
| `tavpbox plugin remove <name>` | Hapus plugin |
| `tavpbox images list` | Lihat cached images |
| `tavpbox images pull <image>` | Download & cache image |
| `tavpbox images remove <fingerprint>` | Hapus cached image |

### Custom Tooling

Jika `.tavpbox.yml` punya `tooling:` section:

```yaml
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer
  npm:
    cmd: npm
```

Maka bisa langsung:
```bash
tavpbox artisan migrate
tavpbox composer install
tavpbox npm run dev
```

---

## Config File: `.tavpbox.yml`

```yaml
# Nama box (wajib)
name: my-app

# Stack: tavp, laravel, symfony, wordpress, node, python, go, ruby, blank
stack: tavp

# Services (opsional)
services:
  - mariadb
  - redis
  - mailpit

# Environment variables (opsional)
env:
  APP_NAME: "My App"
  APP_ENV: local
  APP_DEBUG: "true"

# Webroot folder (default: .)
webroot: .

# Custom tooling commands (opsional)
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer

# Resource limits (opsional)
ram: 512MB
cpu: 1

# Distro override (opsional, default dari global config)
# distro: ubuntu/24.04
```

---

## Stacks

| Stack | Description | Components |
|-------|-------------|------------|
| `tavp` | TAVP Stack (PHP + Nginx + Node.js) | PHP 8.3, Nginx, Node 20 |
| `laravel` | Laravel | PHP 8.3, Nginx |
| `node` | Node.js | Node 20, Nginx |
| `python` | Python | Python 3, Nginx |
| `blank` | Empty container | Basic tools |

## Services

| Service | Description | Port |
|---------|-------------|------|
| `mariadb` | MariaDB database | 3306 |
| `mysql` | MySQL database | 3306 |
| `postgres` | PostgreSQL database | 5432 |
| `redis` | Redis cache | 6379 |
| `mailpit` | Email testing | 8025, 1025 |
| `phpmyadmin` | Database admin UI | 8080 |
| `adminer` | Database admin UI | 8080 |
| `elasticsearch` | Search engine | 9200 |
| `meilisearch` | Search engine | 7700 |
| `minio` | S3-compatible storage | 9000 |
| `rabbitmq` | Message queue | 5672 |

---

## RAM Comparison

```
Scenario: 20 development projects running simultaneously

Docker (Lando):
  dockerd         :  ~200MB
  20 containers   :  ~20 × 150MB = ~3000MB
  Total           :  ~3.2GB

LXC (TAVPBox):
  lxd daemon      :  ~30MB
  20 containers   :  ~20 × 35MB = ~700MB
  Caddy + dnsmasq :  ~15MB
  Total           :  ~745MB

  Savings: ~2.4GB (75% less RAM!)
```

---

## Plugins

### Stack Plugin Format

```yaml
name: my-stack
display_name: "My Stack"
description: "Custom stack"
version: "1.0.0"
category: stack

components:
  php:
    label: "PHP Version"
    type: select
    versions: ["8.1", "8.2", "8.3"]
    default: "8.3"

provision: scripts/my-stack.sh
```

### Service Plugin Format

```yaml
name: my-service
display_name: "My Service"
description: "Custom service"
version: "1.0.0"
category: service

provision: scripts/my-service.sh
ports:
  - 8080
```

### Install Plugin

```bash
tavpbox plugin install my-plugin.yml
```

---

## API (Desktop-Ready)

TAVPBox punya API layer untuk integrasi dengan desktop app:

```bash
# Start API server
tavpbox api

# Endpoints:
# GET  /api/boxes              → list all boxes
# POST /api/boxes/:name/start  → start box
# POST /api/boxes/:name/stop   → stop box
# DELETE /api/boxes/:name      → destroy box
# GET  /api/status             → system status
# GET  /api/plugins            → list plugins
```

Socket: `/tmp/tavpbox.sock`

---

## Development

### Build from Source

```bash
git clone https://github.com/tavp-stack/tavpbox.git
cd tavpbox
go build -o tavpbox .
```

### Cross-Compile

```bash
make cross
# Output: dist/tavpbox-{os}-{arch}
```

### Run Tests

```bash
make test
```

---

## Troubleshooting

### "lxc: not found"

```bash
# Install LXD
sudo snap install lxd
sudo lxd init --auto
```

### "WSL not available" (Windows)

```powershell
# Enable WSL2
wsl --install --no-distribution
# Restart computer
wsl --install -d Ubuntu
```

### Domain tidak resolve

```bash
# Restart dnsmasq
sudo systemctl restart dnsmasq

# Atau tambah manual di /etc/hosts
echo "127.0.0.1 my-app.tavp.local" | sudo tee -a /etc/hosts
```

### Container tidak bisa start

```bash
# Check logs
tavpbox logs my-app

# Rebuild container
tavpbox rebuild my-app
```

---

## License

MIT

---

## Links

- **Gitea (primary)**: https://git.glotama.com/tavp-stack/tavp-box
- **GitHub (mirror)**: https://github.com/tavp-stack/tavpbox
- **Docs**: https://docs.tavp.web.id/guide/tavpbox.html
