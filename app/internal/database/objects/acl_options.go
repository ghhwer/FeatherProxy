package objects

import (
	"encoding/json"
	"time"

	"FeatherProxy/app/internal/database/schema"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ACLOptions is the database object (ORM entity) for the acl_options table.
// AllowList and DenyList are stored as JSON strings.
type ACLOptions struct {
	SourceServerUUID uuid.UUID      `gorm:"primaryKey"`
	Mode             string         `gorm:"column:mode;default:off"`
	ClientIPHeader   string         `gorm:"column:client_ip_header"`
	AllowListJSON    string         `gorm:"column:allow_list"`
	DenyListJSON     string         `gorm:"column:deny_list"`
	CreatedAt        time.Time      `gorm:"not null"`
	UpdatedAt        time.Time      `gorm:"not null"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

// TableName overrides the default table name.
func (ACLOptions) TableName() string {
	return "acl_options"
}

func parseStringList(jsonStr string) []string {
	if jsonStr == "" {
		return nil
	}
	var out []string
	_ = json.Unmarshal([]byte(jsonStr), &out)
	if out == nil {
		return []string{}
	}
	return out
}

func marshalStringList(list []string) string {
	if len(list) == 0 {
		return "[]"
	}
	b, _ := json.Marshal(list)
	return string(b)
}

// ACLOptionsToSchema maps the database object to the domain schema.
func ACLOptionsToSchema(o *ACLOptions) schema.ACLOptions {
	return schema.ACLOptions{
		SourceServerUUID: o.SourceServerUUID,
		Mode:             o.Mode,
		ClientIPHeader:   o.ClientIPHeader,
		AllowList:        parseStringList(o.AllowListJSON),
		DenyList:         parseStringList(o.DenyListJSON),
		CreatedAt:        o.CreatedAt,
		UpdatedAt:        o.UpdatedAt,
	}
}

// SchemaToACLOptions maps the domain schema to the database object.
func SchemaToACLOptions(s schema.ACLOptions) ACLOptions {
	return ACLOptions{
		SourceServerUUID: s.SourceServerUUID,
		Mode:             s.Mode,
		ClientIPHeader:   s.ClientIPHeader,
		AllowListJSON:    marshalStringList(s.AllowList),
		DenyListJSON:     marshalStringList(s.DenyList),
		CreatedAt:        s.CreatedAt,
		UpdatedAt:        s.UpdatedAt,
	}
}
