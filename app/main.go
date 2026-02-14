package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/server"

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

	repo := database.NewRepository(db.DB())
	srv := server.NewServer(":4545", repo)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("server: listening on http://localhost:4545")
	if err := srv.Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
	log.Println("server: stopped")
}
