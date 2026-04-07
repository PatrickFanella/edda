-- +goose Up
CREATE TABLE session_summaries (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
  from_turn INT NOT NULL,
  to_turn INT NOT NULL,
  summary TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_session_summaries_campaign ON session_summaries(campaign_id, created_at DESC);

-- +goose Down
DROP TABLE IF EXISTS session_summaries;
