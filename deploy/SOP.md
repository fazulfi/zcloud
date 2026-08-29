# zcloud — Deployment SOP (Standard Operating Procedure)

> **Access model: TAILNET-ONLY.** No public IP, no public DNS, no nginx/Caddy vhosts.
> Access URL: `https://app.zrouter.dev` (SPA) / `https://api.zrouter.dev` (API). Tailnet `https://app.zrouter.dev` = admin fallback only.
> Domain `zrouter.dev` is parked at Hostinger and is **NOT** used.

---

## 1. Architecture

```
[Admin device on Tailnet]  (e.g. faizzzzz 100.112.201.124, Windows)
        │  Tailscale (tailscale0, UFW allow-all on this interface)
        ▼
https://app.zrouter.dev
        │
        ▼
[container: zcloud-zrouter]  (image zcloud/zrouter:0.1.186-zcloud, port 18080)
   │  single binary: Go API + embedded Vue SPA (index-*.js chunks)
   │  healthcheck: wget localhost:18080/health
   │
   ├──► [postgres]  (zcloud-postgres, DB: sub2api, 108 tables, private docker net)
   ├──► [redis]     (zcloud-redis, private docker net)
   │
   └──► QRIS payments ──► http://172.21.0.1:30158  (socat gomerch-proxy.service)
                          └──► 127.0.0.1:3015      (gomerch systemd service)
                               ├── /etc/gomerch.env        (X-API-Key, 64 hex)
                               └── /opt/gomerch/data/payments.json
```

Key facts:
- DB access: `docker exec zcloud-postgres psql -U zcloud -d sub2api`
- DB user/pass come from `deploy/.env.zcloud` (live values — **do not rename**).
- gomerch proxy: socat `0.0.0.0:30158 → 127.0.0.1:3015` (ufw allows 172.21.0.0/16 → 30158 so the container can reach the host).
- Admin account: `faizzulfikar1080@gmail.com` / `ZcloudAdmin2026!` (AUTO_SETUP). Admin token file: `/tmp/zadmin.token` (Bearer).

---

## 2. Prerequisites (first time only)

1. Tailscale installed + logged in on the VPS:
   `tailscale up` → node `ggl-vps` (100.100.17.99). Public app access is via domain; tailnet is admin fallback.
2. SSH alias `faiz-prod` → `root@82.25.62.204` (key: `~/.ssh/id_ed25519`), used only for SSH transport (never as an app URL).
3. Local repo: `C:\Users\faizz\new-saas\sub2api`, branch `feat/branding-zcloud`, Node/pnpm + Go toolchain available.

---

## 3. Access (who can reach what)

| What | How |
|---|---|
| App (UI + API) | `https://app.zrouter.dev` (public, Cloudflare TLS) |
| API base (VITE_API_BASE_URL) | `https://app.zrouter.dev/api/v1` |
| Websocket (VITE_WS_BASE_URL) | `wss://api.zrouter.dev` |
| Health | `https://app.zrouter.dev/health` |
| Admin panel | `https://app.zrouter.dev/admin/dashboard` (login as admin) |

No firewall rule exposes 18080 publicly (ufw: allow 22,80,443,6080,5900, docker-net→30158, **all on tailscale0**; default deny incoming).

---

## 4. Deploy flow (update)

```bash
# 1. Build frontend (produces frontend/dist, embedded into binary)
cd frontend
pnpm install --frozen-lockfile
pnpm build

# 2. Sync repo → VPS (tar-over-ssh; excludes secrets/build artifacts)
cd ..
tar czf - --exclude=.git --exclude=node_modules --exclude=.sisyphus \
         --exclude=.env.zcloud --exclude=backups --exclude=__pycache__ \
  | ssh faiz-prod "cd /opt/zcloud && tar xzf - && chown -R 197609:197609"

# NOTE: tar does NOT delete stale files. If a file was deleted locally,
# remove it on the VPS manually (e.g. rm -f /opt/zcloud/frontend/src/...).

# 3. Rebuild image + restart container
ssh faiz-prod "cd /opt/zcloud/deploy && \
  docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud build zrouter && \
  docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud up -d --force-recreate zrouter"

# 4. Migrations auto-apply on container start (runner is internal, no schema_migrations table).
#    Current: 245 (plan.model_id), 246 (68 plans seed), 247 (currency→USD repair).
```

Health gate after `up`:
```bash
ssh faiz-prod "docker ps --filter name=zcloud-zrouter --format '{{.Status}}'"   # must show (healthy)
curl -s https://app.zrouter.dev/health
```

---

## 5. Post-deploy verification

