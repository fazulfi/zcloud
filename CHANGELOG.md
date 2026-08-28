# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-08-28

### Added (M1.12 — Paid Beta Release)

- **Support contact channel**
  - `POST /api/v1/support/contact` backend handler with validation, rate limiting, and SMTP delivery.
  - Frontend `/contact` page (`ContactView.vue`) with name/email/subject/message form and success state.
  - `support@zrouter.dev` as the from address; email delivery verified end-to-end on staging.
- **Paid-beta runbook** (`docs/runbooks/paid-beta.md`) covering pre-launch checklist, release procedure, canary customer checklist, and rollback path.
- **Backup / restore / rollback tooling** (`deploy/backup.sh`, `deploy/restore.sh`, `deploy/rollback.sh`) with Redis `--no-auth-warning` fix and health-polled rollback.
- **Release workflow gates** (`.github/workflows/release.yml`): immutable tag input, frontend build stage, staging smoke deploy over SSH, production approval environment, and release evidence artifact (commit SHA, version, image digest).
- **Docker build pipeline**: three-stage `deploy/Dockerfile` with a `frontend-builder` stage (`FRONTEND_MODE` build arg, pinned `pnpm@10.33.0`) so images carry the correct API base URL for the target environment.
- **Domain migration to `zrouter.dev`**: `deploy/.env.zcloud.example`, `deploy/Caddyfile.zcloud`, and `deploy/docker-compose.zcloud.yml` updated; staging env now uses a relative `/api/v1` base URL.

### Changed

- `deploy/Dockerfile` now embeds the frontend at build time instead of copying a stale local artifact.
- `frontend/package.json` pins `packageManager: pnpm@10.33.0` to avoid corepack breaking on Node 20.

### Fixed

- Support-contact handler no longer returns 503 when SMTP settings are present in the database settings store.
- Staging frontend no longer hardcodes the production API origin (relative `/api/v1`), eliminating CORS failures on the staging host.

### Security

- Secrets are excluded from Git (staging env files are gitignored); only `.env.example` templates are committed.
