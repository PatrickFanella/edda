-- name: ListJournalEntries :many
SELECT id, campaign_id, title, content, created_at, updated_at
FROM player_journal_entries
WHERE campaign_id = $1
ORDER BY created_at DESC;

-- name: CreateJournalEntry :one
INSERT INTO player_journal_entries (campaign_id, title, content)
VALUES ($1, $2, $3)
RETURNING id, campaign_id, title, content, created_at, updated_at;

-- name: DeleteJournalEntry :exec
DELETE FROM player_journal_entries WHERE id = $1;
