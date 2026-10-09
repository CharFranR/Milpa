ALTER TABLE liquidations ADD COLUMN assigned_buyer_id UUID NULL REFERENCES users(id);

CREATE TABLE liquidation_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    liquidation_id UUID NOT NULL REFERENCES liquidations(id),
    buyer_id UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT unique_liquidation_interest UNIQUE (liquidation_id, buyer_id)
);

CREATE INDEX idx_liquidation_interests_liquidation ON liquidation_interests(liquidation_id, created_at);
