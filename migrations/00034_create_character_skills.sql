-- +goose Up
CREATE TABLE character_skills (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  character_id UUID NOT NULL REFERENCES player_characters(id) ON DELETE CASCADE,
  skill_id UUID NOT NULL REFERENCES skill_definitions(id) ON DELETE CASCADE,
  points INT NOT NULL DEFAULT 0,
  UNIQUE(character_id, skill_id)
);

-- +goose Down
DROP TABLE IF EXISTS character_skills;
