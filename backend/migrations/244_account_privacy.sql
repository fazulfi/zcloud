SELECT pg_advisory_xact_lock(244247);

ALTER TABLE users ADD COLUMN IF NOT EXISTS anonymized_at TIMESTAMPTZ;
CREATE INDEX IF NOT EXISTS idx_usage_logs_user_created_at ON usage_logs (user_id, created_at);
