CREATE TABLE offerings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    company_id  UUID NOT NULL REFERENCES companies(id) ON DELETE CASCADE,
    type        SMALLINT NOT NULL,
    name        VARCHAR(40) NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    price       DOUBLE PRECISION NOT NULL DEFAULT 0,
    image_url   VARCHAR(2048) NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
