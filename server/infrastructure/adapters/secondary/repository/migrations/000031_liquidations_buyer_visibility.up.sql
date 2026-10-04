ALTER TABLE liquidations DROP CONSTRAINT valid_visibility;

UPDATE liquidations SET visibility = 'wholesale' WHERE visibility = 'private';

ALTER TABLE liquidations
    ADD CONSTRAINT valid_visibility
    CHECK (visibility IN ('public', 'wholesale', 'wholesale_retail', 'wholesale_corporate'));

ALTER TABLE liquidations DROP CONSTRAINT valid_allocation;

ALTER TABLE liquidations
    ADD CONSTRAINT valid_allocation
    CHECK (allocation_method IN ('manual', 'first_come'));

UPDATE liquidations SET closed_at = updated_at WHERE status <> 'open' AND closed_at IS NULL;
