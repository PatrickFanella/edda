-- name: CreateCampaign :one
INSERT INTO campaigns (
  name,
  description,
  genre,
  tone,
  themes,
  world_type,
  danger_level,
  political_complexity,
  status,
  created_by
) VALUES (
  sqlc.arg(name),
  sqlc.arg(description),
  sqlc.arg(genre),
  sqlc.arg(tone),
  COALESCE(sqlc.narg(themes)::text[], '{}'::text[]),
  sqlc.arg(world_type),
  sqlc.arg(danger_level),
  sqlc.arg(political_complexity),
  sqlc.arg(status),
  sqlc.arg(created_by)
)
RETURNING *;

-- name: GetCampaignByID :one
SELECT *
FROM campaigns
WHERE id = sqlc.arg(id);

-- name: ListCampaignsByUser :many
SELECT *
FROM campaigns
WHERE created_by = sqlc.arg(created_by)
ORDER BY created_at, id;

-- name: UpdateCampaign :one
UPDATE campaigns
SET
  name = sqlc.arg(name),
  description = sqlc.arg(description),
  genre = sqlc.arg(genre),
  tone = sqlc.arg(tone),
  themes = COALESCE(sqlc.narg(themes)::text[], '{}'::text[]),
  world_type = sqlc.arg(world_type),
  danger_level = sqlc.arg(danger_level),
  political_complexity = sqlc.arg(political_complexity),
  updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET
  status = sqlc.arg(status),
  updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING *;

-- name: DeleteCampaign :exec
DELETE FROM campaigns
WHERE id = sqlc.arg(id);
