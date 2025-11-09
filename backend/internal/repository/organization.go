package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) Create(ctx context.Context, org *models.Organization) error {
	query := `
		INSERT INTO organizations (id, name, slug, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	org.ID = uuid.New().String()
	org.CreatedAt = time.Now()
	org.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		org.ID,
		org.Name,
		org.Slug,
		org.OwnerID,
		org.CreatedAt,
		org.UpdatedAt,
	).Scan(&org.ID, &org.CreatedAt, &org.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create organization: %w", err)
	}

	return nil
}

func (r *OrganizationRepository) GetByID(ctx context.Context, id string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at, deleted_at
		FROM organizations
		WHERE id = $1 AND deleted_at IS NULL
	`

	org := &models.Organization{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

func (r *OrganizationRepository) GetBySlug(ctx context.Context, slug string) (*models.Organization, error) {
	query := `
		SELECT id, name, slug, owner_id, created_at, updated_at, deleted_at
		FROM organizations
		WHERE slug = $1 AND deleted_at IS NULL
	`

	org := &models.Organization{}
	err := r.db.QueryRowContext(ctx, query, slug).Scan(
		&org.ID,
		&org.Name,
		&org.Slug,
		&org.OwnerID,
		&org.CreatedAt,
		&org.UpdatedAt,
		&org.DeletedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("organization not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	return org, nil
}

func (r *OrganizationRepository) ListByUserID(ctx context.Context, userID string) ([]*models.Organization, error) {
	query := `
		SELECT DISTINCT o.id, o.name, o.slug, o.owner_id, o.created_at, o.updated_at, o.deleted_at
		FROM organizations o
		INNER JOIN organization_members om ON o.id = om.organization_id
		WHERE om.user_id = $1 AND o.deleted_at IS NULL
		ORDER BY o.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	defer rows.Close()

	orgs := []*models.Organization{}
	for rows.Next() {
		org := &models.Organization{}
		err := rows.Scan(
			&org.ID,
			&org.Name,
			&org.Slug,
			&org.OwnerID,
			&org.CreatedAt,
			&org.UpdatedAt,
			&org.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan organization: %w", err)
		}
		orgs = append(orgs, org)
	}

	return orgs, nil
}

func (r *OrganizationRepository) Update(ctx context.Context, org *models.Organization) error {
	query := `
		UPDATE organizations
		SET name = $1, slug = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	org.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		org.Name,
		org.Slug,
		org.UpdatedAt,
		org.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id string) error {
	query := `
		UPDATE organizations
		SET deleted_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, now, id)
	if err != nil {
		return fmt.Errorf("failed to delete organization: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("organization not found")
	}

	return nil
}
