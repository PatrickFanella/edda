-- name: CreateUser :one
INSERT INTO users (
  name
) VALUES (
  $1
)
RETURNING *;

-- name: GetUserByID :one
SELECT *
FROM users
WHERE id = $1;

-- name: GetUserByName :one
SELECT *
FROM users
WHERE name = $1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY created_at, id;

-- name: UpdateUser :one
UPDATE users
SET
  name = $2,
  updated_at = now()
WHERE id = $1
RETURNING *;

-- name: CreateUserWithAuth :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at, updated_at;

-- name: GetUserByEmail :one
SELECT id, name, email, password_hash, created_at, updated_at
FROM users
WHERE email = $1;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
