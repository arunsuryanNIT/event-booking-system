CREATE TABLE IF NOT EXISTS events (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title         VARCHAR(255) NOT NULL,
    description   TEXT,
    capacity      INT NOT NULL CHECK (capacity > 0),
    booked_count  INT NOT NULL DEFAULT 0 CHECK (booked_count >= 0),
    event_date    TIMESTAMPTZ NOT NULL,
    location      VARCHAR(255),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT booked_within_capacity CHECK (booked_count <= capacity)
);

CREATE INDEX IF NOT EXISTS idx_events_date ON events(event_date);
