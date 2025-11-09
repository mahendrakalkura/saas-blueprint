package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
)

type OAuthProviderRepository struct {
	db *sql.DB
}

func NewOAuthProviderRepository(db *sql.DB) *OAuthProviderRepository {
	return &OAuthProviderRepository{db: db}
}

func (r *OAuthProviderRepository) Create(ctx context.Context, provider *models.OAuthProvider) error {
	query := `
		INSERT INTO oauth_providers (
			id, user_id, provider, provider_user_id, email, name, avatar_url,
			access_token, refresh_token, token_expiry, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	provider.ID = uuid.New().String()
	now := time.Now()
	provider.CreatedAt = now
	provider.UpdatedAt = now

	err := r.db.QueryRowContext(ctx, query,
		provider.ID,
		provider.UserID,
		provider.Provider,
		provider.ProviderUserID,
		provider.Email,
		provider.Name,
		provider.AvatarURL,
		provider.AccessToken,
		provider.RefreshToken,
		provider.TokenExpiry,
		provider.CreatedAt,
		provider.UpdatedAt,
	).Scan(&provider.ID, &provider.CreatedAt, &provider.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create oauth provider: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) GetByProviderAndUserID(ctx context.Context, provider, providerUserID string) (*models.OAuthProvider, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url,
		       access_token, refresh_token, token_expiry, created_at, updated_at
		FROM oauth_providers
		WHERE provider = $1 AND provider_user_id = $2
	`

	oauthProvider := &models.OAuthProvider{}
	err := r.db.QueryRowContext(ctx, query, provider, providerUserID).Scan(
		&oauthProvider.ID,
		&oauthProvider.UserID,
		&oauthProvider.Provider,
		&oauthProvider.ProviderUserID,
		&oauthProvider.Email,
		&oauthProvider.Name,
		&oauthProvider.AvatarURL,
		&oauthProvider.AccessToken,
		&oauthProvider.RefreshToken,
		&oauthProvider.TokenExpiry,
		&oauthProvider.CreatedAt,
		&oauthProvider.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get oauth provider: %w", err)
	}

	return oauthProvider, nil
}

func (r *OAuthProviderRepository) GetByUserID(ctx context.Context, userID string) ([]*models.OAuthProvider, error) {
	query := `
		SELECT id, user_id, provider, provider_user_id, email, name, avatar_url,
		       access_token, refresh_token, token_expiry, created_at, updated_at
		FROM oauth_providers
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list oauth providers: %w", err)
	}
	defer rows.Close()

	providers := []*models.OAuthProvider{}
	for rows.Next() {
		provider := &models.OAuthProvider{}
		err := rows.Scan(
			&provider.ID,
			&provider.UserID,
			&provider.Provider,
			&provider.ProviderUserID,
			&provider.Email,
			&provider.Name,
			&provider.AvatarURL,
			&provider.AccessToken,
			&provider.RefreshToken,
			&provider.TokenExpiry,
			&provider.CreatedAt,
			&provider.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan oauth provider: %w", err)
		}
		providers = append(providers, provider)
	}

	return providers, nil
}

func (r *OAuthProviderRepository) UpdateTokens(ctx context.Context, id, accessToken, refreshToken string, tokenExpiry *time.Time) error {
	query := `
		UPDATE oauth_providers
		SET access_token = $1, refresh_token = $2, token_expiry = $3, updated_at = $4
		WHERE id = $5
	`

	_, err := r.db.ExecContext(ctx, query, accessToken, refreshToken, tokenExpiry, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update oauth tokens: %w", err)
	}

	return nil
}

func (r *OAuthProviderRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM oauth_providers WHERE id = $1`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete oauth provider: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("oauth provider not found")
	}

	return nil
}
