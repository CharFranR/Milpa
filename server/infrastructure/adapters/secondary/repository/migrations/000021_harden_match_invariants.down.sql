DROP INDEX IF EXISTS uq_matches_one_active_per_offer;
ALTER TABLE matches DROP CONSTRAINT IF EXISTS ck_matches_matched_amount;
ALTER TABLE supply_requests DROP CONSTRAINT IF EXISTS ck_supply_requests_amounts;
-- Restoring the original UNIQUE(supply_offer_id) requires deduplicating first; not automatic.
