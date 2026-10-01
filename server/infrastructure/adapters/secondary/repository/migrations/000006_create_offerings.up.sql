CREATE TABLE offerings (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id             UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type                SMALLINT NOT NULL,
    name                VARCHAR(40) NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    price               DOUBLE PRECISION NOT NULL DEFAULT 0,
    image_url           VARCHAR(2048) NOT NULL DEFAULT '',
    created_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    variety             VARCHAR(60) NOT NULL DEFAULT '',
    unit_of_measure_id  UUID REFERENCES units_of_measure(id) ON DELETE SET NULL,
    quantity_available  DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    expires_at          TIMESTAMPTZ,
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,
    category_id         UUID REFERENCES categories(id) ON DELETE SET NULL,
    company_id          UUID REFERENCES companies(id) ON DELETE SET NULL,
    latitude            DOUBLE PRECISION,
    longitude           DOUBLE PRECISION,

    CONSTRAINT ck_offerings_coordinates CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    )
);

CREATE UNIQUE INDEX uq_offerings_farmer_active_product
    ON offerings (user_id, name)
    WHERE is_active;

CREATE INDEX idx_offerings_active_expires ON offerings (is_active, expires_at);
CREATE INDEX idx_offerings_category ON offerings (category_id);
