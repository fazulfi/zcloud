# M1.11 load and soak tests

These scripts target a staging gateway configured with a local supplier mock or replay fixture. **Never point them at InferHub** (the shared bucket is 30 requests/minute).

```sh
SUB2API_TARGET_URL=https://staging.example.test \
SUB2API_API_KEY=sk-test \
X_TEST_RUN_ID=m111-$(date +%s) \
k6 run backend/loadtest/m111_gateway.js
```

The default run covers ramp-up to 200 RPS, sustained 200 VUs, and a 400 RPS burst. For the milestone soak run, use `SOAK_DURATION=24h` (use `30m` for a preflight). The service should expose `X-Admission-Wait-Ms` and `X-Auth-Decision-Ms`; absent headers are recorded as zero so the script remains compatible with staging builds.

Every request carries `X-Test-Run-Id` and `X-Request-Id`; mutation iterations also carry `Idempotency-Key`. Required thresholds are encoded in the script: failed requests <1%, TTFT p95 <10s/p99 <15s, admission p99 <2s, and auth decision p95 <10ms.

For G3, additionally verify gateway overhead p95 <50ms excluding upstream time, 5xx <1%, no balance race/overspend, recovery after burst, no memory/connection leak during soak, and balance drift <0.1% from the seeded ledger.

Validate syntax without traffic with `k6 run --dry-run backend/loadtest/m111_gateway.js` when k6 is installed. The load test is intentionally not part of CI because it requires staging and a mock supplier.
