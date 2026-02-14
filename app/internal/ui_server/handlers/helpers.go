package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// parseUUIDParam parses idStr as a UUID. On failure it writes a 400 response and returns (uuid.Nil, false).
func parseUUIDParam(w http.ResponseWriter, idStr, errMsg string) (uuid.UUID, bool) {
	id, err := uuid.Parse(idStr)
	if err != nil {
		respondJSONError(w, http.StatusBadRequest, errMsg)
		return uuid.Nil, false
	}
	return id, true
}

// handleRepoGetError handles repo Get errors: 404 for not found, 500 otherwise. Returns false if a response was written.
func handleRepoGetError(w http.ResponseWriter, err error) bool {
	if err == nil {
		return true
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.NotFound(w, nil)
		return false
	}
	respondJSONError(w, http.StatusInternalServerError, err.Error())
	return false
}

// decodeJSON decodes r.Body into v. On failure it writes a 400 response and returns false.
func decodeJSON(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		respondJSONError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return false
	}
	return true
}
