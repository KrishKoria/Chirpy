-- name: CreateRefreshToken :exec
INSERT INTO refresh_tokens (
    token,
    created_at,
    updated_at,
    user_id,
    expires_at,
    revoked_at
) VALUES (
    $1, -- token
    $2, -- created_at
    $3, -- updated_at
    $4, -- user_id
    $5, -- expires_at
    $6  -- revoked_at (can be NULL)
);

-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens WHERE token = $1 LIMIT 1;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens
SET revoked_at = $2, updated_at = $3
WHERE token = $1;
