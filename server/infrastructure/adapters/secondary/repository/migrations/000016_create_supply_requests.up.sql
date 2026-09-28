CREATE TABLE supply_requests (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    buyer_id                UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name            VARCHAR(255) NOT NULL,
    total_amount            DOUBLE PRECISION NOT NULL,
    actual_amount           DOUBLE PRECISION NOT NULL,
    amount_unit             SMALLINT NOT NULL,
    number_of_units         DOUBLE PRECISION NOT NULL,
    amount_per_unit         DOUBLE PRECISION NOT NULL,
    unit_of_measure         SMALLINT NOT NULL,
    address_id              UUID REFERENCES addresses(id) ON DELETE SET NULL,
    request_deadline        TIMESTAMPTZ NOT NULL,
    delivery_deadline       TIMESTAMPTZ NOT NULL,
    description             TEXT NOT NULL DEFAULT '',
    multiple_providers      BOOLEAN NOT NULL DEFAULT FALSE,
    min_amount_per_provider DOUBLE PRECISION NOT NULL DEFAULT 0,
    status                  SMALLINT NOT NULL DEFAULT 0,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_supply_requests_buyer_status ON supply_requests(buyer_id, status);
