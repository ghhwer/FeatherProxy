package token

import (
	"encoding/base64"
	"errors"
	"os"
	"strings"
	"testing"
)

// testKeyBase64 is a valid 32-byte key (base64). Generated for tests only.
const testKeyBase64 = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

func setTestKey(t *testing.T, key string) {
	t.Helper()
	if key != "" {
		os.Setenv("AUTH_ENCRYPTION_KEY", key)
	} else {
		os.Unsetenv("AUTH_ENCRYPTION_KEY")
	}
	t.Cleanup(func() {
		os.Unsetenv("AUTH_ENCRYPTION_KEY")
	})
}

func TestEncryptToken_DecryptToken_roundtrip(t *testing.T) {
	setTestKey(t, testKeyBase64)
	plain := "secret-token-123"
	ciphertext, salt, err := EncryptToken(plain)
	if err != nil {
		t.Fatalf("EncryptToken: %v", err)
	}
	if ciphertext == "" || salt == "" {
		t.Error("EncryptToken returned empty ciphertext or salt")
	}
	got, err := DecryptToken(ciphertext, salt)
	if err != nil {
		t.Fatalf("DecryptToken: %v", err)
	}
	if got != plain {
		t.Errorf("DecryptToken = %q, want %q", got, plain)
	}
}

func TestEncryptToken_keyMissing(t *testing.T) {
	setTestKey(t, "")
	_, _, err := EncryptToken("x")
	if err == nil {
		t.Fatal("EncryptToken: want error when key missing")
	}
	if !errors.Is(err, ErrEncryptionKeyMissing) {
		t.Errorf("EncryptToken: got %v, want ErrEncryptionKeyMissing", err)
	}
}

func TestEncryptToken_keyTooShort(t *testing.T) {
	// Key must be at least 32 bytes. Base64 of 16 bytes is too short after decode.
	setTestKey(t, base64.StdEncoding.EncodeToString(make([]byte, 16)))
	_, _, err := EncryptToken("x")
	if err == nil {
		t.Fatal("EncryptToken: want error when key too short")
	}
	if !errors.Is(err, ErrEncryptionKeyMissing) {
		t.Errorf("EncryptToken: got %v, want ErrEncryptionKeyMissing", err)
	}
}

func TestDecryptToken_keyMissing(t *testing.T) {
	setTestKey(t, "")
	_, err := DecryptToken("any", "any")
	if err == nil {
		t.Fatal("DecryptToken: want error when key missing")
	}
	if !errors.Is(err, ErrEncryptionKeyMissing) {
		t.Errorf("DecryptToken: got %v, want ErrEncryptionKeyMissing", err)
	}
}

func TestDecryptToken_badCiphertextBase64(t *testing.T) {
	setTestKey(t, testKeyBase64)
	_, err := DecryptToken("not-valid-base64!!!", "AAAAAAAAAAAAAAAAAAAAAA==")
	if err == nil {
		t.Fatal("DecryptToken: want error for bad ciphertext base64")
	}
	if err != nil && !strings.Contains(err.Error(), "decode ciphertext") {
		t.Errorf("DecryptToken: expected decode ciphertext error, got %v", err)
	}
}

func TestDecryptToken_badSaltBase64(t *testing.T) {
	setTestKey(t, testKeyBase64)
	// Valid ciphertext length base64 (at least SaltSize + 16 for GCM tag); salt invalid
	cipherB64 := base64.StdEncoding.EncodeToString(make([]byte, SaltSize+16))
	_, err := DecryptToken(cipherB64, "!!!")
	if err == nil {
		t.Fatal("DecryptToken: want error for bad salt base64")
	}
	if err != nil && !strings.Contains(err.Error(), "decode salt") {
		t.Errorf("DecryptToken: expected decode salt error, got %v", err)
	}
}

func TestDecryptToken_wrongSaltLength(t *testing.T) {
	setTestKey(t, testKeyBase64)
	// Salt must be exactly SaltSize (12). Use 8 bytes base64.
	shortSalt := base64.StdEncoding.EncodeToString(make([]byte, 8))
	cipherB64 := base64.StdEncoding.EncodeToString(make([]byte, SaltSize+16))
	_, err := DecryptToken(cipherB64, shortSalt)
	if err == nil {
		t.Fatal("DecryptToken: want error for wrong salt length")
	}
	if err != nil && !strings.Contains(err.Error(), "invalid salt length") {
		t.Errorf("DecryptToken: expected invalid salt length, got %v", err)
	}
}

func TestDecryptToken_tamperedCiphertext(t *testing.T) {
	setTestKey(t, testKeyBase64)
	ciphertext, salt, err := EncryptToken("secret")
	if err != nil {
		t.Fatal(err)
	}
	// Tamper: decode, flip a byte, re-encode
	raw, _ := base64.StdEncoding.DecodeString(ciphertext)
	if len(raw) > 0 {
		raw[0] ^= 0xff
	}
	tampered := base64.StdEncoding.EncodeToString(raw)
	_, err = DecryptToken(tampered, salt)
	if err == nil {
		t.Fatal("DecryptToken: want error for tampered ciphertext")
	}
	if err != nil && !strings.Contains(err.Error(), "decrypt") {
		t.Errorf("DecryptToken: expected decrypt error, got %v", err)
	}
}

func TestDecryptToken_wrongKey(t *testing.T) {
	setTestKey(t, testKeyBase64)
	ciphertext, salt, err := EncryptToken("secret")
	if err != nil {
		t.Fatal(err)
	}
	// Decrypt with different key (32 bytes)
	setTestKey(t, base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")))
	_, err = DecryptToken(ciphertext, salt)
	if err == nil {
		t.Fatal("DecryptToken: want error when decrypting with wrong key")
	}
}
