package models

import "time"

type OAuthProvider struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	Provider       string     `json:"provider"` // 'google', 'github'
	ProviderUserID string     `json:"provider_user_id"`
	Email          string     `json:"email"`
	Name           string     `json:"name"`
	AvatarURL      string     `json:"avatar_url,omitempty"`
	AccessToken    string     `json:"-"` // Never expose in JSON
	RefreshToken   string     `json:"-"` // Never expose in JSON
	TokenExpiry    *time.Time `json:"-"` // Never expose in JSON
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// OAuthProviderType constants
const (
	OAuthProviderGoogle = "google"
	OAuthProviderGitHub = "github"
)
