ALTER TABLE categories ADD COLUMN default_expiry_days INT NULL;

ALTER TABLE categories
    ADD CONSTRAINT ck_categories_default_expiry_days
    CHECK (default_expiry_days IS NULL OR default_expiry_days > 0);
