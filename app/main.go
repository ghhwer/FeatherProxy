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
		repo = database.NewCachedRepository(db.DB(), c, ttl)
		log.Println("cache: enabled")
	} else {
		repo = database.NewRepository(db.DB())
	}
	srv := server.NewServer(":4545", repo)
	proxyService := proxy.NewService(repo)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Println("server: listening on http://localhost:4545")
		if err := srv.Run(ctx); err != nil {
			log.Printf("server: %v", err)
		}
		log.Println("server: stopped")
	}()

	if err := proxyService.Run(ctx); err != nil {
		log.Fatalf("proxy: %v", err)
	}
	log.Println("proxy: stopped")
}
