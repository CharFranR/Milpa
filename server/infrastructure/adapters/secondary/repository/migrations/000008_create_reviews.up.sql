CREATE TABLE reviews (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id   UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    company_id  UUID REFERENCES companies(id) ON DELETE CASCADE,
    target_type  TEXT NOT NULL DEFAULT 'company',
    target_id   UUID NOT NULL,
    rating      SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    comment     TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    transaction_id UUID NOT NULL,

    CONSTRAINT ck_reviews_target_type CHECK (target_type IN ('company', 'user')),
    CONSTRAINT ck_reviews_company_mirror CHECK (target_type <> 'company' OR (company_id IS NOT NULL AND target_id = company_id)),
    CONSTRAINT uq_reviews_author_target UNIQUE (author_id, target_type, target_id),
    CONSTRAINT uq_reviews_transaction_author UNIQUE (transaction_id, author_id)
);

CREATE INDEX idx_reviews_target_rating ON reviews (target_type, target_id) INCLUDE (rating);
