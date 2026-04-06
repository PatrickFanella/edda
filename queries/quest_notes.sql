-- name: CreateQuestNote :one
INSERT INTO quest_notes (quest_id, content)
VALUES (sqlc.arg(quest_id), sqlc.arg(content))
RETURNING id, quest_id, content, created_at, updated_at;

-- name: ListQuestNotes :many
SELECT id, quest_id, content, created_at, updated_at
FROM quest_notes
WHERE quest_id = sqlc.arg(quest_id)
ORDER BY created_at DESC;

-- name: DeleteQuestNote :exec
DELETE FROM quest_notes
WHERE id = sqlc.arg(id) AND quest_id = sqlc.arg(quest_id);

-- name: CreateQuestHistoryEntry :one
INSERT INTO quest_history (quest_id, snapshot)
VALUES (sqlc.arg(quest_id), sqlc.arg(snapshot))
RETURNING id, quest_id, snapshot, created_at;

-- name: ListQuestHistory :many
SELECT id, quest_id, snapshot, created_at
FROM quest_history
WHERE quest_id = sqlc.arg(quest_id)
ORDER BY created_at DESC;
