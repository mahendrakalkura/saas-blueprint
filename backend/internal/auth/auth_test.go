package auth

import (
	"strings"
	"testing"
	"time"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{
			name:     "valid password",
			password: "SecurePassword123!",
			wantErr:  false,
		},
		{
			name:     "short password",
			password: "short",
			wantErr:  false, // bcrypt doesn't validate length
		},
		{
			name:     "empty password",
			password: "",
			wantErr:  false, // bcrypt allows empty passwords
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Verify the hash is valid
				if hash == "" {
					t.Error("HashPassword() returned empty hash")
				}

				// Verify hash starts with bcrypt prefix
				if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
					t.Errorf("HashPassword() returned invalid bcrypt hash: %s", hash)
				}
			}
		})
	}
}

func TestCheckPassword(t *testing.T) {
	password := "TestPassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		wantErr  bool
	}{
		{
			name:     "correct password",
			password: password,
			hash:     hash,
			wantErr:  false,
		},
		{
			name:     "incorrect password",
			password: "WrongPassword",
			hash:     hash,
			wantErr:  true,
		},
		{
			name:     "empty password",
			password: "",
			hash:     hash,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckPassword(tt.password, tt.hash)

			if (err != nil) != tt.wantErr {
				t.Errorf("CheckPassword() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateAccessToken(t *testing.T) {
	userID := "test-user-id"
	email := "test@example.com"
	secret := "test-secret-key"

	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		t.Fatalf("GenerateAccessToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateAccessToken() returned empty token")
	}

	// Validate the token
	claims, err := ValidateAccessToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateAccessToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("Token UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Email != email {
		t.Errorf("Token Email = %v, want %v", claims.Email, email)
	}

	// Check expiration is in the future
	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token is already expired")
	}

	// Check expiration is within expected range
	expectedExpiry := time.Now().Add(AccessTokenDuration)
	if claims.ExpiresAt.Time.After(expectedExpiry.Add(time.Second)) {
		t.Error("Token expiration is too far in the future")
	}
}

func TestValidateAccessToken(t *testing.T) {
	userID := "test-user-id"
	email := "test@example.com"
	secret := "test-secret-key"

	validToken, _ := GenerateAccessToken(userID, email, secret)

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid token",
			token:   validToken,
			secret:  secret,
			wantErr: false,
		},
		{
			name:    "invalid secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "malformed token",
			token:   "not.a.valid.jwt",
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateAccessToken(tt.token, tt.secret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAccessToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims == nil {
				t.Error("ValidateAccessToken() returned nil claims for valid token")
			}
		})
	}
}

func TestGenerateRefreshToken(t *testing.T) {
	token1, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token1 == "" {
		t.Error("GenerateRefreshToken() returned empty token")
	}

	// Generate another token and ensure they're different
	token2, err := GenerateRefreshToken()
	if err != nil {
		t.Fatalf("GenerateRefreshToken() error = %v", err)
	}

	if token1 == token2 {
		t.Error("GenerateRefreshToken() returned same token twice")
	}

	// Verify token length (base64 of 32 bytes should be ~44 chars)
	if len(token1) < 40 {
		t.Errorf("GenerateRefreshToken() token too short: %d chars", len(token1))
	}
}

func TestGenerateMFAToken(t *testing.T) {
	userID := "test-user-id"
	secret := "test-secret-key"

	token, err := GenerateMFAToken(userID, secret)
	if err != nil {
		t.Fatalf("GenerateMFAToken() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateMFAToken() returned empty token")
	}

	// Validate the MFA token
	claims, err := ValidateMFAToken(token, secret)
	if err != nil {
		t.Fatalf("ValidateMFAToken() error = %v", err)
	}

	if claims.UserID != userID {
		t.Errorf("MFA Token UserID = %v, want %v", claims.UserID, userID)
	}

	if claims.Type != "mfa_pending" {
		t.Errorf("MFA Token Type = %v, want mfa_pending", claims.Type)
	}

	// Check expiration is correct
	expectedExpiry := time.Now().Add(MFATokenDuration)
	if claims.ExpiresAt.Time.After(expectedExpiry.Add(time.Second)) {
		t.Error("MFA token expiration is too far in the future")
	}
}

func TestValidateMFAToken(t *testing.T) {
	userID := "test-user-id"
	secret := "test-secret-key"

	validToken, _ := GenerateMFAToken(userID, secret)

	tests := []struct {
		name    string
		token   string
		secret  string
		wantErr bool
	}{
		{
			name:    "valid MFA token",
			token:   validToken,
			secret:  secret,
			wantErr: false,
		},
		{
			name:    "invalid secret",
			token:   validToken,
			secret:  "wrong-secret",
			wantErr: true,
		},
		{
			name:    "wrong token type (access token)",
			token:   mustGenerateAccessToken(userID, "test@example.com", secret),
			secret:  secret,
			wantErr: true,
		},
		{
			name:    "empty token",
			token:   "",
			secret:  secret,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := ValidateMFAToken(tt.token, tt.secret)

			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMFAToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && claims == nil {
				t.Error("ValidateMFAToken() returned nil claims for valid token")
			}
		})
	}
}

// Helper function for tests
func mustGenerateAccessToken(userID, email, secret string) string {
	token, err := GenerateAccessToken(userID, email, secret)
	if err != nil {
		panic(err)
	}
	return token
}
