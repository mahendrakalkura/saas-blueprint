-- name: CreateSession :one
INSERT INTO sessions (
  id, user_id, refresh_token, refresh_token_expires_at, user_agent, ip_address, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: GetSessionByRefreshToken :one
SELECT * FROM sessions
WHERE refresh_token = $1 AND is_revoked = false;

-- name: GetSessionByID :one
SELECT * FROM sessions
WHERE id = $1;

-- name: RevokeSession :exec
UPDATE sessions
SET is_revoked = true, updated_at = $2
WHERE refresh_token = $1;

-- name: RevokeAllUserSessions :exec
UPDATE sessions
SET is_revoked = true, updated_at = $2
WHERE user_id = $1 AND is_revoked = false;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions
WHERE refresh_token_expires_at < $1;

-- name: ListUserSessions :many
SELECT * FROM sessions
WHERE user_id = $1 AND is_revoked = false
ORDER BY created_at DESC;

-- name: CountActiveSessions :one
SELECT COUNT(*) FROM sessions
WHERE user_id = $1 AND is_revoked = false;
