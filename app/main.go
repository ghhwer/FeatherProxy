package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"FeatherProxy/app/internal/cache"
	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/proxy"
	server "FeatherProxy/app/internal/ui_server"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	db, err := database.NewHandler()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("database close: %v", err)
		}
	}()

	if err := db.AutoMigrate(); err != nil {
		log.Fatalf("database migrate: %v", err)
	}
	log.Println("database: connected and migrated")

	// Initialize the repository and cache.
	var repo database.Repository
	var sharedCache cache.Cache
	var cacheTTL time.Duration
	sharedCache, cacheTTL, err = cache.FromEnv()

	if err != nil {
		log.Printf("cache: %v (cache not enabled)", err)
	}
	repo = database.NewRepository(db.DB(), sharedCache, cacheTTL)

	// UI server.
	// Reload channel to restart the proxy when the UI is reloaded.
	reloadChan := make(chan struct{}, 1)
	srv := server.NewServer(":4545", repo, "internal/ui_server/static", func() {
		select {
		case reloadChan <- struct{}{}:
		default:
		}
	})
	// Proxy service.
	proxyService := proxy.NewService(repo, sharedCache, cacheTTL)
	// Context to stop the server and proxy.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Cancel the whole process if the UI server fails (e.g. port in use).
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		log.Println("server: listening on http://localhost:4545")
		if err := srv.Run(runCtx); err != nil {
			log.Printf("server: %v", err)
			cancel() // so proxy exits and process terminates
		}
		log.Println("server: stopped")
	}()

	// Run proxy under a cancellable context so "reload" from UI can restart it (pick up new source servers).
	for {
		proxyCtx, proxyCancel := context.WithCancel(runCtx)
		proxyDone := make(chan struct{})
		go func() {
			_ = proxyService.Run(proxyCtx)
			close(proxyDone)
		}()
		select {
		case <-runCtx.Done():
			proxyCancel()
			<-proxyDone
			log.Println("proxy: stopped")
			return
		case <-reloadChan:
			log.Println("proxy: reload requested, restarting…")
			proxyCancel()
			<-proxyDone
			log.Println("proxy: restarted")
		}
	}
}
