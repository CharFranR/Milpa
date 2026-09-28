-- Active matches are unique per offer; cancelled matches are retained for history.
ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_supply_offer_id_key;

CREATE UNIQUE INDEX uq_matches_one_active_per_offer
    ON matches (supply_offer_id)
    WHERE status = 0;          -- MatchActive

ALTER TABLE supply_requests
    ADD CONSTRAINT ck_supply_requests_amounts
    CHECK (total_amount >= 0 AND actual_amount >= 0 AND actual_amount <= total_amount);

ALTER TABLE matches
    ADD CONSTRAINT ck_matches_matched_amount
    CHECK (matched_amount >= 0);
