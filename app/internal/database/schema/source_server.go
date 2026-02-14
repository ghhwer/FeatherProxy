package schema

import (
	"time"

	"github.com/google/uuid"
)

// SourceServer is the domain schema for a source server.
type SourceServer struct {
	SourceServerUUID uuid.UUID `json:"source_server_uuid"`
	Name             string    `json:"name"`
	Protocol         string    `json:"protocol"`
	Host             string    `json:"host"`
	Port             int       `json:"port"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
