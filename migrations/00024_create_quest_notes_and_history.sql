-- +goose Up
CREATE TABLE quest_notes (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  content TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quest_notes_quest_id ON quest_notes(quest_id);

CREATE TABLE quest_history (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  quest_id UUID NOT NULL REFERENCES quests(id) ON DELETE CASCADE,
  snapshot TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_quest_history_quest_id ON quest_history(quest_id);

-- +goose Down
DROP TABLE IF EXISTS quest_history;
DROP TABLE IF EXISTS quest_notes;
