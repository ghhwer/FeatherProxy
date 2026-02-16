package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// stubRepo implements database.Repository with no-op / empty responses for route tests.
type stubRepo struct{}

func (stubRepo) ListSourceServers() ([]schema.SourceServer, error)          { return nil, nil }
func (stubRepo) CreateSourceServer(schema.SourceServer) error               { return nil }
func (stubRepo) GetSourceServer(uuid.UUID) (schema.SourceServer, error)     { return schema.SourceServer{}, gorm.ErrRecordNotFound }
func (stubRepo) UpdateSourceServer(schema.SourceServer) error               { return nil }
func (stubRepo) DeleteSourceServer(uuid.UUID) error                         { return nil }
func (stubRepo) GetServerOptions(uuid.UUID) (schema.ServerOptions, error)   { return schema.ServerOptions{}, nil }
func (stubRepo) SetServerOptions(schema.ServerOptions) error                { return nil }
func (stubRepo) GetACLOptions(uuid.UUID) (schema.ACLOptions, error)         { return schema.ACLOptions{}, nil }
func (stubRepo) SetACLOptions(schema.ACLOptions) error                      { return nil }
func (stubRepo) ListTargetServers() ([]schema.TargetServer, error)          { return nil, nil }
func (stubRepo) CreateTargetServer(schema.TargetServer) error               { return nil }
func (stubRepo) GetTargetServer(uuid.UUID) (schema.TargetServer, error)     { return schema.TargetServer{}, gorm.ErrRecordNotFound }
func (stubRepo) UpdateTargetServer(schema.TargetServer) error               { return nil }
func (stubRepo) DeleteTargetServer(uuid.UUID) error                         { return nil }
func (stubRepo) ListRoutes() ([]schema.Route, error)                        { return nil, nil }
func (stubRepo) CreateRoute(schema.Route) error                             { return nil }
func (stubRepo) GetRoute(uuid.UUID) (schema.Route, error)                   { return schema.Route{}, gorm.ErrRecordNotFound }
func (stubRepo) UpdateRoute(schema.Route) error                              { return nil }
func (stubRepo) DeleteRoute(uuid.UUID) error                                { return nil }
func (stubRepo) GetRouteFromSourcePath(string) (schema.Route, error)         { return schema.Route{}, gorm.ErrRecordNotFound }
func (stubRepo) GetRouteFromTargetPath(string) (schema.Route, error)         { return schema.Route{}, gorm.ErrRecordNotFound }
func (stubRepo) FindRouteBySourceMethodPath(uuid.UUID, string, string) (schema.Route, error) {
	return schema.Route{}, gorm.ErrRecordNotFound
}
func (stubRepo) ListAuthentications() ([]schema.Authentication, error)        { return nil, nil }
func (stubRepo) CreateAuthentication(schema.Authentication) error            { return nil }
func (stubRepo) GetAuthentication(uuid.UUID) (schema.Authentication, error) { return schema.Authentication{}, gorm.ErrRecordNotFound }
func (stubRepo) GetAuthenticationWithPlainToken(uuid.UUID) (schema.Authentication, error) {
	return schema.Authentication{}, gorm.ErrRecordNotFound
}
func (stubRepo) UpdateAuthentication(schema.Authentication) error           { return nil }
func (stubRepo) DeleteAuthentication(uuid.UUID) error                      { return nil }
func (stubRepo) ListSourceAuthsForRoute(uuid.UUID) ([]schema.RouteSourceAuth, error) {
	return nil, nil
}
func (stubRepo) SetSourceAuthsForRoute(uuid.UUID, []uuid.UUID) error         { return nil }
func (stubRepo) GetTargetAuthForRoute(uuid.UUID) (uuid.UUID, bool, error)   { return uuid.Nil, false, nil }
func (stubRepo) SetTargetAuthForRoute(uuid.UUID, *uuid.UUID) error           { return nil }
func (stubRepo) GetTargetAuthenticationWithPlainToken(uuid.UUID) (schema.Authentication, bool, error) {
	return schema.Authentication{}, false, nil
}
func (stubRepo) CreateProxyStats([]schema.ProxyStat) error { return nil }
func (stubRepo) ListProxyStats(int, int, *time.Time) ([]schema.ProxyStat, int64, error) {
	return nil, 0, nil
}
func (stubRepo) DeleteProxyStatsOlderThan(time.Time) error { return nil }
func (stubRepo) ClearAllProxyStats() error                 { return nil }
func (stubRepo) StatsSummary() (schema.StatsSummary, error) { return schema.StatsSummary{}, nil }
func (stubRepo) StatsByRoute(*time.Time, int) ([]schema.RouteCount, error) { return nil, nil }
func (stubRepo) StatsByCaller(*time.Time, int) ([]schema.CallerCount, error) { return nil, nil }
func (stubRepo) StatsBySourceServer(*time.Time) ([]schema.ServerCount, error) { return nil, nil }
func (stubRepo) StatsByTargetServer(*time.Time) ([]schema.ServerCount, error) { return nil, nil }
func (stubRepo) StatsTPS(time.Time, time.Duration) ([]schema.BucketCount, error) { return nil, nil }

