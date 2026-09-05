CREATE TABLE addresses (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    department   VARCHAR(254) NOT NULL,
    municipality VARCHAR(254) NOT NULL,
    address_line VARCHAR(2048) NOT NULL,
    latitude     DOUBLE PRECISION,
    longitude    DOUBLE PRECISION
);
