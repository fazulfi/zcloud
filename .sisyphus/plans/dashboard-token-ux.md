# Plan: Customer Dashboard Overhaul — zcloud Model-Token Experience

**Project**: zcloud/zrouter (sub2api fork) · **Repo**: `C:\Users\faizz\new-saas\sub2api` · **Branch base**: `feat/branding-zcloud` (797d9b4bd)
**Goal**: Align customer-facing UI with the zcloud business model — sell MODELs via TOKENs (plans $1/$2/$5/$10 per model), remove currency-"saldo" framing, remove "per platform" grouping, hide non-MVP sub2api features (D22), and make the model catalog the center of the customer experience.
**Date**: 2026-08-28 · **Author**: Sisyphus · **Review**: Momus

---

## 0. Evidence (audits m0024 + b4)

Five user complaints mapped to root causes:

| # | Complaint | Root cause (file:line) |
|---|---|---|
| 1 | "Dashboard masih subdomain api" (default sub2api feel) | Generic "Model balances" title, $ currency cards, default layout. `api_base_url` shown only on Keys page (legit), NOT on dashboard. |
| 2 | "Dimana katalog produknya" | No catalog in dashboard; Model Plaza is a price catalog with NO purchase CTA; PaymentView plans bind to group, not model. |
| 3 | "Kenapa ada saldo padahal sistem token" | `UserDashboardStats.vue:4-18` Balance card `$` (user.balance USD); `AppHeader.vue` header balance `$`; Profile/Payment/Redeem/KeyUsage all `$`. |
| 4 | "Redeem nggak guna" | `AppSidebar.vue:708-719` static nav item, no feature flag; QuickActions `addBalanceWithCode`; route `/redeem` open. |
| 5 | "Per platform maksudnya apa? kita cuma jual model" | `UserDashboardStats.vue:135-222` platform breakdown + USD quotas (data: stats.by_platform + GET /user/platform-quotas). |

---

## 1. Design Invariants (from spec, non-negotiable)

- **D22**: hide non-MVP features (redeem, promo, affiliate, announcement, batch image) via config/flag + route guard; **never delete code/views/APIs**.
- **D6/D7**: per-model token balance; blocked per model; token no-expiry.
- **D8**: $ prices are display layer only; supplier prices never shown. Plans $1/$2/$5/$10 per model.
- **Clean light design** (11-product-surface.md): white bg, #2563EB primary, Inter, radius 10px, WCAG AA.
- **Terminology** (§8): "saldo token per model", "% per 1M token", meter per model. NEVER global wallet, prepaid, reset, recurring.
- Existing backend APIs (GET /user/model-balances, GET /model-plaza, GET /payment/checkout-info) are used as-is; **NO backend changes in this milestone** (per-model purchase order binding is a known gap — see §6 Out of Scope).
- Preserve upstream-diff surface: prefer flags/settings and component-level changes over rewriting shared code.

---

## 2. Work Breakdown (delegatable units)

### W1 — Dashboard overhaul: "Models" first (DashboardView.vue + new component)
**Files**: `frontend/src/views/user/DashboardView.vue`, NEW `frontend/src/components/user/dashboard/ModelCatalogSection.vue` (or rework of inline :6-16 block), `frontend/src/components/user/dashboard/UserDashboardStats.vue`.

1. Replace generic "Model balances" section with **Model Catalog hero section**:
   - Each model card: model name (monospace), status badge (`active`/`blocked`/`not_purchased`), token balance (`formatTokens` K/M, floor-rounded per user rule "token bulatkan kebawah"), usage %, progress bar.
   - **Add purchase CTA per card**: if backend checkout-info exposes plans, deep-link `/purchase?tab=subscription&model=<name>` (fallback: `/purchase`); if model has no balance → "Buy plan" primary CTA; if active → "Top up" secondary CTA + "Use model" link to /keys.
   - Fetch model display metadata (tier token allocation per $1/$2/$5/$10) from Model Plaza API if already enabled; if not available, show plan prices statically from checkout-info plans (no hard failure if APIs empty → graceful empty state with CTA to /purchase).
