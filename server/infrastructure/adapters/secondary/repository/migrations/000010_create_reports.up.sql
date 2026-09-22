CREATE TYPE report_target_type AS ENUM ('offering', 'user');
CREATE TYPE report_status AS ENUM ('pending', 'approved', 'rejected');

CREATE TABLE reports (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    reporter_id UUID NOT NULL REFERENCES users(id),
    target_type report_target_type NOT NULL,
    target_id   UUID NOT NULL,
    reason      VARCHAR(500) NOT NULL,
    status      report_status NOT NULL DEFAULT 'pending',
    resolved_by UUID REFERENCES users(id),
    resolved_at TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_reports_pending_unique
    ON reports (reporter_id, target_type, target_id)
    WHERE status = 'pending';

CREATE INDEX idx_reports_status ON reports (status);
CREATE INDEX idx_reports_target ON reports (target_type, target_id);
