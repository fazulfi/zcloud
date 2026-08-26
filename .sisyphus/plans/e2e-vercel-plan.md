# zcloud/zrouter — End-to-End Implementation Plan (Frontend Terpisah di Vercel)

> **Keputusan user (m1028-m1029):** Frontend **TERPISAH** dari backend (bukan single-binary embed). Hosting: **Vercel**. Domain produksi: **sudah punya** (diisi `[DOMAIN]`, konfirmasi sebelum go-live).
> Base plan ini: spek pack `docs/spec/*` (kernel v0.11.2), F0.1-F0.5 done, M1.2-M1.4 data layer done (PR #5/#6 open).

---

## 0. Arsitektur Target (Frontend Terpisah)

```
┌─────────────────────┐        ┌──────────────────────┐
│  Vercel (Edge/Global)│        │  VPS SG (82.25.62.204)│
│  app.[DOMAIN]        │  HTTPS │  api.[DOMAIN]         │
│  Vue3+Vite static    │───────▶│  Caddy/nginx ────────▶│ 127.0.0.1:18080 (zrouter container)
│  dist/ (build)       │  CORS  │  + postgres:18 + redis│
│  vercel.json rewrite │        └──────────────────────┘
│  /api/* → api.[DOMAIN]│              ▲
└─────────────────────┘              │ localhost:3015
                                     └── gomerch (QRIS, VPS sama)
```

- **Frontend**: Vercel static (Vue3+Vite+TS+Tailwind). Build `pnpm build` → `dist/` (vite.config butuh outDir override). Env: `VITE_API_BASE_URL=https://api.[DOMAIN]`, `VITE_WS_BASE_URL=wss://api.[DOMAIN]`. Rewrite `/api/*` + `/v1/*` + `/setup/*` → `api.[DOMAIN]` (CORS di backend diizinkan origin `https://app.[DOMAIN]`).
- **Backend**: VPS SG, Docker compose (zrouter + postgres:18-alpine + redis + volume), Caddy/nginx TLS, port 18080/18443 internal. Staging: `zcloud.82.25.62.204.sslip.io` (nginx host, tanpa TLS public cert).
- **QRIS**: gomerch di VPS sama (`127.0.0.1:3015`), sub2api poll `/api/qris/status` ≥4s.
- **Arsitektur Ops**: monorepo tetap (backend/ + frontend/ + deploy/ + docs/), CI pada repo sama, Vercel auto-deploy dari `frontend/` (rootDirectory).

---

## 1. Fase A — Finalisasi YANG SEDANG BERJALAN (sebelum frontend split)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| A1 | Merge PR #5 (M1.3) | `gh pr merge 5 --squash --delete-branch` | merged, main = M1.3 |
| A2 | Merge PR #6 (M1.4) | `gh pr merge 6 --squash --delete-branch` | merged, main = M1.4 |
| A3 | Update todo (M1.3/M1.4 completed) | todowrite | todo konsisten |
| A4 | Rebase branch berikutnya ke main baru | feature branches (jika ada) | no drift |

---

## 2. Fase B — Frontend Split ke Vercel (F0.7)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| B1 | **Audit frontend** (agent) | Baca vite.config, package.json, src/api/client.ts, router, views, CSP, OAuth callbacks. Output: daftar perubahan yang dibutuhkan utk deploy Vercel (outDir, env, CORS, proxy, auth redirect). | Laporan audit |
| B2 | **vite.config.ts** — outDir `dist/` (bukan `../backend/internal/web/dist`) | `build.outDir: 'dist'`, `emptyOutDir: true`; pastikan `injectPublicSettings` dev-only (sudah). | build lokal OK |
| B3 | **vercel.json** (baru di frontend/) | `{ "rewrites": [{ "source": "/api/:path*", "destination": "https://api.[DOMAIN]/api/:path*" }, { "source": "/v1/:path*", ... }, { "source": "/setup/:path*", ... }] }` — proxy dev → prod API. | file ada |
| B4 | **Env**: frontend/.env.production | `VITE_API_BASE_URL=https://api.[DOMAIN]`, `VITE_WS_BASE_URL=wss://api.[DOMAIN]` (fallback `/api/v1` sudah ada di client.ts). | build pakai env |
| B5 | **CORS backend**: allow origin `https://app.[DOMAIN]` (config `cors.allowed_origins`) | Config + validasi existing cors config (F0.1 audit). | curl OPTIONS 200 |
| B6 | **OAuth redirect** (Google/GitHub) | Callback URL → `https://app.[DOMAIN]/callback/...` (bukan api). Cek views/auth/*CallbackView.vue. | redirect OK |
| B7 | **CI**: frontend job build `pnpm build` (outDir dist) + deploy | ci.yml frontend job verifikasi build (bukan ke backend dir). | CI frontend green |
| B8 | **Deploy Vercel** (manual/CLI) | Vercel project `zcloud-frontend`, rootDirectory `frontend/`, framework Vite, env production, domain `app.[DOMAIN]`. Deploy pertama via `vercel --prod` atau GitHub Integration (prefer: GitHub auto-deploy). | app.[DOMAIN] live |
| B9 | **Smoke test terpisah**: frontend di Vercel, backend di VPS | Load app.[DOMAIN] → login → key → usage. CORS + API OK. | smoke pass |
| B10 | Update deploy/Dockerfile + runbook | Dockerfile zrouter TANPA frontend embed (hapus `-tags embed`, hapus stage frontend); docs/runbooks/deployment.md + vercel section. | docs updated |

> **Catatan**: `deploy/Dockerfile` (yang direncanakan single-binary) TIDAK dibuat — backend standalone. `backend/internal/web/dist` tidak dipakai lagi (gitignore).

---

## 3. Fase C — M1.5: API Key Management (migration 239)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| C1 | Migration 239_api_key_scopes.sql | ALTER api_keys ADD scope VARCHAR(20) DEFAULT 'all_models' (all_models\|selected_models\|single_model), allowed_model_ids UUID[] DEFAULT '{}', ip_allowlist TEXT[] DEFAULT '{}', ip_denylist TEXT[] DEFAULT '{}', rate_limit_rpm INT DEFAULT 60, rate_limit_tpm INT DEFAULT 100000, is_zcloud BOOLEAN DEFAULT TRUE, metadata JSONB DEFAULT '{}'; + CHECK scope IN (...), CHECK (rate_limit_rpm > 0). Partial unique index: zcloud keys. | migration ada, konvensi |
| C2 | Ent schema api_key.go edit + regenerate | Field baru sesuai 239 (UUID array → ent `[]string` + SchemaType uuid array; pastikan ent dukung). | build/vet OK |
| C3 | **Handler API key management**: backend/internal/handler + service | CRUD key: create (pilih scope/model/allowed models/IP/rate), list, rotate, revoke, update scope. Zcloud path (key created via UI). | endpoint test |
| C4 | **Frontend**: halaman keys zcloud (Vercel) | Buat/lihat/rotate/revoke key + scope/IP/rate setting (per plan model yang dibeli). | UI functional |
| C5 | Validasi: key zcloud beda dari upstream (key lama tetap jalan utk legacy) | Prefix/flag is_zcloud; gateway bedakan. | test manual |

---

## 4. Fase D — M1.6: Gateway + Routing Supplier Termurah (migration 240)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| D1 | Migration 240_gateway_routes.sql | CREATE TABLE gateway_routes (id UUID PK, model_id UUID FK model_catalog, supplier_code VARCHAR(20), priority INT, enabled BOOLEAN DEFAULT TRUE, failover_order INT, cost_in NUMERIC(20,8), cost_out NUMERIC(20,8), effective_from/to, UNIQUE(model_id, supplier_code, priority)); seed dari supplier_pricing (route termurah per model). | migration + seed |
| D2 | Ent schema gateway_route.go + regenerate | Sesuai 240. | build/vet OK |
| D3 | **Service routing** (fail-closed): pilih supplier termurah per (model, region) | `zcloud/service/routing`: resolve model → supplier list (termurah dulu) → health check → pilih; fallback supplier berikutnya. **Fail-closed**: kalau semua mati → 503, jangan kirim ke upstream lain. | unit logic |
| D4 | **Integrasi gateway handler** | Di openai_chat_completions.go + gateway_handler_chat_completions.go: setelah ResolveChannelMappingAndRestrict → routing zcloud → ganti base_url/provider; sebelum SelectAccount*. Data: supplier config (base_url, api key) dari config/suppliers.* | integrasi |
| D5 | **Frontend**: model-plaza tampilkan harga route termurah (display blend) | Per model: harga resmi (model_pricing) + harga route (supplier_pricing). | UI price |

---

## 5. Fase E — M1.7: Dual Metering + Meter % per Model (migration 241-242)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| E1 | Migration 241_usage_ledger.sql | CREATE TABLE usage_ledger (id UUID PK, user_id BIGINT FK users, model_id UUID FK model_catalog, api_key_id BIGINT FK api_keys, supplier_code VARCHAR(20), input_tokens BIGINT, output_tokens BIGINT, cache_read_tokens BIGINT, cache_creation_tokens BIGINT, cost_usd NUMERIC(20,8), amount_idr NUMERIC(20,8), request_id VARCHAR(100), created_at TIMESTAMPTZ DEFAULT NOW()); index (user_id, model_id, created_at), (api_key_id, created_at). Append-only. | migration |
| E2 | Migration 242_usage_daily_rollup.sql | CREATE TABLE usage_daily (id UUID PK, user_id BIGINT, model_id UUID, date DATE, input_tokens BIGINT, output_tokens BIGINT, cache_read_tokens BIGINT, cache_creation_tokens BIGINT, cost_usd NUMERIC(20,8), amount_idr NUMERIC(20,8), UNIQUE(user_id, model_id, date)); upsert harian. | migration |
| E3 | Ent schemas usage_ledger.go + usage_daily.go + regenerate | Sesuai E1/E2. | build/vet OK |
| E4 | **Metering service**: dual meter (USD & IDR) | Saat request selesai (streaming/partial/failover): tulis usage_ledger + model_balances (tokens_consumed += ...; usage_percent = consumed × pct_per_1m / 1M; balance CHECK). Hitung cost: USD (rate supplier) + IDR (exchange rate config). | unit logic |
| E5 | **Integrasi**: hook ke usage-log path existing | Pakai usage_logs existing (input/output/cache_read/cache_creation) sebagai sumber; sinkronkan ke usage_ledger + model_balances (jangan double count). | integrasi |
| E6 | **Frontend**: usage dashboard (key-usage, usage) tampilkan per model + meter % | Chart token & cost per model + meter progress (0-100%). | UI |

---

## 6. Fase F — M1.8: Block per Model 402/403 (migration 243, gate G2)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| F1 | Migration 243_model_balance_blocks.sql | CREATE TABLE model_balance_blocks (id UUID PK, user_id BIGINT, model_id UUID FK model_catalog, reason VARCHAR(50) (balance_exhausted\|admin_block\|manual_review\|fraud), blocked_at TIMESTAMPTZ DEFAULT NOW(), unblocked_at TIMESTAMPTZ, UNIQUE(user_id, model_id, reason)); index (user_id, status-ish). | migration |
| F2 | Ent schema model_balance_block.go + regenerate | Sesuai F1. | build/vet OK |
| F3 | **Reseller admission check** (handler ChatCompletions) | SETELAH ResolveChannelMappingAndRestrict, SEBELUM SelectAccount*: cek model_balances(user, model): status blocked → 403 model_unavailable; balance <= 0 → **402 usage_cap_exhausted** (dengan detail model); token_expiry_at null (never). Sesuai audit gateway. | 402/403 benar |
| F4 | **Auto-block**: balance exhausted → set status blocked + reason balance_exhausted | Saat metering (E5) balance turun ke 0 → block. | auto-block |
| F5 | Admin unblock (manual review) | Endpoint admin: unblock, adjust balance (compensating). | endpoint |
| F6 | Frontend: notif + status model (blocked) di dashboard | Tampilkan model ter-block + alasan. | UI |

---

## 7. Fase G — M1.9: Admin Dashboard Owner-Only (migration 244)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| G1 | Migration 244_admin_audit.sql | CREATE TABLE admin_audit (id UUID PK, admin_user_id BIGINT, action VARCHAR(50), target_type VARCHAR(50), target_id VARCHAR(100), payload JSONB DEFAULT '{}', created_at); index admin_user_id + created_at. | migration |
| G2 | Ent schema admin_audit.go + regenerate | Sesuai G1. | build/vet OK |
| G3 | **Owner gate**: role check (1 owner) | Middleware owner-only untuk /api/v1/zcloud/admin/* (validasi user role; 1 owner dari M1.1/config). | 403 non-owner |
| G4 | **Admin endpoints**: users, balances, blocks, orders, keys, audit | List/search users, adjust balance (compensating), unblock, refund QRIS, audit trail. | endpoint test |
| G5 | **Frontend admin**: dashboard owner (Vercel) | Overview (revenue, users, models), users table, balance adjust, orders, blocks, audit. | UI functional |
| G6 | Admin audit log semua aksi | Setiap aksi admin → admin_audit. | audit trail |

---

## 8. Fase H — M1.10: Customer Dashboard (migration 266)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| H1 | Migration 266_customer_export.sql | CREATE TABLE data_export (id UUID PK, user_id BIGINT FK users, export_type VARCHAR(20) (usage\|orders\|keys\|all), status VARCHAR(20) (pending\|ready\|failed), file_path VARCHAR(255), requested_at, completed_at); index user_id. | migration |
| H2 | Ent schema data_export.go + regenerate | Sesuai H1. | build/vet OK |
| H3 | **Export service** (CSV/JSON) | Usage/orders/keys export → file (R2/S3 atau local volume) + link. | export |
| H4 | **Delete account** (GDPR) | Endpoint: hapus akun + data (anonymize/soft delete sesuai kebijakan) + revoke keys + cancel subs. | delete |
| H5 | **Frontend**: dashboard customer | Model-plaza (beli plan per model), usage per model, meter %, export data, delete account, orders/payment QRIS status. | UI functional |

---

## 9. Fase I — QRIS Payment Integration (gomerch, MENYELIPKAN sebelum M1.5 commit per keputusan user)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| I1 | **Config**: gomerch client | `zcloud/config` + env: `zcloud.gomerch.base_url=http://127.0.0.1:3015`, `gomerch_api_key` (secret), `status_poll_min_ms=4000`, `qris_expiry_sec=900`. | config |
| I2 | **Service**: zcloud/service/qris | create (POST /api/qris/create {amount, Idempotency-Key: order_id} → qr_string+image_base64+payment_ref+expires_at; simpan qris_payments + payment_orders), status poll (POST /api/qris/status {payment_ref} min 4s, throttle-aware; PAID→credited; EXPIRED→expired; AMBIGUOUS→review_required), reconcile (opsional, satu kali). | service logic |
| I3 | **Background worker** (poll loop) | Goroutine/worker: poll pending qris_payments tiap 5s (rate-limit 30/min di gomerch, jangan banjir) → update status → settle. Graceful shutdown. | worker |
| I4 | **Credit settlement** | PAID → create payment_orders paid → kredit model_balances.tokens_purchased (plan token = plan/2 model tokens) → usage_logs/subs. Idempotent (payment_ref unique + idempotency_key). | credit benar |
| I5 | **API**: POST /api/v1/zcloud/payment/qris/create + /status | Frontend call: create (amount, plan model) → QR display; poll status. | endpoint |
| I6 | **Frontend**: payment/qrcode page (Vercel) | Tampilkan QR (image_base64), countdown expiry 15m, poll status otomatis, sukses → redirect key/balance. | UI functional |
| I7 | **Security**: amount 1:1 IDR major, validasi, rate limit, anti-spoof | Amount integer IDR (bukan sen), verifikasi payment_ref milik user, jangan kredit sebelum PAID. | secure |

---

## 10. Fase J — M1.1: Auth/Roles (validasi existing + config)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| J1 | **Validasi auth existing** (agent) | Audit: register/login/forgot/OAuth (Google/GitHub/WeChat), JWT, middleware session, user table roles. Output: gap utk 1-owner + role. | laporan |
| J2 | **Owner provisioning** | Config/env `zcloud.owner_email` — user pertama dengan email tsb jadi owner (role=owner); register lain = customer. | owner |
| J3 | **Role enforcement** | Middleware: admin routes owner-only; customer routes login. | 403 |
| J4 | **Frontend**: login/register (Vercel) | Halaman auth existing (sudah ada) dipakai; pastikan OAuth redirect ke app.[DOMAIN]. | login OK |

---

## 11. Fase K — F0.6: Deployment Rehearsal VPS (staging)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| K1 | **Build image zrouter** di VPS | `docker compose -f deploy/docker-compose.zcloud.yml build` (backend only, tanpa frontend embed). | image build |
| K2 | **Env staging** | `.env.zcloud.example` → `.env` staging (api key gomerch, db, redis, owner email, CORS app.[DOMAIN]). | env |
| K3 | **Boot + migration** | `docker compose up -d` → migration 1-244 (via migration runner existing) → seed. | boot OK |
| K4 | **Smoke staging** | `curl /health/live`, `/health/ready`, `/metrics`, login, key create, model-plaza list, QRIS create (gomerch live). | smoke pass |
| K5 | **Rehearsal rollback** | Simulasi: stop container, restart, cek data. | recovery OK |

---

## 12. Fase L — M1.11: Test (49 skenario + 6 k6, gate G3)

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| L1 | **Integration tests** (49 skenario) | Sesuai docs/spec/13-test-plan.md: auth, keys, balances, metering, block, QRIS flow, admin. | 49 pass |
| L2 | **k6 load tests** (6 skenario) | Sesuai spec: concurrent key auth, gateway routing, metering throughput, QRIS create rate, dashboard read, mixed. | k6 pass |
| L3 | **Abuse/security test** | Rate limit key, IP allowlist, amount spoof, idempotency replay, QRIS double credit, 402/403. | secure |
| L4 | **Gate G3 review** | Semua test hijau + review. | gate pass |

---

## 13. Fase M — M1.12: Paid Beta + Produksi

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| M1 | **Domain finalisasi** | Konfirmasi [DOMAIN] user; DNS: `app.[DOMAIN]` → Vercel, `api.[DOMAIN]` → VPS (A record 82.25.62.204). | DNS |
| M2 | **TLS** | Caddy auto TLS (api.[DOMAIN]); Vercel managed (app.[DOMAIN]). | HTTPS |
| M3 | **Env produksi** | Vercel env (VITE_API_BASE_URL, VITE_WS_BASE_URL); backend env produksi (owner, gomerch key, CORS, exchange rate). | env |
| M4 | **Deploy produksi** | Vercel prod deploy + backend compose prod + migration + seed. | live |
| M5 | **Smoke produksi** | Health, login, beli plan (QRIS), key, call model, usage, block. | smoke pass |
| M6 | **Monitoring** | /metrics scrape (Prometheus/Grafana opsional), log, alerting basic (health check external). | monitoring |
| M7 | **Rollback plan** | Image tag lama + DB backup teruji (rehearsal K5). | rollback |

---

## 14. Fase N — Public Repo README + Dokumentasi

| # | Task | Detail | Kriteria Selesai |
|---|------|--------|------------------|
| N1 | README.md utama | Deskripsi zcloud/zrouter, fitur, arsitektur (Vercel+VPS), quickstart, env, deploy, kontribusi. | README |
| N2 | Docs update | docs/runbooks/deployment.md (Vercel + VPS), spec final, changelog. | docs |
| N3 | License/security | LICENSE (MIT?), SECURITY.md, code of conduct. | docs |

---

## 15. Dependensi & Urutan Kritis

```
A (merge #5/#6) → B (frontend split Vercel) ─┬─→ C (M1.5 keys)
                                              ├─→ D (M1.6 routing)
                                              ├─→ E (M1.7 metering)
                                              ├─→ F (M1.8 block)
                                              ├─→ I (QRIS payment) ← G (M1.9 admin)
                                              ├─→ J (M1.1 auth) ───────┘ (owner gate utk G)
                                              ├─→ H (M1.10 dashboard)
                                              └─→ K (rehearsal) → L (test) → M (prod) → N (docs)
```

**Kritis:**
- **G (M1.9 admin) butuh J (M1.1 owner gate)** — kerjakan J sebelum G.
- **F (M1.8 block) butuh E (M1.7 metering)** (auto-block saat balance 0).
- **E butuh D (routing)** (tahu supplier untuk cost).
- **D butuh C?** Tidak wajib — routing bisa jalan tanpa scope key; tapi C sebelum D memudahkan test key per model.
- **I (QRIS) independen** dari C-F — bisa paralel; tapi settlement kredit butuh model_balances (M1.3 ✓) + plan token (sudah di spek).
- **K (rehearsal) butuh B** (frontend prod siap) untuk smoke end-to-end staging.

## 16. Estimasi & Delegasi
- A: 10 menit (manual, orchestration).
- B: 1 agent visual-engineering + 1 agent deep (vite/vercel config) — 2-4 jam.
- C-F: tiap 1 agent deep (migration+ent+service) + 1 agent visual-engineering (frontend) — tiap 3-6 jam.
- I: 1 agent deep (service+worker) + 1 agent visual-engineering (QR page) — 4-8 jam (termasuk integrasi gomerch live test).
- J: 1 agent deep (audit+config) — 2-4 jam.
- G, H: tiap 1 agent deep + 1 visual — 4-6 jam.
- K: manual orchestration + 1 agent (ops) — 2-4 jam.
- L: 1 agent testing — 6-10 jam (49 skenario + k6).
- M: manual + 1 agent ops — 3-5 jam.
- N: 1 agent writing — 2-3 jam.

## 17. Risiko & Mitigasi
| Risiko | Mitigasi |
|--------|----------|
| Vercel CORS/rewrite misconfig | Smoke test B9 + CORS config eksplisit; fallback proxy rewrite. |
| vite outDir ke backend dir (embed) — conflict | B2 ubah ke dist/ + gitignore backend/internal/web/dist. |
| OAuth redirect salah (api vs app) | B6 audit + test. |
| gomerch ban risk / unofficial | Terima risiko (user setuju); polling ≥4s; rate limit client; kill-switch ENABLE_GOBIZ. |
| QRIS AMBIGUOUS | Status review_required + admin manual (F5/G). |
| Balance CHECK constraint salah saat concurrency | Version optimistic lock + retry; transaction serializable (spek). |
| CI PR #5/#6 fail lint | Fix gofmt/errcheck pola F0.5; tunggu CI sebelum merge. |

## 18. Status Tracker (akhir plan ini)
- ✅ Selesai: F0.1-F0.5, M1.2 (PR #4 merged), M1.3 (PR #5 open), M1.4 (PR #6 open)
- 🔄 Berjalan: PR #5/#6 CI
- ⏳ Berikutnya: A1-A2 (merge) → B (frontend split Vercel) → C-I (M1.5-M1.8 + QRIS) → J (auth) → G/H (admin+customer) → K (rehearsal) → L (test) → M (prod) → N (docs)
