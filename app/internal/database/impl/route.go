package impl

import (
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/repo"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func (r *repository) CreateRoute(route schema.Route) error {
	source, err := r.GetSourceServer(route.SourceServerUUID)
	if err != nil {
		return err
	}
	target, err := r.GetTargetServer(route.TargetServerUUID)
	if err != nil {
		return err
	}
	if !protocolsCompatible(source.Protocol, target.Protocol) {
		return repo.ErrProtocolMismatch
	}
	dbRoute := objects.SchemaToRoute(route)
	return r.invalidate(r.db.Create(&dbRoute).Error, []string{keyListRoutes}, []string{keyPrefixRoute})
}

func (r *repository) GetRoute(routeUUID uuid.UUID) (schema.Route, error) {
	return getCached(r, keyRoute(routeUUID), func() (schema.Route, error) {
		var dbRoute objects.Route
		if err := r.db.Where("route_uuid = ?", routeUUID).First(&dbRoute).Error; err != nil {
			return schema.Route{}, err
		}
		return objects.RouteToSchema(&dbRoute), nil
	})
}

func (r *repository) UpdateRoute(route schema.Route) error {
	source, err := r.GetSourceServer(route.SourceServerUUID)
	if err != nil {
		return err
	}
	target, err := r.GetTargetServer(route.TargetServerUUID)
	if err != nil {
		return err
	}
	if !protocolsCompatible(source.Protocol, target.Protocol) {
		return repo.ErrProtocolMismatch
	}
	dbRoute := objects.SchemaToRoute(route)
	return r.invalidate(r.db.Save(&dbRoute).Error, []string{keyListRoutes}, []string{keyPrefixRoute})
}

func (r *repository) DeleteRoute(routeUUID uuid.UUID) error {
	err := r.db.Delete(&objects.Route{RouteUUID: routeUUID}).Error
	return r.invalidate(err,
		[]string{keyListRoutes, keyRouteSourceAuths(routeUUID), keyTargetAuthForRoute(routeUUID)},
		[]string{keyPrefixRoute})
}

func (r *repository) ListRoutes() ([]schema.Route, error) {
	return getCached(r, keyListRoutes, func() ([]schema.Route, error) {
		var list []objects.Route
		if err := r.db.Find(&list).Error; err != nil {
			return nil, err
		}
		out := make([]schema.Route, len(list))
		for i := range list {
			out[i] = objects.RouteToSchema(&list[i])
		}
		return out, nil
	})
}

func (r *repository) GetRouteFromSourcePath(sourcePath string) (schema.Route, error) {
	return getCached(r, keyRouteSourcePath(sourcePath), func() (schema.Route, error) {
		var dbRoute objects.Route
		if err := r.db.Where("source_path = ?", sourcePath).First(&dbRoute).Error; err != nil {
			return schema.Route{}, err
		}
		return objects.RouteToSchema(&dbRoute), nil
	})
}

func (r *repository) GetRouteFromTargetPath(targetPath string) (schema.Route, error) {
	return getCached(r, keyRouteTargetPath(targetPath), func() (schema.Route, error) {
		var dbRoute objects.Route
		if err := r.db.Where("target_path = ?", targetPath).First(&dbRoute).Error; err != nil {
			return schema.Route{}, err
		}
		return objects.RouteToSchema(&dbRoute), nil
	})
}

func (r *repository) FindRouteBySourceMethodPath(sourceServerUUID uuid.UUID, method, sourcePath string) (schema.Route, error) {
	return getCached(r, keyRouteMethodPath(sourceServerUUID, method, sourcePath), func() (schema.Route, error) {
		var dbRoute objects.Route
		if err := r.db.Where("source_server_uuid = ? AND method = ? AND source_path = ?", sourceServerUUID, method, sourcePath).First(&dbRoute).Error; err != nil {
			return schema.Route{}, err
		}
		return objects.RouteToSchema(&dbRoute), nil
	})
}
