-- name: CreateOrganizationMember :one
INSERT INTO organization_members (
  id, organization_id, user_id, role, invited_by, invited_at, joined_at, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9
) RETURNING *;

-- name: GetOrganizationMember :one
SELECT * FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: GetOrganizationMemberByID :one
SELECT * FROM organization_members
WHERE id = $1;

-- name: UpdateOrganizationMemberRole :exec
UPDATE organization_members
SET role = $2, updated_at = $3
WHERE id = $1;

-- name: RemoveOrganizationMember :exec
DELETE FROM organization_members
WHERE id = $1;

-- name: ListOrganizationMembers :many
SELECT om.*, u.email, u.first_name, u.last_name, u.avatar_url
FROM organization_members om
INNER JOIN users u ON om.user_id = u.id
WHERE om.organization_id = $1
ORDER BY om.created_at ASC;

-- name: CountOrganizationMembers :one
SELECT COUNT(*) FROM organization_members
WHERE organization_id = $1;

-- name: GetUserOrganizationRole :one
SELECT role FROM organization_members
WHERE organization_id = $1 AND user_id = $2;

-- name: ListPendingInvitations :many
SELECT om.*, u.email, u.first_name, u.last_name
FROM organization_members om
INNER JOIN users u ON om.user_id = u.id
WHERE om.organization_id = $1 AND om.joined_at IS NULL
ORDER BY om.invited_at DESC;
