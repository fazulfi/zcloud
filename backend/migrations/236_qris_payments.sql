SELECT pg_advisory_xact_lock(236237);

CREATE TABLE IF NOT EXISTS qris_payments (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         BIGINT NOT NULL,
    amount_idr      BIGINT NOT NULL,
    payment_ref     VARCHAR(100) NOT NULL,
    qr_string       TEXT NOT NULL,
    image_base64    TEXT NOT NULL,
    expires_at      TIMESTAMPTZ NOT NULL,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    gomerch_payload JSONB DEFAULT '{}',
    idempotency_key VARCHAR(100),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_qris_payments_payment_ref UNIQUE (payment_ref),
    CONSTRAINT uq_qris_payments_idempotency UNIQUE (idempotency_key),
    CHECK (status IN ('pending', 'paid', 'expired', 'review_required', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_qris_payments_user_id ON qris_payments (user_id);
CREATE INDEX IF NOT EXISTS idx_qris_payments_status ON qris_payments (status);
