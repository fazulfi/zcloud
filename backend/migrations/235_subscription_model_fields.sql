SELECT pg_advisory_xact_lock(234235);

ALTER TABLE user_subscriptions
    ADD COLUMN IF NOT EXISTS model_id UUID REFERENCES model_catalog(id),
    ADD COLUMN IF NOT EXISTS purchased_tokens BIGINT,
    ADD COLUMN IF NOT EXISTS token_expiry_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS credit_ledger_id UUID,
    ADD COLUMN IF NOT EXISTS plan_name_snapshot VARCHAR(200);
