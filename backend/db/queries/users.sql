-- name: CreateUser :one
INSERT INTO users (
  id, email, password_hash, first_name, last_name, email_verified, is_active, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1 AND deleted_at IS NULL;

-- name: UpdateUser :one
UPDATE users
SET email = $2, first_name = $3, last_name = $4, avatar_url = $5,
    email_verified = $6, updated_at = $7
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users
SET password_hash = $2, updated_at = $3
WHERE id = $1 AND deleted_at IS NULL;

-- name: SetEmailVerificationToken :exec
UPDATE users
SET email_verification_token = $2, email_verification_expires_at = $3, updated_at = $4
WHERE id = $1 AND deleted_at IS NULL;

-- name: VerifyEmail :exec
UPDATE users
SET email_verified = true, email_verification_token = NULL,
    email_verification_expires_at = NULL, updated_at = $1
WHERE email_verification_token = $2
  AND email_verification_expires_at > $1
  AND deleted_at IS NULL;

-- name: SetPasswordResetToken :exec
UPDATE users
SET password_reset_token = $2, password_reset_expires_at = $3, updated_at = $4
WHERE email = $1 AND deleted_at IS NULL;

-- name: ResetPassword :exec
UPDATE users
SET password_hash = $2, password_reset_token = NULL,
    password_reset_expires_at = NULL, updated_at = $1
WHERE password_reset_token = $3
  AND password_reset_expires_at > $1
  AND deleted_at IS NULL;

-- name: DeactivateUser :exec
UPDATE users
SET is_active = false, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: DeleteUser :exec
UPDATE users
SET deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUsers :many
SELECT * FROM users
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT COUNT(*) FROM users
WHERE deleted_at IS NULL;

-- name: SearchUsersByEmail :many
SELECT * FROM users
WHERE email ILIKE $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
