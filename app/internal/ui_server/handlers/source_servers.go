package handlers

import (
	"net/http"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func ListSourceServers(repo database.Repository, w http.ResponseWriter, _ *http.Request) {
	list, err := repo.ListSourceServers()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if list == nil {
		list = []schema.SourceServer{}
	}
	respondJSON(w, http.StatusOK, list)
}

func CreateSourceServer(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.CreateSourceServer(svc); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusCreated, svc)
}

func GetSourceServer(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid source server UUID")
	if !ok {
		return
	}
	svc, err := repo.GetSourceServer(id)
	if !handleRepoGetError(w, err) {
		return
	}
	respondJSON(w, http.StatusOK, svc)
}

func UpdateSourceServer(repo database.Repository, w http.ResponseWriter, r *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid source server UUID")
	if !ok {
		return
	}
	existing, err := repo.GetSourceServer(id)
	if !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		Name     string `json:"name"`
		Protocol string `json:"protocol"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
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
	if err := repo.UpdateSourceServer(existing); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, existing)
}

func DeleteSourceServer(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	id, ok := parseUUIDParam(w, idStr, "invalid source server UUID")
	if !ok {
		return
	}
	if err := repo.DeleteSourceServer(id); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
