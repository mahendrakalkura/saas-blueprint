package models

import (
	"time"
)

type User struct {
	ID                         string     `json:"id"`
	Email                      string     `json:"email"`
	PasswordHash               string     `json:"-"`
	FirstName                  *string    `json:"first_name,omitempty"`
	LastName                   *string    `json:"last_name,omitempty"`
	AvatarURL                  *string    `json:"avatar_url,omitempty"`
	EmailVerified              bool       `json:"email_verified"`
	EmailVerificationToken     *string    `json:"-"`
	EmailVerificationExpiresAt *time.Time `json:"-"`
	PasswordResetToken         *string    `json:"-"`
	PasswordResetExpiresAt     *time.Time `json:"-"`
	IsActive                   bool       `json:"is_active"`
	CreatedAt                  time.Time  `json:"created_at"`
	UpdatedAt                  time.Time  `json:"updated_at"`
	DeletedAt                  *time.Time `json:"-"`
}

type Session struct {
	ID                    string     `json:"id"`
	UserID                string     `json:"user_id"`
	RefreshToken          string     `json:"-"`
	RefreshTokenExpiresAt time.Time  `json:"refresh_token_expires_at"`
	UserAgent             *string    `json:"user_agent,omitempty"`
	IPAddress             *string    `json:"ip_address,omitempty"`
	IsRevoked             bool       `json:"is_revoked"`
	CreatedAt             time.Time  `json:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at"`
}
