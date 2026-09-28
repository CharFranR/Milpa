CREATE TABLE matches (
    id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supply_offer_id    UUID NOT NULL UNIQUE REFERENCES supply_offers(id) ON DELETE CASCADE,
    supply_request_id  UUID NOT NULL REFERENCES supply_requests(id) ON DELETE CASCADE,
    status             SMALLINT NOT NULL DEFAULT 0,
    matched_amount     DOUBLE PRECISION NOT NULL,
    amount_unit        SMALLINT NOT NULL,
    created_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at         TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_matches_request_status ON matches(supply_request_id, status);
