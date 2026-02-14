package database

import (
	"fmt"
	"os"
	"sync"

	"FeatherProxy/app/internal/database/objects"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Handler holds the ORM connection and provides access to the database.
// Config is read from .env (DB_DRIVER, DB_DSN). Safe for concurrent use.
type Handler struct {
	mu sync.RWMutex
	db *gorm.DB
}

// NewHandler opens a database connection using DB_DRIVER and DB_DSN from the environment.
// DB_DRIVER: "postgres" or "sqlite". DB_DSN: driver-specific connection string.
// Returns error if env is missing or connection fails.
func NewHandler() (*Handler, error) {
	driver := os.Getenv("DB_DRIVER")
	dsn := os.Getenv("DB_DSN")

	if driver == "" || dsn == "" {
		return nil, fmt.Errorf("database: DB_DRIVER and DB_DSN must be set in .env")
	}

	var dialector gorm.Dialector
	switch driver {
	case "postgres", "postgresql":
		dialector = postgres.Open(dsn)
	case "sqlite":
		dialector = sqlite.Open(dsn)
	default:
		return nil, fmt.Errorf("database: unsupported DB_DRIVER %q (use postgres or sqlite)", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("database: open: %w", err)
	}

	return &Handler{db: db}, nil
}

// DB returns the GORM DB instance for running queries and migrations.
// Use this only to work with database objects (entities); expose data to the app via schemas.
func (h *Handler) DB() *gorm.DB {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.db
}

// AutoMigrate creates/updates tables for all registered database objects.
func (h *Handler) AutoMigrate() error {
	return h.DB().AutoMigrate(
		&objects.SourceServer{},
		&objects.ServerOptions{},
		&objects.TargetServer{},
		&objects.Route{},
		&objects.Authentication{},
		&objects.RouteSourceAuth{},
		&objects.RouteTargetAuth{},
	)
}

// Close closes the underlying database connection. Call from main on shutdown.
func (h *Handler) Close() error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.db == nil {
		return nil
	}
	sqlDB, err := h.db.DB()
	if err != nil {
		return err
	}
	h.db = nil
	return sqlDB.Close()
}