2. `UserDashboardStats.vue`:
   - **Remove/hide Balance `$` card (:4-18)** — replace with "Active models" count card (active/blocked/not_purchased counts from model-balances).
   - **Hide platform breakdown (:135-222)** behind `isSimple`-style internal flag: render only when `publicSettings.show_platform_breakdown === true` (default false; add to PublicSettings type default in stores/app.ts:349-374). Keep Today Cost card but relabel to muted "est. cost" secondary row (D8: display layer) — move under token stats, not a headline card.
3. `DashboardView.vue`: skip `loadPlatformQuotas()` call when breakdown hidden (avoid useless request).

### W2 — QuickActions + header token framing
**Files**: `frontend/src/components/user/dashboard/UserDashboardQuickActions.vue`, `frontend/src/components/layout/AppHeader.vue`, `frontend/src/utils/format.ts`.

1. QuickActions: remove "Redeem Code" action; replace with "Buy model tokens" → /purchase (primary) and "Model catalog" → /model-plaza (if enabled).
2. AppHeader desktop (:57-101) + mobile (:142-151): replace $ available/frozen/total balance block with **total token balance across models** (sum of /user/model-balances balances; fetch once in a shared store or reuse dashboard cache — implement lightweight module-level cache to avoid refetch per route). Fallback: hide the block while loading. Keep frozen/currency values OUT of the header entirely.
   - New util in format.ts: `formatTokenBalance(n)` → compact K/M/B, floor rounding.

### W3 — Hide redeem + non-MVP nav (D22)
**Files**: `frontend/src/components/layout/AppSidebar.vue`, `frontend/src/router/index.ts`, `frontend/src/composables/useRoutePrefetch.ts`, `frontend/src/stores/featureFlags.ts` (or wherever registry lives), i18n en/zh.

1. Add feature flag `redeem_enabled` (default **false**) to featureFlags registry + publicSettings defaults (stores/app.ts).
2. AppSidebar: wrap redeem item (:708-719) in flag; also gate `batch-image` and `affiliate` items behind their existing flags if not already (verify each).
3. Router: `/redeem` guard redirects to /dashboard when flag false (keep route defined). Same pattern as existing simple-mode guard (:959-974).
4. useRoutePrefetch: exclude /redeem when flag false.
5. Admin pages/controls untouched.
6. Backend flag plumbing: check whether `redeem_enabled` needs a new OEM setting key in `backend/internal/service/domain_constants.go` OEM block + `settings/public` response + setting_parse.go default false (small, safe addition following existing key patterns; NO migrations — settings table is KV).

### W4 — Cross-page token copy consistency (customer surfaces only)
**Files**: `frontend/src/views/user/PaymentView.vue` (:42 Current Balance → keep but relabel "Account credit" in muted style OR hide if balance_disabled), `views/user/RedeemView.vue` (leave — hidden), `components/user/profile/ProfileInfoCard.vue` (:63-70 Account Balance → "Account credit"), `components/user/profile/ProfileBalanceNotifyCard.vue` (hide when currency wallet unused — gate behind same platform-breakdown-style flag), i18n `frontend/src/i18n/locales/en/dashboard.ts`, `frontend/src/i18n/locales/zh/dashboard.ts`, `frontend/src/i18n/locales/en/common.ts`, `frontend/src/i18n/locales/zh/common.ts`.

1. i18n: dashboard strings → token framing ("Model tokens", "Tokens remaining", "Buy plan"); common balance strings stay for account credit (internal recharge still exists) but are no longer shown as headline.
2. Do NOT touch admin i18n, provider accounts, affiliate internals.

