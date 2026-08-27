-- 241_usage_dual_metering.sql
-- M1.7: dual metering columns on usage_logs.
-- display columns are customer-visible (D8); cost columns are internal-only.
SELECT pg_advisory_xact_lock(240241);

ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS display_input_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS display_output_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS display_cache_read_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS display_cache_write_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS display_total_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS display_blend_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_input NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_output NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_cache_read NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_cache_write NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_total NUMERIC(20,8) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS cost_supplier_code VARCHAR(20) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS pricing_version INT NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS reservation_status VARCHAR(16) NOT NULL DEFAULT 'finalized';

-- Reservation ledger: atomic reserve/finalize/release by request_id (idempotent).
CREATE TABLE IF NOT EXISTS usage_reservations (
    id              BIGSERIAL PRIMARY KEY,
    request_id      VARCHAR(64)  NOT NULL,
    user_id         BIGINT       NOT NULL,
    api_key_id      BIGINT       NOT NULL,
    account_id      BIGINT       NOT NULL DEFAULT 0,
    model_id        VARCHAR(64)  NOT NULL DEFAULT '',
    model           VARCHAR(100) NOT NULL,
    fingerprint     VARCHAR(128) NOT NULL DEFAULT '',
    reserved_cost   NUMERIC(20,8) NOT NULL DEFAULT 0,
    status          VARCHAR(16)  NOT NULL DEFAULT 'pending',
    pricing_version INT          NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    CONSTRAINT usage_reservations_request_id_key UNIQUE (request_id)
);

CREATE INDEX IF NOT EXISTS idx_usage_reservations_user_created
    ON usage_reservations (user_id, created_at);
CREATE INDEX IF NOT EXISTS idx_usage_reservations_status
    ON usage_reservations (status);
