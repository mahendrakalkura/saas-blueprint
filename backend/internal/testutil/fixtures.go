package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/mahendrakalkura/saas-blueprint/internal/auth"
	"github.com/mahendrakalkura/saas-blueprint/internal/models"
	"github.com/mahendrakalkura/saas-blueprint/internal/repository"
)

// CreateTestUser creates a test user in the database
func CreateTestUser(t *testing.T, userRepo *repository.UserRepository, email string) *models.User {
	t.Helper()

	passwordHash, err := auth.HashPassword("TestPassword123!")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	firstName := "Test"
	lastName := "User"

	user := &models.User{
		Email:        email,
		PasswordHash: passwordHash,
		FirstName:    &firstName,
		LastName:     &lastName,
		IsActive:     true,
	}

	err = userRepo.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// CreateTestUserWithMFA creates a test user with MFA enabled
func CreateTestUserWithMFA(t *testing.T, userRepo *repository.UserRepository, email, secret string) *models.User {
	t.Helper()

	user := CreateTestUser(t, userRepo, email)

	user.MFAEnabled = true
	user.MFASecret = &secret
	user.MFABackupCodes = []string{
		"AAAA-BBBB-CCCC",
		"DDDD-EEEE-FFFF",
		"GGGG-HHHH-IIII",
	}

	err := userRepo.Update(context.Background(), user)
	if err != nil {
		t.Fatalf("Failed to update test user with MFA: %v", err)
	}

	return user
}

// CreateTestSession creates a test session for a user
func CreateTestSession(t *testing.T, sessionRepo *repository.SessionRepository, userID string) *models.Session {
	t.Helper()

	refreshToken, err := auth.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("Failed to generate refresh token: %v", err)
	}

	userAgent := "Test User Agent"
	ipAddress := "127.0.0.1"

	session := &models.Session{
		UserID:                userID,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		UserAgent:             &userAgent,
		IPAddress:             &ipAddress,
	}

	err = sessionRepo.Create(context.Background(), session)
	if err != nil {
		t.Fatalf("Failed to create test session: %v", err)
	}

	return session
}

// GenerateTestJWT generates a test JWT token for a user
func GenerateTestJWT(t *testing.T, userID, email, secret string) string {
	t.Helper()

	token, err := auth.GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("Failed to generate test JWT: %v", err)
	}

	return token
}
