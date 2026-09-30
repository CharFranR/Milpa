ALTER TABLE categories DROP CONSTRAINT IF EXISTS uq_categories_name;

ALTER TABLE categories
    DROP COLUMN IF EXISTS default_unit_of_measure_id,
    DROP COLUMN IF EXISTS is_active,
    DROP COLUMN IF EXISTS main_category;
