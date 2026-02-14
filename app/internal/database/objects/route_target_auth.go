package objects

import (
	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RouteTargetAuth is the database object for the route_target_auths table (one per route).
type RouteTargetAuth struct {
	RouteUUID          uuid.UUID      `gorm:"primaryKey"`
	AuthenticationUUID uuid.UUID      `gorm:"not null"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (RouteTargetAuth) TableName() string {
	return "route_target_auths"
}

// RouteTargetAuthToSchema maps the database object to the domain schema.
func RouteTargetAuthToSchema(r *RouteTargetAuth) schema.RouteTargetAuth {
	return schema.RouteTargetAuth{
		RouteUUID:          r.RouteUUID,
		AuthenticationUUID: r.AuthenticationUUID,
	}
}

// SchemaToRouteTargetAuth maps the domain schema to the database object.
func SchemaToRouteTargetAuth(r schema.RouteTargetAuth) RouteTargetAuth {
	return RouteTargetAuth{
		RouteUUID:          r.RouteUUID,
		AuthenticationUUID: r.AuthenticationUUID,
	}
}
