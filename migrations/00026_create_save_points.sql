-- +goose Up
CREATE TABLE save_points (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  name TEXT NOT NULL DEFAULT '',
  turn_number INT NOT NULL DEFAULT 0,
  is_auto BOOLEAN NOT NULL DEFAULT false,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_save_points_campaign ON save_points(campaign_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS save_points;
