package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/cache"
	"FeatherProxy/app/internal/proxy"
	"FeatherProxy/app/internal/ui_server"

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

	var repo database.Repository
	if c, ttl, err := cache.FromEnv(); err != nil {
		log.Printf("cache: %v (using repo without cache)", err)
		repo = database.NewRepository(db.DB())
	} else if c != nil {
		defer c.Close() // stop cache goroutines so process can exit on Ctrl+C
		repo = database.NewCachedRepository(db.DB(), c, ttl)
		log.Println("cache: enabled")
	} else {
		repo = database.NewRepository(db.DB())
	}
	reloadChan := make(chan struct{}, 1)
	srv := server.NewServer(":4545", repo, "internal/ui_server/static", func() {
		select {
		case reloadChan <- struct{}{}:
		default:
		}
	})
	proxyService := proxy.NewService(repo)
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
