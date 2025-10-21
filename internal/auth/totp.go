package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"strings"
	"time"
)

// TOTP implements TOTP per RFC 6238 as specified in CLAUDE.md
type TOTP struct {
	Secret string
	Digits int
	Period int
}

// NewTOTP creates a new TOTP instance per CLAUDE.md
func NewTOTP() (*TOTP, error) {
	secret, err := generateSecret()
	if err != nil {
		return nil, err
	}

	return &TOTP{
		Secret: secret,
		Digits: 6, // Standard 6 digits
		Period: 30, // 30 seconds per RFC 6238
	}, nil
}

// Generate generates a TOTP code for current time per RFC 6238
func (t *TOTP) Generate() (string, error) {
	return t.GenerateAtTime(time.Now())
}

// GenerateAtTime generates TOTP code for specific time per RFC 6238
func (t *TOTP) GenerateAtTime(tm time.Time) (string, error) {
	counter := uint64(tm.Unix()) / uint64(t.Period)
	return t.generateHOTP(counter)
}

// Verify verifies a TOTP code per RFC 6238
func (t *TOTP) Verify(code string, window int) bool {
	now := time.Now()
	counter := uint64(now.Unix()) / uint64(t.Period)

	// Check current time window and adjacent windows
	for i := -window; i <= window; i++ {
		testCounter := uint64(int64(counter) + int64(i))
		expected, err := t.generateHOTP(testCounter)
		if err != nil {
			continue
		}

		if code == expected {
			return true
		}
	}

	return false
}

// generateHOTP generates HOTP code per RFC 4226
func (t *TOTP) generateHOTP(counter uint64) (string, error) {
	// Decode base32 secret
	key, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.ToUpper(t.Secret))
	if err != nil {
		return "", fmt.Errorf("invalid secret: %w", err)
	}

	// Convert counter to bytes
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	// HMAC-SHA1
	h := hmac.New(sha1.New, key)
	h.Write(buf)
	hash := h.Sum(nil)

	// Dynamic truncation per RFC 4226
	offset := hash[len(hash)-1] & 0x0f
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff

	// Generate code with specified digits
	format := fmt.Sprintf("%%0%dd", t.Digits)
	codeStr := fmt.Sprintf(format, code%uint32(pow(10, t.Digits)))

	return codeStr, nil
}

// GetProvisioningURI generates otpauth:// URI for QR code per RFC 6238
func (t *TOTP) GetProvisioningURI(accountName, issuer string) string {
	return fmt.Sprintf("otpauth://totp/%s:%s?secret=%s&issuer=%s&algorithm=SHA1&digits=%d&period=%d",
		issuer, accountName, t.Secret, issuer, t.Digits, t.Period)
}

// generateSecret generates a random base32 secret
func generateSecret() (string, error) {
	// Generate 20 random bytes (160 bits)
	secret := make([]byte, 20)
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}

	// Encode as base32
	encoded := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(secret)
	return encoded, nil
}

// pow calculates base^exp for integers
func pow(base, exp int) int {
	result := 1
	for i := 0; i < exp; i++ {
		result *= base
	}
	return result
}

// TOTPService manages 2FA for users per CLAUDE.md
type TOTPService struct {
	authService *AuthService
}

// NewTOTPService creates TOTP service
func NewTOTPService(authService *AuthService) *TOTPService {
	return &TOTPService{
		authService: authService,
	}
}

// EnableTOTP enables 2FA for a user per CLAUDE.md
func (s *TOTPService) EnableTOTP(userID int) (*TOTP, error) {
	totp, err := NewTOTP()
	if err != nil {
		return nil, err
	}

	// Save secret to user record (encrypted in production)
	query := "UPDATE users SET totp_secret = ?, totp_enabled = TRUE WHERE id = ?"
	_, err = s.authService.db.Exec(query, totp.Secret, userID)
	if err != nil {
		return nil, err
	}

	return totp, nil
}

// DisableTOTP disables 2FA for a user per CLAUDE.md
func (s *TOTPService) DisableTOTP(userID int) error {
	query := "UPDATE users SET totp_secret = NULL, totp_enabled = FALSE WHERE id = ?"
	_, err := s.authService.db.Exec(query, userID)
	return err
}

// VerifyTOTP verifies a user's TOTP code per CLAUDE.md
func (s *TOTPService) VerifyTOTP(userID int, code string) (bool, error) {
	var secret string
	var enabled bool

	query := "SELECT totp_secret, totp_enabled FROM users WHERE id = ?"
	err := s.authService.db.QueryRow(query, userID).Scan(&secret, &enabled)
	if err != nil {
		return false, err
	}

	if !enabled || secret == "" {
		return false, fmt.Errorf("2FA not enabled for user")
	}

	totp := &TOTP{
		Secret: secret,
		Digits: 6,
		Period: 30,
	}

	// Allow 1 time window before/after for clock drift
	return totp.Verify(code, 1), nil
}

// GetTOTPStatus checks if user has 2FA enabled
func (s *TOTPService) GetTOTPStatus(userID int) (bool, error) {
	var enabled bool
	query := "SELECT COALESCE(totp_enabled, FALSE) FROM users WHERE id = ?"
	err := s.authService.db.QueryRow(query, userID).Scan(&enabled)
	return enabled, err
}
