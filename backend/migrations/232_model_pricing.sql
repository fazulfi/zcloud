-- Versioned official display pricing snapshots. Snapshots are immutable; add a new version instead of updating.
CREATE TABLE model_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES model_catalog(id),
    version INT NOT NULL,
    input_rate NUMERIC(20,8),
    output_rate NUMERIC(20,8),
    cached_read_rate NUMERIC(20,8),
    cached_write_rate NUMERIC(20,8),
    context_tier VARCHAR(20),
    tokens_per_dollar BIGINT,
    pct_per_1m_tokens NUMERIC(10,6),
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    source_ref VARCHAR(100),
    UNIQUE(model_id, version)
);
