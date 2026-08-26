CREATE TABLE model_catalog (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    canonical_name VARCHAR(100) UNIQUE NOT NULL,
    public_name VARCHAR(200) NOT NULL,
    context_window BIGINT,
    source_suppliers JSONB DEFAULT '[]',
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
