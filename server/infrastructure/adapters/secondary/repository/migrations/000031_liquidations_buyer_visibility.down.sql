ALTER TABLE liquidations DROP CONSTRAINT valid_visibility;

UPDATE liquidations SET visibility = 'private' WHERE visibility <> 'public';

ALTER TABLE liquidations
    ADD CONSTRAINT valid_visibility
    CHECK (visibility IN ('public', 'private'));

ALTER TABLE liquidations DROP CONSTRAINT valid_allocation;

UPDATE liquidations SET allocation_method = 'manual' WHERE allocation_method <> 'manual';

ALTER TABLE liquidations
    ADD CONSTRAINT valid_allocation
    CHECK (allocation_method IN ('manual'));