### W5 — Build + deploy + visual verification
1. Local: `pnpm -C frontend typecheck` + `pnpm -C frontend build`; `go build ./...` if W3 touches backend.
2. Per-unit QA (run BEFORE W5, evidence = tool output):
   - **W1**: `pnpm -C frontend typecheck` clean; local dev server → Playwright: dashboard renders model cards (token count, status badge, CTA present for both "has balance" and "no balance" states); platform breakdown section absent; no `$`-headline card. Empty-state: with 0 model balances, catalog shows graceful empty CTA to /purchase.
   - **W2**: Playwright: header shows total token count (not `$`); with model-balances API failing/slow → block hidden (no layout break, no NaN). QuickActions: no "Redeem Code", has "Buy model tokens".
   - **W3**: with flag off: sidebar has no redeem item; navigating to /redeem redirects to /dashboard; prefetch does not request redeem route; with flag on (dev toggle) item reappears. Admin redeem page still reachable for admin.
   - **W4**: grep customer i18n files for leftover headline "Balance"/"余额" framing in changed keys; Profile/Payment show "Account credit" copy; `pnpm -C frontend build` succeeds (i18n key integrity).
3. Commit on `feat/branding-zcloud`.
4. **Deploy (existing procedure, verbatim)**:
   - Sync: `tar --exclude=.git --exclude=node_modules --exclude=frontend/dist --exclude=frontend/vendor -czf - . | ssh faiz-prod "mkdir -p /opt/zcloud/build && tar -xzf - -C /opt/zcloud/build"` (docs/legal/ MUST be included — required by Dockerfile COPY).
   - Build on VPS: `ssh faiz-prod "cd /opt/zcloud/build && docker build --build-arg FRONTEND_MODE=production --build-arg VERSION=0.1.185-zcloud -f deploy/Dockerfile -t zcloud/zrouter:0.1.185-zcloud-prod . > /tmp/zcloud-build.log 2>&1 && tail -5 /tmp/zcloud-build.log"`.
   - Deploy: `ssh faiz-prod "cd /opt/zcloud/deploy && ZCLOUD_VERSION=0.1.185-zcloud-prod docker compose -f docker-compose.zcloud.yml --env-file .env.zcloud up -d --no-build zrouter"`.
   - Verify: `curl -s https://app.zrouter.dev | head -c 200` + Playwright at https://app.zrouter.dev: dashboard model catalog w/ token framing, no $ headline card, no platform breakdown, no redeem in sidebar, catalog CTAs present; login page still branded zcloud.

---

## 3. Execution Order & Delegation

- W1, W2, W3, W4 are **largely independent files** → delegate in parallel (4 × deep/unspecified-high agents) after this review. Shared touchpoints to coordinate: stores/app.ts defaults (W1+W3), format.ts (W2 only), i18n dashboard.ts (W1+W4). Mitigation: W1 owns stores/app.ts + dashboard i18n; W4 owns only profile/payment i18n keys; W2 owns format.ts.
- W5 sequential after merge of all.

**Success criteria (E2E)**:
- [ ] Dashboard: model catalog cards with token balance + status + buy CTA; zero `$`-headline cards; no platform breakdown; no redeem quick action.
- [ ] Sidebar: no redeem item; route /redeem redirects when flag off; admin redeem unaffected.
- [ ] Header: token total (or hidden), no USD saldo.
- [ ] All changed files typecheck-clean; no `as any`/`@ts-ignore`.
- [ ] Prod visual verification passes via Playwright.

## 4. Risks
- Header token total needs aggregation — mitigate with module cache + graceful hide on failure.
- Backend flag addition must follow existing OEM key pattern exactly (setting_parse.go defaults + public settings serialization).
- Upstream merge hygiene: all hides are flag-gated, no deletions.

## 5. Out of Scope (explicit)
- Per-model ORDER binding in payment backend (model→order→balance write path) — real purchase flow backend; next milestone.
- Subscription/plan data model changes; pricing engine changes.
- Admin UI changes; affiliate/promo internals; landing page rebuild (separate).
- Deploying already-built 0.1.184 image (it will be superseded by 0.1.185; only deploy as interim if user asks).
