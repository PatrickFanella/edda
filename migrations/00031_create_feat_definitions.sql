-- +goose Up
CREATE TABLE feat_definitions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  prerequisites TEXT NOT NULL DEFAULT '',
  bonus_type TEXT NOT NULL DEFAULT '',
  bonus_value INT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_feat_definitions_campaign ON feat_definitions(campaign_id);

-- +goose Down
DROP TABLE IF EXISTS feat_definitions;
