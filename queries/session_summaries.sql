-- name: ListSessionSummaries :many
SELECT id, campaign_id, from_turn, to_turn, summary, created_at
FROM session_summaries
WHERE campaign_id = $1
ORDER BY created_at DESC;

-- name: CreateSessionSummary :one
INSERT INTO session_summaries (campaign_id, from_turn, to_turn, summary)
VALUES ($1, $2, $3, $4)
RETURNING id, campaign_id, from_turn, to_turn, summary, created_at;
