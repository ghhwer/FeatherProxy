package handlers

import (
	"log"
	"net/http"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func GetRouteSourceAuth(repo database.Repository, w http.ResponseWriter, _ *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: GET /api/routes/%s/source-auth", routeIDStr)
	routeID, ok := parseUUIDParam(w, routeIDStr, "invalid route UUID")
	if !ok {
		return
	}
	list, err := repo.ListSourceAuthsForRoute(routeID)
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

func PutRouteSourceAuth(repo database.Repository, w http.ResponseWriter, r *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: PUT /api/routes/%s/source-auth", routeIDStr)
	routeID, ok := parseUUIDParam(w, routeIDStr, "invalid route UUID")
	if !ok {
		return
	}
	if _, err := repo.GetRoute(routeID); !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		AuthenticationUUIDs []string `json:"authentication_uuids"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.SetSourceAuthsForRoute(routeID, authUUIDs); err != nil {
		log.Printf("api/route_auth: put source-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	list, _ := repo.ListSourceAuthsForRoute(routeID)
	log.Printf("api/route_auth: put source-auth ok route=%s count=%d", routeID, len(list))
	respondJSON(w, http.StatusOK, list)
}

func GetRouteTargetAuth(repo database.Repository, w http.ResponseWriter, _ *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: GET /api/routes/%s/target-auth", routeIDStr)
	routeID, ok := parseUUIDParam(w, routeIDStr, "invalid route UUID")
	if !ok {
		return
	}
	authUUID, ok, err := repo.GetTargetAuthForRoute(routeID)
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

func PutRouteTargetAuth(repo database.Repository, w http.ResponseWriter, r *http.Request, routeIDStr string) {
	log.Printf("api/route_auth: PUT /api/routes/%s/target-auth", routeIDStr)
	routeID, ok := parseUUIDParam(w, routeIDStr, "invalid route UUID")
	if !ok {
		return
	}
	if _, err := repo.GetRoute(routeID); !handleRepoGetError(w, err) {
		return
	}
	var body struct {
		AuthenticationUUID string `json:"authentication_uuid"`
	}
	if !decodeJSON(w, r, &body) {
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
	if err := repo.SetTargetAuthForRoute(routeID, authUUID); err != nil {
		log.Printf("api/route_auth: put target-auth error: %v", err)
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	authUUIDOut, ok, _ := repo.GetTargetAuthForRoute(routeID)
	out := struct {
		AuthenticationUUID string `json:"authentication_uuid,omitempty"`
	}{}
	if ok && authUUIDOut != uuid.Nil {
		out.AuthenticationUUID = authUUIDOut.String()
	}
	log.Printf("api/route_auth: put target-auth ok route=%s auth_set=%v", routeID, ok)
	respondJSON(w, http.StatusOK, out)
}
