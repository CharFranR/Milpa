ALTER TABLE offerings
    ADD COLUMN variety VARCHAR(60) NOT NULL DEFAULT '',
    ADD COLUMN unit_of_measure_id UUID REFERENCES units_of_measure(id) ON DELETE SET NULL,
    ADD COLUMN quantity_available DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (quantity_available >= 0),
    ADD COLUMN expires_at TIMESTAMPTZ,
    ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
    ADD COLUMN company_id UUID REFERENCES companies(id) ON DELETE SET NULL,
    ADD COLUMN latitude DOUBLE PRECISION,
    ADD COLUMN longitude DOUBLE PRECISION,
    ADD CONSTRAINT ck_offerings_coordinates CHECK (
        (latitude IS NULL AND longitude IS NULL) OR
        (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
    );

CREATE UNIQUE INDEX uq_offerings_farmer_product_window
    ON offerings (user_id, name, created_at)
    WHERE is_active;

CREATE INDEX idx_offerings_active_expires ON offerings (is_active, expires_at);
CREATE INDEX idx_offerings_category ON offerings (category_id);
