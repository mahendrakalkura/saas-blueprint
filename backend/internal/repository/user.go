package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, first_name, last_name, email_verified, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`

	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.EmailVerified = false
	user.IsActive = true

	err := r.db.QueryRowContext(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.FirstName,
		user.LastName,
		user.EmailVerified,
		user.IsActive,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified,
		       mfa_enabled, mfa_secret, mfa_backup_codes, is_active, created_at, updated_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	var backupCodes []sql.NullString
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.MFAEnabled,
		&user.MFASecret,
		pq.Array(&backupCodes),
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert backup codes from []sql.NullString to []string
	if len(backupCodes) > 0 {
		user.MFABackupCodes = make([]string, 0, len(backupCodes))
		for _, code := range backupCodes {
			if code.Valid {
				user.MFABackupCodes = append(user.MFABackupCodes, code.String)
			}
		}
	}

	return user, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT id, email, password_hash, first_name, last_name, avatar_url, email_verified,
		       mfa_enabled, mfa_secret, mfa_backup_codes, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	user := &models.User{}
	var backupCodes []sql.NullString
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.FirstName,
		&user.LastName,
		&user.AvatarURL,
		&user.EmailVerified,
		&user.MFAEnabled,
		&user.MFASecret,
		pq.Array(&backupCodes),
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Convert backup codes from []sql.NullString to []string
	if len(backupCodes) > 0 {
		user.MFABackupCodes = make([]string, 0, len(backupCodes))
		for _, code := range backupCodes {
			if code.Valid {
				user.MFABackupCodes = append(user.MFABackupCodes, code.String)
			}
		}
	}

	return user, nil
}

func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users
		SET email = $1, first_name = $2, last_name = $3, avatar_url = $4,
		    email_verified = $5, mfa_enabled = $6, mfa_secret = $7,
		    mfa_backup_codes = $8, updated_at = $9
		WHERE id = $10 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Email,
		user.FirstName,
		user.LastName,
		user.AvatarURL,
		user.EmailVerified,
		user.MFAEnabled,
		user.MFASecret,
		pq.Array(user.MFABackupCodes),
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) SetEmailVerificationToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
		UPDATE users
		SET email_verification_token = $1, email_verification_expires_at = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, token, expiresAt, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to set email verification token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) VerifyEmail(ctx context.Context, token string) error {
	query := `
		UPDATE users
		SET email_verified = true, email_verification_token = NULL,
		    email_verification_expires_at = NULL, updated_at = $1
		WHERE email_verification_token = $2
		  AND email_verification_expires_at > $1
		  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("invalid or expired verification token")
	}

	return nil
}

func (r *UserRepository) SetPasswordResetToken(ctx context.Context, email, token string, expiresAt time.Time) error {
	query := `
		UPDATE users
		SET password_reset_token = $1, password_reset_expires_at = $2, updated_at = $3
		WHERE email = $4 AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, token, expiresAt, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to set password reset token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *UserRepository) ResetPassword(ctx context.Context, token, newPasswordHash string) error {
	query := `
		UPDATE users
		SET password_hash = $1, password_reset_token = NULL,
		    password_reset_expires_at = NULL, updated_at = $2
		WHERE password_reset_token = $3
		  AND password_reset_expires_at > $2
		  AND deleted_at IS NULL
	`

	result, err := r.db.ExecContext(ctx, query, newPasswordHash, time.Now(), token)
	if err != nil {
		return fmt.Errorf("failed to reset password: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("invalid or expired reset token")
	}

	return nil
}
