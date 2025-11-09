-- name: CreateOrganization :one
INSERT INTO organizations (
  id, name, slug, logo_url, owner_id, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations
WHERE id = $1 AND deleted_at IS NULL;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations
WHERE slug = $1 AND deleted_at IS NULL;

-- name: UpdateOrganization :one
UPDATE organizations
SET name = $2, slug = $3, logo_url = $4, updated_at = $5
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: DeleteOrganization :exec
UPDATE organizations
SET deleted_at = $2, updated_at = $2
WHERE id = $1 AND deleted_at IS NULL;

-- name: ListUserOrganizations :many
SELECT o.* FROM organizations o
INNER JOIN organization_members om ON o.id = om.organization_id
WHERE om.user_id = $1 AND o.deleted_at IS NULL
ORDER BY o.created_at DESC;

-- name: ListOrganizations :many
SELECT * FROM organizations
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountOrganizations :one
SELECT COUNT(*) FROM organizations
WHERE deleted_at IS NULL;
