-- +goose Up
CREATE TABLE character_feats (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  character_id UUID NOT NULL REFERENCES player_characters(id) ON DELETE CASCADE,
  feat_id UUID NOT NULL REFERENCES feat_definitions(id) ON DELETE CASCADE,
  granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE(character_id, feat_id)
);

-- +goose Down
DROP TABLE IF EXISTS character_feats;
