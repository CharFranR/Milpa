CREATE TABLE supplier_inventory (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_name VARCHAR(255) NOT NULL,
    quantity     DOUBLE PRECISION NOT NULL,
    amount_unit  SMALLINT NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT uq_supplier_inventory_supplier_product UNIQUE (supplier_id, product_name),
    CONSTRAINT ck_supplier_inventory_quantity CHECK (quantity >= 0)
);
