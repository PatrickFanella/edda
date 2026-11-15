-- +goose Up
-- No character conversion: existing campaigns keep the legacy rules engine.
ALTER TABLE campaigns ADD COLUMN rules_version TEXT NOT NULL DEFAULT 'legacy'
  CHECK (rules_version IN ('legacy', 'srd-5.2.1'));

-- +goose Down
ALTER TABLE campaigns DROP COLUMN rules_version;
