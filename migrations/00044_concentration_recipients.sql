-- +goose Up
CREATE TABLE concentration_casts (
 id uuid PRIMARY KEY,
 cast_order bigint GENERATED ALWAYS AS IDENTITY UNIQUE,
 campaign_id uuid NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
 caster_id uuid NOT NULL REFERENCES player_characters(id) ON DELETE CASCADE,
 started_at integer NOT NULL CHECK (started_at >= 0),
 expires_at integer NOT NULL CHECK (expires_at = started_at + 600)
);
CREATE TABLE concentration_recipients (
 cast_id uuid PRIMARY KEY REFERENCES concentration_casts(id) ON DELETE CASCADE,
 recipient_id uuid NOT NULL
);
CREATE INDEX concentration_casts_campaign ON concentration_casts(campaign_id, caster_id);
CREATE INDEX concentration_recipients_actor ON concentration_recipients(recipient_id);
-- +goose Down
DROP TABLE concentration_recipients;
DROP TABLE concentration_casts;
