package cache

import (
	"context"
	"time"
)

// Redis is a stub implementation of Cache for CACHING_STRATEGY=redis.
// Get always misses; Set and Delete are no-ops so the app does not crash when redis is selected.
// Replace with a real Redis-backed implementation when REDIS_* env is used.
type Redis struct{}

// Get always returns (nil, false).
func (Redis) Get(_ context.Context, _ string) ([]byte, bool) {
	return nil, false
}

// Set is a no-op.
func (Redis) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

// Delete is a no-op.
func (Redis) Delete(_ context.Context, _ string) error {
	return nil
}

// DeleteByPrefix is a no-op.
func (Redis) DeleteByPrefix(_ context.Context, _ string) error {
	return nil
}

// Close is a no-op.
func (Redis) Close() {}
