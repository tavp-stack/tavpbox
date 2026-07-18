# NEXT_STEPS.md

> Snapshot kondisi terakhir session — dibaca AI/session berikutnya untuk langsung lanjut tanpa reka ulang konteks.

**Terakhir diupdate:** 2026-07-18 14:30 WIB

---

## Branch Aktif

- `main` — semua development langsung di main
- Remote: `origin` = Gitea (git.glotama.com/tavp-stack/tavp-box), `github` = GitHub mirror (github.com/tavp-stack/tavpbox)

## File yang Diubah di Session Ini

| File | Perubahan |
|------|-----------|
| `cmd/create.go` | Fix phpMyAdmin world-writable (#7), add mysqli (#8), add adminer support, fix port mapping |
| `images/php/Containerfile` | Add `mysqli` to docker-php-ext-install |
| `CHANGELOG.md` | ZeroVer migration: v1.x.y → 0.x.y, add 0.11.2 |
| `README.md` | Status Terkini → 0.11.2, fix panel port 5000→8080 |
| `WIKI.md` | Version 0.11.2, session 2026-07-18 |
| `SESSION_LOG.md` | Add session 2026-07-18 |
| `NEXT_STEPS.md` | Update session snapshot |
| `.gitignore` | Add tavpbox.exe, dist/, proxy-*.log, release-notes-*.md |
| `fix-nginx.sh` | New utility script for container nginx config fix |

## Progress Fitur/Task

| Task | Status |
|------|--------|
| phpMyAdmin world-writable fix (#7) | ✅ Selesai — symlink to /etc (non-drvfs) |
| phpMyAdmin mysqli missing (#8) | ✅ Selesai — add mysqli to Containerfile |
| Adminer support + CSS v5.5.0 | ✅ Selesai — nginx port 8081, proxy route |
| ZeroVer migration (0.11.2) | ✅ Selesai — CHANGELOG/README/WIKI |
| Remote rename (origin=Gitea, github=GitHub) | ✅ Selesai |
| Rebuild + push image to ghcr.io | ✅ Selesai — `ghcr.io/tavp-stack/tavpbox-php:latest` |
| GitHub Release 0.11.2 | ✅ Selesai — 6 binaries uploaded |
| Proxy routes fix | ✅ Selesai — rewrite routes.json |
| WSL2 SSH port forwarding fix | ✅ Selesai (session sebelumnya) |
| Auto-fix Podman on start | ✅ Selesai (session sebelumnya) |
| TAVP stack webroot issue (#9) | ⚠️ Partial — ubah ke `webroot: public` tapi masih 403/404 |
| `events.post-start` auto-execution (#4) | ❌ Belum — user tunda |
| Windows Task Scheduler auto-start | ❌ Belum — user belum setup |

## Blocker Terakhir

### Issue #9: TAVP stack webroot issue (Lando migration)

**Masalah:** TAVP stack projects (migrasi Lando) punya `index.php` di `public/` bukan root. TAVPBox generate nginx config hardcoded `root /var/www/html` → 403/404.

**Sudah dicoba:**
1. Ubah `.tavpbox.yml` dari `webroot: .` ke `webroot: public` → rebuild → test
2. Fix nginx config manual via base64 → test
3. Restart container → test

**Hasil:**
- `lula`: HTTP 404 (progress dari 403)
- `tavp-web-id`: HTTP 403 (nginx config perlu fix)

**Solusi yang belum dilakukan:**
- Update `cmd/create.go` agar auto-detect `public/index.php` dan set nginx root ke `/var/www/html/public`

### Issue #4: events.post-start not auto-executed

**Status:** Ditunda (user: "nanti dulu kalo project project gw semuanya mau gw convert")

## Kerjaan Setengah Jadi

- **Webroot fix:** `.tavpbox.yml` sudah diubah ke `webroot: public` untuk `lula` dan `tavp-web-id`, tapi masih bermasalah (403/404)
- **Nginx config:** Sudah fix via base64 untuk `tavp-web-id` (HTTP 200 sebelum rebuild), tapi setelah rebuild kembali 403

## TODO Prioritas untuk Sesi Berikutnya

1. **Fix webroot issue (Issue #9)** — Update `cmd/create.go` agar auto-detect `public/index.php` → set nginx root `/var/www/html/public`
2. **Fix `lula` (HTTP 404)** — Investigasi kenapa 404 setelah `webroot: public`
3. **Fix `tavp-web-id` (HTTP 403)** — Pastikan nginx config benar setelah rebuild
4. **Implement events.post-start (Issue #4)** — Setelah user siap convert project
5. **Windows Task Scheduler** — Setup auto-start Podman saat Windows boot
6. **Test full restart cycle** — Matikan Windows → nyalakan → `tavpbox start` → verify works

## Referensi Issue/PR

- **#1** [closed] Port binding fix
- **#2** [closed] Post-start events + port binding
- **#3** [closed] HTTP→HTTPS + Service unavailable
- **#4** [open] events.post-start not auto-executed (DITUNDA)
- **#5** [closed] Mailpit not started after container restart
- **#6** [closed] Mailpit process dies silently
- **#7** [closed] phpMyAdmin world-writable
- **#8** [closed] mysqli extension missing
- **#9** [open] TAVP stack webroot issue (Lando migration) ← **PRIORITAS**

## Release Info

- **0.11.2** (ZeroVer) — Commits: `73b9745`, `7ba228a`, `5dfb1be`, `675e505`, `ca707a5`, `6398e3a`, `841843b`, `0f945ba`
- **Pre-built image** — `ghcr.io/tavp-stack/tavpbox-php:latest` (rebuilt with mysqli + adminer)
- **GitHub Release** — https://github.com/tavp-stack/tavpbox/releases/tag/0.11.2

## ZeroVer Convention

- Major version: **selalu 0** (tidak pernah naik ke 1.x)
- Patch (angka belakang): utama, naik tiap bug fix
- Minor (angka tengah): mengikuti kalau patch reset
- Contoh: 0.11.1 → 0.11.2 (patch naik), 0.11.9 → 0.12.0 (minor naik, patch reset)

## Active Projects

| Project | Container | Status | URL |
|---------|-----------|--------|-----|
| tavp-web-id | tavp-tavp-web-id | ⚠️ HTTP 403 | https://tavp-web-id.tavp.my.id/ |
| lula | tavp-lula | ⚠️ HTTP 404 | https://lula.tavp.my.id/ |
| test-tavp | tavp-test-tavp | ✅ HTTP 200 | https://test-tavp.tavp.my.id/ |
