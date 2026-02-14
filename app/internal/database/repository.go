package database

import (
	"errors"
	"log"

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
	FindRouteBySourceMethodPath(sourceServerUUID uuid.UUID, method, sourcePath string) (schema.Route, error)
	// Authentications
	CreateAuthentication(a schema.Authentication) error
	GetAuthentication(id uuid.UUID) (schema.Authentication, error)
	GetAuthenticationWithPlainToken(id uuid.UUID) (schema.Authentication, error) // For proxy only; returns decrypted Token
	UpdateAuthentication(a schema.Authentication) error
	DeleteAuthentication(id uuid.UUID) error
	ListAuthentications() ([]schema.Authentication, error)
	// Route auth mappings
	ListSourceAuthsForRoute(routeUUID uuid.UUID) ([]schema.RouteSourceAuth, error)
	SetSourceAuthsForRoute(routeUUID uuid.UUID, authUUIDs []uuid.UUID) error
	GetTargetAuthForRoute(routeUUID uuid.UUID) (uuid.UUID, bool, error)
	SetTargetAuthForRoute(routeUUID uuid.UUID, authUUID *uuid.UUID) error
	GetTargetAuthenticationWithPlainToken(routeUUID uuid.UUID) (schema.Authentication, bool, error) // For proxy
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

func (r *repository) FindRouteBySourceMethodPath(sourceServerUUID uuid.UUID, method, sourcePath string) (schema.Route, error) {
	var dbRoute objects.Route
	if err := r.db.Where("source_server_uuid = ? AND method = ? AND source_path = ?", sourceServerUUID, method, sourcePath).First(&dbRoute).Error; err != nil {
		return schema.Route{}, err
	}
	return objects.RouteToSchema(&dbRoute), nil
}

const tokenMaskedPlaceholder = "***"

func (r *repository) CreateAuthentication(a schema.Authentication) error {
	log.Printf("auth/repo: CreateAuthentication id=%s name=%q token_type=%s", a.AuthenticationUUID, a.Name, a.TokenType)
	if a.Token == "" {
		return errors.New("token is required for create")
	}
	encrypted, salt, err := EncryptToken(a.Token)
	if err != nil {
		return err
	}
	obj := objects.SchemaToAuthentication(a, encrypted, salt)
	err = r.db.Create(&obj).Error
	if err != nil {
		log.Printf("auth/repo: CreateAuthentication db error: %v", err)
		return err
	}
	log.Printf("auth/repo: CreateAuthentication ok id=%s", a.AuthenticationUUID)
	return nil
}

func (r *repository) GetAuthentication(id uuid.UUID) (schema.Authentication, error) {
	log.Printf("auth/repo: GetAuthentication id=%s", id)
	var obj objects.Authentication
	if err := r.db.Where("authentication_uuid = ?", id).First(&obj).Error; err != nil {
		log.Printf("auth/repo: GetAuthentication not found or error: %v", err)
		return schema.Authentication{}, err
	}
	out := objects.AuthenticationToSchema(&obj)
	out.TokenMasked = tokenMaskedPlaceholder
	return out, nil
}

func (r *repository) GetAuthenticationWithPlainToken(id uuid.UUID) (schema.Authentication, error) {
	log.Printf("auth/repo: GetAuthenticationWithPlainToken id=%s (for proxy)", id)
	var obj objects.Authentication
	if err := r.db.Where("authentication_uuid = ?", id).First(&obj).Error; err != nil {
		log.Printf("auth/repo: GetAuthenticationWithPlainToken not found: %v", err)
		return schema.Authentication{}, err
	}
	plain, err := DecryptToken(obj.TokenEncrypted, obj.TokenSalt)
	if err != nil {
		log.Printf("auth/repo: GetAuthenticationWithPlainToken decrypt error: %v", err)
		return schema.Authentication{}, err
	}
	out := objects.AuthenticationToSchema(&obj)
	out.Token = plain
	return out, nil
}

func (r *repository) UpdateAuthentication(a schema.Authentication) error {
	log.Printf("auth/repo: UpdateAuthentication id=%s name=%q token_provided=%v", a.AuthenticationUUID, a.Name, a.Token != "")
	var obj objects.Authentication
	if err := r.db.Where("authentication_uuid = ?", a.AuthenticationUUID).First(&obj).Error; err != nil {
		log.Printf("auth/repo: UpdateAuthentication not found: %v", err)
		return err
	}
	obj.Name = a.Name
	obj.TokenType = a.TokenType
	if a.Token != "" {
		encrypted, salt, err := EncryptToken(a.Token)
		if err != nil {
			return err
		}
		obj.TokenEncrypted = encrypted
		obj.TokenSalt = salt
	}
	err := r.db.Save(&obj).Error
	if err != nil {
		log.Printf("auth/repo: UpdateAuthentication save error: %v", err)
		return err
	}
	log.Printf("auth/repo: UpdateAuthentication ok id=%s", a.AuthenticationUUID)
	return nil
}

func (r *repository) DeleteAuthentication(id uuid.UUID) error {
	log.Printf("auth/repo: DeleteAuthentication id=%s", id)
	err := r.db.Delete(&objects.Authentication{AuthenticationUUID: id}).Error
	if err != nil {
		log.Printf("auth/repo: DeleteAuthentication error: %v", err)
		return err
	}
	return nil
}

func (r *repository) ListAuthentications() ([]schema.Authentication, error) {
	log.Printf("auth/repo: ListAuthentications")
	var list []objects.Authentication
	if err := r.db.Find(&list).Error; err != nil {
		log.Printf("auth/repo: ListAuthentications error: %v", err)
		return nil, err
	}
	out := make([]schema.Authentication, len(list))
	for i := range list {
		out[i] = objects.AuthenticationToSchema(&list[i])
		out[i].TokenMasked = tokenMaskedPlaceholder
	}
	log.Printf("auth/repo: ListAuthentications ok count=%d", len(out))
	return out, nil
}

func (r *repository) ListSourceAuthsForRoute(routeUUID uuid.UUID) ([]schema.RouteSourceAuth, error) {
	log.Printf("route_auth/repo: ListSourceAuthsForRoute route=%s", routeUUID)
	var list []objects.RouteSourceAuth
	if err := r.db.Where("route_uuid = ?", routeUUID).Order("position").Find(&list).Error; err != nil {
		log.Printf("route_auth/repo: ListSourceAuthsForRoute error: %v", err)
		return nil, err
	}
	out := make([]schema.RouteSourceAuth, len(list))
	for i := range list {
		out[i] = objects.RouteSourceAuthToSchema(&list[i])
	}
	log.Printf("route_auth/repo: ListSourceAuthsForRoute ok route=%s count=%d", routeUUID, len(out))
	return out, nil
}

func (r *repository) SetSourceAuthsForRoute(routeUUID uuid.UUID, authUUIDs []uuid.UUID) error {
	log.Printf("route_auth/repo: SetSourceAuthsForRoute route=%s count=%d", routeUUID, len(authUUIDs))
	if err := r.db.Unscoped().Where("route_uuid = ?", routeUUID).Delete(&objects.RouteSourceAuth{}).Error; err != nil {
		log.Printf("route_auth/repo: SetSourceAuthsForRoute delete error: %v", err)
		return err
	}
	for i, authUUID := range authUUIDs {
		obj := objects.RouteSourceAuth{
			RouteUUID:          routeUUID,
			AuthenticationUUID: authUUID,
			Position:           i,
		}
		if err := r.db.Create(&obj).Error; err != nil {
			log.Printf("route_auth/repo: SetSourceAuthsForRoute create error: %v", err)
			return err
		}
	}
	log.Printf("route_auth/repo: SetSourceAuthsForRoute ok route=%s", routeUUID)
	return nil
}

func (r *repository) GetTargetAuthForRoute(routeUUID uuid.UUID) (uuid.UUID, bool, error) {
	log.Printf("route_auth/repo: GetTargetAuthForRoute route=%s", routeUUID)
	var obj objects.RouteTargetAuth
	if err := r.db.Where("route_uuid = ?", routeUUID).First(&obj).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("route_auth/repo: GetTargetAuthForRoute route=%s no target auth", routeUUID)
			return uuid.Nil, false, nil
		}
		log.Printf("route_auth/repo: GetTargetAuthForRoute error: %v", err)
		return uuid.Nil, false, err
	}
	log.Printf("route_auth/repo: GetTargetAuthForRoute ok route=%s auth=%s", routeUUID, obj.AuthenticationUUID)
	return obj.AuthenticationUUID, true, nil
}