var _ database.Repository = (*stubRepo)(nil)

func TestNewServer(t *testing.T) {
	s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
	if s == nil {
		t.Fatal("NewServer returned nil")
	}
	if s.Routes() == nil {
		t.Fatal("Routes() returned nil")
	}
}

func TestServer_Routes_reload(t *testing.T) {
	handler := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil).Routes()

	t.Run("POST /api/reload when onReload nil returns 503", func(t *testing.T) {
		s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
		h := s.Routes()
		req := httptest.NewRequest(http.MethodPost, "/api/reload", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusServiceUnavailable {
			t.Errorf("status = %d, want 503", rec.Code)
		}
		var body map[string]string
		if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["error"] != "reload not configured" {
			t.Errorf("body = %v", body)
		}
	})

	t.Run("POST /api/reload when onReload set returns 200", func(t *testing.T) {
		called := false
		onReload := func() { called = true }
		s := NewServer(":0", stubRepo{}, "internal/ui_server/static", onReload)
		h := s.Routes()
		req := httptest.NewRequest(http.MethodPost, "/api/reload", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		if !called {
			t.Error("onReload was not called")
		}
	})

	t.Run("GET /api/reload returns 405", func(t *testing.T) {
		s := NewServer(":0", stubRepo{}, "internal/ui_server/static", func() {})
		h := s.Routes()
		req := httptest.NewRequest(http.MethodGet, "/api/reload", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("status = %d, want 405", rec.Code)
		}
	})

	_ = handler
}

func TestServer_Routes_sourceServers(t *testing.T) {
	s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
	h := s.Routes()

	t.Run("GET /api/source-servers returns 200", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/source-servers", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
	})

	t.Run("POST /api/source-servers returns 201", func(t *testing.T) {
		body := `{"name":"s1","protocol":"http","host":"localhost","port":8080}`
		req := httptest.NewRequest(http.MethodPost, "/api/source-servers", bytes.NewReader([]byte(body)))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Errorf("status = %d, want 201", rec.Code)
		}
	})

	t.Run("GET /api/source-servers/{valid-uuid} returns 404 when not found", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/source-servers/"+uuid.New().String(), nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})
}

func TestServer_Routes_targetServers(t *testing.T) {
	s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
	h := s.Routes()
	req := httptest.NewRequest(http.MethodGet, "/api/target-servers", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/target-servers status = %d, want 200", rec.Code)
	}
}

func TestServer_Routes_routes(t *testing.T) {
	s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
	h := s.Routes()
	req := httptest.NewRequest(http.MethodGet, "/api/routes", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("GET /api/routes status = %d, want 200", rec.Code)
	}
}

func TestServer_Run_shutdown(t *testing.T) {
	s := NewServer(":0", stubRepo{}, "internal/ui_server/static", nil)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- s.Run(ctx) }()
	// Give server a moment to start
	time.Sleep(50 * time.Millisecond)
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("Run: %v", err)
	}
}
