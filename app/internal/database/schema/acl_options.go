package schema

import (
	"time"

	"github.com/google/uuid"
)

// ACLOptions is the domain schema for per-source-server ACL (allow/deny by client IP/CIDR).
// When Mode is not "off", ClientIPHeader (if set) names the request header used for client IP; empty = use RemoteAddr.
type ACLOptions struct {
	SourceServerUUID uuid.UUID `json:"source_server_uuid"`
	Mode             string    `json:"mode"` // "off", "allow_only", "deny_only"
	ClientIPHeader   string    `json:"client_ip_header"`
	AllowList        []string  `json:"allow_list"`
	DenyList         []string  `json:"deny_list"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
