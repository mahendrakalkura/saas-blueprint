-- name: CreateAuditLog :one
INSERT INTO audit_logs (
  id, user_id, organization_id, action, resource_type, resource_id,
  metadata, ip_address, user_agent, created_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
) RETURNING *;

-- name: GetAuditLog :one
SELECT * FROM audit_logs
WHERE id = $1;

-- name: ListAuditLogs :many
SELECT * FROM audit_logs
WHERE organization_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUserAuditLogs :many
SELECT * FROM audit_logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountAuditLogs :one
SELECT COUNT(*) FROM audit_logs
WHERE organization_id = $1;

-- name: SearchAuditLogs :many
SELECT * FROM audit_logs
WHERE organization_id = $1 AND action = $2
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;
