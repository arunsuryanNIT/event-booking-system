CREATE TABLE IF NOT EXISTS bookings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    UUID NOT NULL REFERENCES events(id) ON DELETE RESTRICT,
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    status      VARCHAR(20) NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'cancelled')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_active_booking
    ON bookings (event_id, user_id)
    WHERE (status = 'active');

CREATE INDEX IF NOT EXISTS idx_bookings_user_status
    ON bookings (user_id, status);

CREATE INDEX IF NOT EXISTS idx_bookings_event_status
    ON bookings (event_id, status);
