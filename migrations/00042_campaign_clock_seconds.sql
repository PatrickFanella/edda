-- +goose Up
-- Preserve six-second combat rounds without rounding spell durations to minutes.
ALTER TABLE campaign_time ADD COLUMN second INT NOT NULL DEFAULT 0 CHECK (second BETWEEN 0 AND 59);

-- +goose Down
ALTER TABLE campaign_time DROP COLUMN second;
