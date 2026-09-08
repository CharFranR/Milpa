CREATE TABLE users (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email         VARCHAR(254) NOT NULL UNIQUE,
    first_name    VARCHAR(40) NOT NULL,
    last_name     VARCHAR(40) NOT NULL,
    role          SMALLINT NOT NULL,
    password_hash VARCHAR(254) NOT NULL,
    phone_number  VARCHAR(20) NOT NULL,
    address_id    UUID REFERENCES addresses(id) ON DELETE SET NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
