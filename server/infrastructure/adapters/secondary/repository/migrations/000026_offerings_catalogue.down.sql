DROP INDEX IF EXISTS idx_offerings_category;
DROP INDEX IF EXISTS idx_offerings_active_expires;
DROP INDEX IF EXISTS uq_offerings_farmer_product_window;

ALTER TABLE offerings DROP CONSTRAINT IF EXISTS ck_offerings_coordinates;

ALTER TABLE offerings
    DROP COLUMN IF EXISTS longitude,
    DROP COLUMN IF EXISTS latitude,
    DROP COLUMN IF EXISTS company_id,
    DROP COLUMN IF EXISTS category_id,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS quantity_available,
    DROP COLUMN IF EXISTS unit_of_measure_id,
    DROP COLUMN IF EXISTS variety;
