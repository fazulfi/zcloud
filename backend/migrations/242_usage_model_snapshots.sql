-- 242_usage_model_snapshots.sql
-- M1.7: per-model usage meter snapshots (kernel §5).
-- display$ columns are customer-visible (D8); cost$ columns are internal-only.
SELECT pg_advisory_xact_lock(242243);

CREATE TABLE IF NOT EXISTS usage_model_snapshots (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    model VARCHAR(100) NOT NULL,
    pricing_version INT NOT NULL,
    display_input_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    display_output_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    display_cache_read_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    display_cache_write_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    display_total_cost NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_input NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_output NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_cache_read NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_cache_write NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_total NUMERIC(20,8) NOT NULL DEFAULT 0,
    cost_supplier_code VARCHAR(20) NOT NULL DEFAULT '',
    input_tokens BIGINT NOT NULL DEFAULT 0,
    output_tokens BIGINT NOT NULL DEFAULT 0,
    cache_read_tokens BIGINT NOT NULL DEFAULT 0,
    cache_write_tokens BIGINT NOT NULL DEFAULT 0,
    usage_model_pct NUMERIC(10,6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_usage_model_snapshots_user_model
    ON usage_model_snapshots (user_id, model, pricing_version);
CREATE INDEX IF NOT EXISTS idx_usage_model_snapshots_created
    ON usage_model_snapshots (created_at);
