package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TargetServer is the database object (ORM entity) for the target_servers table.
type TargetServer struct {
	TargetServerUUID uuid.UUID      `gorm:"primaryKey"`
	Name             string         `gorm:"not null"`
	Protocol         string         `gorm:"not null"`
	Host             string         `gorm:"not null"`
	Port             int            `gorm:"not null"`
	BasePath         string         `gorm:"not null"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (TargetServer) TableName() string {
	return "target_servers"
}

// TargetServerToSchema maps the database object to the domain schema.
func TargetServerToSchema(t *TargetServer) schema.TargetServer {
	return schema.TargetServer{
		TargetServerUUID: t.TargetServerUUID,
		Name:             t.Name,
		Protocol:         t.Protocol,
		Host:             t.Host,
		Port:             t.Port,
		BasePath:         t.BasePath,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}

// SchemaToTargetServer maps the domain schema to the database object.
func SchemaToTargetServer(t schema.TargetServer) TargetServer {
	return TargetServer{
		TargetServerUUID: t.TargetServerUUID,
		Name:             t.Name,
		Protocol:         t.Protocol,
		Host:             t.Host,
		Port:             t.Port,
		BasePath:         t.BasePath,
		CreatedAt:        t.CreatedAt,
		UpdatedAt:        t.UpdatedAt,
	}
}
