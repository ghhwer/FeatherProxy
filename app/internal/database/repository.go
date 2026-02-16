package database

import (
	"time"

	"FeatherProxy/app/internal/cache"
	"FeatherProxy/app/internal/database/impl"
	"FeatherProxy/app/internal/database/repo"

	"gorm.io/gorm"
)

// Repository is the persistence interface for source/target servers, routes, and authentications.
// It is an alias for repo.Repository so the rest of the app can keep using database.Repository.
type Repository = repo.Repository

// ErrProtocolMismatch is returned when a route links a source and target server with incompatible protocols.
var ErrProtocolMismatch = repo.ErrProtocolMismatch

// NewCachedRepository returns a Repository implementation backed by the given DB and cache.
func NewCachedRepository(db *gorm.DB, c cache.Cache, ttl time.Duration) Repository {
	return impl.NewWithCache(db, c, ttl)
}

func NewRepository(db *gorm.DB, c cache.Cache, ttl time.Duration) Repository {
	if c == nil {
		return impl.New(db)
	}
	return impl.NewWithCache(db, c, ttl)
}