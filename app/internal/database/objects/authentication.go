package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Authentication is the database object (ORM entity) for the authentications table.
// Token is stored encrypted with salt; use database.EncryptToken/DecryptToken.
type Authentication struct {
	AuthenticationUUID uuid.UUID `gorm:"primaryKey"`
	Name               string    `gorm:"not null"`
	TokenType          string    `gorm:"not null"`
	TokenEncrypted     string    `gorm:"not null;column:token_encrypted"`
	TokenSalt          string    `gorm:"not null;column:token_salt"`
	CreatedAt          time.Time `gorm:"not null"`
	UpdatedAt          time.Time `gorm:"not null"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (Authentication) TableName() string {
	return "authentications"
}

// AuthenticationToSchema maps the database object to the domain schema (no decryption).
// For API responses, set Token to empty and TokenMasked to "***" in the handler.
func AuthenticationToSchema(a *Authentication) schema.Authentication {
	return schema.Authentication{
		AuthenticationUUID: a.AuthenticationUUID,
		Name:               a.Name,
		TokenType:          a.TokenType,
		Token:              "", // Never fill from DB in this path; use repo GetAuthenticationWithPlainToken for proxy
		CreatedAt:          a.CreatedAt,
		UpdatedAt:          a.UpdatedAt,
	}
}

// SchemaToAuthentication maps the domain schema to the database object.
// Caller must have already set TokenEncrypted and TokenSalt (via crypto helper); this only copies other fields.
func SchemaToAuthentication(s schema.Authentication, tokenEncrypted, tokenSalt string) Authentication {
	return Authentication{
		AuthenticationUUID: s.AuthenticationUUID,
		Name:               s.Name,
		TokenType:          s.TokenType,
		TokenEncrypted:     tokenEncrypted,
		TokenSalt:          tokenSalt,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}
