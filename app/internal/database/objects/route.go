package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Route is the database object (ORM entity) for the routes table.
// This package defines only persistence shape; use schema.Route for app and cache layers.
type Route struct {
	RouteUUID        uuid.UUID      `gorm:"primaryKey"`
	SourceServerUUID uuid.UUID      `gorm:"not null"`
	TargetServerUUID uuid.UUID      `gorm:"not null"`
	Method           string         `gorm:"not null"`
	SourcePath       string         `gorm:"not null"`
	TargetPath       string         `gorm:"not null"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (Route) TableName() string {
	return "routes"
}

// RouteToSchema maps the database object to the domain schema (for app and cache use).
func RouteToSchema(r *Route) schema.Route {
	return schema.Route{
		RouteUUID:        r.RouteUUID,
		SourceServerUUID:  r.SourceServerUUID,
		TargetServerUUID:  r.TargetServerUUID,
		Method:            r.Method,
		SourcePath:        r.SourcePath,
		TargetPath:        r.TargetPath,
		CreatedAt:         r.CreatedAt,
		UpdatedAt:         r.UpdatedAt,
	}
}

// SchemaToRoute maps the domain schema to the database object (for database use).
func SchemaToRoute(r schema.Route) Route {
	return Route{
		RouteUUID:        r.RouteUUID,
		SourceServerUUID: r.SourceServerUUID,
		TargetServerUUID: r.TargetServerUUID,
		Method:           r.Method,
		SourcePath:       r.SourcePath,
		TargetPath:       r.TargetPath,
		CreatedAt:        r.CreatedAt,
		UpdatedAt:        r.UpdatedAt,
	}
}