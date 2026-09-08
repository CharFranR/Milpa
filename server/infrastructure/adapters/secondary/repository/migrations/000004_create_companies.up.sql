CREATE TABLE companies (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name         VARCHAR(40) NOT NULL,
    owner_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    address_id   UUID REFERENCES addresses(id) ON DELETE SET NULL,
    description  TEXT NOT NULL DEFAULT '',
    phone_number VARCHAR(20) NOT NULL,
    email        VARCHAR(254) NOT NULL UNIQUE,
    website      VARCHAR(254) NOT NULL DEFAULT '',
    verified     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
