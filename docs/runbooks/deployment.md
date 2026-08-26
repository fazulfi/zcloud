# zcloud deployment runbook

> Baseline: sub2api v0.1.183 (e8cb019fa) + zcloud fork boundary. See `docs/upstream-baseline.md`.
> Kernel: `spec/00-overview.md` v0.11.2. Milestones: `spec/09-roadmap.md`, `spec/10-implementation-plan.md`.

## Environments

| Env | URL | Host | Notes |
|---|---|---|---|
| staging | zcloud.82.25.62.204.sslip.io | 82.25.62.204 (SG) | Port 18080/18443, VPS host PG+Redis, nginx TLS |
| production | zcloud.dev / api.zcloud.dev / app.zcloud.dev | 82.25.62.204 (SG) | Caddy/nginx TLS, managed PG/Redis or same VPS |

## Staging deploy (rehearsal)

```bash
# On VPS (root@82.25.62.204) or from CI artifact:
cd /opt/zcloud/deploy
cp .env.zcloud.example .env.zcloud && chmod 600 .env.zcloud
# edit .env.zcloud: DATABASE_PASSWORD, REDIS_PASSWORD, JWT_SECRET, TOTP_ENCRYPTION_KEY,
#                  INFERHUB_API_KEY, TRON_API_KEY, ADMIN_PASSWORD, TELEGRAM_BOT_TOKEN
docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud build
docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud up -d
docker compose -f docker-compose.zcloud.yml ps
curl -s http://localhost:18080/health
```

Host nginx site (staging):

```nginx
# /etc/nginx/sites-available/zcloud-staging
server {
    listen 80;
    server_name zcloud.82.25.62.204.sslip.io;
    underscores_in_headers on;              # required: sticky-session underscore headers
    client_max_body_size 256m;
    location / {
        proxy_pass http://127.0.0.1:18080;
        proxy_http_version 1.1;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_buffering off;                # SSE streaming relay
        proxy_read_timeout 600s;
    }
}
```

## Production deploy

```bash
cd /opt/zcloud/deploy
cp .env.zcloud.example .env.zcloud && chmod 600 .env.zcloud
# rotate ALL secrets; set DATABASE_HOST/REDIS_HOST to managed endpoints or keep compose services
docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud up -d --build
docker compose -f docker-compose.zcloud.yml exec zrouter /app/sub2api migrate  # if manual migration needed
```

TLS: Caddyfile.zcloud (Caddy handles ACME) OR host nginx + certbot. SSE: keep `proxy_buffering off`.

## Migration policy

- 001–230 upstream unchanged; zcloud migrations start at 231 (`backend/migrations/231_*.sql`).
- Runner: sort filename, SHA-256 checksum, advisory lock, transactional per file; forward-only, NO Down SQL.
- New schema files: `ent/schema/*.go` → `go generate ./ent` (never edit generated).

## Secrets inventory (rotate before prod)

INFERHUB_API_KEY · TRON_API_KEY + wallet private key · JWT_SECRET · TOTP_ENCRYPTION_KEY ·
API_KEY_SECRET (sub2api default — MUST change) · DATABASE_PASSWORD · REDIS_PASSWORD ·
SMTP/Resend · TELEGRAM_BOT_TOKEN · webhook signing secret

## Backup (prod)

- PG daily logical 14d + weekly 8w + monthly 6m; WAL/PITR RPO ≤15m, RTO ≤60m; monthly restore drill.
- Redis is NOT billing source of truth; on loss → fail closed for authorization.

## Rollback

- Image tag pinned (`ZCLOUD_VERSION`); rollback = redeploy previous tag + DB migration rollback (compensating,
  forward-only) or point-in-time restore. Canary: shift 1 instance at a time; smoke /health/ready before promote.

## Health

- `/health/live` = process alive; `/health/ready` = deps + drain state.
- Upstream probe: `GET /v1/models` + `GET /v1/me/usage` (InferHub has no /health).
