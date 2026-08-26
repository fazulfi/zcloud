-- Internal supplier route pricing. This data is never exposed to customers.
CREATE TABLE supplier_pricing (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    model_id UUID REFERENCES model_catalog(id),
    supplier_code VARCHAR(20) NOT NULL,
    version INT NOT NULL,
    tier_label VARCHAR(50),
    availability VARCHAR(20),
    input_rate NUMERIC(20,8),
    output_rate NUMERIC(20,8),
    cached_read_rate NUMERIC(20,8),
    cached_write_rate NUMERIC(20,8),
    cache_capabilities JSONB DEFAULT '{}',
    effective_from TIMESTAMPTZ,
    effective_to TIMESTAMPTZ,
    UNIQUE(model_id, supplier_code, version, tier_label)
);

-- Seed the canonical InferHub catalog and its version-1 pricing snapshots.
WITH models(canonical_name, public_name, context_window, source_suppliers) AS (
    VALUES
        ('gpt-5.6-luna', 'GPT-5.6 Luna', 400000000, '["cx"]'::jsonb),
        ('deepseek-v4-flash', 'DeepSeek V4 Flash', 400000000, '["cbcn"]'::jsonb),
        ('gpt-5.4-mini', 'GPT-5.4 Mini', 100000000, '["cx"]'::jsonb),
        ('minimax-m3', 'MiniMax M3', 100000000, '["cb"]'::jsonb),
        ('minimax-m2.7', 'MiniMax M2.7', 100000000, '["cbcn"]'::jsonb),
        ('gpt-5.6-terra', 'GPT-5.6 Terra', 60000000, '["cx"]'::jsonb),
        ('deepseek-v4-pro', 'DeepSeek V4 Pro', 40000000, '["cbcn"]'::jsonb),
        ('gpt-5.4', 'GPT-5.4', 30000000, '["cx"]'::jsonb),
        ('glm-5.2', 'GLM-5.2', 30000000, '["cb"]'::jsonb),
        ('glm-5.3', 'GLM-5.3', 30000000, '["cb"]'::jsonb),
        ('kimi-k2.6', 'Kimi K2.6', 30000000, '["cbcn"]'::jsonb),
        ('kimi-k2.7', 'Kimi K2.7', 30000000, '["cbcn"]'::jsonb),
        ('gemini-3.1-pro', 'Gemini 3.1 Pro', 20000000, '["cb"]'::jsonb),
        ('gpt-5.6-sol', 'GPT-5.6 Sol', 20000000, '["cx"]'::jsonb),
        ('kimi-k3', 'Kimi K3', 20000000, '["cb"]'::jsonb),
        ('gpt-5.5', 'GPT-5.5', 20000000, '["cx"]'::jsonb),
        ('gpt-5.3-codex', 'GPT-5.3 Codex', 20000000, '["cb"]'::jsonb)
)
INSERT INTO model_catalog (canonical_name, public_name, context_window, source_suppliers)
SELECT canonical_name, public_name, context_window, source_suppliers
FROM models
WHERE NOT EXISTS (SELECT 1 FROM model_catalog existing WHERE existing.canonical_name = models.canonical_name);

WITH pricing(canonical_name, input_rate, output_rate, tokens_per_dollar, pct_per_1m_tokens) AS (
    VALUES
        ('gpt-5.6-luna', 0.20::numeric, 1.20::numeric, 400000000, 0.250000::numeric),
        ('deepseek-v4-flash', 0.22::numeric, 0.66::numeric, 400000000, 0.250000::numeric),
        ('gpt-5.4-mini', 0.75::numeric, 4.50::numeric, 100000000, 1.000000::numeric),
        ('minimax-m3', 0.30::numeric, 1.20::numeric, 100000000, 1.000000::numeric),
        ('minimax-m2.7', 0.30::numeric, 1.20::numeric, 100000000, 1.000000::numeric),
        ('gpt-5.6-terra', 2.00::numeric, 12.00::numeric, 60000000, 1.666667::numeric),
        ('deepseek-v4-pro', 0.66::numeric, 1.98::numeric, 40000000, 2.500000::numeric),
        ('gpt-5.4', 2.50::numeric, 15.00::numeric, 30000000, 3.333333::numeric),
        ('glm-5.2', 1.40::numeric, 4.40::numeric, 30000000, 3.333333::numeric),
        ('glm-5.3', 1.40::numeric, 4.40::numeric, 30000000, 3.333333::numeric),
        ('kimi-k2.6', 0.95::numeric, 4.00::numeric, 30000000, 3.333333::numeric),
        ('kimi-k2.7', 0.95::numeric, 4.00::numeric, 30000000, 3.333333::numeric),
        ('gemini-3.1-pro', 2.00::numeric, 12.00::numeric, 20000000, 5.000000::numeric),
        ('gpt-5.6-sol', 5.00::numeric, 30.00::numeric, 20000000, 5.000000::numeric),
        ('kimi-k3', 3.00::numeric, 15.00::numeric, 20000000, 5.000000::numeric),
        ('gpt-5.5', 5.00::numeric, 30.00::numeric, 20000000, 5.000000::numeric),
        ('gpt-5.3-codex', 1.75::numeric, 14.00::numeric, 20000000, 5.000000::numeric)
)
INSERT INTO model_pricing (model_id, version, input_rate, output_rate, tokens_per_dollar, pct_per_1m_tokens, effective_from, source_ref)
SELECT catalog.id, 1, pricing.input_rate, pricing.output_rate, pricing.tokens_per_dollar, pricing.pct_per_1m_tokens, NOW(), 'InferHub'
FROM pricing
JOIN model_catalog catalog ON catalog.canonical_name = pricing.canonical_name
WHERE NOT EXISTS (
    SELECT 1 FROM model_pricing existing
    WHERE existing.model_id = catalog.id AND existing.version = 1
);

WITH suppliers(canonical_name, supplier_code) AS (
    VALUES
        ('gpt-5.6-luna', 'cx'), ('deepseek-v4-flash', 'cbcn'), ('gpt-5.4-mini', 'cx'),
        ('minimax-m3', 'cb'), ('minimax-m2.7', 'cbcn'), ('gpt-5.6-terra', 'cx'),
        ('deepseek-v4-pro', 'cbcn'), ('gpt-5.4', 'cx'), ('glm-5.2', 'cb'),
        ('glm-5.3', 'cb'), ('kimi-k2.6', 'cbcn'), ('kimi-k2.7', 'cbcn'),
        ('gemini-3.1-pro', 'cb'), ('gpt-5.6-sol', 'cx'), ('kimi-k3', 'cb'),
        ('gpt-5.5', 'cx'), ('gpt-5.3-codex', 'cb')
)
INSERT INTO supplier_pricing (model_id, supplier_code, version, availability, cache_capabilities, effective_from)
SELECT catalog.id, suppliers.supplier_code, 1, 'active', '{"cache_read": true, "cache_write": true}'::jsonb, NOW()
FROM suppliers
JOIN model_catalog catalog ON catalog.canonical_name = suppliers.canonical_name
WHERE NOT EXISTS (
    SELECT 1 FROM supplier_pricing existing
    WHERE existing.model_id = catalog.id
      AND existing.supplier_code = suppliers.supplier_code
      AND existing.version = 1
      AND existing.tier_label IS NULL
);
