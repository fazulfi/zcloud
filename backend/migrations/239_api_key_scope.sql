SELECT pg_advisory_xact_lock(239240);

CREATE TABLE IF NOT EXISTS api_key_model_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    model_id UUID NOT NULL REFERENCES model_catalog(id),
    rate_limit_per_min INT NOT NULL DEFAULT 60,
    rate_limit_per_hour INT NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(api_key_id, model_id),
    CHECK (rate_limit_per_min >= 0),
    CHECK (rate_limit_per_hour >= 0)
);

CREATE INDEX IF NOT EXISTS idx_api_key_model_scopes_api_key_id ON api_key_model_scopes(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_model_scopes_model_id ON api_key_model_scopes(model_id);

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS model_scope_mode VARCHAR(20) NOT NULL DEFAULT 'all'
    CHECK (model_scope_mode IN ('all', 'explicit'));

ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS enabled_models_count INT NOT NULL DEFAULT 0;
