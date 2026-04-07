-- name: CreateSavePoint :one
INSERT INTO save_points (campaign_id, name, turn_number, is_auto)
VALUES ($1, $2, $3, $4)
RETURNING id, campaign_id, name, turn_number, is_auto, created_at;

-- name: ListSavePointsByCampaign :many
SELECT id, campaign_id, name, turn_number, is_auto, created_at
FROM save_points
WHERE campaign_id = $1
ORDER BY created_at DESC;

-- name: DeleteSavePoint :exec
DELETE FROM save_points WHERE id = $1;

-- name: DeleteOldAutoSaves :exec
DELETE FROM save_points
WHERE campaign_id = $1
  AND is_auto = true
  AND id NOT IN (
    SELECT id FROM save_points
    WHERE campaign_id = $1 AND is_auto = true
    ORDER BY created_at DESC
    LIMIT 3
  );
