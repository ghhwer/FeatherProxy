package handlers

import (
	"net/http"
	"strconv"
	"time"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"
)

const (
	defaultStatsLimit  = 100
	defaultStatsOffset = 0
	defaultTPSWindow   = time.Hour
	defaultTPSBucket   = time.Minute
)

func ListStats(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	limit := defaultStatsLimit
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 1000 {
			limit = n
		}
	}
	offset := defaultStatsOffset
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			offset = n
		}
	}
	var since *time.Time
	if v := r.URL.Query().Get("since"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			respondJSONError(w, http.StatusBadRequest, "invalid since (use RFC3339)")
			return
		}
		since = &t
	}
	stats, total, err := repo.ListProxyStats(limit, offset, since)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if stats == nil {
		stats = []schema.ProxyStat{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
		"total": total,
	})
}

func GetStatsSummary(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	sum, err := repo.StatsSummary()
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, sum)
}

func GetStatsByRoute(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	since, limit := parseStatsSinceLimit(r)
	items, err := repo.StatsByRoute(since, limit)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []schema.RouteCount{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func GetStatsByCaller(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	since, limit := parseStatsSinceLimit(r)
	items, err := repo.StatsByCaller(since, limit)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []schema.CallerCount{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"items": items})
}

func GetStatsBySourceServer(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	since, _ := parseStatsSinceLimit(r)
	items, err := repo.StatsBySourceServer(since)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []schema.ServerCount{}
	}
	// Map to API shape with source_server_uuid key
	out := make([]schema.SourceServerCountItem, len(items))
	for i := range items {
		out[i] = schema.SourceServerCountItem{SourceServerUUID: items[i].ServerUUID, Count: items[i].Count}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"items": out})
}

func GetStatsByTargetServer(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	since, _ := parseStatsSinceLimit(r)
	items, err := repo.StatsByTargetServer(since)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if items == nil {
		items = []schema.ServerCount{}
	}
	out := make([]schema.TargetServerCountItem, len(items))
	for i := range items {
		out[i] = schema.TargetServerCountItem{TargetServerUUID: items[i].ServerUUID, Count: items[i].Count}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"items": out})
}

func GetStatsTPS(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	window := defaultTPSWindow
	if v := r.URL.Query().Get("window"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			window = d
		}
	}
	bucket := defaultTPSBucket
	if v := r.URL.Query().Get("bucket"); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			bucket = d
		}
	}
	since := time.Now().Add(-window)
	buckets, err := repo.StatsTPS(since, bucket)
	if err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if buckets == nil {
		buckets = []schema.BucketCount{}
	}
	respondJSON(w, http.StatusOK, map[string]interface{}{"buckets": buckets})
}

func ClearStats(repo database.Repository, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := repo.ClearAllProxyStats(); err != nil {
		respondJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	respondJSON(w, http.StatusOK, map[string]string{"ok": "true"})
}

func parseStatsSinceLimit(r *http.Request) (since *time.Time, limit int) {
	limit = 0
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if v := r.URL.Query().Get("since"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err == nil {
			since = &t
		}
	}
	return since, limit
}
