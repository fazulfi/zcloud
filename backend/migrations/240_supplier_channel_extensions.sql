SELECT pg_advisory_xact_lock(240241);

-- M1.6: supplier routing support index.
-- Accelerates the active-pricing lookup used by cheapest-healthy-supplier routing:
--   model_id + availability='active' + effective_from<=now + (effective_to IS NULL OR effective_to>now)
CREATE INDEX IF NOT EXISTS idx_supplier_pricing_active_lookup
    ON supplier_pricing (model_id, availability, effective_from, effective_to);

-- Optional operator override: NULL means the channel is unrestricted.
-- Not used by default cheapest routing; reserved for explicit channel-level
-- supplier pinning as a future operator override.
ALTER TABLE channels
    ADD COLUMN IF NOT EXISTS supplier_code VARCHAR(20) DEFAULT NULL;

CREATE INDEX IF NOT EXISTS idx_channels_supplier_code
    ON channels (supplier_code) WHERE supplier_code IS NOT NULL;

-- Guard: if a channel carries a supplier code it must be a recognized code.
ALTER TABLE channels
    DROP CONSTRAINT IF EXISTS channels_supplier_code_check;

ALTER TABLE channels
    ADD CONSTRAINT channels_supplier_code_check
    CHECK (supplier_code IS NULL OR supplier_code IN ('cb', 'cbcn', 'cx'));
