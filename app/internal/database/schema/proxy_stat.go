package schema

import (
	"time"

	"github.com/google/uuid"
)

// ProxyStat is the domain schema for one proxied request metric.
// Used by the stats service and API; do not depend on database objects.
type ProxyStat struct {
	ID                 uuid.UUID `json:"id"`
	Timestamp          time.Time `json:"timestamp"`
	SourceServerUUID   uuid.UUID `json:"source_server_uuid"`
	RouteUUID          uuid.UUID `json:"route_uuid"`
	TargetServerUUID   uuid.UUID `json:"target_server_uuid"`
	Method             string    `json:"method"`
	Path               string    `json:"path"`
	StatusCode         *int     `json:"status_code,omitempty"`
	DurationMs         *int64   `json:"duration_ms,omitempty"`
	ClientIP           string   `json:"client_ip,omitempty"`
}

// StatsSummary holds aggregated counts for the summary endpoint.
type StatsSummary struct {
	Total          int64 `json:"total"`
	Last24h        int64 `json:"last_24h"`
	Status2xx      int64 `json:"status_2xx"`
	Status4xx      int64 `json:"status_4xx"`
	Status5xx      int64 `json:"status_5xx"`
	TpsLastMinute  int64 `json:"tps_last_minute,omitempty"`
}

// RouteCount is one row from StatsByRoute aggregation.
type RouteCount struct {
	RouteUUID   uuid.UUID `json:"route_uuid"`
	Method      string    `json:"method"`
	SourcePath  string    `json:"source_path"`
	Count       int64     `json:"count"`
}

// CallerCount is one row from StatsByCaller aggregation.
type CallerCount struct {
	ClientIP string `json:"client_ip"`
	Count    int64  `json:"count"`
}

// ServerCount is one row from StatsBySourceServer or StatsByTargetServer (internal aggregation).
type ServerCount struct {
	ServerUUID uuid.UUID `json:"server_uuid"`
	Count      int64     `json:"count"`
}

// SourceServerCountItem is the API shape for by-source-server (source_server_uuid key).
type SourceServerCountItem struct {
	SourceServerUUID uuid.UUID `json:"source_server_uuid"`
	Count            int64     `json:"count"`
}

// TargetServerCountItem is the API shape for by-target-server (target_server_uuid key).
type TargetServerCountItem struct {
	TargetServerUUID uuid.UUID `json:"target_server_uuid"`
	Count            int64     `json:"count"`
}

// BucketCount is one time bucket for TPS.
type BucketCount struct {
	At    time.Time `json:"at"`
	Count int64     `json:"count"`
}
