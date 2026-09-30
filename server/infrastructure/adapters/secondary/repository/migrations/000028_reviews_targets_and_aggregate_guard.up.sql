-- RF-15 asks for reviews of both counterparties. The table could only express
-- one: company_id was NOT NULL and there was no way to say who a review was
-- about other than which company owned them, so a farmer could not be reviewed
-- at all.
--
-- The shape is Amazon's: an author, a target type, a target id. Anything else --
-- a transaction id, a per-type nullable column set -- re-introduces the same
-- one-target-per-row problem the triple is meant to remove.
--
-- user_id becomes author_id. A rename, not a second column: they hold the same
-- value and two columns holding the same thing is a question nobody can answer
-- later. The Go field is AuthorID for the same reason; the DTO keeps its
-- published user_id, because that is the API contract.
--
-- The backfill is not a guess. Every existing review is, by construction, a
-- review of a company -- that was the only thing the column could mean.
ALTER TABLE reviews RENAME COLUMN user_id TO author_id;

ALTER TABLE reviews ADD COLUMN target_type TEXT NOT NULL DEFAULT 'company';

-- No foreign key, and that is the price of the triple rather than an oversight.
-- target_id is polymorphic: on a company review it names a row in companies and
-- on a user review a row in users, and SQL cannot express one FK over two
-- tables. A FK to users(id) would reject every company review, and a pair of
-- nullable per-type columns with a CHECK is the schema this migration exists to
-- replace.
--
-- Referential integrity is not lost, only relocated. For a company review
-- ck_reviews_company_mirror ties target_id to company_id, which keeps its own FK
-- to companies(id), so the company case is still fully enforced -- and it is the
-- case the listing and the aggregate depend on. A user review's target is
-- enforced by the application, and a rating nobody can collect is a rating the
-- aggregate reports as belonging to a target that no longer exists.
ALTER TABLE reviews ADD COLUMN target_id UUID;

UPDATE reviews SET target_type = 'company', target_id = company_id;

-- The backfill left no NULL, so NOT NULL can be asserted rather than assumed.
-- The DEFAULT goes with it: one would let a future INSERT skip the target and
-- create a company review with no company, which is the row the check below
-- exists to reject.
ALTER TABLE reviews ALTER COLUMN target_id SET NOT NULL;
ALTER TABLE reviews ALTER COLUMN target_type DROP DEFAULT;

ALTER TABLE reviews ADD CONSTRAINT ck_reviews_target_type
    CHECK (target_type IN ('company', 'user'));

-- A user review has no company to mirror, so the mirror has to be able to be
-- absent. Dropping NOT NULL loses nothing: no row written before this migration
-- is a user review.
ALTER TABLE reviews ALTER COLUMN company_id DROP NOT NULL;

-- Mandatory, not advisory: on a company review target_id and company_id are the
-- same company, and on a user review there is no company to mirror. This is what
-- keeps FindByCompany and the aggregate cheap.
ALTER TABLE reviews ADD CONSTRAINT ck_reviews_company_mirror
    CHECK (target_type <> 'company' OR (company_id IS NOT NULL AND target_id = company_id));

-- Without this a buyer could post the same rating twice and move the average
-- without anybody having changed their mind.
ALTER TABLE reviews ADD CONSTRAINT uq_reviews_author_target UNIQUE (author_id, target_type, target_id);

-- The aggregate is AVG(rating) grouped by target, read on every profile view and
-- written on every review. Its leading column is the group key and rating is
-- carried along, so the aggregate is answered from the index.
CREATE INDEX idx_reviews_target_rating ON reviews (target_type, target_id) INCLUDE (rating);
