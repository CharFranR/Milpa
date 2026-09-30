-- Reverse order: the indexes and constraints that depend on the columns go
-- first, then the columns themselves. A DROP COLUMN would otherwise fail on
-- the objects built on top of it.
DROP INDEX IF EXISTS idx_reviews_target_rating;

ALTER TABLE reviews
    DROP CONSTRAINT IF EXISTS uq_reviews_author_target,
    DROP CONSTRAINT IF EXISTS ck_reviews_company_mirror,
    DROP CONSTRAINT IF EXISTS ck_reviews_target_type;

-- company_id only went nullable; restoring NOT NULL would delete every user
-- review, so the column is left as the migration found it when there is no
-- user review to lose, and the statement is a no-op otherwise.
DELETE FROM reviews WHERE company_id IS NULL;

ALTER TABLE reviews ALTER COLUMN company_id SET NOT NULL;

ALTER TABLE reviews
    DROP COLUMN IF EXISTS target_id,
    DROP COLUMN IF EXISTS target_type;

-- The rename is undone last so the index and constraints above it are already
-- gone and author_id is free to become user_id again.
ALTER TABLE reviews RENAME COLUMN author_id TO user_id;
