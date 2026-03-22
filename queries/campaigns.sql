-- name: CreateCampaign :one
INSERT INTO campaigns (
  name,
  description,
  genre,
  tone,
  themes,
  status,
  created_by
) VALUES (
  sqlc.arg(name),
  sqlc.arg(description),
  sqlc.arg(genre),
  sqlc.arg(tone),
  COALESCE(sqlc.narg(themes)::text[], '{}'::text[]),
  sqlc.arg(status),
  sqlc.arg(created_by)
)
RETURNING id, name, description, genre, tone, themes, status, created_by, created_at, updated_at;

-- name: GetCampaignByID :one
SELECT id, name, description, genre, tone, themes, status, created_by, created_at, updated_at
FROM campaigns
WHERE id = sqlc.arg(id);

-- name: ListCampaignsByUser :many
SELECT id, name, description, genre, tone, themes, status, created_by, created_at, updated_at
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
  updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, name, description, genre, tone, themes, status, created_by, created_at, updated_at;

-- name: UpdateCampaignStatus :one
UPDATE campaigns
SET
  status = sqlc.arg(status),
  updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, name, description, genre, tone, themes, status, created_by, created_at, updated_at;

-- name: DeleteCampaign :exec
DELETE FROM campaigns
WHERE id = sqlc.arg(id);
