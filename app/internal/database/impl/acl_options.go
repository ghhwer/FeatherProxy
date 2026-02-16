package impl

import (
	"time"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) GetACLOptions(sourceServerUUID uuid.UUID) (schema.ACLOptions, error) {
	return getCached(r, keyACLOptions(sourceServerUUID), func() (schema.ACLOptions, error) {
		var obj objects.ACLOptions
		if err := r.db.Where("source_server_uuid = ?", sourceServerUUID).First(&obj).Error; err != nil {
			return schema.ACLOptions{}, err
		}
		return objects.ACLOptionsToSchema(&obj), nil
	})
}

func (r *repository) SetACLOptions(opts schema.ACLOptions) error {
	now := time.Now()
	var obj objects.ACLOptions
	err := r.db.Where("source_server_uuid = ?", opts.SourceServerUUID).First(&obj).Error
	if err != nil {
		// Create new
		obj = objects.SchemaToACLOptions(opts)
		if obj.CreatedAt.IsZero() {
			obj.CreatedAt = now
		}
		if obj.UpdatedAt.IsZero() {
			obj.UpdatedAt = now
		}
		return r.invalidate(r.db.Create(&obj).Error, []string{keyACLOptions(opts.SourceServerUUID)}, nil)
	}
	// Update existing
	updated := objects.SchemaToACLOptions(opts)
	obj.Mode = updated.Mode
	obj.ClientIPHeader = updated.ClientIPHeader
	obj.AllowListJSON = updated.AllowListJSON
	obj.DenyListJSON = updated.DenyListJSON
	obj.UpdatedAt = now
	return r.invalidate(r.db.Save(&obj).Error, []string{keyACLOptions(opts.SourceServerUUID)}, nil)
}
