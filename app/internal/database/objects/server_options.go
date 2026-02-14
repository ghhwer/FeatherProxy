package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ServerOptions is the database object (ORM entity) for the server_options table.
type ServerOptions struct {
	SourceServerUUID uuid.UUID      `gorm:"primaryKey"`
	TLSCertPath      string         `gorm:"column:tls_cert_path"`
	TLSKeyPath       string         `gorm:"column:tls_key_path"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (ServerOptions) TableName() string {
	return "server_options"
}

// ServerOptionsToSchema maps the database object to the domain schema.
func ServerOptionsToSchema(o *ServerOptions) schema.ServerOptions {
	return schema.ServerOptions{
		SourceServerUUID: o.SourceServerUUID,
		TLSCertPath:      o.TLSCertPath,
		TLSKeyPath:       o.TLSKeyPath,
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}

// SchemaToServerOptions maps the domain schema to the database object.
func SchemaToServerOptions(s schema.ServerOptions) ServerOptions {
	return ServerOptions{
		SourceServerUUID: s.SourceServerUUID,
		TLSCertPath:      s.TLSCertPath,
		TLSKeyPath:       s.TLSKeyPath,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
