CREATE TABLE conversations (
    id UUID PRIMARY KEY,
    buyer_id UUID NOT NULL REFERENCES users(id),
    farmer_id UUID NOT NULL REFERENCES users(id),
    offering_id UUID NOT NULL REFERENCES offerings(id),
    visibility BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    CONSTRAINT different_users CHECK (buyer_id <> farmer_id)
);