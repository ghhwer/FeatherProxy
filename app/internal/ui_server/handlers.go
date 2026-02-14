package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"FeatherProxy/app/internal/database"
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
		SourceServerUUID string `json:"source_server_uuid"`
		TargetServerUUID string `json:"target_server_uuid"`
		Method           string `json:"method"`
		SourcePath       string `json:"source_path"`
		TargetPath       string `json:"target_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	sourceID, err := uuid.Parse(body.SourceServerUUID)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid source_server_uuid")
		return
	}
	targetID, err := uuid.Parse(body.TargetServerUUID)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid target_server_uuid")
		return
	}
	if body.Method == "" || body.SourcePath == "" || body.TargetPath == "" {
		respondJSONError(w, http.StatusBadRequest, "method, source_path, target_path required")
		return
	}
	route := schema.Route{
		RouteUUID:        uuid.New(),
		SourceServerUUID: sourceID,
		TargetServerUUID: targetID,
		Method:           body.Method,
		SourcePath:       body.SourcePath,
		TargetPath:       body.TargetPath,
	}
	if err := s.repo.CreateRoute(route); err != nil {
		if errors.Is(err, database.ErrProtocolMismatch) {
			respondJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondJSONError(w, http.StatusBadRequest, "source or target server not found")
			return
		}
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
		SourceServerUUID string `json:"source_server_uuid"`
		TargetServerUUID string `json:"target_server_uuid"`
		Method           string `json:"method"`
		SourcePath       string `json:"source_path"`
		TargetPath       string `json:"target_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	sourceID, err := uuid.Parse(body.SourceServerUUID)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid source_server_uuid")
		return
	}
	targetID, err := uuid.Parse(body.TargetServerUUID)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid target_server_uuid")
		return
	}
	if body.Method == "" || body.SourcePath == "" || body.TargetPath == "" {
		respondJSONError(w, http.StatusBadRequest, "method, source_path, target_path required")
		return
	}
	route := schema.Route{
		RouteUUID:        id,
		SourceServerUUID: sourceID,
		TargetServerUUID: targetID,
		Method:           body.Method,
		SourcePath:       body.SourcePath,
		TargetPath:       body.TargetPath,
		CreatedAt:        existing.CreatedAt,
		UpdatedAt:        existing.UpdatedAt,
	}
	if err := s.repo.UpdateRoute(route); err != nil {
		if errors.Is(err, database.ErrProtocolMismatch) {
			respondJSONError(w, http.StatusBadRequest, err.Error())
			return
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			respondJSONError(w, http.StatusBadRequest, "source or target server not found")
			return
		}
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

func (s *Server) listSourceServers(w http.ResponseWriter, _ *http.Request) {
	list, err := s.repo.ListSourceServers()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.SourceServer{}
	}
	respondJSON(w, http.StatusOK, list)
}

func (s *Server) createSourceServer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Protocol == "" || body.Host == "" || body.Port <= 0 {
		respondJSONError(w, http.StatusBadRequest, "protocol, host, and port (positive) required")
		return
	}
	svc := schema.SourceServer{
		SourceServerUUID: uuid.New(),
		Name:             body.Name,
		Protocol:         body.Protocol,
		Host:             body.Host,
		Port:             body.Port,
	}
	if err := s.repo.CreateSourceServer(svc); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, svc)
}

func (s *Server) getSourceServer(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid source server UUID")
		return
	}
	svc, err := s.repo.GetSourceServer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

func (s *Server) updateSourceServer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid source server UUID")
		return
	}
	existing, err := s.repo.GetSourceServer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Protocol == "" || body.Host == "" || body.Port <= 0 {
		respondJSONError(w, http.StatusBadRequest, "protocol, host, and port (positive) required")
		return
	}
	existing.Name = body.Name
	existing.Protocol = body.Protocol
	existing.Host = body.Host
	existing.Port = body.Port
	if err := s.repo.UpdateSourceServer(existing); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

func (s *Server) deleteSourceServer(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid source server UUID")
		return
	}
	if err := s.repo.DeleteSourceServer(id); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) listTargetServers(w http.ResponseWriter, _ *http.Request) {
	list, err := s.repo.ListTargetServers()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.TargetServer{}
	}
	respondJSON(w, http.StatusOK, list)
}

func (s *Server) createTargetServer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		BasePath string `json:"base_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Protocol == "" || body.Host == "" || body.Port <= 0 {
		respondJSONError(w, http.StatusBadRequest, "protocol, host, and port (positive) required")
		return
	}
	svc := schema.TargetServer{
		TargetServerUUID: uuid.New(),
		Name:             body.Name,
		Protocol:         body.Protocol,
		Host:             body.Host,
		Port:             body.Port,
		BasePath:         body.BasePath,
	}
	if err := s.repo.CreateTargetServer(svc); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, svc)
}

func (s *Server) getTargetServer(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid target server UUID")
		return
	}
	svc, err := s.repo.GetTargetServer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

func (s *Server) updateTargetServer(w http.ResponseWriter, r *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid target server UUID")
		return
	}
	existing, err := s.repo.GetTargetServer(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		BasePath string `json:"base_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.Protocol == "" || body.Host == "" || body.Port <= 0 {
		respondJSONError(w, http.StatusBadRequest, "protocol, host, and port (positive) required")
		return
	}
	existing.Name = body.Name
	existing.Protocol = body.Protocol
	existing.Host = body.Host
	existing.Port = body.Port
	existing.BasePath = body.BasePath
	if err := s.repo.UpdateTargetServer(existing); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

func (s *Server) deleteTargetServer(w http.ResponseWriter, _ *http.Request, idStr string) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid target server UUID")
		return
	}
	if err := s.repo.DeleteTargetServer(id); err != nil {
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
