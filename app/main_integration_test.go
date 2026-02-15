//go:build integration

package main

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/ui_server"
)

const integrationTestPort = ":19545"

func TestMainIntegration_serverResponds(t *testing.T) {
	// Same env as plan: sqlite in-memory
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DSN", "file::memory:?cache=shared")
	os.Unsetenv("CACHING_STRATEGY")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_DSN")
	}()

	db, err := database.NewHandler()
	if err != nil {
		t.Fatalf("NewHandler: %v", err)
	}
	defer db.Close()

	if err := db.AutoMigrate(); err != nil {
		t.Fatalf("AutoMigrate: %v", err)
	}

	repo := database.NewRepository(db.DB())
	srv := server.NewServer(integrationTestPort, repo, "internal/ui_server/static", nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = srv.Run(ctx)
	}()

	// Wait for server to listen
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get("http://localhost" + integrationTestPort + "/api/source-servers")
	if err != nil {
		t.Fatalf("GET /api/source-servers: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GET /api/source-servers status = %d, want 200", resp.StatusCode)
	}
}
