-- +goose Up
-- +goose StatementBegin
CREATE TABLE events (
  id               UUID PRIMARY KEY DEFAULT uuidv4(),
  name             TEXT NOT NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE ticket_tiers (
  id            UUID PRIMARY KEY DEFAULT uuidv4(),
  event_id      UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  name          TEXT NOT NULL,
  price_cents   INTEGER NOT NULL CHECK (price_cents >= 0),
  UNIQUE (event_id, name)
);

CREATE TABLE tickets (
  id              UUID PRIMARY KEY DEFAULT uuidv4(),
  event_id        UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
  tier_id         UUID NOT NULL REFERENCES ticlet_tiers(id) ON DELETE RESTRICT,
  is_sold         BOOLEAN NOT NULL DEFAULT FALSE,
  hold_token      UUID,
  hold_expires_at TIMESTAMPTZ

  -- Extra constraint for hold rules
  CONSTRAINT chk_hold_valid CHECK (
    (hold_token IS NULL AND hold_expires_at IS NULL)
    OR (hold_token IS NOT NULL AND hold_expires_at IS NOT NULL)
  )
);

-- Useful indexes for performance.
-- Can be a hot path during peaks as is_sold and hold_expires_at freq updates.
-- But since we have at peak 500 concurrent users at (1k DAU) - should be cheap.
CREATE INDEX idx_tickets_tier_sold
  ON tickets (event_id, tier_id, is_sold);

CREATE INDEX idx_tickets_hold_expiry
  ON tickets (hold_expires_at);

CREATE INDEX idx_tickets_hold_token
  ON tickets (hold_token);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_tickets_hold_token;
DROP INDEX IF EXISTS idx_tickets_hold_expiry;
DROP INDEX IF EXISTS idx_tickets_tier_sold;

DROP TABLE IF EXISTS tickets;
DROP TABLE IF EXISTS ticket_tiers;
DROP TABLE IF EXISTS events;
-- +goose StatementEnd
