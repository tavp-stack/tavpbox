# SESSION_LOG.md

> Histori permanen tiap sesi — append di paling atas (reverse-chronological). JANGAN hapus/timpa entri lama.

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
