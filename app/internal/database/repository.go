package database

import (
	"errors"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ErrProtocolMismatch is returned when a route links a source and target server with incompatible protocols.
var ErrProtocolMismatch = errors.New("source and target server must have the same protocol")

// protocolsCompatible returns true if source and target protocols can be linked (http and https are allowed together).
func protocolsCompatible(source, target string) bool {
	if source == target {
		return true
	}
	return (source == "http" && target == "https") || (source == "https" && target == "http")
}

type Repository interface {
	// Source servers
	CreateSourceServer(s schema.SourceServer) error
	GetSourceServer(uuid uuid.UUID) (schema.SourceServer, error)
	UpdateSourceServer(s schema.SourceServer) error
	DeleteSourceServer(uuid uuid.UUID) error
	ListSourceServers() ([]schema.SourceServer, error)
	// Target servers
	CreateTargetServer(t schema.TargetServer) error
	GetTargetServer(uuid uuid.UUID) (schema.TargetServer, error)
	UpdateTargetServer(t schema.TargetServer) error
	DeleteTargetServer(uuid uuid.UUID) error
	ListTargetServers() ([]schema.TargetServer, error)
	// Routes
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

func (r *repository) CreateSourceServer(s schema.SourceServer) error {
	obj := objects.SchemaToSourceServer(s)
	return r.db.Create(&obj).Error
}

func (r *repository) GetSourceServer(id uuid.UUID) (schema.SourceServer, error) {
	var obj objects.SourceServer
	if err := r.db.Where("source_server_uuid = ?", id).First(&obj).Error; err != nil {
		return schema.SourceServer{}, err
	}
	return objects.SourceServerToSchema(&obj), nil
}

func (r *repository) UpdateSourceServer(s schema.SourceServer) error {
	obj := objects.SchemaToSourceServer(s)
	return r.db.Save(&obj).Error
}

func (r *repository) DeleteSourceServer(id uuid.UUID) error {
	return r.db.Delete(&objects.SourceServer{SourceServerUUID: id}).Error
}

func (r *repository) ListSourceServers() ([]schema.SourceServer, error) {
	var list []objects.SourceServer
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]schema.SourceServer, len(list))
	for i := range list {
		out[i] = objects.SourceServerToSchema(&list[i])
	}
	return out, nil
}

func (r *repository) CreateTargetServer(t schema.TargetServer) error {
	obj := objects.SchemaToTargetServer(t)
	return r.db.Create(&obj).Error
}

func (r *repository) GetTargetServer(id uuid.UUID) (schema.TargetServer, error) {
	var obj objects.TargetServer
	if err := r.db.Where("target_server_uuid = ?", id).First(&obj).Error; err != nil {
		return schema.TargetServer{}, err
	}
	return objects.TargetServerToSchema(&obj), nil
}

func (r *repository) UpdateTargetServer(t schema.TargetServer) error {
	obj := objects.SchemaToTargetServer(t)
	return r.db.Save(&obj).Error
}

func (r *repository) DeleteTargetServer(id uuid.UUID) error {
	return r.db.Delete(&objects.TargetServer{TargetServerUUID: id}).Error
}

func (r *repository) ListTargetServers() ([]schema.TargetServer, error) {
	var list []objects.TargetServer
	if err := r.db.Find(&list).Error; err != nil {
		return nil, err
	}
	out := make([]schema.TargetServer, len(list))
	for i := range list {
		out[i] = objects.TargetServerToSchema(&list[i])
	}
	return out, nil
}

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
		return ErrProtocolMismatch
	}
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
	source, err := r.GetSourceServer(route.SourceServerUUID)
	if err != nil {
		return err
	}
	target, err := r.GetTargetServer(route.TargetServerUUID)
	if err != nil {
		return err
	}
	if !protocolsCompatible(source.Protocol, target.Protocol) {
		return ErrProtocolMismatch
	}
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