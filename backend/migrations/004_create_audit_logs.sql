CREATE TABLE IF NOT EXISTS audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    operation       VARCHAR(20) NOT NULL CHECK (operation IN ('book', 'cancel')),
    event_id        UUID NOT NULL REFERENCES events(id),
    user_id         UUID NOT NULL REFERENCES users(id),
    booking_id      UUID,
    outcome         VARCHAR(20) NOT NULL CHECK (outcome IN ('success', 'failure')),
    failure_reason  VARCHAR(255),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_event ON audit_logs(event_id);
CREATE INDEX IF NOT EXISTS idx_audit_user ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_created ON audit_logs(created_at DESC);
