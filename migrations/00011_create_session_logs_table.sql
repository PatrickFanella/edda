-- +goose Up
CREATE TABLE session_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE RESTRICT,
  turn_number INTEGER NOT NULL,
  player_input TEXT NOT NULL,
  input_type TEXT NOT NULL CHECK (input_type IN ('game_action', 'meta', 'narrative')),
  llm_response TEXT NOT NULL,
  tool_calls JSONB NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(tool_calls) = 'array'),
  location_id UUID REFERENCES locations(id) ON DELETE RESTRICT,
  npcs_involved UUID[] NOT NULL DEFAULT '{}',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_session_logs_campaign_id ON session_logs(campaign_id);
CREATE INDEX idx_session_logs_location_id ON session_logs(location_id);
CREATE INDEX idx_session_logs_campaign_turn_number ON session_logs(campaign_id, turn_number);

-- +goose Down
DROP TABLE IF EXISTS session_logs;
