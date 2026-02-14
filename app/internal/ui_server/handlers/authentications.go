package handlers

import (
	"errors"
	"log"
	"net/http"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func ListAuthentications(repo database.Repository, w http.ResponseWriter, _ *http.Request) {
	log.Printf("api/auth: GET /api/authentications")
	list, err := repo.ListAuthentications()
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

func CreateAuthentication(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	log.Printf("api/auth: POST /api/authentications")
	var body struct {
		Name      string `json:"name"`
		TokenType string `json:"token_type"`
		Token     string `json:"token"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.CreateAuthentication(a); err != nil {
		if errors.Is(err, database.ErrEncryptionKeyMissing) {
			log.Printf("api/auth: create AUTH_ENCRYPTION_KEY missing")
			respondJSONError(w, http.StatusServiceUnavailable, "authentication encryption not configured: set AUTH_ENCRYPTION_KEY")
			return
		}
		log.Printf("api/auth: create error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	out, _ := repo.GetAuthentication(a.AuthenticationUUID)
	log.Printf("api/auth: create ok id=%s", a.AuthenticationUUID)
	respondJSON(w, http.StatusCreated, out)
}

func GetAuthentication(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	log.Printf("api/auth: GET /api/authentications/%s", idStr)
	id, ok := parseUUIDParam(w, idStr, "invalid authentication UUID")
	if !ok {
		return
	}
	a, err := repo.GetAuthentication(id)
	if !handleRepoGetError(w, err) {
		return
	}
	respondJSON(w, http.StatusOK, a)
}

func UpdateAuthentication(repo database.Repository, w http.ResponseWriter, r *http.Request, idStr string) {
	log.Printf("api/auth: PUT /api/authentications/%s", idStr)
	id, ok := parseUUIDParam(w, idStr, "invalid authentication UUID")
	if !ok {
		return
	}
	existing, err := repo.GetAuthentication(id)
	if !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		Name      string `json:"name"`
		TokenType string `json:"token_type"`
		Token     string `json:"token"`
	}
	if !decodeJSON(w, r, &body) {
		return
	}
	existing.Name = body.Name
	if body.TokenType != "" {
		existing.TokenType = body.TokenType
	}
	if body.Token != "" {
		existing.Token = body.Token
	}
	if err := repo.UpdateAuthentication(existing); err != nil {
		if errors.Is(err, database.ErrEncryptionKeyMissing) {
			log.Printf("api/auth: update AUTH_ENCRYPTION_KEY missing")
			respondJSONError(w, http.StatusServiceUnavailable, "authentication encryption not configured: set AUTH_ENCRYPTION_KEY")
			return
		}
		log.Printf("api/auth: update error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	updated, _ := repo.GetAuthentication(id)
	log.Printf("api/auth: update ok id=%s", id)
	respondJSON(w, http.StatusOK, updated)
}

func DeleteAuthentication(repo database.Repository, w http.ResponseWriter, _ *http.Request, idStr string) {
	log.Printf("api/auth: DELETE /api/authentications/%s", idStr)
	id, ok := parseUUIDParam(w, idStr, "invalid authentication UUID")
	if !ok {
		return
	}
	if err := repo.DeleteAuthentication(id); err != nil {
		log.Printf("api/auth: delete error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("api/auth: delete ok id=%s", id)
	w.WriteHeader(http.StatusNoContent)
}
