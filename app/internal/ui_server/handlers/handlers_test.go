package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// mockRepo implements database.Repository for handler tests. Set Fn* fields to inject behavior.
type mockRepo struct {
	FnListSourceServers    func() ([]schema.SourceServer, error)
	FnCreateSourceServer   func(schema.SourceServer) error
	FnGetSourceServer      func(uuid.UUID) (schema.SourceServer, error)
	FnUpdateSourceServer   func(schema.SourceServer) error
	FnDeleteSourceServer   func(uuid.UUID) error
	FnGetServerOptions     func(uuid.UUID) (schema.ServerOptions, error)
	FnSetServerOptions     func(schema.ServerOptions) error
	FnGetACLOptions        func(uuid.UUID) (schema.ACLOptions, error)
	FnSetACLOptions        func(schema.ACLOptions) error
	FnListTargetServers    func() ([]schema.TargetServer, error)
	FnCreateTargetServer   func(schema.TargetServer) error
	FnGetTargetServer      func(uuid.UUID) (schema.TargetServer, error)
	FnUpdateTargetServer   func(schema.TargetServer) error
	FnDeleteTargetServer   func(uuid.UUID) error
	FnListRoutes           func() ([]schema.Route, error)
	FnCreateRoute          func(schema.Route) error
	FnGetRoute             func(uuid.UUID) (schema.Route, error)
	FnUpdateRoute          func(schema.Route) error
	FnDeleteRoute          func(uuid.UUID) error
	FnListAuthentications  func() ([]schema.Authentication, error)
	FnCreateAuthentication func(schema.Authentication) error
	FnGetAuthentication    func(uuid.UUID) (schema.Authentication, error)
	FnUpdateAuthentication func(schema.Authentication) error
	FnDeleteAuthentication func(uuid.UUID) error
	FnListSourceAuthsForRoute  func(uuid.UUID) ([]schema.RouteSourceAuth, error)
	FnSetSourceAuthsForRoute   func(uuid.UUID, []uuid.UUID) error
	FnGetTargetAuthForRoute    func(uuid.UUID) (uuid.UUID, bool, error)
	FnSetTargetAuthForRoute    func(uuid.UUID, *uuid.UUID) error
}

