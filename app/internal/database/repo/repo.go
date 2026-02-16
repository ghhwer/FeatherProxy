package repo

import (
	"errors"
	"time"

	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

// ErrProtocolMismatch is returned when a route links a source and target server with incompatible protocols.
var ErrProtocolMismatch = errors.New("source and target server must have the same protocol")

// Repository defines persistence for source/target servers, routes, and authentications.
type Repository interface {
	// Source servers
	CreateSourceServer(s schema.SourceServer) error
	GetSourceServer(uuid uuid.UUID) (schema.SourceServer, error)
	UpdateSourceServer(s schema.SourceServer) error
	DeleteSourceServer(uuid uuid.UUID) error
	ListSourceServers() ([]schema.SourceServer, error)
	// Server options (1:1 with source server; e.g. TLS for HTTPS)
	GetServerOptions(sourceServerUUID uuid.UUID) (schema.ServerOptions, error)
	SetServerOptions(opts schema.ServerOptions) error
	// ACL options (1:1 with source server; allow/deny by client IP/CIDR)
	GetACLOptions(sourceServerUUID uuid.UUID) (schema.ACLOptions, error)
	SetACLOptions(opts schema.ACLOptions) error
	// Target servers
	CreateTargetServer(t schema.TargetServer) error
	GetTargetServer(uuid uuid.UUID) (schema.TargetServer, error)
	UpdateTargetServer(t schema.TargetServer) error
	DeleteTargetServer(uuid uuid.UUID) error
	ListTargetServers() ([]schema.TargetServer, error)
	// Routes
	CreateRoute(route schema.Route) error
	GetRoute(routeUUID uuid.UUID) (schema.Route, error)
	UpdateRoute(route schema.Route) error
	DeleteRoute(routeUUID uuid.UUID) error
	ListRoutes() ([]schema.Route, error)
	GetRouteFromSourcePath(sourcePath string) (schema.Route, error)
	GetRouteFromTargetPath(targetPath string) (schema.Route, error)
	FindRouteBySourceMethodPath(sourceServerUUID uuid.UUID, method, sourcePath string) (schema.Route, error)
	// Authentications
	CreateAuthentication(a schema.Authentication) error
	GetAuthentication(id uuid.UUID) (schema.Authentication, error)
	GetAuthenticationWithPlainToken(id uuid.UUID) (schema.Authentication, error) // For proxy only; returns decrypted Token
	UpdateAuthentication(a schema.Authentication) error
	DeleteAuthentication(id uuid.UUID) error
	ListAuthentications() ([]schema.Authentication, error)
	// Route auth mappings
	ListSourceAuthsForRoute(routeUUID uuid.UUID) ([]schema.RouteSourceAuth, error)
	SetSourceAuthsForRoute(routeUUID uuid.UUID, authUUIDs []uuid.UUID) error
	GetTargetAuthForRoute(routeUUID uuid.UUID) (uuid.UUID, bool, error)
	SetTargetAuthForRoute(routeUUID uuid.UUID, authUUID *uuid.UUID) error
	GetTargetAuthenticationWithPlainToken(routeUUID uuid.UUID) (schema.Authentication, bool, error) // For proxy

	// Proxy stats (no cache; write-heavy)
	CreateProxyStats(stats []schema.ProxyStat) error
	ListProxyStats(limit, offset int, since *time.Time) ([]schema.ProxyStat, int64, error)
	DeleteProxyStatsOlderThan(until time.Time) (int64, error)
	ClearAllProxyStats() error
	StatsSummary() (schema.StatsSummary, error)
	StatsByRoute(since *time.Time, limit int) ([]schema.RouteCount, error)
	StatsByCaller(since *time.Time, limit int) ([]schema.CallerCount, error)
	StatsBySourceServer(since *time.Time) ([]schema.ServerCount, error)
	StatsByTargetServer(since *time.Time) ([]schema.ServerCount, error)
	StatsTPS(since time.Time, bucketDuration time.Duration) ([]schema.BucketCount, error)
}
