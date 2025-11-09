package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, session *models.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, refresh_token, refresh_token_expires_at, user_agent, ip_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	session.ID = uuid.New().String()
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	session.IsRevoked = false

	err := r.db.QueryRowContext(ctx, query,
		session.ID,
		session.UserID,
		session.RefreshToken,
		session.RefreshTokenExpiresAt,
		session.UserAgent,
		session.IPAddress,
		session.CreatedAt,
		session.UpdatedAt,
	).Scan(&session.ID, &session.CreatedAt, &session.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

func (r *SessionRepository) GetByRefreshToken(ctx context.Context, refreshToken string) (*models.Session, error) {
	query := `
		SELECT id, user_id, refresh_token, refresh_token_expires_at, user_agent, ip_address,
		       is_revoked, created_at, updated_at
		FROM sessions
		WHERE refresh_token = $1 AND is_revoked = false
	`

	session := &models.Session{}
	err := r.db.QueryRowContext(ctx, query, refreshToken).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshToken,
		&session.RefreshTokenExpiresAt,
		&session.UserAgent,
		&session.IPAddress,
		&session.IsRevoked,
		&session.CreatedAt,
		&session.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, refreshToken string) error {
	query := `
		UPDATE sessions
		SET is_revoked = true, updated_at = $1
		WHERE refresh_token = $2
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), refreshToken)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *SessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	query := `
		UPDATE sessions
		SET is_revoked = true, updated_at = $1
		WHERE user_id = $2 AND is_revoked = false
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	return nil
}

func (r *SessionRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM sessions
		WHERE refresh_token_expires_at < $1
	`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return nil
}
