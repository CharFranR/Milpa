-- ListOpen scans every open request for every supplier browsing the open
-- marketplace: it filters on status alone and sorts on created_at DESC.
--
-- The only index carrying status in a leading position was
-- (buyer_id, status), which a global status predicate cannot use, and nothing
-- indexed created_at, so every call both filtered and sorted with a sequential
-- scan. Leading with status and following with created_at DESC lets PostgreSQL
-- satisfy the filter and the sort from the index alone.
CREATE INDEX IF NOT EXISTS idx_supply_requests_status_created
    ON supply_requests (status, created_at DESC);
