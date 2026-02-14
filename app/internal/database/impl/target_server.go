package impl

import (
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) CreateTargetServer(t schema.TargetServer) error {
	obj := objects.SchemaToTargetServer(t)
	return r.invalidate(r.db.Create(&obj).Error, []string{keyTargetServer(t.TargetServerUUID), keyListTargetServers}, nil)
}

func (r *repository) GetTargetServer(id uuid.UUID) (schema.TargetServer, error) {
	return getCached(r, keyTargetServer(id), func() (schema.TargetServer, error) {
		var obj objects.TargetServer
		if err := r.db.Where("target_server_uuid = ?", id).First(&obj).Error; err != nil {
			return schema.TargetServer{}, err
		}
		return objects.TargetServerToSchema(&obj), nil
	})
}

func (r *repository) UpdateTargetServer(t schema.TargetServer) error {
	obj := objects.SchemaToTargetServer(t)
	return r.invalidate(r.db.Save(&obj).Error, []string{keyTargetServer(t.TargetServerUUID), keyListTargetServers}, nil)
}

func (r *repository) DeleteTargetServer(id uuid.UUID) error {
	return r.invalidate(r.db.Delete(&objects.TargetServer{TargetServerUUID: id}).Error, []string{keyTargetServer(id), keyListTargetServers}, nil)
}

func (r *repository) ListTargetServers() ([]schema.TargetServer, error) {
	return getCached(r, keyListTargetServers, func() ([]schema.TargetServer, error) {
		var list []objects.TargetServer
		if err := r.db.Find(&list).Error; err != nil {
			return nil, err
		}
		out := make([]schema.TargetServer, len(list))
		for i := range list {
			out[i] = objects.TargetServerToSchema(&list[i])
		}
		return out, nil
	})
}
