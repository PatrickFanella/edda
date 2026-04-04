-- +goose Up
ALTER TABLE campaigns ADD COLUMN world_type TEXT;
ALTER TABLE campaigns ADD COLUMN danger_level TEXT;
ALTER TABLE campaigns ADD COLUMN political_complexity TEXT;

-- +goose Down
ALTER TABLE campaigns DROP COLUMN IF EXISTS political_complexity;
ALTER TABLE campaigns DROP COLUMN IF EXISTS danger_level;
ALTER TABLE campaigns DROP COLUMN IF EXISTS world_type;
