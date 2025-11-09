package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type OrganizationMemberRepository struct {
	db *sql.DB
}

func NewOrganizationMemberRepository(db *sql.DB) *OrganizationMemberRepository {
	return &OrganizationMemberRepository{db: db}
}

func (r *OrganizationMemberRepository) Create(ctx context.Context, member *models.OrganizationMember) error {
	query := `
		INSERT INTO organization_members (id, organization_id, user_id, role, invited_by, invited_at, joined_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	member.ID = uuid.New().String()
	member.CreatedAt = time.Now()
	member.UpdatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		member.ID,
		member.OrganizationID,
		member.UserID,
		member.Role,
		member.InvitedBy,
		member.InvitedAt,
		member.JoinedAt,
		member.CreatedAt,
		member.UpdatedAt,
	).Scan(&member.ID, &member.CreatedAt, &member.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create member: %w", err)
	}

	return nil
}

func (r *OrganizationMemberRepository) GetByID(ctx context.Context, id string) (*models.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, invited_by, invited_at, joined_at, created_at, updated_at
		FROM organization_members
		WHERE id = $1
	`

	member := &models.OrganizationMember{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.InvitedBy,
		&member.InvitedAt,
		&member.JoinedAt,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return member, nil
}

func (r *OrganizationMemberRepository) GetByOrganizationAndUser(ctx context.Context, orgID, userID string) (*models.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, invited_by, invited_at, joined_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1 AND user_id = $2
	`

	member := &models.OrganizationMember{}
	err := r.db.QueryRowContext(ctx, query, orgID, userID).Scan(
		&member.ID,
		&member.OrganizationID,
		&member.UserID,
		&member.Role,
		&member.InvitedBy,
		&member.InvitedAt,
		&member.JoinedAt,
		&member.CreatedAt,
		&member.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("member not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get member: %w", err)
	}

	return member, nil
}

func (r *OrganizationMemberRepository) ListByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationMember, error) {
	query := `
		SELECT id, organization_id, user_id, role, invited_by, invited_at, joined_at, created_at, updated_at
		FROM organization_members
		WHERE organization_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list members: %w", err)
	}
	defer rows.Close()

	members := []*models.OrganizationMember{}
	for rows.Next() {
		member := &models.OrganizationMember{}
		err := rows.Scan(
			&member.ID,
			&member.OrganizationID,
			&member.UserID,
			&member.Role,
			&member.InvitedBy,
			&member.InvitedAt,
			&member.JoinedAt,
			&member.CreatedAt,
			&member.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, member)
	}

	return members, nil
}

func (r *OrganizationMemberRepository) UpdateRole(ctx context.Context, id, role string) error {
	query := `
		UPDATE organization_members
		SET role = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, role, now, id)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

func (r *OrganizationMemberRepository) Delete(ctx context.Context, id string) error {
	query := `
		DELETE FROM organization_members
		WHERE id = $1
	`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete member: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("member not found")
	}

	return nil
}

// Invitation methods

func (r *OrganizationMemberRepository) CreateInvitation(ctx context.Context, inv *models.OrganizationInvitation) error {
	query := `
		INSERT INTO organization_invitations (id, organization_id, email, role, token, invited_by, expires_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at
	`

	inv.ID = uuid.New().String()
	inv.CreatedAt = time.Now()

	err := r.db.QueryRowContext(ctx, query,
		inv.ID,
		inv.OrganizationID,
		inv.Email,
		inv.Role,
		inv.Token,
		inv.InvitedBy,
		inv.ExpiresAt,
		inv.CreatedAt,
	).Scan(&inv.ID, &inv.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to create invitation: %w", err)
	}

	return nil
}

func (r *OrganizationMemberRepository) GetInvitationByToken(ctx context.Context, token string) (*models.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, invited_by, expires_at, accepted_at, created_at
		FROM organization_invitations
		WHERE token = $1 AND accepted_at IS NULL
	`

	inv := &models.OrganizationInvitation{}
	err := r.db.QueryRowContext(ctx, query, token).Scan(
		&inv.ID,
		&inv.OrganizationID,
		&inv.Email,
		&inv.Role,
		&inv.Token,
		&inv.InvitedBy,
		&inv.ExpiresAt,
		&inv.AcceptedAt,
		&inv.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invitation not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get invitation: %w", err)
	}

	return inv, nil
}

func (r *OrganizationMemberRepository) AcceptInvitation(ctx context.Context, token string) error {
	query := `
		UPDATE organization_invitations
		SET accepted_at = $1
		WHERE token = $2 AND accepted_at IS NULL
	`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, token)
	if err != nil {
		return fmt.Errorf("failed to accept invitation: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("invitation not found or already accepted")
	}

	return nil
}

func (r *OrganizationMemberRepository) ListInvitationsByOrganization(ctx context.Context, orgID string) ([]*models.OrganizationInvitation, error) {
	query := `
		SELECT id, organization_id, email, role, token, invited_by, expires_at, accepted_at, created_at
		FROM organization_invitations
		WHERE organization_id = $1 AND accepted_at IS NULL AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to list invitations: %w", err)
	}
	defer rows.Close()

	invitations := []*models.OrganizationInvitation{}
	for rows.Next() {
		inv := &models.OrganizationInvitation{}
		err := rows.Scan(
			&inv.ID,
			&inv.OrganizationID,
			&inv.Email,
			&inv.Role,
			&inv.Token,
			&inv.InvitedBy,
			&inv.ExpiresAt,
			&inv.AcceptedAt,
			&inv.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan invitation: %w", err)
		}
		invitations = append(invitations, inv)
	}

	return invitations, nil
}
