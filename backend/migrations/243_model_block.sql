SELECT pg_advisory_xact_lock(243245);

CREATE INDEX IF NOT EXISTS idx_model_balances_user_status
    ON model_balances(user_id, status);
