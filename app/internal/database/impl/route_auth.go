package impl

import (
	"errors"
	"log"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func (r *repository) ListSourceAuthsForRoute(routeUUID uuid.UUID) ([]schema.RouteSourceAuth, error) {
	return getCached(r, keyRouteSourceAuths(routeUUID), func() ([]schema.RouteSourceAuth, error) {
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
	})
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
	return r.invalidate(nil, []string{keyRouteSourceAuths(routeUUID)}, nil)
}

func (r *repository) GetTargetAuthForRoute(routeUUID uuid.UUID) (uuid.UUID, bool, error) {
	v, err := getCached(r, keyTargetAuthForRoute(routeUUID), func() (targetAuthCached, error) {
		log.Printf("route_auth/repo: GetTargetAuthForRoute route=%s", routeUUID)
		var obj objects.RouteTargetAuth
		if err := r.db.Where("route_uuid = ?", routeUUID).First(&obj).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				log.Printf("route_auth/repo: GetTargetAuthForRoute route=%s no target auth", routeUUID)
				return targetAuthCached{}, nil
			}
			log.Printf("route_auth/repo: GetTargetAuthForRoute error: %v", err)
			return targetAuthCached{}, err
		}
		log.Printf("route_auth/repo: GetTargetAuthForRoute ok route=%s auth=%s", routeUUID, obj.AuthenticationUUID)
		return targetAuthCached{AuthUUID: obj.AuthenticationUUID, OK: true}, nil
	})
	return v.AuthUUID, v.OK, err
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
	return r.invalidate(nil, []string{keyTargetAuthForRoute(routeUUID)}, nil)
}

// GetTargetAuthenticationWithPlainToken is not cached (security: decrypted token).
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
