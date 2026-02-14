package database

import (
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Repository interface {
	CreateRoute(route schema.Route) error
	GetRoute(routeUUID uuid.UUID) (schema.Route, error)
	UpdateRoute(route schema.Route) error
	DeleteRoute(routeUUID uuid.UUID) error
	ListRoutes() ([]schema.Route, error)
	GetRouteFromSourcePath(sourcePath string) (schema.Route, error)
	GetRouteFromTargetPath(targetPath string) (schema.Route, error)
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) CreateRoute(route schema.Route) error {
	dbRoute := objects.SchemaToRoute(route)
	return r.db.Create(&dbRoute).Error
}

func (r *repository) GetRoute(routeUUID uuid.UUID) (schema.Route, error) {
	var dbRoute objects.Route
	if err := r.db.Where("route_uuid = ?", routeUUID).First(&dbRoute).Error; err != nil {
		return schema.Route{}, err
	}
	return objects.RouteToSchema(&dbRoute), nil
}

func (r *repository) UpdateRoute(route schema.Route) error {
	dbRoute := objects.SchemaToRoute(route)
	return r.db.Save(&dbRoute).Error
}

func (r *repository) DeleteRoute(routeUUID uuid.UUID) error {
	return r.db.Delete(&objects.Route{RouteUUID: routeUUID}).Error
}

func (r *repository) ListRoutes() ([]schema.Route, error) {
	var list []objects.Route
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]schema.Route, len(list))
	for i := range list {
		out[i] = objects.RouteToSchema(&list[i])
	}
	return out, nil
}

func (r *repository) GetRouteFromSourcePath(sourcePath string) (schema.Route, error) {
	var dbRoute objects.Route
	if err := r.db.Where("source_path = ?", sourcePath).First(&dbRoute).Error; err != nil {
		return schema.Route{}, err
	}
	return objects.RouteToSchema(&dbRoute), nil
}

func (r *repository) GetRouteFromTargetPath(targetPath string) (schema.Route, error) {
	var dbRoute objects.Route
	if err := r.db.Where("target_path = ?", targetPath).First(&dbRoute).Error; err != nil {
		return schema.Route{}, err
	}
	return objects.RouteToSchema(&dbRoute), nil
}