package impl

import (
	"context"
	"encoding/json"
	"time"

	"FeatherProxy/app/internal/cache"

	"github.com/google/uuid"
)

// Cache key prefixes and builders. Keep in sync with invalidation in mutation methods.
const (
	keyPrefixSourceServer       = "source_server:"
	keyPrefixTargetServer       = "target_server:"
	keyPrefixRoute              = "route:"
	keyPrefixRouteSourcePath     = "route:source_path:"
	keyPrefixRouteTargetPath     = "route:target_path:"
	keyPrefixRouteMethodPath    = "route:method_path:"
	keyPrefixAuth               = "auth:"
	keyListSourceServers         = "list:source_servers"
	keyListTargetServers         = "list:target_servers"
	keyListRoutes                = "list:routes"
	keyListAuthentications       = "list:authentications"
	keyPrefixRouteSourceAuths    = "route_source_auths:"
	keyPrefixTargetAuthForRoute  = "target_auth_for_route:"
	keyPrefixServerOptions       = "server_options:"
	keyPrefixACLOptions          = "acl_options:"
)

func keySourceServer(id uuid.UUID) string              { return keyPrefixSourceServer + id.String() }
func keyTargetServer(id uuid.UUID) string              { return keyPrefixTargetServer + id.String() }
func keyRoute(id uuid.UUID) string                     { return keyPrefixRoute + id.String() }
func keyRouteSourcePath(path string) string            { return keyPrefixRouteSourcePath + path }
func keyRouteTargetPath(path string) string             { return keyPrefixRouteTargetPath + path }
func keyRouteMethodPath(sourceID uuid.UUID, method, path string) string {
	return keyPrefixRouteMethodPath + sourceID.String() + ":" + method + ":" + path
}
func keyAuth(id uuid.UUID) string                   { return keyPrefixAuth + id.String() }
func keyRouteSourceAuths(routeID uuid.UUID) string   { return keyPrefixRouteSourceAuths + routeID.String() }
func keyTargetAuthForRoute(routeID uuid.UUID) string { return keyPrefixTargetAuthForRoute + routeID.String() }
func keyServerOptions(sourceID uuid.UUID) string    { return keyPrefixServerOptions + sourceID.String() }
func keyACLOptions(sourceID uuid.UUID) string      { return keyPrefixACLOptions + sourceID.String() }

func (r *repository) cacheCtx() context.Context { return context.Background() }

func (r *repository) cacheTTL() time.Duration {
	if r.ttl > 0 {
		return r.ttl
	}
	return cache.DefaultTTL
}

// getCached returns the value from cache if present; otherwise calls delegate, stores result, and returns it.
func getCached[T any](r *repository, key string, delegate func() (T, error)) (T, error) {
	var zero T
	if b, ok := r.c.Get(r.cacheCtx(), key); ok {
		if err := json.Unmarshal(b, &zero); err != nil {
			return zero, err
		}
		return zero, nil
	}
	val, err := delegate()
	if err != nil {
		return zero, err
	}
	b, _ := json.Marshal(val)
	_ = r.c.Set(r.cacheCtx(), key, b, r.cacheTTL())
	return val, nil
}

// invalidate runs the mutation; on success deletes the given keys and prefixes and returns nil.
func (r *repository) invalidate(err error, keys []string, prefixes []string) error {
	if err != nil {
		return err
	}
	for _, k := range keys {
		_ = r.c.Delete(r.cacheCtx(), k)
	}
	for _, p := range prefixes {
		_ = r.c.DeleteByPrefix(r.cacheCtx(), p)
	}
	return nil
}

// targetAuthCached is the cached shape for GetTargetAuthForRoute (uuid.UUID, bool).
type targetAuthCached struct {
	AuthUUID uuid.UUID `json:"auth_uuid"`
	OK      bool      `json:"ok"`
}
