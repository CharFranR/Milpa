CREATE TABLE liquidations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    supplier_id UUID NOT NULL REFERENCES users(id),
    product_name VARCHAR(255) NOT NULL,
    quantity DECIMAL(10,2) NOT NULL,
    unit_of_measure VARCHAR(50) NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    unit_price DECIMAL(10,2) NOT NULL,
    delivery_time VARCHAR(100),
    location_id UUID NOT NULL,
    visibility VARCHAR(20) NOT NULL DEFAULT 'public',
    allocation_method VARCHAR(20) NOT NULL DEFAULT 'manual',
    status VARCHAR(20) NOT NULL DEFAULT 'open',
    closed_at TIMESTAMP,
    expires_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT valid_status CHECK (status IN ('open', 'closed', 'expired', 'assigned')),
    CONSTRAINT valid_visibility CHECK (visibility IN ('public', 'private')),
    CONSTRAINT valid_allocation CHECK (allocation_method IN ('manual'))
);

CREATE INDEX idx_liquidations_supplier ON liquidations(supplier_id);
CREATE INDEX idx_liquidations_status ON liquidations(status);
