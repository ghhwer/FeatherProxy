package objects

import (
	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// RouteSourceAuth is the database object for the route_source_auths junction table.
type RouteSourceAuth struct {
	RouteUUID          uuid.UUID      `gorm:"primaryKey;uniqueIndex:idx_route_auth"`
	AuthenticationUUID uuid.UUID      `gorm:"primaryKey;uniqueIndex:idx_route_auth"`
	Position           int            `gorm:"not null;default:0"`
	DeletedAt          gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (RouteSourceAuth) TableName() string {
	return "route_source_auths"
}

// RouteSourceAuthToSchema maps the database object to the domain schema.
func RouteSourceAuthToSchema(r *RouteSourceAuth) schema.RouteSourceAuth {
	return schema.RouteSourceAuth{
		RouteUUID:          r.RouteUUID,
		AuthenticationUUID: r.AuthenticationUUID,
		Position:           r.Position,
	}
}

// SchemaToRouteSourceAuth maps the domain schema to the database object.
func SchemaToRouteSourceAuth(r schema.RouteSourceAuth) RouteSourceAuth {
	return RouteSourceAuth{
		RouteUUID:          r.RouteUUID,
		AuthenticationUUID: r.AuthenticationUUID,
		Position:           r.Position,
	}
}
