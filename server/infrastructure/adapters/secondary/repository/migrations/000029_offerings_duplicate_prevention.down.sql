DROP INDEX IF EXISTS uq_offerings_farmer_active_product;

CREATE UNIQUE INDEX uq_offerings_farmer_product_window
    ON offerings (user_id, name, created_at)
    WHERE is_active;
