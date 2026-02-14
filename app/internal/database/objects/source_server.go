package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SourceServer is the database object (ORM entity) for the source_servers table.
type SourceServer struct {
	SourceServerUUID uuid.UUID      `gorm:"primaryKey"`
	Name             string         `gorm:"not null"`
	Protocol         string         `gorm:"not null"`
	Host             string         `gorm:"not null"`
	Port             int            `gorm:"not null"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (SourceServer) TableName() string {
	return "source_servers"
}

// SourceServerToSchema maps the database object to the domain schema.
func SourceServerToSchema(s *SourceServer) schema.SourceServer {
	return schema.SourceServer{
		SourceServerUUID: s.SourceServerUUID,
		Name:             s.Name,
		Protocol:         s.Protocol,
		Host:             s.Host,
		Port:             s.Port,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}

// SchemaToSourceServer maps the domain schema to the database object.
func SchemaToSourceServer(s schema.SourceServer) SourceServer {
	return SourceServer{
		SourceServerUUID: s.SourceServerUUID,
		Name:             s.Name,
		Protocol:         s.Protocol,
		Host:             s.Host,
		Port:             s.Port,
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
