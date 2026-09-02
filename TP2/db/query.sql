-- name: CreateUser :one
INSERT INTO users (name, email)
VALUES ($1, $2)
RETURNING id, name, email, created_at;


-- name: GetUserByID :one
SELECT id, name, email, created_at
FROM users
WHERE id = $1;


-- name: ListUsers :many
SELECT id, name, email, created_at
FROM users
ORDER BY id;


-- name: UpdateUser :one
UPDATE users
SET name = $1,
    email = $2
WHERE id = $3
RETURNING id, name, email, created_at;


-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;