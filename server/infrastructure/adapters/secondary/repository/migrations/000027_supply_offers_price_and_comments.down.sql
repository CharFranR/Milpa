ALTER TABLE supply_offers
    DROP CONSTRAINT IF EXISTS ck_supply_offers_price;

-- The CHECK has to go before the column it guards.
ALTER TABLE supply_offers
    DROP COLUMN IF EXISTS comments,
    DROP COLUMN IF EXISTS price_per_unit;
