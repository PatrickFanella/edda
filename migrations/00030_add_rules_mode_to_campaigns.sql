-- +goose Up
ALTER TABLE campaigns ADD COLUMN rules_mode TEXT NOT NULL DEFAULT 'narrative';

-- +goose Down
ALTER TABLE campaigns DROP COLUMN IF EXISTS rules_mode;
