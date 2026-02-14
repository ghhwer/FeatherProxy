package impl

import (
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) CreateSourceServer(s schema.SourceServer) error {
	obj := objects.SchemaToSourceServer(s)
	return r.invalidate(r.db.Create(&obj).Error, []string{keySourceServer(s.SourceServerUUID), keyListSourceServers}, nil)
}

func (r *repository) GetSourceServer(id uuid.UUID) (schema.SourceServer, error) {
	return getCached(r, keySourceServer(id), func() (schema.SourceServer, error) {
		var obj objects.SourceServer
		if err := r.db.Where("source_server_uuid = ?", id).First(&obj).Error; err != nil {
			return schema.SourceServer{}, err
		}
		return objects.SourceServerToSchema(&obj), nil
	})
}

func (r *repository) UpdateSourceServer(s schema.SourceServer) error {
	obj := objects.SchemaToSourceServer(s)
	return r.invalidate(r.db.Save(&obj).Error, []string{keySourceServer(s.SourceServerUUID), keyListSourceServers}, nil)
}

func (r *repository) DeleteSourceServer(id uuid.UUID) error {
	_ = r.db.Delete(&objects.ServerOptions{SourceServerUUID: id})
	return r.invalidate(r.db.Delete(&objects.SourceServer{SourceServerUUID: id}).Error, []string{keySourceServer(id), keyListSourceServers, keyServerOptions(id)}, nil)
}

func (r *repository) ListSourceServers() ([]schema.SourceServer, error) {
	return getCached(r, keyListSourceServers, func() ([]schema.SourceServer, error) {
		var list []objects.SourceServer
		if err := r.db.Find(&list).Error; err != nil {
			return nil, err
		}
		out := make([]schema.SourceServer, len(list))
		for i := range list {
			out[i] = objects.SourceServerToSchema(&list[i])
		}
		return out, nil
	})
}
