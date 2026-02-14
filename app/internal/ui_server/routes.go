package server

import (
	"net/http"
	"strings"
)

// Routes returns the HTTP handler for the server (API + UI).
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/source-servers", s.handleSourceServersCollection)
	mux.HandleFunc("/api/source-servers/", s.handleSourceServerByID)
	mux.HandleFunc("/api/target-servers", s.handleTargetServersCollection)
	mux.HandleFunc("/api/target-servers/", s.handleTargetServerByID)
	mux.HandleFunc("/api/routes", s.handleRoutesCollection)
	mux.HandleFunc("/api/routes/", s.handleRouteByID)

	// UI (HTML, CSS, JS served from in-memory embedded files)
	mux.HandleFunc("/styles.css", s.handleStyles)
	mux.HandleFunc("/app.js", s.handleScript)
	mux.HandleFunc("/", s.handleUI)

	return mux
}

// handleRoutesCollection: GET /api/routes (list), POST /api/routes (create).
func (s *Server) handleRoutesCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/routes" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.listRoutes(w, r)
	case http.MethodPost:
		s.createRoute(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSourceServersCollection: GET /api/source-servers (list), POST /api/source-servers (create).
func (s *Server) handleSourceServersCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/source-servers" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.listSourceServers(w, r)
	case http.MethodPost:
		s.createSourceServer(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleSourceServerByID: GET/PUT/DELETE /api/source-servers/{uuid}.
func (s *Server) handleSourceServerByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/source-servers/")
	if path == "" || strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getSourceServer(w, r, path)
	case http.MethodPut:
		s.updateSourceServer(w, r, path)
	case http.MethodDelete:
		s.deleteSourceServer(w, r, path)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTargetServersCollection: GET /api/target-servers (list), POST /api/target-servers (create).
func (s *Server) handleTargetServersCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/target-servers" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.listTargetServers(w, r)
	case http.MethodPost:
		s.createTargetServer(w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleTargetServerByID: GET/PUT/DELETE /api/target-servers/{uuid}.
func (s *Server) handleTargetServerByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/target-servers/")
	if path == "" || strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		s.getTargetServer(w, r, path)
	case http.MethodPut:
		s.updateTargetServer(w, r, path)
	case http.MethodDelete:
		s.deleteTargetServer(w, r, path)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRouteByID: GET/PUT/DELETE /api/routes/{uuid}.
func (s *Server) handleRouteByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	if path == "" || strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}
	idStr := path
	switch r.Method {
	case http.MethodGet:
		s.getRoute(w, r, idStr)
	case http.MethodPut:
		s.updateRoute(w, r, idStr)
	case http.MethodDelete:
		s.deleteRoute(w, r, idStr)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