func (m *mockRepo) ListSourceServers() ([]schema.SourceServer, error) {
	if m.FnListSourceServers != nil {
		return m.FnListSourceServers()
	}
	return nil, nil
}
func (m *mockRepo) CreateSourceServer(s schema.SourceServer) error {
	if m.FnCreateSourceServer != nil {
		return m.FnCreateSourceServer(s)
	}
	return nil
}
func (m *mockRepo) GetSourceServer(id uuid.UUID) (schema.SourceServer, error) {
	if m.FnGetSourceServer != nil {
		return m.FnGetSourceServer(id)
	}
	return schema.SourceServer{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) UpdateSourceServer(s schema.SourceServer) error {
	if m.FnUpdateSourceServer != nil {
		return m.FnUpdateSourceServer(s)
	}
	return nil
}
func (m *mockRepo) DeleteSourceServer(id uuid.UUID) error {
	if m.FnDeleteSourceServer != nil {
		return m.FnDeleteSourceServer(id)
	}
	return nil
}
func (m *mockRepo) GetServerOptions(id uuid.UUID) (schema.ServerOptions, error) {
	if m.FnGetServerOptions != nil {
		return m.FnGetServerOptions(id)
	}
	return schema.ServerOptions{}, nil
}
func (m *mockRepo) SetServerOptions(opts schema.ServerOptions) error {
	if m.FnSetServerOptions != nil {
		return m.FnSetServerOptions(opts)
	}
	return nil
}
func (m *mockRepo) GetACLOptions(id uuid.UUID) (schema.ACLOptions, error) {
	if m.FnGetACLOptions != nil {
		return m.FnGetACLOptions(id)
	}
	return schema.ACLOptions{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) SetACLOptions(opts schema.ACLOptions) error {
	if m.FnSetACLOptions != nil {
		return m.FnSetACLOptions(opts)
	}
	return nil
}
func (m *mockRepo) ListTargetServers() ([]schema.TargetServer, error) {
	if m.FnListTargetServers != nil {
		return m.FnListTargetServers()
	}
	return nil, nil
}
func (m *mockRepo) CreateTargetServer(t schema.TargetServer) error {
	if m.FnCreateTargetServer != nil {
		return m.FnCreateTargetServer(t)
	}
	return nil
}
func (m *mockRepo) GetTargetServer(id uuid.UUID) (schema.TargetServer, error) {
	if m.FnGetTargetServer != nil {
		return m.FnGetTargetServer(id)
	}
	return schema.TargetServer{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) UpdateTargetServer(t schema.TargetServer) error {
	if m.FnUpdateTargetServer != nil {
		return m.FnUpdateTargetServer(t)
	}
		return nil
}
func (m *mockRepo) DeleteTargetServer(id uuid.UUID) error {
	if m.FnDeleteTargetServer != nil {
		return m.FnDeleteTargetServer(id)
	}
	return nil
}
func (m *mockRepo) ListRoutes() ([]schema.Route, error) {
	if m.FnListRoutes != nil {
		return m.FnListRoutes()
	}
	return nil, nil
}
func (m *mockRepo) CreateRoute(r schema.Route) error {
	if m.FnCreateRoute != nil {
		return m.FnCreateRoute(r)
	}
	return nil
}
func (m *mockRepo) GetRoute(id uuid.UUID) (schema.Route, error) {
	if m.FnGetRoute != nil {
		return m.FnGetRoute(id)
	}
	return schema.Route{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) UpdateRoute(r schema.Route) error {
	if m.FnUpdateRoute != nil {
		return m.FnUpdateRoute(r)
	}
	return nil
}
func (m *mockRepo) DeleteRoute(id uuid.UUID) error {
	if m.FnDeleteRoute != nil {
		return m.FnDeleteRoute(id)
	}
	return nil
}
func (m *mockRepo) ListAuthentications() ([]schema.Authentication, error) {
	if m.FnListAuthentications != nil {
		return m.FnListAuthentications()
	}
	return nil, nil
}
func (m *mockRepo) CreateAuthentication(a schema.Authentication) error {
	if m.FnCreateAuthentication != nil {
		return m.FnCreateAuthentication(a)
	}
	return nil
}
func (m *mockRepo) GetAuthentication(id uuid.UUID) (schema.Authentication, error) {
	if m.FnGetAuthentication != nil {
		return m.FnGetAuthentication(id)
	}
	return schema.Authentication{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) UpdateAuthentication(a schema.Authentication) error {
	if m.FnUpdateAuthentication != nil {
		return m.FnUpdateAuthentication(a)
	}
	return nil
}
func (m *mockRepo) DeleteAuthentication(id uuid.UUID) error {
	if m.FnDeleteAuthentication != nil {
		return m.FnDeleteAuthentication(id)
	}
	return nil
}
func (m *mockRepo) ListSourceAuthsForRoute(routeID uuid.UUID) ([]schema.RouteSourceAuth, error) {
	if m.FnListSourceAuthsForRoute != nil {
		return m.FnListSourceAuthsForRoute(routeID)
	}
	return nil, nil
}
func (m *mockRepo) SetSourceAuthsForRoute(routeID uuid.UUID, authIDs []uuid.UUID) error {
	if m.FnSetSourceAuthsForRoute != nil {
		return m.FnSetSourceAuthsForRoute(routeID, authIDs)
	}
	return nil
}
func (m *mockRepo) GetTargetAuthForRoute(routeID uuid.UUID) (uuid.UUID, bool, error) {
	if m.FnGetTargetAuthForRoute != nil {
		return m.FnGetTargetAuthForRoute(routeID)
	}
	return uuid.Nil, false, nil
}
func (m *mockRepo) SetTargetAuthForRoute(routeID uuid.UUID, authID *uuid.UUID) error {
	if m.FnSetTargetAuthForRoute != nil {
		return m.FnSetTargetAuthForRoute(routeID, authID)
	}
	return nil
}

// Unused by handlers but required by interface
func (m *mockRepo) GetRouteFromSourcePath(string) (schema.Route, error) {
	return schema.Route{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) GetRouteFromTargetPath(string) (schema.Route, error) {
	return schema.Route{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) FindRouteBySourceMethodPath(uuid.UUID, string, string) (schema.Route, error) {
	return schema.Route{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) GetAuthenticationWithPlainToken(uuid.UUID) (schema.Authentication, error) {
	return schema.Authentication{}, gorm.ErrRecordNotFound
}
func (m *mockRepo) GetTargetAuthenticationWithPlainToken(uuid.UUID) (schema.Authentication, bool, error) {
	return schema.Authentication{}, false, nil
}

func (m *mockRepo) CreateProxyStats([]schema.ProxyStat) error { return nil }
func (m *mockRepo) ListProxyStats(limit, offset int, since *time.Time) ([]schema.ProxyStat, int64, error) {
	return nil, 0, nil
}
func (m *mockRepo) DeleteProxyStatsOlderThan(time.Time) (int64, error) { return 0, nil }
func (m *mockRepo) ClearAllProxyStats() error                 { return nil }
func (m *mockRepo) StatsSummary() (schema.StatsSummary, error) {
	return schema.StatsSummary{}, nil
}
func (m *mockRepo) StatsByRoute(*time.Time, int) ([]schema.RouteCount, error) { return nil, nil }
func (m *mockRepo) StatsByCaller(*time.Time, int) ([]schema.CallerCount, error) { return nil, nil }
func (m *mockRepo) StatsBySourceServer(*time.Time) ([]schema.ServerCount, error) { return nil, nil }
func (m *mockRepo) StatsByTargetServer(*time.Time) ([]schema.ServerCount, error) { return nil, nil }
func (m *mockRepo) StatsTPS(time.Time, time.Duration) ([]schema.BucketCount, error) { return nil, nil }

var _ database.Repository = (*mockRepo)(nil)

// --- Source servers ---

func TestListSourceServers(t *testing.T) {
	repo := &mockRepo{
		FnListSourceServers: func() ([]schema.SourceServer, error) {
			return []schema.SourceServer{{Name: "s1", Host: "localhost", Port: 8080}}, nil
		},
	}
	w := httptest.NewRecorder()
	ListSourceServers(repo, w, httptest.NewRequest(http.MethodGet, "/api/source-servers", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var list []schema.SourceServer
	if err := json.NewDecoder(w.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 || list[0].Name != "s1" {
		t.Errorf("body = %+v", list)
	}
}

func TestListSourceServers_error(t *testing.T) {
	repo := &mockRepo{
		FnListSourceServers: func() ([]schema.SourceServer, error) {
			return nil, errors.New("db error")
		},
	}
	w := httptest.NewRecorder()
	ListSourceServers(repo, w, httptest.NewRequest(http.MethodGet, "/api/source-servers", nil))
	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", w.Code)
	}
}

func TestCreateSourceServer(t *testing.T) {
	var created schema.SourceServer
	repo := &mockRepo{
		FnCreateSourceServer: func(s schema.SourceServer) error {
			created = s
			return nil
		},
	}
	body := `{"name":"api","protocol":"http","host":"0.0.0.0","port":4545}`
	w := httptest.NewRecorder()
	CreateSourceServer(repo, w, httptest.NewRequest(http.MethodPost, "/api/source-servers", bytes.NewReader([]byte(body))))
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if created.Name != "api" || created.Host != "0.0.0.0" || created.Port != 4545 {
		t.Errorf("created = %+v", created)
	}
}

func TestCreateSourceServer_badRequest(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	CreateSourceServer(repo, w, httptest.NewRequest(http.MethodPost, "/api/source-servers", bytes.NewReader([]byte(`{"name":"x"}`))))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGetSourceServer_notFound(t *testing.T) {
	repo := &mockRepo{} // GetSourceServer returns ErrRecordNotFound by default
	id := uuid.New()
	w := httptest.NewRecorder()
	GetSourceServer(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), id.String())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetSourceServer_ok(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		FnGetSourceServer: func(uuid.UUID) (schema.SourceServer, error) {
			return schema.SourceServer{SourceServerUUID: id, Name: "s1", Host: "localhost", Port: 8080}, nil
		},
	}
	w := httptest.NewRecorder()
	GetSourceServer(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), id.String())
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGetSourceServer_invalidUUID(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	GetSourceServer(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), "not-a-uuid")
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestDeleteSourceServer(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{FnDeleteSourceServer: func(uuid.UUID) error { return nil }}
	w := httptest.NewRecorder()
	DeleteSourceServer(repo, w, httptest.NewRequest(http.MethodDelete, "/", nil), id.String())
	if w.Code != http.StatusNoContent {
		t.Errorf("status = %d, want 204", w.Code)
	}
}

func TestGetACLOptions_notFound(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		FnGetSourceServer: func(uuid.UUID) (schema.SourceServer, error) {
			return schema.SourceServer{SourceServerUUID: id}, nil
		},
		// FnGetACLOptions nil => returns ErrRecordNotFound
	}
	w := httptest.NewRecorder()
	GetACLOptions(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), id.String())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

func TestGetACLOptions_ok(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		FnGetSourceServer: func(uuid.UUID) (schema.SourceServer, error) {
			return schema.SourceServer{SourceServerUUID: id}, nil
		},
		FnGetACLOptions: func(uuid.UUID) (schema.ACLOptions, error) {
			return schema.ACLOptions{
				SourceServerUUID: id,
				Mode:             "allow_only",
				ClientIPHeader:   "X-Forwarded-For",
				AllowList:        []string{"192.168.1.0/24"},
			}, nil
		},
	}
	w := httptest.NewRecorder()
	GetACLOptions(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), id.String())
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	var got schema.ACLOptions
	if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Mode != "allow_only" || got.ClientIPHeader != "X-Forwarded-For" || len(got.AllowList) != 1 {
		t.Errorf("got = %+v", got)
	}
}

func TestSetACLOptions_ok(t *testing.T) {
	id := uuid.New()
	var saved schema.ACLOptions
	repo := &mockRepo{
		FnGetSourceServer: func(uuid.UUID) (schema.SourceServer, error) {
			return schema.SourceServer{SourceServerUUID: id}, nil
		},
		FnSetACLOptions: func(opts schema.ACLOptions) error {
			saved = opts
			return nil
		},
		FnGetACLOptions: func(uuid.UUID) (schema.ACLOptions, error) {
			return saved, nil
		},
	}
	body := `{"mode":"deny_only","client_ip_header":"X-Real-IP","allow_list":[],"deny_list":["10.0.0.1"]}`
	w := httptest.NewRecorder()
	SetACLOptions(repo, w, httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(body))), id.String())
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if saved.Mode != "deny_only" || saved.ClientIPHeader != "X-Real-IP" || len(saved.DenyList) != 1 || saved.DenyList[0] != "10.0.0.1" {
		t.Errorf("saved = %+v", saved)
	}
}

func TestSetACLOptions_invalidMode(t *testing.T) {
	id := uuid.New()
	repo := &mockRepo{
		FnGetSourceServer: func(uuid.UUID) (schema.SourceServer, error) {
			return schema.SourceServer{SourceServerUUID: id}, nil
		},
	}
	body := `{"mode":"invalid","allow_list":[],"deny_list":[]}`
	w := httptest.NewRecorder()
	SetACLOptions(repo, w, httptest.NewRequest(http.MethodPut, "/", bytes.NewReader([]byte(body))), id.String())
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

// --- Target servers ---

func TestListTargetServers(t *testing.T) {
	repo := &mockRepo{
		FnListTargetServers: func() ([]schema.TargetServer, error) {
			return []schema.TargetServer{{Name: "t1", Host: "backend", Port: 443}}, nil
		},
	}
	w := httptest.NewRecorder()
	ListTargetServers(repo, w, httptest.NewRequest(http.MethodGet, "/api/target-servers", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCreateTargetServer(t *testing.T) {
	var created schema.TargetServer
	repo := &mockRepo{
		FnCreateTargetServer: func(t schema.TargetServer) error {
			created = t
			return nil
		},
	}
	body := `{"name":"backend","protocol":"https","host":"api.example.com","port":443,"base_path":"/v1"}`
	w := httptest.NewRecorder()
	CreateTargetServer(repo, w, httptest.NewRequest(http.MethodPost, "/api/target-servers", bytes.NewReader([]byte(body))))
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if created.Name != "backend" || created.Port != 443 {
		t.Errorf("created = %+v", created)
	}
}

func TestGetTargetServer_notFound(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	GetTargetServer(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// --- Routes ---

func TestListRoutes(t *testing.T) {
	repo := &mockRepo{
		FnListRoutes: func() ([]schema.Route, error) {
			return []schema.Route{}, nil
		},
	}
	w := httptest.NewRecorder()
	ListRoutes(repo, w, httptest.NewRequest(http.MethodGet, "/api/routes", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCreateRoute(t *testing.T) {
	sourceID := uuid.New()
	targetID := uuid.New()
	var created schema.Route
	repo := &mockRepo{
		FnCreateRoute: func(r schema.Route) error {
			created = r
			return nil
		},
	}
	body := map[string]string{
		"source_server_uuid": sourceID.String(),
		"target_server_uuid": targetID.String(),
		"method":             "GET",
		"source_path":        "/foo",
		"target_path":        "/bar",
	}
	bodyBytes, _ := json.Marshal(body)
	w := httptest.NewRecorder()
	CreateRoute(repo, w, httptest.NewRequest(http.MethodPost, "/api/routes", bytes.NewReader(bodyBytes)))
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if created.Method != "GET" || created.SourcePath != "/foo" || created.TargetPath != "/bar" {
		t.Errorf("created = %+v", created)
	}
}

func TestCreateRoute_protocolMismatch(t *testing.T) {
	repo := &mockRepo{
		FnCreateRoute: func(schema.Route) error { return database.ErrProtocolMismatch },
	}
	body := `{"source_server_uuid":"` + uuid.New().String() + `","target_server_uuid":"` + uuid.New().String() + `","method":"GET","source_path":"/","target_path":"/"}`
	w := httptest.NewRecorder()
	CreateRoute(repo, w, httptest.NewRequest(http.MethodPost, "/api/routes", bytes.NewReader([]byte(body))))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGetRoute_notFound(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	GetRoute(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// --- Authentications ---

func TestListAuthentications(t *testing.T) {
	repo := &mockRepo{FnListAuthentications: func() ([]schema.Authentication, error) { return []schema.Authentication{}, nil }}
	w := httptest.NewRecorder()
	ListAuthentications(repo, w, httptest.NewRequest(http.MethodGet, "/api/authentications", nil))
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestCreateAuthentication(t *testing.T) {
	var created schema.Authentication
	repo := &mockRepo{
		FnCreateAuthentication: func(a schema.Authentication) error {
			created = a
			return nil
		},
		FnGetAuthentication: func(id uuid.UUID) (schema.Authentication, error) {
			return schema.Authentication{AuthenticationUUID: id, Name: created.Name, TokenType: created.TokenType}, nil
		},
	}
	body := `{"name":"api-key","token_type":"Bearer","token":"secret"}`
	w := httptest.NewRecorder()
	CreateAuthentication(repo, w, httptest.NewRequest(http.MethodPost, "/api/authentications", bytes.NewReader([]byte(body))))
	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want 201", w.Code)
	}
	if created.Name != "api-key" || created.Token != "secret" {
		t.Errorf("created = %+v", created)
	}
}

func TestCreateAuthentication_noToken(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	CreateAuthentication(repo, w, httptest.NewRequest(http.MethodPost, "/api/authentications", bytes.NewReader([]byte(`{"name":"x"}`))))
	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", w.Code)
	}
}

func TestGetAuthentication_notFound(t *testing.T) {
	repo := &mockRepo{}
	w := httptest.NewRecorder()
	GetAuthentication(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", w.Code)
	}
}

// --- Route auth ---

func TestGetRouteSourceAuth(t *testing.T) {
	repo := &mockRepo{
		FnListSourceAuthsForRoute: func(uuid.UUID) ([]schema.RouteSourceAuth, error) {
			return []schema.RouteSourceAuth{}, nil
		},
	}
	w := httptest.NewRecorder()
	GetRouteSourceAuth(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestGetRouteTargetAuth(t *testing.T) {
	repo := &mockRepo{
		FnGetTargetAuthForRoute: func(uuid.UUID) (uuid.UUID, bool, error) {
			return uuid.Nil, false, nil
		},
	}
	w := httptest.NewRecorder()
	GetRouteTargetAuth(repo, w, httptest.NewRequest(http.MethodGet, "/", nil), uuid.New().String())
	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", w.Code)
	}
}
