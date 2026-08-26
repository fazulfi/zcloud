SELECT pg_advisory_xact_lock(236237);

-- QRIS payment columns on payment_orders (zcloud flow)
ALTER TABLE payment_orders
    ADD COLUMN IF NOT EXISTS payment_method VARCHAR(20) NOT NULL DEFAULT 'qris',
    ADD COLUMN IF NOT EXISTS qris_payment_id UUID REFERENCES qris_payments(id),
    ADD COLUMN IF NOT EXISTS qris_payment_ref VARCHAR(100),
    ADD COLUMN IF NOT EXISTS qris_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS qris_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS plan_id_uuid UUID REFERENCES model_catalog(id),
    ADD COLUMN IF NOT EXISTS amount_idr BIGINT,
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(100),
    ADD COLUMN IF NOT EXISTS asset VARCHAR(20) NOT NULL DEFAULT 'IDR',
    ADD COLUMN IF NOT EXISTS network VARCHAR(20) NOT NULL DEFAULT 'QRIS';

CREATE UNIQUE INDEX IF NOT EXISTS uq_payment_orders_idempotency
    ON payment_orders (idempotency_key) WHERE idempotency_key IS NOT NULL;
