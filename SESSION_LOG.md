# SESSION_LOG.md

> Histori permanen tiap sesi — append di paling atas (reverse-chronological). JANGAN hapus/timpa entri lama.

---

## 2026-07-20 — Session: v0.12.0 Complete + Docs Update + Closing

**Waktu:** ~3 jam (13:00 - 15:35 WIB)

**Apa yang dikerjakan:**
- HTTP-only mode: Remove all HTTPS/SSL/TLS (proxy, setup, create, cert generation)
- Auto-detect recipe: `detectRecipe()` from composer.json/package.json/go.mod/requirements.txt
- events.post-start: Auto-execution during `tavpbox create` via EventsConfig
- Auto-create on start: `tavpbox start` auto-creates container if missing
- Asia/Jakarta timezone: Default TZ=Asia/Jakarta in config, create.go, lando.go, all Containerfiles
- Nginx webroot fix: Volume mount handles subdir, nginx always serves from /var/www/html
- Heredoc fix: writeNginxConfig via podman cp (avoids bash $ escaping)
- README update: HTTP-only mode, auto-detect, timezone
- docs.tavp.web.id: Updated and deployed via Vercel
- Issues #4 and #9: Closed (fixed in v0.12.0)
- Security audit: No secrets/tokens in tracked files

**Commit penting:**
- `924bf87` refactor: complete HTTP-only + TZ (remove port 443 kill, add TZ to Containerfiles, lando config)
- `5f4d291` fix: nginx webroot + podman cp for config writing
- `e85053a` feat: auto-detect recipe, events.post-start, auto-create on start (v0.12.0)
- `c245732` docs: update README to HTTP-only mode (v0.12.0)
- `12ada5a` docs: update TAVPBox v0.12.0 (tavp-docs repo)

**Issues:**
- #4 [closed] events.post-start not auto-executed ← FIXED
- #9 [closed] TAVP stack webroot issue ← FIXED

**Status:** Selesai — v0.12.0 released + docs updated

**Blocker untuk sesi berikutnya:**
- Wiki Gitea perlu update manual via web interface
- Release v0.12.0 belum dibuat (menunggu konfirmasi)
- English docs masih ada beberapa HTTPS references

---

## 2026-07-18 — Session: phpMyAdmin + Adminer Fix → ZeroVer 0.11.2 + Webroot Issue

**Waktu:** ~6 jam (08:00 - 14:30 WIB)

**Apa yang dikerjakan:**
- Fix phpMyAdmin world-writable (#7): symlink config ke `/etc` (non-drvfs)
- Fix phpMyAdmin mysqli missing (#8): tambah `mysqli` ke `docker-php-ext-install`
- Add Adminer support: nginx port 8081, CSS v5.5.0, proxy route
- ZeroVer migration: v1.x.y → 0.x.y (CHANGELOG/README/WIKI)
- GitHub Release 0.11.2: 6 binaries uploaded
- Proxy routes fix: rewrite `routes.json` dengan format benar
- Webroot fix (partial): ubah `.tavpbox.yml` ke `webroot: public` untuk TAVP stack projects
- Fix nginx config via base64 (bypass PowerShell heredoc issues)

**Commit penting:**
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
- #9 [created] TAVP stack webroot issue (Lando migration) ← **PRIORITAS**
- #4 [commented] events.post-start not auto-executed (ditunda)

**Release:** 0.11.2 (GitHub + Gitea)

**Status:** Masih berjalan — webroot fix belum selesai (HTTP 403/404)

**Blocker untuk sesi berikutnya:**
- Issue #9: TAVP stack webroot issue — `cmd/create.go` perlu update agar auto-detect `public/index.php`
- `lula`: HTTP 404 (progress dari 403)
- `tavp-web-id`: HTTP 403 (nginx config perlu fix)
- Issue #4: events.post-start ditunda (user belum siap convert)

---

## 2026-07-18 — Session: phpMyAdmin + Adminer Fix → ZeroVer 0.11.2

**Waktu:** ~3 jam (08:00 - 11:30 WIB)

**Apa yang dikerjakan:**
- Fix phpMyAdmin world-writable (#7): Root cause = drvfs mount `C:\` → `chmod` diabaikan → `config.inc.php` selalu `0777`. Fix: symlink ke `/etc` (non-drvfs, perms `0644`).
- Fix phpMyAdmin mysqli missing (#8): `images/php/Containerfile` tidak install `mysqli`. Fix: tambah `mysqli` ke `docker-php-ext-install`.
- Add proper Adminer support: nginx config port 8081, drvfs fix, proxy route.
- ZeroVer migration: CHANGELOG/README/WIKI → `0.11.2` (major=0, patch utama).

**Commit penting:**
- `73b9745` fix: phpMyAdmin world-writable config.inc.php on drvfs/WSL mounts (#7)
- `7ba228a` fix: install mysqli PHP extension in php image (#8)
- `5dfb1be` feat: add proper adminer support with dedicated nginx config (#8 follow-up)
- `...` docs: ZeroVer 0.11.2 changelog + README + WIKI

**Issues:**
- #7 [open→fix] phpMyAdmin world-writable (fixed, commented)
- #8 [open→fix] mysqli extension missing (fixed, commented)
- #4 [open] events.post-start not auto-executed

**Status:** Selesai — phpMyAdmin + Adminer HTTP 200 di container `tavp-tavp-web-id`

**Blocker untuk sesi berikutnya:**
- Issue #4 (events.post-start) belum dikerjakan (user minta tunda)
- Rebuild pre-built image `ghcr.io/tavp-stack/tavpbox-php:latest` (mysqli + adminer)

---

## 2026-07-17 — Session: Podman SSH Fix + v1.11.0 Release

**Waktu:** ~2 jam (18:00 - 20:00 WIB)

**Apa yang dikerjakan:**
- Investigasi Podman SSH socket (50312) tidak listen meskipun `podman machine list` show "running"
- Root cause: WSL2 localhost forwarding broken — SSH listen inside WSL tapi tidak di-forward ke Windows
- Fix: Buat `.wslconfig` dengan `networkingMode=mirrored` di `%USERPROFILE%`
- Implement `EnsureRunning()` auto-fix: detect → stop → start → wait for SSH
- Fix startup script: MariaDB sleep 2, PHP-FPM sleep 1, nginx retry 3x
- Rebuild pre-built image `ghcr.io/tavp-stack/tavpbox-php:latest`
- Recreate lula container, verify HTTP 200

**Commit penting:**
- `54a4f5a` feat: auto-fix Podman SSH
- `dd24eaf` fix: comprehensive Podman + nginx fixes
- `a30932f` docs: add v1.11.0 changelog
- `59c598a` docs: add Status Terkini section
- `3062698` docs: add NEXT_STEPS.md

**Release:**
- v1.11.0 — GitHub + Gitea (binaries uploaded)

**Issues:**
- #1 [closed] Port binding fix
- #2 [closed] Post-start events + port binding
- #3 [closed] HTTP→HTTPS + Service unavailable
- #4 [created] events.post-start not auto-executed

**Status:** Selesai — Lula webapp accessible di `http://lula.tavp.my.id/`

**Blocker untuk sesi berikutnya:**
- `events.post-start` belum auto-execute (Issue #4)
- User belum setup Windows Task Scheduler untuk Podman auto-start

---
