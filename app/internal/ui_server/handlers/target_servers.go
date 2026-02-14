package handlers

import (
	"net/http"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func ListTargetServers(repo database.Repository, w http.ResponseWriter, _ *http.Request) {
	list, err := repo.ListTargetServers()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.TargetServer{}
	}
	respondJSON(w, http.StatusOK, list)
}

func CreateTargetServer(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		BasePath string `json:"base_path"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.CreateTargetServer(svc); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, svc)
}

func GetTargetServer(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid target server UUID")
	if !ok {
		return
	}
	svc, err := repo.GetTargetServer(id)
	if !handleRepoGetError(w, err) {
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

func UpdateTargetServer(repo database.Repository, w http.ResponseWriter, r *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid target server UUID")
	if !ok {
		return
	}
	existing, err := repo.GetTargetServer(id)
	if !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		BasePath string `json:"base_path"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.UpdateTargetServer(existing); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

func DeleteTargetServer(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid target server UUID")
	if !ok {
		return
	}
	if err := repo.DeleteTargetServer(id); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
