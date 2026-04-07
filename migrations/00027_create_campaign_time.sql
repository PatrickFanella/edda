-- +goose Up
CREATE TABLE campaign_time (
  campaign_id UUID PRIMARY KEY REFERENCES campaigns(id) ON DELETE CASCADE,
  day INT NOT NULL DEFAULT 1,
  hour INT NOT NULL DEFAULT 8,
  minute INT NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS campaign_time;
