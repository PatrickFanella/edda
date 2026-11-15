-- +goose Up
ALTER TABLE session_logs ADD CONSTRAINT session_logs_campaign_turn_unique UNIQUE (campaign_id, turn_number);

-- +goose Down
ALTER TABLE session_logs DROP CONSTRAINT session_logs_campaign_turn_unique;
