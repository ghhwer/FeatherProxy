package handlers

import (
	"errors"
	"net/http"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func ListRoutes(repo database.Repository, w http.ResponseWriter, _ *http.Request) {
	routes, err := repo.ListRoutes()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if routes == nil {
		routes = []schema.Route{}
	}
	respondJSON(w, http.StatusOK, routes)
}

func CreateRoute(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	var body struct {
		SourceServerUUID string `json:"source_server_uuid"`
		TargetServerUUID string `json:"target_server_uuid"`
		Method           string `json:"method"`
		SourcePath       string `json:"source_path"`
		TargetPath       string `json:"target_path"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	sourceID, ok := parseUUIDParam(w, body.SourceServerUUID, "invalid source_server_uuid")
	if !ok {
		return
	}
	targetID, ok := parseUUIDParam(w, body.TargetServerUUID, "invalid target_server_uuid")
	if !ok {
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
	if err := repo.CreateRoute(route); err != nil {
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

func GetRoute(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid route UUID")
	if !ok {
		return
	}
	route, err := repo.GetRoute(id)
	if !handleRepoGetError(w, err) {
		return
	}
	respondJSON(w, http.StatusOK, route)
}

func UpdateRoute(repo database.Repository, w http.ResponseWriter, r *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid route UUID")
	if !ok {
		return
	}
	existing, err := repo.GetRoute(id)
	if !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		SourceServerUUID string `json:"source_server_uuid"`
		TargetServerUUID string `json:"target_server_uuid"`
		Method           string `json:"method"`
		SourcePath       string `json:"source_path"`
		TargetPath       string `json:"target_path"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	sourceID, ok := parseUUIDParam(w, body.SourceServerUUID, "invalid source_server_uuid")
	if !ok {
		return
	}
	targetID, ok := parseUUIDParam(w, body.TargetServerUUID, "invalid target_server_uuid")
	if !ok {
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
	if err := repo.UpdateRoute(route); err != nil {
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
	updated, _ := repo.GetRoute(id)
	respondJSON(w, http.StatusOK, updated)
}

func DeleteRoute(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid route UUID")
	if !ok {
		return
	}
	if err := repo.DeleteRoute(id); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
