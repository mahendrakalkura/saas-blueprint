-- name: CreateNotification :one
INSERT INTO notifications (
  id, user_id, type, title, message, metadata, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetNotification :one
SELECT * FROM notifications
WHERE id = $1;

-- name: ListUserNotifications :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUnreadNotifications :many
SELECT * FROM notifications
WHERE user_id = $1 AND is_read = false
ORDER BY created_at DESC
LIMIT $2;

-- name: MarkNotificationAsRead :exec
UPDATE notifications
SET is_read = true, read_at = $2
WHERE id = $1;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = true, read_at = $2
WHERE user_id = $1 AND is_read = false;

-- name: CountUnreadNotifications :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1 AND is_read = false;

-- name: DeleteNotification :exec
DELETE FROM notifications
WHERE id = $1;

-- name: GetNotificationPreferences :one
SELECT * FROM notification_preferences
WHERE user_id = $1;

-- name: CreateNotificationPreferences :one
INSERT INTO notification_preferences (
  id, user_id, email_marketing, email_product_updates,
  email_security_alerts, email_billing, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8
) RETURNING *;

-- name: UpdateNotificationPreferences :one
UPDATE notification_preferences
SET email_marketing = $2, email_product_updates = $3,
    email_security_alerts = $4, email_billing = $5, updated_at = $6
WHERE user_id = $1
RETURNING *;
