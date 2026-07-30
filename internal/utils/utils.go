package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// HashToken creates a SHA-256 hash of a token with salt
func HashToken(token, salt string) string {
	h := hmac.New(sha256.New, []byte(salt))
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GenerateInviteCode generates a human-readable invite code
// Format: XXXX-XXXX-XXXX-XXXX using base32 without confusing chars
func GenerateInviteCode() (string, error) {
	// Use alphabet without confusing characters (0, O, 1, I, L)
	alpha := "ABCDEFGHJKMNPQRSTUVWXYZ23456789"
	result := make([]byte, 19) // 16 chars + 3 dashes

	for i := 0; i < 16; i++ {
		if i > 0 && i%4 == 0 {
			result[i+i/4-1] = '-'
		}
		b := make([]byte, 1)
		if _, err := rand.Read(b); err != nil {
			return "", err
		}
		result[i+i/4] = alpha[int(b[0])%len(alpha)]
	}

	return string(result), nil
}

// GenerateFileID generates a secure file ID (128-bit minimum)
func GenerateFileID() (string, error) {
	bytes := make([]byte, 16) // 128 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// SanitizeFilename removes dangerous characters from filename
func SanitizeFilename(name string) string {
	// Remove path separators
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	
	// Remove null bytes
	name = strings.ReplaceAll(name, "\x00", "")
	
	// Trim leading/trailing dots and spaces
	name = strings.Trim(name, ". ")
	
	// Limit length
	if len(name) > 255 {
		ext := filepath.Ext(name)
		base := name[:len(name)-len(ext)]
		if len(ext) > 20 {
			ext = ext[:20]
		}
		maxBase := 255 - len(ext)
		if maxBase > 0 {
			name = base[:maxBase] + ext
		} else {
			name = name[:255]
		}
	}
	
	if name == "" {
		name = "unnamed"
	}
	
	return name
}

// GetFileExtension returns lowercase extension without dot
func GetFileExtension(filename string) string {
	ext := filepath.Ext(filename)
	if len(ext) > 0 && ext[0] == '.' {
		return strings.ToLower(ext[1:])
	}
	return ""
}

// HasDoubleExtension checks for double extensions like .pdf.exe
func HasDoubleExtension(filename string) bool {
	ext := filepath.Ext(filename)
	if ext == "" {
		return false
	}
	
	base := filename[:len(filename)-len(ext)]
	secondExt := filepath.Ext(base)
	
	return secondExt != ""
}

// CopyFile copies a file from src to dst
func CopyFile(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer d.Close()

	if _, err := io.Copy(d, s); err != nil {
		os.Remove(dst)
		return err
	}

	return nil
}

// EncodeBase32 encodes bytes to base32 without padding
func EncodeBase32(data []byte) string {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	return encoding.EncodeToString(data)
}

// DecodeBase32 decodes base32 string
func DecodeBase32(s string) ([]byte, error) {
	encoding := base32.StdEncoding.WithPadding(base32.NoPadding)
	// Add padding if needed
	if len(s)%8 != 0 {
		s += strings.Repeat("=", 8-len(s)%8)
	}
	return encoding.DecodeString(s)
}

// SafeJoin joins paths and prevents path traversal
func SafeJoin(base, target string) (string, error) {
	// Clean the target
	target = filepath.Clean("/" + target)
	if strings.HasPrefix(target, "/..") {
		return "", fmt.Errorf("path traversal detected")
	}
	
	result := filepath.Join(base, filepath.Clean(target))
	
	// Ensure result is within base
	absBase, err := filepath.Abs(base)
	if err != nil {
		return "", err
	}
	absResult, err := filepath.Abs(result)
	if err != nil {
		return "", err
	}
	
	if !strings.HasPrefix(absResult, absBase) {
		return "", fmt.Errorf("path traversal detected")
	}
	
	return result, nil
}