func (r *repository) SetTargetAuthForRoute(routeUUID uuid.UUID, authUUID *uuid.UUID) error {
	authStr := "none"
	if authUUID != nil && *authUUID != uuid.Nil {
		authStr = authUUID.String()
	}
	log.Printf("route_auth/repo: SetTargetAuthForRoute route=%s auth=%s", routeUUID, authStr)
	if err := r.db.Unscoped().Where("route_uuid = ?", routeUUID).Delete(&objects.RouteTargetAuth{}).Error; err != nil {
		log.Printf("route_auth/repo: SetTargetAuthForRoute delete error: %v", err)
		return err
	}
	if authUUID != nil && *authUUID != uuid.Nil {
		err := r.db.Create(&objects.RouteTargetAuth{
			RouteUUID:          routeUUID,
			AuthenticationUUID: *authUUID,
		}).Error
		if err != nil {
			log.Printf("route_auth/repo: SetTargetAuthForRoute create error: %v", err)
			return err
		}
	}
	log.Printf("route_auth/repo: SetTargetAuthForRoute ok route=%s", routeUUID)
	return nil
}

func (r *repository) GetTargetAuthenticationWithPlainToken(routeUUID uuid.UUID) (schema.Authentication, bool, error) {
	log.Printf("route_auth/repo: GetTargetAuthenticationWithPlainToken route=%s", routeUUID)
	authUUID, ok, err := r.GetTargetAuthForRoute(routeUUID)
	if err != nil || !ok {
		return schema.Authentication{}, false, err
	}
	auth, err := r.GetAuthenticationWithPlainToken(authUUID)
	if err != nil {
		return schema.Authentication{}, false, err
	}
	log.Printf("route_auth/repo: GetTargetAuthenticationWithPlainToken ok route=%s auth=%s", routeUUID, authUUID)
	return auth, true, nil
}