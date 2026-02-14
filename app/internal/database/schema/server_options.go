package schema

import (
	"time"

	"github.com/google/uuid"
)

// ServerOptions is the domain schema for protocol-specific options attached to a source server (e.g. TLS for HTTPS).
type ServerOptions struct {
	SourceServerUUID uuid.UUID `json:"source_server_uuid"`
	TLSCertPath      string    `json:"tls_cert_path"`
	TLSKeyPath       string    `json:"tls_key_path"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
