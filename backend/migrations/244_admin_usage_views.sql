SELECT pg_advisory_xact_lock(244245);

CREATE INDEX IF NOT EXISTS idx_usage_logs_model_cost
    ON usage_logs (model, created_at)
    INCLUDE (display_total_cost, cost_total, cost_supplier_code, pricing_version);
