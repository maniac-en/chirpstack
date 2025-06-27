-- name: CreateChirp :one
INSERT INTO chirps (id, created_at, updated_at, body, user_id)
VALUES (
    gen_random_uuid(),
    NOW(),
    NOW(),
    $1,
    $2
)
RETURNING *;

-- name: GetChirps :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
ORDER BY created_at ASC;

-- name: GetChirpByID :one
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE id = $1;

-- name: DeleteChirpByID :exec
DELETE FROM chirps
WHERE id = $1;

-- name: GetChirpsByAuthorID :many
SELECT id, created_at, updated_at, body, user_id
FROM chirps
WHERE user_id = $1
ORDER BY created_at ASC;
