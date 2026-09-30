-- RF-11 requires an offer to carry a price and optional comments, and an offer
-- had neither: the table stored the amount a supplier is willing to provide and
-- nothing about what that amount costs.
--
-- price_per_unit is NULLABLE ON PURPOSE, with no DEFAULT. A DEFAULT 0 would
-- rewrite the meaning of every offer that already exists in a deployed database:
-- a 0 is not "unranked", it is "the cheapest thing on the platform". NULL
-- carries the honest meaning -- the supplier has not quoted a price yet -- and a
-- supplier creating an offer now has to quote one, which the use case enforces.
--
-- The CHECK keeps the nullable branch honest for any path that skips the use
-- case. This mirrors ck_supplier_inventory_quantity from 000023.
ALTER TABLE supply_offers
    ADD COLUMN price_per_unit DOUBLE PRECISION,
    ADD COLUMN comments TEXT NOT NULL DEFAULT '',
    ADD CONSTRAINT ck_supply_offers_price CHECK (price_per_unit IS NULL OR price_per_unit > 0);
