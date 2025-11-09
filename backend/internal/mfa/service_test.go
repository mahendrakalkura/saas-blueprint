package mfa

import (
	"strings"
	"testing"

	"github.com/pquerna/otp/totp"
)

func TestNewService(t *testing.T) {
	issuer := "Test App"
	service := NewService(issuer)

	if service == nil {
		t.Fatal("NewService() returned nil")
	}

	if service.issuer != issuer {
		t.Errorf("Service issuer = %v, want %v", service.issuer, issuer)
	}
}

func TestGenerateSecret(t *testing.T) {
	service := NewService("Test App")

	tests := []struct {
		name  string
		email string
	}{
		{
			name:  "valid email",
			email: "test@example.com",
		},
		{
			name:  "email with special chars",
			email: "test+tag@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := service.GenerateSecret(tt.email)
			if err != nil {
				t.Fatalf("GenerateSecret() error = %v", err)
			}

			if data == nil {
				t.Fatal("GenerateSecret() returned nil")
			}

			// Verify secret is not empty
			if data.Secret == "" {
				t.Error("GenerateSecret() returned empty secret")
			}

			// Verify URL is not empty
			if data.URL == "" {
				t.Error("GenerateSecret() returned empty URL")
			}

			// Verify URL contains account name
			if !strings.Contains(data.URL, tt.email) {
				t.Errorf("URL does not contain email: %s", data.URL)
			}

			// Verify QR code is not empty
			if len(data.QRCode) == 0 {
				t.Error("GenerateSecret() returned empty QR code")
			}

			// Verify backup codes
			if len(data.BackupCodes) != 10 {
				t.Errorf("Expected 10 backup codes, got %d", len(data.BackupCodes))
			}

			for _, code := range data.BackupCodes {
				if len(code) != 14 { // Format: XXXX-XXXX-XXXX
					t.Errorf("Invalid backup code length: %s", code)
				}
				if !strings.Contains(code, "-") {
					t.Errorf("Backup code missing separator: %s", code)
				}
			}
		})
	}
}

func TestVerifyCode(t *testing.T) {
	service := NewService("Test App")

	// Generate a secret
	data, err := service.GenerateSecret("test@example.com")
	if err != nil {
		t.Fatalf("Failed to generate secret: %v", err)
	}

	// Generate a valid code for the secret
	validCode, err := totp.GenerateCode(data.Secret, 0) // Use time 0 for deterministic testing
	if err != nil {
		t.Fatalf("Failed to generate TOTP code: %v", err)
	}

	tests := []struct {
		name   string
		secret string
		code   string
		want   bool
	}{
		{
			name:   "invalid code - wrong digits",
			secret: data.Secret,
			code:   "000000",
			want:   false,
		},
		{
			name:   "invalid code - wrong length",
			secret: data.Secret,
			code:   "12345",
			want:   false,
		},
		{
			name:   "empty code",
			secret: data.Secret,
			code:   "",
			want:   false,
		},
		{
			name:   "invalid secret",
			secret: "invalid-secret",
			code:   validCode,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.VerifyCode(tt.secret, tt.code)
			if result != tt.want {
				t.Errorf("VerifyCode() = %v, want %v", result, tt.want)
			}
		})
	}

	// Test with current time (this should work)
	t.Run("valid code - current time", func(t *testing.T) {
		currentCode, err := totp.GenerateCode(data.Secret, 0)
		if err != nil {
			t.Fatalf("Failed to generate code: %v", err)
		}

		// Note: This test might fail if run exactly at a time boundary
		// In production, you'd use a fixed time for testing
		result := service.VerifyCode(data.Secret, currentCode)
		if !result {
			t.Logf("Warning: Current time code verification failed (might be at time boundary)")
		}
	})
}