```bash
TOKEN=$(ssh faiz-prod "cat /tmp/zadmin.token")

# 5.1 checkout-info: 68 plans, all USD + tokens + model_id + product_name
curl -s http://127.0.0.1:18080/api/v1/payment/checkout-info \
  -H "Authorization: Bearer $TOKEN" \
  | python3 -c "import json,sys; d=json.load(sys.stdin)['data']; ps=d['plans'];
print('plans',len(ps),'| non-USD',sum(1 for p in ps if p.get('currency','USD')!='USD'),
      '| no-tokens',sum(1 for p in ps if not p.get('tokens')),'| no-model',sum(1 for p in ps if not p.get('model_id')))"

# 5.2 SPA serves the new frontend chunk
curl -s http://127.0.0.1:18080/ | grep -o 'index-[A-Za-z0-9_-]*\.js' | head -1

# 5.3 New model-detail route answers
curl -s -o /dev/null -w '%{http_code}\n' http://127.0.0.1:18080/model-plaza/model/gpt-5.5   # 200
```

---

## 6. E2E QRIS smoke test (payment)

1. Mint a customer JWT (user-scope token):
   - On the VPS, read `JWT_SECRET` from `/opt/zcloud/deploy/.env.zcloud`.
   - `token_version` = fingerprint = `int64(sha256(email_lower + "\n" + pw_hash)[:8]) & 0x7fffffffffffffff` XOR `user.TokenVersion` (0 if column absent).
   - Build HS256 JWT `{user_id, email, role:"user", token_version, sid:"e2e", exp:+7200}`.
2. Create order:
   `POST /api/v1/payment/orders` `{"amount":1,"payment_type":"qris","order_type":"subscription","plan_id":13}` → capture `out_trade_no` + `qr_code`.
3. Simulate payment (gomerch has no webhook/cancel):
   - Edit `/opt/gomerch/data/payments.json`: mark the order's `payment_ref` as `paid`.
   - `cp payments.json payments.json.bak && systemctl restart gomerch`
4. Verify:
   `POST /api/v1/payment/orders/verify` `{"out_trade_no":"..."}` → status `COMPLETED`.
5. Assert DB:
   - `model_balances`: 1 row, `tokens_purchased=balance=10,000,000`, status active.
   - `user_subscriptions`: model_id set, purchased_tokens 10M, plan_name_snapshot 'GPT-5.5 $1'.
   - `payment_audit_logs`: ORDER_CREATED → ORDER_PAID → SUBSCRIPTION_ASSIGNED → MODEL_TOKENS_CREDITED → SUBSCRIPTION_SUCCESS.
6. Idempotency: re-run verify → still COMPLETED, balance unchanged.

---

## 7. Rollback

```bash
# Previous image tag is kept in docker images
docker images zcloud/zrouter --format '{{.Tag}} {{.ID}} {{.CreatedAt}}'
# Redeploy the previous tag:
#  (a) restore the previous source from /opt/zcloud/backups/ or git (uncommitted state is NOT in git)
#  (b) rebuild with the old tag: sed the ZCLOUD_VERSION / use the old image directly:
docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud up -d --force-recreate zrouter
```

Backups live in `/opt/zcloud/backups/` (preserved across syncs). Scripts: `deploy/backup.sh`, `deploy/restore.sh`, `deploy/rollback.sh`.

---

## 8. Troubleshooting

| Symptom | Check |
|---|---|
| Container unhealthy | `docker ps` → Restarts; `docker logs zcloud-zrouter --tail 100` |
| App unreachable from device | `tailscale status` on both ends; `ping 100.100.17.99`; UFW `ufw status verbose` (tailscale0 allow) |
| QRIS order stuck PENDING | `systemctl status gomerch gomerch-proxy`; `ss -ltnp | grep -E '30158|3015'`; `/opt/gomerch/data/payments.json` |
| DB errors | `docker exec zcloud-postgres psql -U zcloud -d sub2api -c '\dt'` |
| CORS/redirect issues | `CORS_ALLOWED_ORIGINS` in `deploy/.env.zcloud` must include https://app.zrouter.dev,https://api.zrouter.dev |

---

## 9. Security notes

- **Never** expose port 18080 publicly (no ufw rule, no nginx vhost, no Caddy).
- If public HTTPS is ever needed, use **Tailscale Serve/Funnel** — never a public reverse proxy.
- `DATABASE_PASSWORD` / `POSTGRES_PASSWORD` values contain `zcloud_staging_2026` — these are the **live credentials**; the string is historical, do NOT rename or it breaks the DB.
- Keep `.env.zcloud` (600) and `.env.zcloud.local` out of git. **Do not commit** unless the user explicitly asks.
- Support/contact: `support@zrouter.dev`.
