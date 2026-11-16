-- +goose Up
CREATE TABLE tactical_mind_windows (
 actor_id uuid PRIMARY KEY REFERENCES player_characters(id) ON DELETE CASCADE,
 owner_id uuid NOT NULL REFERENCES users(id),
 campaign_id uuid NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
 check_id uuid NOT NULL,
 base_receipt jsonb NOT NULL,
 difficulty integer NOT NULL CHECK (difficulty > 0),
 revision text NOT NULL,
 history_revision text NOT NULL,
 turn_number integer NOT NULL,
 resolved boolean NOT NULL DEFAULT false
);

-- +goose Down
DROP TABLE tactical_mind_windows;
