package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
)

var (
	ErrInvalidPassword = errors.New("неверный пароль")
	ErrWeakPassword    = errors.New("слишком слабый пароль")
	ErrInvalidTOTP     = errors.New("неверный TOTP код")
)

// PasswordPolicy defines password requirements
type PasswordPolicy struct {
	MinLength int
	MaxLength int
}

var DefaultPolicy = PasswordPolicy{
	MinLength: 12,
	MaxLength: 256,
}

// CommonPasswords is a small list of common passwords to reject
var CommonPasswords = map[string]bool{
	"password":      true,
	"123456789012":  true,
	"qwertyuiop12":  true,
	"administrator": true,
	"admin1234567":  true,
}

// ValidatePassword checks if password meets policy requirements
func ValidatePassword(password, username string) error {
	if len(password) < DefaultPolicy.MinLength {
		return fmt.Errorf("%w: минимальная длина %d символов", ErrWeakPassword, DefaultPolicy.MinLength)
	}
	if len(password) > DefaultPolicy.MaxLength {
		return fmt.Errorf("%w: максимальная длина %d символов", ErrWeakPassword, DefaultPolicy.MaxLength)
	}
	if password == username {
		return fmt.Errorf("%w: пароль не должен совпадать с именем пользователя", ErrWeakPassword)
	}
	if CommonPasswords[password] {
		return fmt.Errorf("%w: этот пароль слишком распространён", ErrWeakPassword)
	}
	return nil
}

// HashPassword creates an Argon2id hash of the password
func HashPassword(password string) (string, error) {
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	params := &argon2.Params{
		Memory:      64 * 1024, // 64 MB
		Iterations:  3,
		Parallelism: 4,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash := argon2.IDKey([]byte(password), salt, params.Time, params.Memory, uint8(params.Parallelism), params.KeyLength)

	// Encode as: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s",
		params.Memory, params.Time, params.Parallelism, b64Salt, b64Hash), nil
}

// VerifyPassword checks if password matches the hash
func VerifyPassword(password, hash string) error {
	// Parse the hash
	var version, m, t, p int
	var salt, hashPart string

	_, err := fmt.Sscanf(hash, "$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		&version, &m, &t, &p, &salt, &hashPart)
	if err != nil {
		return ErrInvalidPassword
	}

	saltBytes, err := base64.RawStdEncoding.DecodeString(salt)
	if err != nil {
		return ErrInvalidPassword
	}

	hashBytes, err := base64.RawStdEncoding.DecodeString(hashPart)
	if err != nil {
		return ErrInvalidPassword
	}

	params := &argon2.Params{
		Memory:      uint32(m),
		Iterations:  uint32(t),
		Parallelism: uint8(p),
		SaltLength:  16,
		KeyLength:   uint8(len(hashBytes)),
	}

	comparisonHash := argon2.IDKey([]byte(password), saltBytes, params.Time, params.Memory, uint8(params.Parallelism), params.KeyLength)

	if !constantTimeCompare(comparisonHash, hashBytes) {
		return ErrInvalidPassword
	}

	return nil
}

func constantTimeCompare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	result := byte(0)
	for i := range a {
		result |= a[i] ^ b[i]
	}
	return result == 0
}

// GenerateTOTPSecret generates a new TOTP secret (base32 encoded)
func GenerateTOTPSecret() (string, error) {
	secret := make([]byte, 20) // 160 bits for SHA-1
	if _, err := rand.Read(secret); err != nil {
		return "", err
	}

	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	return encoding.EncodeToString(secret), nil
}

// GenerateTOTPCode generates a TOTP code for the given secret and time
func GenerateTOTPCode(secret string, t time.Time) (string, error) {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	secretBytes, err := encoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	// Time step is 30 seconds
	counter := t.Unix() / 30

	// Create HMAC-SHA1
	h := hmac.New(sha1.New, secretBytes)
	
	// Counter as big-endian 8 bytes
	var counterBytes [8]byte
	binary.BigEndian.PutUint64(counterBytes[:], uint64(counter))
	
	h.Write(counterBytes[:])
	sum := h.Sum(nil)

	// Dynamic truncation
	offset := sum[len(sum)-1] & 0x0f
	code := binary.BigEndian.Uint32(sum[offset:]) & 0x7fffffff
	
	// Get 6 digit code
	digits := code % 1000000
	return fmt.Sprintf("%06d", digits), nil
}

// VerifyTOTP verifies a TOTP code with ±1 time step tolerance
func VerifyTOTP(secret, code string, t time.Time) (bool, error) {
	// Check current time step
	expected, err := GenerateTOTPCode(secret, t)
	if err != nil {
		return false, err
	}
	if code == expected {
		return true, nil
	}

	// Check previous time step (-30 seconds)
	expectedPrev, _ := GenerateTOTPCode(secret, t.Add(-30*time.Second))
	if code == expectedPrev {
		return true, nil
	}

	// Check next time step (+30 seconds)
	expectedNext, _ := GenerateTOTPCode(secret, t.Add(30*time.Second))
	if code == expectedNext {
		return true, nil
	}

	return false, ErrInvalidTOTP
}
