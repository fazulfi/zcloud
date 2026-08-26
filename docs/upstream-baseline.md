# Upstream Baseline

Record of the upstream sub2api revision this fork is based on, captured during
F0.1 (pin + audit). Any rebase/upgrade must start from this baseline.

## Baseline commit

- **Upstream repo:** https://github.com/Wei-Shaw/sub2api
- **Version tag:** v0.1.183 (latest release, 2026-08-25)
- **Baseline commit:** `e8cb019fa` — Merge PR #6078 (fix/responses-custom-tool-call-id)
- **License:** LGPL-3.0

## Security posture

- **CVE-2026-73079** (path traversal + confused deputy, CVSS 8.5,
  GHSA-vrxq-qm4h-6hgg): affected `>=v0.1.135 <=v0.1.168`; **fixed in v0.1.169**
  via `backend/internal/service/upstream_path_guard.go` (closed-allowlist path
  validation). v0.1.183 includes the fix. Verified present at baseline:
  `git show e2d9b82:backend/internal/service/upstream_path_guard.go` → OK.
- Recommended hardening from v0.1.169 release notes: add
  `security_opt: [no-new-privileges:true]` to container compose.

## Baseline diff policy

- **Migrations:** 001–230 are upstream-owned and MUST NOT be modified.
  zcloud extensions start at migration `231_` (see `spec/12-schema-migrations.md`).
- **Generated code:** `backend/ent/` is generated. Edit `ent/schema/*.go` then
  run `go generate ./ent`.
- **Fork boundary:** zcloud-specific code lives under `backend/zcloud/`.
  Minimal, documented edits to upstream files only where integration demands.

## Verified local toolchain (at baseline)

- VPS: Go 1.27.0 (`/usr/local/go1.27`, symlink `/usr/local/bin/go`),
  Docker 29.6.2, Compose v5.3.1, PostgreSQL 18.4 (host, 127.0.0.1:5432),
  Redis (host).
- Local Windows: go 1.22.12 — cannot build `go.mod` requiring 1.27; build on
  VPS or CI (`golang:1.27`).
