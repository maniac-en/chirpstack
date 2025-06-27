-- name: StoreRefreshToken :one
INSERT INTO refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
VALUES (
    $1,
    NOW(),
    NOW(),
    $2,
    date_add(NOW(), '60 days'),
    NULL
    )
RETURNING *;

-- name: GetUserFromRefreshToken :one
SELECT token, created_at, updated_at, user_id, expires_at, revoked_at
FROM refresh_tokens
WHERE 1=1
AND token = $1
AND revoked_at IS NULL;

-- name: RevokeRefreshToken :one
UPDATE refresh_tokens
SET
    revoked_at = NOW(),
    updated_at = NOW()
WHERE token = $1
RETURNING *;
