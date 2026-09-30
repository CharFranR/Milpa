DROP INDEX IF EXISTS uq_offerings_farmer_product_window;

CREATE UNIQUE INDEX uq_offerings_farmer_active_product
    ON offerings (user_id, name)
    WHERE is_active;
