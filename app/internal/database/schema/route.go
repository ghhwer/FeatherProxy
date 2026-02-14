package schema

import (
	"time"

	"github.com/google/uuid"
)

// Route is the domain schema for a route.
// Use this in application logic and for caching (e.g. Redis); do not depend on database objects.
// Keeps persistence details (columns, soft delete) separate from the rest of the app.
// Protocol is derived from the linked source/target servers; they must match.
type Route struct {
	RouteUUID         uuid.UUID `json:"route_uuid"`
	SourceServerUUID  uuid.UUID `json:"source_server_uuid"`
	TargetServerUUID  uuid.UUID `json:"target_server_uuid"`
	Method            string    `json:"method"`
	SourcePath        string    `json:"source_path"`
	TargetPath        string    `json:"target_path"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
