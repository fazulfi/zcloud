# Paid-beta operational runbook

> Domain of record: `zrouter.dev`. Staging: `zrouter.82.25.62.204.sslip.io`. Owner: release operator. Never put credentials in this document or in Git.

## Scope and success criteria

This runbook covers the first paid-beta release of zrouter: a pinned image, verified dependencies, a one-customer canary, controlled rolling promotion, and an operator-owned rollback path. The release is successful when health, gateway, billing, wallet, email, and support paths are verified and monitoring remains quiet for the observation window.

## Pre-launch checklist

- [ ] Confirm the release commit, immutable image tag, image digest, and migration list. Migrations 001–230 remain unchanged; new migrations start at 231 and are forward-only.
- [ ] DNS/TLS: `zrouter.dev`, `api.zrouter.dev`, and `app.zrouter.dev` resolve to the intended edge; staging resolves to `zrouter.82.25.62.204.sslip.io`; certificates renew successfully; nginx/Caddy has SSE `proxy_buffering off`.
- [ ] SMTP: verify `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASSWORD`, `SMTP_FROM=support@zrouter.dev`, SPF, DKIM, and DMARC. Send and receive a support test message.
- [ ] Secrets: rotate and store `JWT_SECRET`, `TOTP_ENCRYPTION_KEY`, `API_KEY_SECRET`, database/Redis credentials, supplier and Tron keys, wallet credentials, Telegram token, SMTP credentials, and webhook signing secret. Confirm `.env.zcloud` is mode 600 and absent from Git.
- [ ] Staging sign-off: deploy the candidate to `/opt/zcloud`, run `curl -fsS http://localhost:18080/health`, `/health/live`, `/health/ready`, login, model listing, a small gateway request, billing/usage read, wallet deposit observation, and an email test.
- [ ] Backup: confirm the latest PostgreSQL logical backup, Redis snapshot, off-host copy, rotation, and restore-drill record. PG RPO is ≤15 minutes when WAL/PITR is available; RTO is ≤60 minutes.
- [ ] Confirm on-call coverage, release window, customer communication, status-page wording, and a named approver for staging and production environments.

## Release procedure

1. **Build image.** Run the manual `Release` workflow with the intended tag. The workflow builds the frontend, builds and pushes `ghcr.io/<owner>/<repo>:<version>` (no mutable `latest`), records commit SHA, version, and digest, and publishes release evidence.
2. **Staging smoke.** The workflow validates the image healthcheck and, when `STAGING_HOST` and `STAGING_SSH_KEY` are configured, deploys the pinned image to staging over SSH. The staging environment requires manual approval.
3. **Canary.** After staging approval, select one canary customer and route only that customer's traffic (or one instance) to the immutable image. Observe for at least 15 minutes and complete the canary checklist below.
4. **Rolling promotion.** Obtain production approval, update `ZCLOUD_VERSION` to the exact approved tag/digest, apply migrations once, and replace one instance at a time. Verify `/health/ready` and smoke traffic after each instance before continuing.
5. **Record evidence.** Save workflow artifact and step summary links, deploy time, commit SHA, image digest, migration result, health responses, canary metrics, approver, and final version in the release log.

## Canary customer checklist

- [ ] Customer is informed, opted in, and has a rollback contact.
- [ ] Authentication and API key usage work; no unexpected 401/403 or quota changes.
- [ ] `GET /v1/models`, a representative streaming gateway request, cancellation, and retry succeed.
- [ ] Supplier latency/error rate and request cost stay within baseline; no elevated 5xx/timeouts.
- [ ] Wallet deposit/confirmation and balance changes occur once; billing ledger, usage, and displayed balance agree.
- [ ] No duplicate wallet credit, negative balance, or billing drift; compare ledger to supplier usage.
- [ ] Support email reaches `support@zrouter.dev` and replies thread correctly.
- [ ] Watch logs, gateway latency p95/p99, readiness, DB pool, Redis errors, queue/backlog, supplier errors, payment confirmations, and security events for the observation window.

