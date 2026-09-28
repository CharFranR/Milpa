CREATE TABLE supply_offers (
    id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id          UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    supply_request_id    UUID NOT NULL REFERENCES supply_requests(id) ON DELETE CASCADE,
    total_amount         DOUBLE PRECISION NOT NULL,
    amount_unit          SMALLINT NOT NULL,
    proposed_delivery_day TIMESTAMPTZ NOT NULL,
    delivery_available   BOOLEAN NOT NULL DEFAULT FALSE,
    status               SMALLINT NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_supply_offers_supplier_request UNIQUE (supplier_id, supply_request_id)
);

CREATE INDEX idx_supply_offers_request ON supply_offers(supply_request_id);
