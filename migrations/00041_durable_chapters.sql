-- +goose Up
-- Legacy session_summaries and session_logs remain untouched.
CREATE TABLE chapters (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
 from_turn INTEGER NOT NULL CHECK (from_turn > 0),
 to_turn INTEGER NOT NULL CHECK (to_turn >= from_turn),
 source_hash TEXT NOT NULL,
 source_turns JSONB NOT NULL CHECK (jsonb_typeof(source_turns) = 'array'),
 knowledge_boundary TEXT NOT NULL,
 rules_version TEXT NOT NULL,
 rules_mode TEXT NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 UNIQUE (campaign_id, source_hash),
 UNIQUE (id, campaign_id)
);
CREATE TABLE chapter_revisions (
 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 chapter_id UUID NOT NULL,
 campaign_id UUID NOT NULL,
 request_id UUID NOT NULL,
 revision INTEGER NOT NULL CHECK (revision > 0),
 status TEXT NOT NULL CHECK (status IN ('generating', 'completed', 'failed')),
 content TEXT NOT NULL DEFAULT '',
 error TEXT NOT NULL DEFAULT '',
 prompt_version TEXT NOT NULL,
 lease_token UUID NOT NULL,
 lease_until TIMESTAMPTZ NOT NULL,
 created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
 FOREIGN KEY (chapter_id, campaign_id) REFERENCES chapters(id, campaign_id) ON DELETE CASCADE,
 UNIQUE (campaign_id, request_id),
 UNIQUE (chapter_id, revision)
);
-- +goose Down
DROP TABLE chapter_revisions;
DROP TABLE chapters;
