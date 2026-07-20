# NEXT_STEPS.md

> Snapshot kondisi terakhir session — dibaca AI/session berikutnya untuk langsung lanjut tanpa reka ulang konteks.

**Terakhir diupdate:** 2026-07-20 15:35 WIB

---

## Branch Aktif

- `main` — semua development langsung di main
- Remote: `origin` = Gitea (git.glotama.com/tavp-stack/tavp-box), `github` = GitHub mirror (github.com/tavp-stack/tavpbox)

## File yang Diubah di Session Ini

| File | Perubahan |
|------|-----------|
| `cmd/create.go` | HTTP-only (remove cert check), events.post-start execution, writeNginxConfig via podman cp, webroot handling |
| `cmd/init.go` | Auto-detect recipe from composer.json/package.json/go.mod/requirements.txt |
| `cmd/lifecycle.go` | Auto-create on `tavpbox start` when container missing |
| `cmd/setup.go` | Remove HTTPS cert generation step |
| `cmd/proxy.go` | Remove port 443 from kill/restart |
| `internal/proxy/proxy.go` | Remove TLS listener, crypto/tls, certs imports |
| `internal/config/config.go` | Add EventsConfig struct, TZ field |
| `internal/config/lando.go` | Add events.post-start mapping, detectRecipeFromLando(), default TZ |
| `internal/podman/client.go` | Add Copy() and Exists() methods |
| `images/php/Containerfile` | Add tzdata + ENV TZ=Asia/Jakarta |
| `images/node/Containerfile` | Add tzdata + ENV TZ=Asia/Jakarta |
| `images/python/Containerfile` | Add tzdata + ENV TZ=Asia/Jakarta |
| `images/go/Containerfile` | Add tzdata + ENV TZ=Asia/Jakarta |
| `CHANGELOG.md` | v0.12.0 entry |
| `README.md` | HTTP-only mode, auto-detect, timezone |
| `.gitignore` | Add .opencode/ |

## Progress Fitur/Task

| Task | Status |
|------|--------|
| HTTP-only mode (remove HTTPS) | ✅ Selesai — v0.12.0 |
| Auto-detect recipe | ✅ Selesai — v0.12.0 |
| events.post-start auto-execution (#4) | ✅ Selesai — v0.12.0 |
| Auto-create on start | ✅ Selesai — v0.12.0 |
| Asia/Jakarta timezone default | ✅ Selesai — v0.12.0 |
| Nginx webroot fix (#9) | ✅ Selesai — v0.12.0 |
| Heredoc $ escaping fix | ✅ Selesai — podman cp approach |
| README update HTTP-only | ✅ Selesai — c245732 |
| docs.tavp.web.id update | ✅ Selesai — deployed via Vercel |
| Wiki Gitea update | ⚠️ Perlu manual — API tidak support update |
| Release v0.12.0 | ❌ Belum — menunggu konfirmasi |
| Windows Task Scheduler auto-start | ❌ Belum |

## Blocker Terakhir

Tidak ada blocker. Semua Issue (#4, #9) sudah fix di v0.12.0.

## Kerjaan Setengah Jadi

- **Wiki Gitea**: Perlu update manual via web interface (API tidak support create/update pages)
- **Release v0.12.0**: Belum dibuat — menunggu konfirmasi user

## TODO Prioritas untuk Sesi Berikutnya

1. **Update Wiki Gitea** — Update halaman Architecture, Quick Start, Known Issues dengan info v0.12.0 (manual via web)
2. **Release v0.12.0** — Buat Release di Gitea + mirror ke GitHub (jika user konfirmasi)
3. **Windows Task Scheduler** — Setup auto-start Podman saat Windows boot
4. **Test full restart cycle** — Matikan Windows → nyalakan → `tavpbox start` → verify works
5. **English docs cleanup** — Update docs/en/guide/tavpbox.md (masih ada beberapa HTTPS references)

## Referensi Issue/PR

- **#1** [closed] Port binding fix
- **#2** [closed] Post-start events + port binding
- **#3** [closed] HTTP→HTTPS + Service unavailable
- **#4** [closed] events.post-start not auto-executed ← **FIXED di v0.12.0**
- **#5** [closed] Mailpit not started after container restart
- **#6** [closed] Mailpit process dies silently
- **#7** [closed] phpMyAdmin world-writable
- **#8** [closed] mysqli extension missing
- **#9** [closed] TAVP stack webroot issue ← **FIXED di v0.12.0**

## Release Info

- **0.12.0** (current) — Commits: `924bf87`, `5f4d291`, `e85053a`, `c245732`
- **0.11.2** (previous) — GitHub Release: https://github.com/tavp-stack/tavpbox/releases/tag/0.11.2
- **Pre-built image** — `ghcr.io/tavp-stack/tavpbox-php:latest`

## ZeroVer Convention

- Major version: **selalu 0** (tidak pernah naik ke 1.x)
- Patch (angka belakang): utama, naik tiap bug fix
- Minor (angka tengah): mengikuti kalau patch reset
- Contoh: 0.11.1 → 0.11.2 (patch naik), 0.11.9 → 0.12.0 (minor naik, patch reset)

## Active Projects

| Project | Container | Status | URL |
|---------|-----------|--------|-----|
| tavp-web-id | tavp-tavp-web-id | ✅ HTTP 200 | http://tavp-web-id.tavp.my.id |
| lula | tavp-lula | ✅ HTTP 200 | http://lula.tavp.my.id |

## Git Remotes

| Remote | URL | Purpose |
|--------|-----|---------|
| origin | https://git.glotama.com/tavp-stack/tavp-box.git | Gitea (primary) |
| github | https://github.com/tavp-stack/tavpbox.git | GitHub mirror |

## Gitea API Token

- Token: `0e6b86795bb32063035b69a49784a2a438b93e96`
- Scope: Issues, Wiki, Releases
- Hanya untuk operasi Gitea, JANGAN di-commit ke repo

## Domain & DNS

- Domain: `*.tavp.my.id` via Cloudflare
- DNS: A record `*.tavp.my.id` → `127.0.0.1`
- Protocol: HTTP only (port 80)
