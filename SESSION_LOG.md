# SESSION_LOG.md

> Histori permanen tiap sesi — append di paling atas (reverse-chronological). JANGAN hapus/timpa entri lama.

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
