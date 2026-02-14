package impl

import (
	"errors"
	"log"

	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"
	"FeatherProxy/app/internal/database/token"

	"github.com/google/uuid"
)

const tokenMaskedPlaceholder = "***"

func (r *repository) CreateAuthentication(a schema.Authentication) error {
	log.Printf("auth/repo: CreateAuthentication id=%s name=%q token_type=%s", a.AuthenticationUUID, a.Name, a.TokenType)
	if a.Token == "" {
		return errors.New("token is required for create")
	}
	encrypted, salt, err := token.EncryptToken(a.Token)
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
	return r.invalidate(nil, []string{keyAuth(a.AuthenticationUUID), keyListAuthentications}, nil)
}

func (r *repository) GetAuthentication(id uuid.UUID) (schema.Authentication, error) {
	return getCached(r, keyAuth(id), func() (schema.Authentication, error) {
		log.Printf("auth/repo: GetAuthentication id=%s", id)
		var obj objects.Authentication
		if err := r.db.Where("authentication_uuid = ?", id).First(&obj).Error; err != nil {
			log.Printf("auth/repo: GetAuthentication not found or error: %v", err)
			return schema.Authentication{}, err
		}
		out := objects.AuthenticationToSchema(&obj)
		out.TokenMasked = tokenMaskedPlaceholder
		return out, nil
	})
}

// GetAuthenticationWithPlainToken is not cached (security: decrypted token).
func (r *repository) GetAuthenticationWithPlainToken(id uuid.UUID) (schema.Authentication, error) {
	log.Printf("auth/repo: GetAuthenticationWithPlainToken id=%s (for proxy)", id)
	var obj objects.Authentication
	if err := r.db.Where("authentication_uuid = ?", id).First(&obj).Error; err != nil {
		log.Printf("auth/repo: GetAuthenticationWithPlainToken not found: %v", err)
		return schema.Authentication{}, err
	}
	plain, err := token.DecryptToken(obj.TokenEncrypted, obj.TokenSalt)
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
		encrypted, salt, err := token.EncryptToken(a.Token)
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
	return r.invalidate(nil, []string{keyAuth(a.AuthenticationUUID), keyListAuthentications}, []string{keyPrefixTargetAuthForRoute})
}

func (r *repository) DeleteAuthentication(id uuid.UUID) error {
	log.Printf("auth/repo: DeleteAuthentication id=%s", id)
	return r.invalidate(r.db.Delete(&objects.Authentication{AuthenticationUUID: id}).Error,
		[]string{keyAuth(id), keyListAuthentications}, []string{keyPrefixTargetAuthForRoute})
}

func (r *repository) ListAuthentications() ([]schema.Authentication, error) {
	return getCached(r, keyListAuthentications, func() ([]schema.Authentication, error) {
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
	})
}
