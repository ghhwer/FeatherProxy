package objects

import (
	"time"

	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

// ProxyStat is the database object (ORM entity) for the proxy_stats table.
type ProxyStat struct {
	ID                 uuid.UUID  `gorm:"primaryKey;type:uuid"`
	Timestamp          time.Time  `gorm:"not null;index"`
	SourceServerUUID   uuid.UUID  `gorm:"not null;index"`
	RouteUUID          uuid.UUID  `gorm:"not null;index"`
	TargetServerUUID   uuid.UUID  `gorm:"not null;index"`
	Method             string     `gorm:"not null"`
	Path               string     `gorm:"not null"`
	StatusCode         *int
	DurationMs         *int64
	ClientIP           string     `gorm:"index"`
}

// TableName overrides the default table name.
func (ProxyStat) TableName() string {
	return "proxy_stats"
}

// ProxyStatToSchema maps the database object to the domain schema.
func ProxyStatToSchema(p *ProxyStat) schema.ProxyStat {
	return schema.ProxyStat{
		ID:               p.ID,
		Timestamp:        p.Timestamp,
		SourceServerUUID: p.SourceServerUUID,
		RouteUUID:        p.RouteUUID,
		TargetServerUUID: p.TargetServerUUID,
		Method:           p.Method,
		Path:             p.Path,
		StatusCode:       p.StatusCode,
		DurationMs:       p.DurationMs,
		ClientIP:         p.ClientIP,
	}
}

// SchemaToProxyStat maps the domain schema to the database object.
func SchemaToProxyStat(p schema.ProxyStat) ProxyStat {
	return ProxyStat{
		ID:               p.ID,
		Timestamp:        p.Timestamp,
		SourceServerUUID: p.SourceServerUUID,
		RouteUUID:        p.RouteUUID,
		TargetServerUUID: p.TargetServerUUID,
		Method:           p.Method,
		Path:             p.Path,
		StatusCode:       p.StatusCode,
		DurationMs:       p.DurationMs,
		ClientIP:         p.ClientIP,
	}
}
