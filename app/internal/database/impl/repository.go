package impl

import (
	"FeatherProxy/app/internal/database/cache"
	"FeatherProxy/app/internal/database/repo"
	"time"

	"gorm.io/gorm"
)

// protocolsCompatible returns true if source and target protocols can be linked (http and https are allowed together).
func protocolsCompatible(source, target string) bool {
	if source == target {
		return true
	}
	return (source == "http" && target == "https") || (source == "https" && target == "http")
}

type repository struct {
	db    *gorm.DB
	c     cache.Cache
	ttl   time.Duration
}

// New returns a Repository implementation backed by the given DB (no cache).
func New(db *gorm.DB) repo.Repository {
	return &repository{db: db, c: cache.NoOp{}, ttl: cache.DefaultTTL}
}

// NewWithCache returns a Repository implementation backed by the given DB and cache.
func NewWithCache(db *gorm.DB, c cache.Cache, ttl time.Duration) repo.Repository {
	if c == nil {
		c = cache.NoOp{}
	}
	if ttl <= 0 {
		ttl = cache.DefaultTTL
	}
	return &repository{db: db, c: c, ttl: ttl}
}
