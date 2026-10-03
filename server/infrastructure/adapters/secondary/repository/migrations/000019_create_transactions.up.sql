CREATE TABLE transactions (
    id                           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    match_id                     UUID NOT NULL UNIQUE REFERENCES matches(id) ON DELETE CASCADE,
    status                       SMALLINT NOT NULL DEFAULT 0,
    buyer_start_confirmed_at     TIMESTAMPTZ,
    supplier_start_confirmed_at  TIMESTAMPTZ,
    buyer_delivery_confirmed_at  TIMESTAMPTZ,
    supplier_delivery_confirmed_at TIMESTAMPTZ,
    cancelled_by                 UUID REFERENCES users(id) ON DELETE SET NULL,
    cancel_reason                TEXT NOT NULL DEFAULT '',
    created_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
