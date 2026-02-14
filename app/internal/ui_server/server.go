package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"FeatherProxy/app/internal/database"
)

// Server runs the HTTP service for the route management UI and API.
type Server struct {
	addr       string
	staticDir  string
	httpServer *http.Server
	repo       database.Repository
}

// NewServer builds a server that serves the UI and route API on the given address.
// staticDir is the path to the directory containing static files (e.g. index.html, app.js); served from disk, not embedded.
func NewServer(addr string, repo database.Repository, staticDir string) *Server {
	s := &Server{addr: addr, staticDir: staticDir, repo: repo}
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	return s
}

// Run starts the HTTP server and blocks until the context is cancelled or the server errors.
func (s *Server) Run(ctx context.Context) error {
	go func() {
		<-ctx.Done()
		_ = s.httpServer.Shutdown(context.Background())
	}()
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}
	return nil
}