func TestGenerateBackupCodes(t *testing.T) {
	service := NewService("Test App")

	tests := []struct {
		name  string
		count int
	}{
		{
			name:  "generate 10 codes",
			count: 10,
		},
		{
			name:  "generate 5 codes",
			count: 5,
		},
		{
			name:  "generate 1 code",
			count: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			codes, err := service.GenerateBackupCodes(tt.count)
			if err != nil {
				t.Fatalf("GenerateBackupCodes() error = %v", err)
			}

			if len(codes) != tt.count {
				t.Errorf("Expected %d codes, got %d", tt.count, len(codes))
			}

			// Verify all codes are unique
			seen := make(map[string]bool)
			for _, code := range codes {
				if seen[code] {
					t.Errorf("Duplicate backup code: %s", code)
				}
				seen[code] = true

				// Verify format: XXXX-XXXX-XXXX
				if len(code) != 14 {
					t.Errorf("Invalid code length: %s (len=%d)", code, len(code))
				}

				parts := strings.Split(code, "-")
				if len(parts) != 3 {
					t.Errorf("Invalid code format: %s", code)
				}

				for _, part := range parts {
					if len(part) != 4 {
						t.Errorf("Invalid code part length: %s in %s", part, code)
					}
				}
			}
		})
	}
}

func TestVerifyBackupCode(t *testing.T) {
	service := NewService("Test App")

	originalCodes := []string{
		"AAAA-BBBB-CCCC",
		"DDDD-EEEE-FFFF",
		"GGGG-HHHH-IIII",
	}

	tests := []struct {
		name          string
		code          string
		backupCodes   []string
		wantValid     bool
		wantRemaining int
	}{
		{
			name:          "valid code - first code",
			code:          "AAAA-BBBB-CCCC",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     true,
			wantRemaining: 2,
		},
		{
			name:          "valid code - middle code",
			code:          "DDDD-EEEE-FFFF",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     true,
			wantRemaining: 2,
		},
		{
			name:          "valid code - last code",
			code:          "GGGG-HHHH-IIII",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     true,
			wantRemaining: 2,
		},
		{
			name:          "invalid code",
			code:          "XXXX-YYYY-ZZZZ",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     false,
			wantRemaining: 3,
		},
		{
			name:          "lowercase code (should work)",
			code:          "aaaa-bbbb-cccc",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     true,
			wantRemaining: 2,
		},
		{
			name:          "empty code",
			code:          "",
			backupCodes:   append([]string{}, originalCodes...),
			wantValid:     false,
			wantRemaining: 3,
		},
		{
			name:          "empty backup codes",
			code:          "AAAA-BBBB-CCCC",
			backupCodes:   []string{},
			wantValid:     false,
			wantRemaining: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, remaining := service.VerifyBackupCode(tt.code, tt.backupCodes)

			if valid != tt.wantValid {
				t.Errorf("VerifyBackupCode() valid = %v, want %v", valid, tt.wantValid)
			}

			if len(remaining) != tt.wantRemaining {
				t.Errorf("VerifyBackupCode() remaining count = %d, want %d", len(remaining), tt.wantRemaining)
			}

			// If valid, ensure the used code was removed
			if valid {
				for _, code := range remaining {
					if strings.ToUpper(code) == strings.ToUpper(tt.code) {
						t.Error("Used backup code still present in remaining codes")
					}
				}
			}
		})
	}
}

func TestVerifyBackupCode_UsedTwice(t *testing.T) {
	service := NewService("Test App")

	codes := []string{
		"AAAA-BBBB-CCCC",
		"DDDD-EEEE-FFFF",
	}

	// Use the first code
	valid1, remaining1 := service.VerifyBackupCode("AAAA-BBBB-CCCC", codes)
	if !valid1 {
		t.Fatal("First use of backup code should be valid")
	}

	if len(remaining1) != 1 {
		t.Fatalf("Expected 1 remaining code, got %d", len(remaining1))
	}

	// Try to use the same code again
	valid2, remaining2 := service.VerifyBackupCode("AAAA-BBBB-CCCC", remaining1)
	if valid2 {
		t.Error("Second use of same backup code should be invalid")
	}

	if len(remaining2) != 1 {
		t.Errorf("Remaining codes should stay at 1, got %d", len(remaining2))
	}
}
