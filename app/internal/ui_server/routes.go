package server

import (
	"net/http"
	"strings"

	"FeatherProxy/app/internal/ui_server/handlers"
)

// Routes returns the HTTP handler for the server (API + UI).
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()

	// API
	mux.HandleFunc("/api/source-servers", s.handleSourceServersCollection)
	mux.HandleFunc("/api/source-servers/", s.handleSourceServerByID)
	mux.HandleFunc("/api/target-servers", s.handleTargetServersCollection)
	mux.HandleFunc("/api/target-servers/", s.handleTargetServerByID)
	mux.HandleFunc("/api/authentications", s.handleAuthenticationsCollection)
	mux.HandleFunc("/api/authentications/", s.handleAuthenticationByID)
	mux.HandleFunc("/api/routes", s.handleRoutesCollection)
	mux.HandleFunc("/api/routes/", s.handleRouteOrRouteAuth)

	// UI: serve anything under static from disk (no embed)
	mux.HandleFunc("/", s.handleStatic)

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
		handlers.ListRoutes(s.repo, w, r)
	case http.MethodPost:
		handlers.CreateRoute(s.repo, w, r)
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
		handlers.ListSourceServers(s.repo, w, r)
	case http.MethodPost:
		handlers.CreateSourceServer(s.repo, w, r)
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
		handlers.GetSourceServer(s.repo, w, r, path)
	case http.MethodPut:
		handlers.UpdateSourceServer(s.repo, w, r, path)
	case http.MethodDelete:
		handlers.DeleteSourceServer(s.repo, w, r, path)
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
		handlers.ListTargetServers(s.repo, w, r)
	case http.MethodPost:
		handlers.CreateTargetServer(s.repo, w, r)
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
		handlers.GetTargetServer(s.repo, w, r, path)
	case http.MethodPut:
		handlers.UpdateTargetServer(s.repo, w, r, path)
	case http.MethodDelete:
		handlers.DeleteTargetServer(s.repo, w, r, path)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleRouteOrRouteAuth: GET/PUT/DELETE /api/routes/{uuid} or .../source-auth or .../target-auth.
func (s *Server) handleRouteOrRouteAuth(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/routes/")
	if path == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.SplitN(path, "/", 2)
	routeIDStr := parts[0]
	subPath := ""
	if len(parts) == 2 {
		subPath = parts[1]
	}
	if subPath == "source-auth" {
		switch r.Method {
		case http.MethodGet:
			handlers.GetRouteSourceAuth(s.repo, w, r, routeIDStr)
		case http.MethodPut:
			handlers.PutRouteSourceAuth(s.repo, w, r, routeIDStr)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if subPath == "target-auth" {
		switch r.Method {
		case http.MethodGet:
			handlers.GetRouteTargetAuth(s.repo, w, r, routeIDStr)
		case http.MethodPut:
			handlers.PutRouteTargetAuth(s.repo, w, r, routeIDStr)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
		return
	}
	if subPath != "" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handlers.GetRoute(s.repo, w, r, routeIDStr)
	case http.MethodPut:
		handlers.UpdateRoute(s.repo, w, r, routeIDStr)
	case http.MethodDelete:
		handlers.DeleteRoute(s.repo, w, r, routeIDStr)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAuthenticationsCollection: GET /api/authentications (list), POST /api/authentications (create).
func (s *Server) handleAuthenticationsCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/authentications" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handlers.ListAuthentications(s.repo, w, r)
	case http.MethodPost:
		handlers.CreateAuthentication(s.repo, w, r)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// handleAuthenticationByID: GET/PUT/DELETE /api/authentications/{uuid}.
func (s *Server) handleAuthenticationByID(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/authentications/")
	if path == "" || strings.Contains(path, "/") {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		handlers.GetAuthentication(s.repo, w, r, path)
	case http.MethodPut:
		handlers.UpdateAuthentication(s.repo, w, r, path)
	case http.MethodDelete:
		handlers.DeleteAuthentication(s.repo, w, r, path)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}
