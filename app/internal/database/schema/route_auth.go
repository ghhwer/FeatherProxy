package schema

import "github.com/google/uuid"

// RouteSourceAuth links a route to one of its allowed source authentications (list).
type RouteSourceAuth struct {
	RouteUUID         uuid.UUID `json:"route_uuid"`
	AuthenticationUUID uuid.UUID `json:"authentication_uuid"`
	Position          int       `json:"position"` // Order in the list
}

// RouteTargetAuth links a route to its single target authentication.
type RouteTargetAuth struct {
	RouteUUID          uuid.UUID `json:"route_uuid"`
	AuthenticationUUID uuid.UUID `json:"authentication_uuid"`
}
