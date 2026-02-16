package proxy

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"strings"
	"time"

	"FeatherProxy/app/internal/cache"
)

// Caching key prefix
const cachingKeyPrefix = "dns:"

// HostnameResolver performs reverse DNS lookups for client IPs.
// Implementations should be safe for concurrent use.
type HostnameResolver interface {
	// ReverseLookup returns the hostnames associated with the given IP.
	// Implementations may return a cached result.
	ReverseLookup(ctx context.Context, ip net.IP) ([]string, error)
}

type cachingResolver struct {
	cache cache.Cache
	ttl   time.Duration
}

type noCacheResolver struct{}

func ReverseLookup(ctx context.Context, ip net.IP) ([]string, error) {
	log.Println("ReverseLookup", "ip", ip.String())
	if ip == nil {
		return nil, nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	ptrs, err := net.DefaultResolver.LookupAddr(ctx, ip.String())
	hostnames := make([]string, 0, len(ptrs))
	if err == nil {
		for _, h := range ptrs {
			h = strings.TrimSpace(h)
			if h == "" {
				continue
			}
			// Trim trailing dot from FQDN and lower-case for comparison.
			h = strings.TrimSuffix(h, ".")
			h = strings.ToLower(h)
			hostnames = append(hostnames, h)
		}
	}
	log.Println("ReverseLookup", "hostnames", hostnames)
	log.Println("ReverseLookup", "err", err)
	return hostnames, err
}

func NewResolver(cache cache.Cache, ttl time.Duration) HostnameResolver {
	if cache == nil {
		return &noCacheResolver{}
	}
	return &cachingResolver{cache: cache, ttl: ttl}
}

func (r *cachingResolver) ReverseLookup(ctx context.Context, ip net.IP) ([]string, error) {
	var key = cachingKeyPrefix + ip.String()
	if b, ok := r.cache.Get(ctx, key); ok {
		var hostnames []string
		if err := json.Unmarshal(b, &hostnames); err == nil {
			return hostnames, nil
		}
		// On unmarshal error, fall through and refresh the cache below.
	}

	// Cache did not hit (or was invalid), lookup the IP and cache the result.
	result, err := ReverseLookup(ctx, ip)
	if err != nil {
		return nil, err
	}
	if b, err := json.Marshal(result); err == nil {
		_ = r.cache.Set(ctx, key, b, r.ttl)
	}
	return result, nil
}

func (r *noCacheResolver) ReverseLookup(ctx context.Context, ip net.IP) ([]string, error) {
	return ReverseLookup(ctx, ip)
}
