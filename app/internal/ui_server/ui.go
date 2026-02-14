package server

import (
	"net/http"
)

// handleStatic serves files from s.staticDir. No in-memory storage; reads from disk.
// Serves index.html for / and any other file under static (e.g. /app.js, /styles.css, /api.js).
func (s *Server) handleStatic(w http.ResponseWriter, r *http.Request) {
	if s.staticDir == "" {
		http.NotFound(w, r)
		return
	}
	http.FileServer(http.Dir(s.staticDir)).ServeHTTP(w, r)
}
