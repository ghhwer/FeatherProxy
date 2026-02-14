package cache

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type memoryItem struct {
	value    []byte
	expireAt time.Time
}

// Memory is an in-memory cache with per-entry TTL.
// Expired entries are removed on Get (lazy) and by a periodic cleanup goroutine.
// Stats (hits, misses, sets, deletes) are tracked and logged periodically.
type Memory struct {
	mu         sync.RWMutex
	items      map[string]memoryItem
	defaultTTL time.Duration
	stop       chan struct{}

	hits   atomic.Uint64
	misses atomic.Uint64
	sets   atomic.Uint64
	deletesOperations atomic.Uint64
	evictionsOperations atomic.Uint64
}

// NewMemory returns an in-memory cache with the given default TTL.
// Call Close() to stop the cleanup goroutine when the cache is no longer needed.
func NewMemory(defaultTTL time.Duration) *Memory {
	if defaultTTL <= 0 {
		defaultTTL = DefaultTTL
	}
	m := &Memory{
		items:      make(map[string]memoryItem),
		defaultTTL: defaultTTL,
		stop:       make(chan struct{}),
	}
	go m.cleanupLoop()
	return m
}

// Get returns the value for key if present and not expired. Expired entries are removed.
func (m *Memory) Get(ctx context.Context, key string) ([]byte, bool) {
	m.mu.Lock()
	item, ok := m.items[key]
	if ok && time.Now().After(item.expireAt) {
		delete(m.items, key)
		ok = false
	}
	m.mu.Unlock()
	if !ok {
		m.misses.Add(1)
		return nil, false
	}
	m.hits.Add(1)
	return item.value, true
}

// Set stores value for key with the given ttl. If ttl <= 0, defaultTTL is used.
func (m *Memory) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = m.defaultTTL
	}
	expireAt := time.Now().Add(ttl)
	// Copy value so caller can reuse the slice
	valCopy := make([]byte, len(value))
	copy(valCopy, value)
	m.mu.Lock()
	m.items[key] = memoryItem{value: valCopy, expireAt: expireAt}
	m.mu.Unlock()
	m.sets.Add(1)
	return nil
}

// Delete removes key from the cache.
func (m *Memory) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	m.deletesOperations.Add(1)
	return nil
}

// DeleteByPrefix removes all keys that start with prefix.
func (m *Memory) DeleteByPrefix(ctx context.Context, prefix string) error {
	m.mu.Lock()
	n := 0
	for k := range m.items {
		if len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			delete(m.items, k)
			n++
		}
	}
	m.mu.Unlock()
	m.deletesOperations.Add(uint64(n))
	return nil
}

// cleanupLoop periodically removes expired entries and logs cache stats.
func (m *Memory) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stop:
			return
		case <-ticker.C:
			m.evictExpired()
			m.logStats()
		}
	}
}

// logStats logs current cache statistics (hits, misses, sets, deletes, entries).
func (m *Memory) logStats() {
	h := m.hits.Load()
	miss := m.misses.Load()
	s := m.sets.Load()
	d := m.deletesOperations.Load()
	e := m.evictionsOperations.Load()
	m.mu.RLock()
	entries := len(m.items)
	m.mu.RUnlock()
	log.Printf("cache/memory: stats hits=%d misses=%d sets=%d deletes=%d evictions=%d entries=%d",
		h, miss, s, d, e, entries)
	m.hits.Store(0)
	m.misses.Store(0)
	m.sets.Store(0)
	m.deletesOperations.Store(0)
	m.evictionsOperations.Store(0)
}

func (m *Memory) evictExpired() {
	now := time.Now()
	m.mu.Lock()
	deleted := 0
	for k, item := range m.items {
		if now.After(item.expireAt) {
			delete(m.items, k)
			deleted++
		}
	}
	m.evictionsOperations.Add(uint64(deleted))
	log.Printf("cache/memory: evicted %d expired entries", deleted)
	m.mu.Unlock()
}

// Close stops the cleanup goroutine. Call when the cache is no longer needed.
func (m *Memory) Close() {
	close(m.stop)
}
