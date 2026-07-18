# NEXT_STEPS.md

> Snapshot kondisi terakhir session — dibaca AI/session berikutnya untuk langsung lanjut tanpa reka ulang konteks.

**Terakhir diupdate:** 2026-07-18 11:30 WIB

---

## Branch Aktif

- `main` — semua development langsung di main
- Remote: `origin` = Gitea (git.glotama.com/tavp-stack/tavp-box), `github` = GitHub mirror (github.com/tavp-stack/tavpbox)

## File yang Diubah di Session Ini

| File | Perubahan |
|------|-----------|
| `cmd/create.go` | Fix phpMyAdmin world-writable (#7), add mysqli (#8), add adminer support |
| `images/php/Containerfile` | Add `mysqli` to docker-php-ext-install |
| `CHANGELOG.md` | ZeroVer migration: v1.x.y → 0.x.y, add 0.11.2 |
| `README.md` | Status Terkini → 0.11.2, fix panel port 5000→8080 |
| `WIKI.md` | Version 0.11.2, session 2026-07-18 |
| `SESSION_LOG.md` | Add session 2026-07-18 |
| `NEXT_STEPS.md` | Update session snapshot |

## Progress Fitur/Task

| Task | Status |
|------|--------|
| phpMyAdmin world-writable fix (#7) | ✅ Selesai — symlink to /etc (non-drvfs) |
| phpMyAdmin mysqli missing (#8) | ✅ Selesai — add mysqli to Containerfile |
| Adminer support | ✅ Selesai — nginx port 8081, proxy route |
| ZeroVer migration (0.11.2) | ✅ Selesai — CHANGELOG/README/WIKI |
| Remote rename (origin=Gitea, github=GitHub) | ✅ Selesai |
| Rebuild + push image to ghcr.io | ✅ Selesai — `ghcr.io/tavp-stack/tavpbox-php:latest` |
| WSL2 SSH port forwarding fix | ✅ Selesai (session sebelumnya) |
| Auto-fix Podman on start | ✅ Selesai (session sebelumnya) |
| `events.post-start` auto-execution | ❌ Belum — Issue #4 (user tunda) |
| Windows Task Scheduler auto-start | ❌ Belum — user belum setup |

## Blocker Terakhir

- **Issue #4** — `events.post-start` belum auto-execute. User minta tunda sampai project-project siap convert ke TAVP stack.
- **GitHub mirror** — Belum push ke GitHub (aturan: hanya main + CHANGELOG versi resmi + konfirmasi user).

## TODO Prioritas untuk Sesi Berikutnya

1. **Implement `events.post-start` executor** (Issue #4) — Setelah user siap convert project
2. **Close Issue #7 & #8** — Buat PR formal, test, close
3. **Push ke GitHub mirror** — Setelah user konfirmasi (main branch + CHANGELOG 0.11.2 resmi)
4. **Auto-start via Windows Task Scheduler** — Podman machine auto-start saat Windows boot
5. **Test full restart cycle** — Matikan Windows → nyalakan → `tavpbox start` → verify works

## Referensi Issue/PR

- **#1** [closed] Port binding fix
- **#2** [closed] Post-start events + port binding
- **#3** [closed] HTTP→HTTPS + Service unavailable
- **#4** [open] events.post-start not auto-executed (DITUNDA)
- **#7** [open→fix] phpMyAdmin world-writable (fixed, commented, PR menyusul)
- **#8** [open→fix] mysqli extension missing (fixed, commented, PR menyusul)

## Release Info

- **0.11.2** (ZeroVer) — Commits: `73b9745`, `7ba228a`, `5dfb1be`, `6398e3a`
- **Pre-built image** — `ghcr.io/tavp-stack/tavpbox-php:latest` (rebuilt with mysqli + adminer)
- **User binary** — `C:\Users\JT\AppData\Local\tavpbox\tavpbox.exe` (0.11.2 dev build)

## ZeroVer Convention

- Major version: **selalu 0** (tidak pernah naik ke 1.x)
- Patch (angka belakang): utama, naik tiap bug fix
- Minor (angka tengah): mengikuti kalau patch reset
- Contoh: 0.11.1 → 0.11.2 (patch naik), 0.11.9 → 0.12.0 (minor naik, patch reset)
