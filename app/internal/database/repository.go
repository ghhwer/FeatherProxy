package database

import (
	"FeatherProxy/app/internal/database/impl"
	"FeatherProxy/app/internal/database/repo"

	"gorm.io/gorm"
)

// Repository is the persistence interface for source/target servers, routes, and authentications.
// It is an alias for repo.Repository so the rest of the app can keep using database.Repository.
type Repository = repo.Repository

// ErrProtocolMismatch is returned when a route links a source and target server with incompatible protocols.
var ErrProtocolMismatch = repo.ErrProtocolMismatch

// NewRepository returns a Repository implementation backed by the given DB.
func NewRepository(db *gorm.DB) Repository {
	return impl.New(db)
}
