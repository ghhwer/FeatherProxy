package database

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

const (
	// SaltSize is the nonce/salt size for AES-GCM (12 bytes recommended).
	SaltSize = 12
	// KeySize is the AES-256 key size in bytes.
	KeySize = 32
)

// ErrEncryptionKeyMissing is returned when AUTH_ENCRYPTION_KEY is not set or too short.
var ErrEncryptionKeyMissing = errors.New("AUTH_ENCRYPTION_KEY must be set and at least 32 bytes (base64 or raw)")

// encryptionKey returns the key bytes from AUTH_ENCRYPTION_KEY env.
// Supports base64-encoded key or raw hex/string; must be at least KeySize bytes.
func encryptionKey() ([]byte, error) {
	raw := os.Getenv("AUTH_ENCRYPTION_KEY")
	if raw == "" {
		return nil, ErrEncryptionKeyMissing
	}
	// Try base64 first (e.g. openssl rand -base64 32)
	if key, err := base64.StdEncoding.DecodeString(raw); err == nil {
		if len(key) >= KeySize {
			return key[:KeySize], nil
		}
		return nil, ErrEncryptionKeyMissing
	}
	// Raw string: use as-is but must be at least KeySize bytes
	if len(raw) >= KeySize {
		return []byte(raw)[:KeySize], nil
	}
	return nil, ErrEncryptionKeyMissing
}

// EncryptToken encrypts plaintext with a new random salt using AES-GCM.
// Returns (ciphertextBase64, saltBase64, error). Salt is SaltSize bytes.
func EncryptToken(plaintext string) (ciphertextBase64, saltBase64 string, err error) {
	log.Printf("auth/token: EncryptToken called (plaintext len=%d)", len(plaintext))
	key, err := encryptionKey()
	if err != nil {
		log.Printf("auth/token: EncryptToken encryption key error: %v", err)
		return "", "", err
	}
	salt := make([]byte, SaltSize)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		log.Printf("auth/token: EncryptToken random salt error: %v", err)
		return "", "", fmt.Errorf("token: random salt: %w", err)
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", fmt.Errorf("token: aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", fmt.Errorf("token: gcm: %w", err)
	}
	ciphertext := aead.Seal(nil, salt, []byte(plaintext), nil)
	log.Printf("auth/token: EncryptToken success (ciphertext len=%d)", len(ciphertext))
	return base64.StdEncoding.EncodeToString(ciphertext), base64.StdEncoding.EncodeToString(salt), nil
}

// DecryptToken decrypts ciphertext using the given salt. Both can be base64-encoded.
func DecryptToken(ciphertextBase64, saltBase64 string) (plaintext string, err error) {
	log.Printf("auth/token: DecryptToken called (ciphertext base64 len=%d)", len(ciphertextBase64))
	key, err := encryptionKey()
	if err != nil {
		log.Printf("auth/token: DecryptToken encryption key error: %v", err)
		return "", err
	}
	ciphertext, err := base64.StdEncoding.DecodeString(ciphertextBase64)
	if err != nil {
		log.Printf("auth/token: DecryptToken decode ciphertext error: %v", err)
		return "", fmt.Errorf("token: decode ciphertext: %w", err)
	}
	salt, err := base64.StdEncoding.DecodeString(saltBase64)
	if err != nil {
		return "", fmt.Errorf("token: decode salt: %w", err)
	}
	if len(salt) != SaltSize {
		return "", fmt.Errorf("token: invalid salt length %d", len(salt))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("token: aes: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("token: gcm: %w", err)
	}
	plain, err := aead.Open(nil, salt, ciphertext, nil)
	if err != nil {
		log.Printf("auth/token: DecryptToken decrypt failed: %v", err)
		return "", fmt.Errorf("token: decrypt: %w", err)
	}
	log.Printf("auth/token: DecryptToken success (plaintext len=%d)", len(plain))
	return string(plain), nil
}
