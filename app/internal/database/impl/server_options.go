package impl

import (
	"time"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) GetServerOptions(sourceServerUUID uuid.UUID) (schema.ServerOptions, error) {
	return getCached(r, keyServerOptions(sourceServerUUID), func() (schema.ServerOptions, error) {
		var obj objects.ServerOptions
		if err := r.db.Where("source_server_uuid = ?", sourceServerUUID).First(&obj).Error; err != nil {
			return schema.ServerOptions{}, err
		}
		return objects.ServerOptionsToSchema(&obj), nil
	})
}

func (r *repository) SetServerOptions(opts schema.ServerOptions) error {
	now := time.Now()
	var obj objects.ServerOptions
	err := r.db.Where("source_server_uuid = ?", opts.SourceServerUUID).First(&obj).Error
	if err != nil {
		// Create new
		obj = objects.SchemaToServerOptions(opts)
		if obj.CreatedAt.IsZero() {
			obj.CreatedAt = now
		}
		if obj.UpdatedAt.IsZero() {
			obj.UpdatedAt = now
		}
		return r.invalidate(r.db.Create(&obj).Error, []string{keyServerOptions(opts.SourceServerUUID)}, nil)
	}
	// Update existing
	obj.TLSCertPath = opts.TLSCertPath
	obj.TLSKeyPath = opts.TLSKeyPath
	obj.UpdatedAt = now
	return r.invalidate(r.db.Save(&obj).Error, []string{keyServerOptions(opts.SourceServerUUID)}, nil)
}
