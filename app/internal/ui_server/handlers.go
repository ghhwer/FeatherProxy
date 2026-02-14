package server

import (
	"encoding/json"
	"errors"
	"log"
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

// --- Authentications ---
func (s *Server) listAuthentications(w http.ResponseWriter, _ *http.Request) {
	log.Printf("api/auth: GET /api/authentications")
	list, err := s.repo.ListAuthentications()
	if err != nil {
		log.Printf("api/auth: list authentications error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.Authentication{}
	}
	log.Printf("api/auth: list ok count=%d", len(list))
	respondJSON(w, http.StatusOK, list)
}

func (s *Server) createAuthentication(w http.ResponseWriter, r *http.Request) {
	log.Printf("api/auth: POST /api/authentications")
	var body struct {
		Name      string `json:"name"`
		TokenType string `json:"token_type"`
		Token     string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("api/auth: create decode error: %v", err)
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if body.TokenType == "" {
		body.TokenType = "bearer"
	}
	if body.Token == "" {
		respondJSONError(w, http.StatusBadRequest, "token is required")
		return
	}
	a := schema.Authentication{
		AuthenticationUUID: uuid.New(),
		Name:                body.Name,
		TokenType:           body.TokenType,
		Token:               body.Token,
	}
	if err := s.repo.CreateAuthentication(a); err != nil {
		if errors.Is(err, database.ErrEncryptionKeyMissing) {
			log.Printf("api/auth: create AUTH_ENCRYPTION_KEY missing")
			respondJSONError(w, http.StatusServiceUnavailable, "authentication encryption not configured: set AUTH_ENCRYPTION_KEY")
			return
		}
		log.Printf("api/auth: create error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out, _ := s.repo.GetAuthentication(a.AuthenticationUUID)
	log.Printf("api/auth: create ok id=%s", a.AuthenticationUUID)
	respondJSON(w, http.StatusCreated, out)
}

func (s *Server) getAuthentication(w http.ResponseWriter, _ *http.Request, idStr string) {
	log.Printf("api/auth: GET /api/authentications/%s", idStr)
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid authentication UUID")
		return
	}
	a, err := s.repo.GetAuthentication(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		log.Printf("api/auth: get error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func (s *Server) updateAuthentication(w http.ResponseWriter, r *http.Request, idStr string) {
	log.Printf("api/auth: PUT /api/authentications/%s", idStr)
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid authentication UUID")
		return
	}
	existing, err := s.repo.GetAuthentication(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		Name      string `json:"name"`
		TokenType string `json:"token_type"`
		Token     string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("api/auth: update decode error: %v", err)
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	existing.Name = body.Name
	if body.TokenType != "" {
		existing.TokenType = body.TokenType
	}
	if body.Token != "" {
		existing.Token = body.Token
	}
	if err := s.repo.UpdateAuthentication(existing); err != nil {
		if errors.Is(err, database.ErrEncryptionKeyMissing) {
			log.Printf("api/auth: update AUTH_ENCRYPTION_KEY missing")
			respondJSONError(w, http.StatusServiceUnavailable, "authentication encryption not configured: set AUTH_ENCRYPTION_KEY")
			return
		}
		log.Printf("api/auth: update error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := s.repo.GetAuthentication(id)
	log.Printf("api/auth: update ok id=%s", id)
	respondJSON(w, http.StatusOK, updated)
}

func (s *Server) deleteAuthentication(w http.ResponseWriter, _ *http.Request, idStr string) {
	log.Printf("api/auth: DELETE /api/authentications/%s", idStr)
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid authentication UUID")
		return
	}
	if err := s.repo.DeleteAuthentication(id); err != nil {
		log.Printf("api/auth: delete error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("api/auth: delete ok id=%s", id)
	w.WriteHeader(http.StatusNoContent)
}

// --- Route source/target auth ---
func (s *Server) getRouteSourceAuth(w http.ResponseWriter, _ *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: GET /api/routes/%s/source-auth", routeIDStr)
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	list, err := s.repo.ListSourceAuthsForRoute(routeID)
	if err != nil {
		log.Printf("api/route_auth: get source-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.RouteSourceAuth{}
	}
	log.Printf("api/route_auth: get source-auth ok route=%s count=%d", routeID, len(list))
	respondJSON(w, http.StatusOK, list)
}

func (s *Server) putRouteSourceAuth(w http.ResponseWriter, r *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: PUT /api/routes/%s/source-auth", routeIDStr)
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	if _, err := s.repo.GetRoute(routeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		AuthenticationUUIDs []string `json:"authentication_uuids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("api/route_auth: put source-auth decode error: %v", err)
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	authUUIDs := make([]uuid.UUID, 0, len(body.AuthenticationUUIDs))
	for _, idStr := range body.AuthenticationUUIDs {
		if idStr == "" {
			continue
		}
		u, err := uuid.Parse(idStr)
		if err != nil {
			respondJSONError(w, http.StatusBadRequest, "invalid authentication_uuid in list")
			return
		}
		authUUIDs = append(authUUIDs, u)
	}
	if err := s.repo.SetSourceAuthsForRoute(routeID, authUUIDs); err != nil {
		log.Printf("api/route_auth: put source-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	list, _ := s.repo.ListSourceAuthsForRoute(routeID)
	log.Printf("api/route_auth: put source-auth ok route=%s count=%d", routeID, len(list))
	respondJSON(w, http.StatusOK, list)
}

func (s *Server) getRouteTargetAuth(w http.ResponseWriter, _ *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: GET /api/routes/%s/target-auth", routeIDStr)
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	authUUID, ok, err := s.repo.GetTargetAuthForRoute(routeID)
	if err != nil {
		log.Printf("api/route_auth: get target-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	type resp struct {
		AuthenticationUUID string `json:"authentication_uuid,omitempty"`
	}
	out := resp{}
	if ok && authUUID != uuid.Nil {
		out.AuthenticationUUID = authUUID.String()
	}
	log.Printf("api/route_auth: get target-auth ok route=%s auth_set=%v", routeID, ok)
	respondJSON(w, http.StatusOK, out)
}

func (s *Server) putRouteTargetAuth(w http.ResponseWriter, r *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: PUT /api/routes/%s/target-auth", routeIDStr)
	routeID, err := uuid.Parse(routeIDStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid route UUID")
		return
	}
	if _, err := s.repo.GetRoute(routeID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.NotFound(w, nil)
			return
		}
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	var body struct {
		AuthenticationUUID string `json:"authentication_uuid"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("api/route_auth: put target-auth decode error: %v", err)
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	var authUUID *uuid.UUID
	if body.AuthenticationUUID != "" {
		u, err := uuid.Parse(body.AuthenticationUUID)
		if err != nil {
			respondJSONError(w, http.StatusBadRequest, "invalid authentication_uuid")
			return
		}
		authUUID = &u
	}
	if err := s.repo.SetTargetAuthForRoute(routeID, authUUID); err != nil {
		log.Printf("api/route_auth: put target-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	authUUIDOut, ok, _ := s.repo.GetTargetAuthForRoute(routeID)
	out := struct {
		AuthenticationUUID string `json:"authentication_uuid,omitempty"`
	}{}
	if ok && authUUIDOut != uuid.Nil {
		out.AuthenticationUUID = authUUIDOut.String()
	}
	log.Printf("api/route_auth: put target-auth ok route=%s auth_set=%v", routeID, ok)
	respondJSON(w, http.StatusOK, out)
}

func respondJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondJSONError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}
