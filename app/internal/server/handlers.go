package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (s *Server) listRoutes(w http.ResponseWriter, _ *http.Request) {
	routes, err := s.repo.ListRoutes()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if routes == nil {
		routes = []schema.Route{}
	}
	respondJSON(w, http.StatusOK, routes)
}

func (s *Server) createRoute(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Protocol   string `json:"protocol"`
		Method     string `json:"method"`
		SourcePath string `json:"source_path"`
		TargetPath string `json:"target_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Protocol == "" || body.Method == "" || body.SourcePath == "" || body.TargetPath == "" {
		respondJSONError(w, http.StatusBadRequest, "protocol, method, source_path, target_path required")
		return
	}
	route := schema.Route{
		RouteUUID:  uuid.New(),
		Protocol:   body.Protocol,
		Method:     body.Method,
		SourcePath: body.SourcePath,
		TargetPath: body.TargetPath,
	}
	if err := s.repo.CreateRoute(route); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, route)
}

func (s *Server) getRoute(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	route, err := s.repo.GetRoute(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, route)
}

func (s *Server) updateRoute(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	existing, err := s.repo.GetRoute(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Protocol   string `json:"protocol"`
		Method     string `json:"method"`
		SourcePath string `json:"source_path"`
		TargetPath string `json:"target_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	route := schema.Route{
		RouteUUID:  id,
		Protocol:   body.Protocol,
		Method:     body.Method,
		SourcePath: body.SourcePath,
		TargetPath: body.TargetPath,
		CreatedAt:  existing.CreatedAt,
		UpdatedAt:  existing.UpdatedAt,
	}
	if err := s.repo.UpdateRoute(route); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := s.repo.GetRoute(id)
	respondJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteRoute(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	if err := s.repo.DeleteRoute(id); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondJSONError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
