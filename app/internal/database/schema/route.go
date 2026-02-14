package schema

import (
	"time"

	"github.com/google/uuid"
)

// Route is the domain schema for a route.
// Use this in application logic and for caching (e.g. Redis); do not depend on database objects.
// Keeps persistence details (columns, soft delete) separate from the rest of the app.
type Route struct {
	RouteUUID 		uuid.UUID `json:"route_uuid"`
	Protocol  		string    `json:"protocol"`
	Method          string    `json:"method"`
	SourcePath      string    `json:"source_path"`
	TargetPath      string    `json:"target_path"`
	CreatedAt 		time.Time `json:"created_at"`
	UpdatedAt 		time.Time `json:"updated_at"`
}
