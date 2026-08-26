SELECT pg_advisory_xact_lock(234235);

CREATE TABLE IF NOT EXISTS model_balances (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES model_catalog(id),
    tokens_purchased BIGINT NOT NULL DEFAULT 0,
    tokens_consumed BIGINT NOT NULL DEFAULT 0,
    balance BIGINT NOT NULL DEFAULT 0,
    usage_percent NUMERIC(20, 8) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    version BIGINT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, model_id),
    CHECK (balance = tokens_purchased - tokens_consumed),
    CHECK (status IN ('active', 'blocked', 'not_purchased'))
);

CREATE INDEX IF NOT EXISTS idx_model_balances_user_id ON model_balances(user_id);
CREATE INDEX IF NOT EXISTS idx_model_balances_status ON model_balances(status);
