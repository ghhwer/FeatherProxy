package database

import (
	"FeatherProxy/app/internal/database/token"
)

// Re-export token package for callers that use database.EncryptToken / database.DecryptToken.

var (
	ErrEncryptionKeyMissing = token.ErrEncryptionKeyMissing
	EncryptToken            = token.EncryptToken
	DecryptToken            = token.DecryptToken
)

const (
	SaltSize = token.SaltSize
	KeySize  = token.KeySize
)
