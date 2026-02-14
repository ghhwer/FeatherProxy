package schema

import (
	"time"

	"github.com/google/uuid"
)

// Authentication is the domain schema for an authentication credential.
// Token is used for API input (create/update) and for proxy use (decrypted); never persisted in plain form.
// TokenMasked is set for API responses (e.g. "***") when returning to the UI.
type Authentication struct {
	AuthenticationUUID uuid.UUID `json:"authentication_uuid"`
	Name               string    `json:"name"`
	TokenType          string    `json:"token_type"`
	Token              string    `json:"token,omitempty"`       // Input on create/update; decrypted only when needed for proxy
	TokenMasked        string    `json:"token_masked,omitempty"` // Set in API responses; never stored
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
