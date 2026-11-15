-- +goose Up
CREATE TABLE turn_attempts (
 id UUID PRIMARY KEY,
 campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
 turn_number INTEGER NOT NULL CHECK (turn_number > 0),
 player_input TEXT NOT NULL,
 status TEXT NOT NULL CHECK (status IN ('pending_mechanics','pending_narration','completed','failed')),
 calls JSONB NOT NULL DEFAULT '[]',
 candidate_narrative TEXT NOT NULL DEFAULT '',
 final_narrative TEXT NOT NULL DEFAULT '',
 final_result JSONB,
 recovered BOOLEAN NOT NULL DEFAULT false,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX turn_attempts_active_number ON turn_attempts(campaign_id,turn_number) WHERE status <> 'failed';
CREATE UNIQUE INDEX turn_attempts_one_pending ON turn_attempts(campaign_id) WHERE status IN ('pending_mechanics','pending_narration');
-- +goose Down
DROP TABLE turn_attempts;