## Monitoring and alert response

Monitor `/health/live` and `/health/ready`, request rate/5xx, gateway p50/p95/p99 latency, streaming disconnects, supplier availability and latency, PostgreSQL saturation/replication/WAL lag, Redis memory/errors, billing reconciliation drift, wallet confirmation age/duplicates, SMTP delivery, and authentication/security anomalies. Page the on-call for readiness failure, sustained 5xx or latency breach, supplier outage, payment/wallet inconsistency, data-loss risk, or suspected compromise. Preserve timestamps, request/correlation IDs, deployment version, dashboards, and logs before changing state.

## Incident response

Use the incident skeleton in `spec/08-deployment-ops.md` (if present in the release/spec bundle) and declare an incident owner, impact, start time, severity, and communications lead.

- **Gateway latency:** freeze promotion, compare app/supplier/DB/Redis latency, reduce load or disable affected route, and canary a known-good image.
- **Supplier outage:** fail closed for unavailable supplier operations, do not fabricate success or charge, communicate degraded service, and replay only idempotent work after recovery.
- **Billing drift:** stop automated credit/refund actions, make the ledger the source of truth, export affected accounts, reconcile supplier usage and ledger, and issue compensating forward-only entries with approval.
- **Wallet duplicate:** suspend affected crediting/withdrawal path, preserve transaction IDs and audit logs, deduplicate by chain transaction/hash, reconcile balances, and require two-person approval to correct balances.
- **Security event:** revoke/rotate affected credentials, isolate the workload, preserve evidence, disable compromised sessions/keys, notify the incident lead, and follow legal/customer notification requirements.

## Rollback procedure

Rollback is an image redeploy, not a destructive migration rollback. Stop promotion, identify the last known-good immutable tag, and run `deploy/rollback.sh <previous-tag> --dry-run` first. Take a fresh backup, then run `deploy/rollback.sh <previous-tag>` and confirm interactively. Verify `/health/ready`, login, model listing, gateway smoke, billing read, and logs. For schema incompatibility, use an approved compensating forward migration or PITR; do not edit or delete applied migrations. Escalate if RTO ≤60 minutes or data integrity is at risk.

## Backup and restore

Run `deploy/backup.sh` on the host from `/opt/zcloud`; it writes timestamped PostgreSQL custom dumps and Redis snapshots under `/opt/zcloud/backups/`, optionally encrypting with `gpg` when configured/available. Retain daily backups for 14 days, weekly for 8 weeks, and monthly for 6 months. Redis is not billing truth; after Redis loss, fail closed for authorization and restore only as operationally appropriate. Test `deploy/restore.sh <backupfile> --dry-run` before an approved restore; restore PG and Redis only with a maintenance window, confirmation, health checks, row-count verification, and a monthly restore drill.

## Support flow

Customer reports go to **support@zrouter.dev**. Support records customer, UTC timestamp, plan, request/correlation ID, endpoint, release version, symptoms, and screenshots/log excerpts without secrets. Acknowledge within the paid-beta SLA, classify billing/wallet/security incidents as urgent, and hand off to on-call with the incident template. Never request API keys, passwords, wallet private keys, or SMTP credentials by email.

## Post-launch checklist

- [ ] Keep canary and production dashboards watched through the agreed window; no unresolved page or unexplained error spike.
- [ ] Reconcile billing, supplier usage, wallet deposits, and support tickets; document exceptions and compensating entries.
- [ ] Confirm backups, off-host copy, retention labels, and WAL/PITR status; schedule the next monthly restore drill.
- [ ] Publish release notes/status update and notify beta customers of material changes.
- [ ] Close or hand off incidents, capture timeline/metrics/evidence, and schedule a blameless review.
- [ ] Update the release log with final version/digest, approvals, rollback readiness, and follow-up owners.
