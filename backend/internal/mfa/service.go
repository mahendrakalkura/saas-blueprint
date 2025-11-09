package mfa

import (
	"crypto/rand"
	"encoding/base32"
	"fmt"
	"image/png"
	"io"
	"strings"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

type Service struct {
	issuer string
}

// QRCodeData represents the data needed to generate a QR code
type QRCodeData struct {
	Secret string
	URL    string
	QRCode []byte
}

func NewService(issuer string) *Service {
	return &Service{
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret for a user
func (s *Service) GenerateSecret(email string) (*QRCodeData, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	// Generate QR code image
	var buf []byte
	img, err := key.Image(200, 200)
	if err != nil {
		return nil, fmt.Errorf("failed to generate QR code image: %w", err)
	}

	// Encode image to PNG bytes
	// Create a buffer to write the PNG data
	writer := &bytesWriter{data: make([]byte, 0, 4096)}
	if err := png.Encode(writer, img); err != nil {
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	buf = writer.data

	return &QRCodeData{
		Secret: key.Secret(),
		URL:    key.URL(),
		QRCode: buf,
	}, nil
}

// VerifyCode verifies a TOTP code against a secret
func (s *Service) VerifyCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates backup codes for MFA recovery
func (s *Service) GenerateBackupCodes(count int) ([]string, error) {
	codes := make([]string, count)
	for i := 0; i < count; i++ {
		code, err := generateBackupCode()
		if err != nil {
			return nil, fmt.Errorf("failed to generate backup code: %w", err)
		}
		codes[i] = code
	}
	return codes, nil
}

// VerifyBackupCode checks if a backup code is valid
func (s *Service) VerifyBackupCode(code string, backupCodes []string) (bool, []string) {
	// Remove hyphens for comparison
	code = strings.ReplaceAll(code, "-", "")
	code = strings.ToUpper(code)

	for i, backupCode := range backupCodes {
		// Remove hyphens from stored code
		storedCode := strings.ReplaceAll(backupCode, "-", "")
		storedCode = strings.ToUpper(storedCode)

		if code == storedCode {
			// Remove used backup code
			newCodes := append(backupCodes[:i], backupCodes[i+1:]...)
			return true, newCodes
		}
	}

	return false, backupCodes
}

// generateBackupCode generates a random backup code (format: XXXX-XXXX-XXXX)
func generateBackupCode() (string, error) {
	// Generate 9 random bytes (72 bits)
	b := make([]byte, 9)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// Encode to base32 (will give us 12+ characters)
	code := base32.StdEncoding.EncodeToString(b)
	// Take first 12 characters and format as XXXX-XXXX-XXXX
	code = code[:12]
	return fmt.Sprintf("%s-%s-%s", code[0:4], code[4:8], code[8:12]), nil
}

// bytesWriter is a simple io.Writer that writes to a byte slice
type bytesWriter struct {
	data []byte
}

func (w *bytesWriter) Write(p []byte) (n int, err error) {
	w.data = append(w.data, p...)
	return len(p), nil
}
