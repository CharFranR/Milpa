-- Quantity is a reserved amount: it must never be negative. The entity rejects
-- negatives in Go, but that leaves the invariant unprotected against a direct
-- write or a future migration. This mirrors the discipline already applied to
-- supply_requests.actual_amount and matches.matched_amount in 000021.
ALTER TABLE supplier_inventory
    ADD CONSTRAINT ck_supplier_inventory_quantity
    CHECK (quantity >= 0);
