package cache

import (
	"context"
	"time"
)

// NoOp is a cache implementation that does nothing: Get always misses, Set and Delete are no-ops.
type NoOp struct{}

// Get always returns (nil, false).
func (NoOp) Get(_ context.Context, _ string) ([]byte, bool) {
	return nil, false
}

// Set is a no-op.
func (NoOp) Set(_ context.Context, _ string, _ []byte, _ time.Duration) error {
	return nil
}

// Delete is a no-op.
func (NoOp) Delete(_ context.Context, _ string) error {
	return nil
}

// DeleteByPrefix is a no-op.
func (NoOp) DeleteByPrefix(_ context.Context, _ string) error {
	return nil
}

// Close is a no-op.
func (NoOp) Close() {}

