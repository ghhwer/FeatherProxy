package cache

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

// Cache is the abstract caching interface. Implementations may be in-memory, Redis, or no-op.
// Values are opaque bytes; callers use encoding/json for schema types.
// Call Close() when the cache is no longer needed so background goroutines and connections can exit.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, bool)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	// DeleteByPrefix removes all keys that start with prefix. Used for bulk invalidation (e.g. "route:").
	DeleteByPrefix(ctx context.Context, prefix string) error
	Close()
}

// DefaultTTL is the default cache TTL when CACHE_TTL is not set.
const DefaultTTL = 5 * time.Minute

// FromEnv returns a Cache and TTL from environment variables.
// CACHING_STRATEGY: "none", "memory", or "redis". Empty or "none" returns a no-op cache.
// CACHE_TTL: optional duration (e.g. "5m", "1h"). Uses DefaultTTL if unset or invalid.
// Returns (nil, 0, error) for unsupported or not-yet-implemented strategies so the caller can skip wrapping.
func FromEnv() (Cache, time.Duration, error) {
	strategy := strings.ToLower(strings.TrimSpace(os.Getenv("CACHING_STRATEGY")))
	ttl := parseTTL(os.Getenv("CACHE_TTL"))

	switch strategy {
	case "", "none":
		return nil, ttl, nil // No cache; caller uses repo as-is
	case "memory":
		return NewMemory(ttl), ttl, nil
	case "redis":
		// Stub: behaves like no-op until a real Redis client is wired from REDIS_* env.
		return Redis{}, ttl, nil
	default:
		return nil, 0, fmt.Errorf("unknown CACHING_STRATEGY %q (use none, memory, or redis)", strategy)
	}
}

func parseTTL(s string) time.Duration {
	s = strings.TrimSpace(s)
	if s == "" {
		return DefaultTTL
	}
	d, err := time.ParseDuration(s)
	if err != nil {
		return DefaultTTL
	}
	if d <= 0 {
		return DefaultTTL
	}
	return d
}

