# NEXT_STEPS.md

> Snapshot kondisi terakhir session — dibaca AI/session berikutnya untuk langsung lanjut tanpa reka ulang konteks.

**Terakhir diupdate:** 2026-07-17 19:30 WIB

---

## Branch Aktif

- `main` — semua development langsung di main

## File yang Diubah di Session Ini

| File | Perubahan |
|------|-----------|
| `internal/podman/client.go` | `EnsureRunning()` — auto-detect + auto-fix Podman SSH (stop+start machine) |
| `cmd/lifecycle.go` | `start` command calls `EnsureRunning()` |
| `cmd/create.go` | `buildStartupScript()` — added service delays + nginx retry loop |
| `cmd/proxy.go` | `isProxyRunning()` — port 80 check, `restartProxy()` — kill by port |
| `images/php/start.sh` | MariaDB sleep 2, PHP-FPM sleep 1, nginx retry 3x |
| `.wslconfig` (user profile) | Created with `networkingMode=mirrored` for WSL2 SSH fix |
| `CHANGELOG.md` | Added v1.11.0 and v1.10.x entries |
| `README.md` | Added "Status Terkini" section |

## Progress Fitur/Task

| Task | Status |
|------|--------|
| WSL2 SSH port forwarding fix | ✅ Selesai — `.wslconfig` mirrored mode |
| Auto-fix Podman on start | ✅ Selesai — stop+start machine automatically |
| Startup script reliability | ✅ Selesai — service delays + nginx retry |
| Pre-built image rebuilt | ✅ Selesai — `ghcr.io/tavp-stack/tavpbox-php:latest` |
| Luto Laundry webapp working | ✅ HTTP 200 verified |
| 8/8 Lando projects migrated | ✅ Selesai |
| `events.post-start` auto-execution | ❌ Belum — Issue #4 dibuat |
| Windows Task Scheduler auto-start | ❌ Belum — user belum setup |

## Blocker Terakhir

- **Podman SSH socket (port 50312)** — Root cause: WSL2 localhost forwarding broken. Fix: `.wslconfig` dengan `networkingMode=mirrored`. Kalau user restart Windows, `.wslconfig` sudah ada dan seharusnya langsung works.
- **`events.post-start`** — Belum diauto-execute. User harus manual `tavpbox ssh` lalu jalankan command sendiri. Issue #4 tercatat.

## TODO Prioritas untuk Sesi Berikutnya

1. **Implement `events.post-start` executor** (Issue #4) — Execute commands from `.tavpbox.yml` after container starts
2. **Auto-start via Windows Task Scheduler** — Podman machine auto-start saat Windows boot
3. **Podman image rebuild automation** — Push new image setelah start.sh diupdate
4. **Test full restart cycle** — Matikan Windows → nyalakan → `tavpbox start` → verify works
5. **Close Issue #4** setelah events.post-start diimplement

## Referensi Issue/PR

- **#1** [closed] Port binding fix
- **#2** [closed] Post-start events + port binding
- **#3** [closed] HTTP→HTTPS + Service unavailable
- **#4** [open] events.post-start not auto-executed

## Release Info

- **v1.11.0** — Released to GitHub + Gitea (binaries uploaded)
- **Pre-built image** — `ghcr.io/tavp-stack/tavpbox-php:latest` (rebuilt with fixed start.sh)
- **User binary** — `C:\Users\JT\AppData\Local\tavpbox\tavpbox.exe` (v1.11.0 dev build)
