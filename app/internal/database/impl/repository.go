package impl

import (
	"FeatherProxy/app/internal/database/repo"

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
	db *gorm.DB
}

// New returns a Repository implementation backed by the given DB.
func New(db *gorm.DB) repo.Repository {
	return &repository{db: db}
}
