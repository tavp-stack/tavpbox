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
│ CLI language     │ Go (single binary)                   │
│ Desktop app      │ Tauri (coming soon)                  │
└──────────────────┴──────────────────────────────────────┘
```

---

## Tentang Project Ini

TAVPBox dikembangkan menggunakan bantuan AI (MiMo-V2.5-Pro via OpenCode) untuk mempercepat proses development. Arsitektur dan keputusan desain dibuat berdasarkan analisis mendalam terhadap:

- [WSL UI](https://github.com/octasoft-ltd/wsl-ui) — pattern untuk install WSL tanpa terminal
- [Lando](https://lando.dev) — pattern untuk CLI dan plugin system
- [TAVP Stack](https://docs.tavp.web.id) — ekosistem TAVP yang sudah ada

---

## Install

### Windows (PowerShell as Administrator)

```powershell
# Cara cepat
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

# 4. Download binary dari GitHub Releases
# https://github.com/tavp-stack/tavpbox/releases

# 5. Jalankan
.\tavpbox.exe init
```

### macOS

```bash
curl -fsSL https://get.tavp.dev/setup-tavpbox.sh | bash
```

### Linux

```bash
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

### Custom Tooling

Jika `.tavpbox.yml` punya `tooling:` section:

```yaml
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer
```

Maka bisa langsung:
```bash
tavpbox artisan migrate
tavpbox composer install
```

---

## Config File: `.tavpbox.yml`

```yaml
name: my-project
stack: tavp
services:
  - mariadb
  - redis
  - mailpit
webroot: .
env:
  APP_NAME: "My Project"
  APP_ENV: local
tooling:
  artisan:
    cmd: php artisan
  composer:
    cmd: composer
ram: 512MB
cpu: 1
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
| `postgres` | PostgreSQL database | 5432 |
| `redis` | Redis cache | 6379 |
| `mailpit` | Email testing | 8025, 1025 |
| `phpmyadmin` | Database admin UI | 8080 |

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

## Development

### Build from Source

```bash
git clone https://git.glotama.com/tavp-stack/tavp-box.git
cd tavp-box
go build -o tavpbox .
```

### Cross-Compile

```bash
make cross
# Output: dist/tavpbox-{os}-{arch}
```

---

## Troubleshooting

### "lxc: not found"

```bash
sudo snap install lxd
sudo lxd init --auto
```

### "WSL not available" (Windows)

```powershell
wsl --install --no-distribution
# Restart computer
wsl --install -d Ubuntu
```

### Domain tidak resolve

```bash
sudo systemctl restart dnsmasq
```

---

## License

MIT

---

## Links

- **Gitea (primary)**: https://git.glotama.com/tavp-stack/tavp-box
- **GitHub (mirror)**: https://github.com/tavp-stack/tavpbox
- **Docs**: https://docs.tavp.web.id/guide/tavpbox.html
