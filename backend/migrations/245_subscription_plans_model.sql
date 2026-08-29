SELECT pg_advisory_xact_lock(245246);

ALTER TABLE subscription_plans ADD COLUMN IF NOT EXISTS model_id UUID REFERENCES model_catalog(id);
CREATE INDEX IF NOT EXISTS idx_subscription_plans_model_id ON subscription_plans(model_id);
