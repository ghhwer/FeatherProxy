package schema

import (
	"time"

	"github.com/google/uuid"
)

// TargetServer is the domain schema for a target server.
type TargetServer struct {
	TargetServerUUID uuid.UUID `json:"target_server_uuid"`
	Name             string    `json:"name"`
	Protocol         string    `json:"protocol"`
	Host             string    `json:"host"`
	Port             int       `json:"port"`
	BasePath         string    `json:"base_path"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
