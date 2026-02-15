package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"FeatherProxy/app/internal/database/repo"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

func Test_joinHostPort(t *testing.T) {
	tests := []struct {
		name string
		host string
		port int
		want string
	}{
		{"port zero", "localhost", 0, "localhost"},
		{"with port", "localhost", 4545, "localhost:4545"},
		{"host only", "example.com", 0, "example.com"},
		{"host and port", "127.0.0.1", 8080, "127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinHostPort(tt.host, tt.port)
			if got != tt.want {
				t.Errorf("joinHostPort(%q, %d) = %q, want %q", tt.host, tt.port, got, tt.want)
			}
		})
	}
}

func Test_joinPath(t *testing.T) {
	tests := []struct {
		name string
		base string
		p    string
		want string
	}{
		{"both empty", "", "", "/"},
		{"base empty", "", "api/foo", "/api/foo"},
		{"path empty", "/base", "", "/base/"}, // joinPath(base, "") => base + "/" + "" => "/base/"
		{"no trailing slash", "/base", "path", "/base/path"},
		{"base with trailing", "/base/", "path", "/base/path"},
		{"path with leading", "/base", "/path", "/base/path"},
		{"both slashes", "/base/", "/path", "/base/path"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinPath(tt.base, tt.p)
			if got != tt.want {
				t.Errorf("joinPath(%q, %q) = %q, want %q", tt.base, tt.p, got, tt.want)
			}
		})
	}
}

func Test_buildTargetURL(t *testing.T) {
	target := &schema.TargetServer{
		Protocol: "https",
		Host:     "backend.example.com",
		Port:     443,
		BasePath: "/api/v1",
	}
	route := &schema.Route{
		TargetPath: "/users",
	}
	u := buildTargetURL(target, route, "q=1")
	if u.Scheme != "https" {
		t.Errorf("Scheme = %q, want https", u.Scheme)
	}
	if u.Host != "backend.example.com:443" {
		t.Errorf("Host = %q, want backend.example.com:443", u.Host)
	}
	if u.Path != "/api/v1/users" {
		t.Errorf("Path = %q, want /api/v1/users", u.Path)
	}
	if u.RawQuery != "q=1" {
		t.Errorf("RawQuery = %q, want q=1", u.RawQuery)
	}
}

func Test_buildTargetURL_portZero(t *testing.T) {
	target := &schema.TargetServer{
		Protocol: "https",
		Host:     "backend.example.com",
		Port:     0,
		BasePath: "/api",
	}
	route := &schema.Route{TargetPath: "/foo"}
	u := buildTargetURL(target, route, "")
	if u.Host != "backend.example.com" {
		t.Errorf("Host with Port 0 = %q, want host only (no :0)", u.Host)
	}
	if u.Scheme != "https" || u.Path != "/api/foo" {
		t.Errorf("Scheme = %q Path = %q", u.Scheme, u.Path)
	}
}

func Test_director(t *testing.T) {
	target, _ := url.Parse("https://backend.example.com/api")
	incoming := httptest.NewRequest(http.MethodGet, "http://proxy.example.com/foo", nil)
	incoming.RemoteAddr = "192.168.1.1:12345"

	t.Run("no target auth forwards Authorization", func(t *testing.T) {
		incoming.Header.Set("Authorization", "Bearer client-token")
		dir := director(target, incoming, nil)
		out := &http.Request{Header: http.Header{}, URL: &url.URL{}}
		dir(out)
		if out.URL.Scheme != "https" || out.URL.Host != "backend.example.com" || out.URL.Path != "/api" {
			t.Errorf("out.URL = %v", out.URL)
		}
		if v := out.Header.Get("X-Forwarded-For"); v != "192.168.1.1:12345" {
			t.Errorf("X-Forwarded-For = %q", v)
		}
		if v := out.Header.Get("X-Forwarded-Proto"); v != "http" {
			t.Errorf("X-Forwarded-Proto = %q", v)
		}
		if v := out.Header.Get("Authorization"); v != "Bearer client-token" {
			t.Errorf("Authorization = %q", v)
		}
	})

	t.Run("X-Forwarded-Proto https when incoming TLS", func(t *testing.T) {
		incomingTLS := httptest.NewRequest(http.MethodGet, "https://proxy.example.com/foo", nil)
		incomingTLS.TLS = &tls.ConnectionState{}
		incomingTLS.RemoteAddr = "10.0.0.1:443"
		dir := director(target, incomingTLS, nil)
		out := &http.Request{Header: http.Header{}, URL: &url.URL{}}
		dir(out)
		if v := out.Header.Get("X-Forwarded-Proto"); v != "https" {
			t.Errorf("X-Forwarded-Proto = %q, want https", v)
		}
	})

	t.Run("target auth Bearer overwrites Authorization", func(t *testing.T) {
		auth := &schema.Authentication{TokenType: "Bearer", Token: "backend-secret"}
		dir := director(target, incoming, auth)
		out := &http.Request{Header: http.Header{}, URL: &url.URL{}}
		dir(out)
		if v := out.Header.Get("Authorization"); v != "Bearer backend-secret" {
			t.Errorf("Authorization = %q", v)
		}
	})

	t.Run("target auth raw token type", func(t *testing.T) {
		auth := &schema.Authentication{TokenType: "Basic", Token: "dXNlcjpwYXNz"}
		dir := director(target, incoming, auth)
		out := &http.Request{Header: http.Header{}, URL: &url.URL{}}
		dir(out)
		if v := out.Header.Get("Authorization"); v != "dXNlcjpwYXNz" {
			t.Errorf("Authorization = %q", v)
		}
	})
}

// mockRepo implements database.Repository for tests; only proxy-used methods are set.
type mockRepo struct {
	listSources     func() ([]schema.SourceServer, error)
	getServerOpts   func(uuid.UUID) (schema.ServerOptions, error)
	findRoute       func(sourceUUID uuid.UUID, method, path string) (schema.Route, error)
	getTarget       func(uuid.UUID) (schema.TargetServer, error)
	getTargetAuth   func(routeUUID uuid.UUID) (schema.Authentication, bool, error)
}

func (m *mockRepo) ListSourceServers() ([]schema.SourceServer, error) {
	if m.listSources != nil {
		return m.listSources()
	}
	return nil, nil
}
func (m *mockRepo) GetServerOptions(uuid.UUID) (schema.ServerOptions, error) {
	if m.getServerOpts != nil {
		return m.getServerOpts(uuid.UUID{})
	}
	return schema.ServerOptions{}, nil
}
func (m *mockRepo) FindRouteBySourceMethodPath(s uuid.UUID, method, path string) (schema.Route, error) {
	if m.findRoute != nil {
		return m.findRoute(s, method, path)
	}
	return schema.Route{}, errors.New("not found")
}
func (m *mockRepo) GetTargetServer(u uuid.UUID) (schema.TargetServer, error) {
	if m.getTarget != nil {
		return m.getTarget(u)
	}
	return schema.TargetServer{}, errors.New("not found")
}
func (m *mockRepo) GetTargetAuthenticationWithPlainToken(routeUUID uuid.UUID) (schema.Authentication, bool, error) {
	if m.getTargetAuth != nil {
		return m.getTargetAuth(routeUUID)
	}
	return schema.Authentication{}, false, nil
}

func (m *mockRepo) CreateSourceServer(schema.SourceServer) error              { return nil }
func (m *mockRepo) GetSourceServer(uuid.UUID) (schema.SourceServer, error)    { return schema.SourceServer{}, nil }
func (m *mockRepo) UpdateSourceServer(schema.SourceServer) error              { return nil }
func (m *mockRepo) DeleteSourceServer(uuid.UUID) error                       { return nil }
func (m *mockRepo) SetServerOptions(schema.ServerOptions) error              { return nil }
func (m *mockRepo) CreateTargetServer(schema.TargetServer) error              { return nil }
func (m *mockRepo) UpdateTargetServer(schema.TargetServer) error              { return nil }
func (m *mockRepo) DeleteTargetServer(uuid.UUID) error                       { return nil }
func (m *mockRepo) ListTargetServers() ([]schema.TargetServer, error)         { return nil, nil }
func (m *mockRepo) CreateRoute(schema.Route) error                            { return nil }
func (m *mockRepo) GetRoute(uuid.UUID) (schema.Route, error)                  { return schema.Route{}, nil }
func (m *mockRepo) UpdateRoute(schema.Route) error                            { return nil }
func (m *mockRepo) DeleteRoute(uuid.UUID) error                               { return nil }
func (m *mockRepo) ListRoutes() ([]schema.Route, error)                       { return nil, nil }
func (m *mockRepo) GetRouteFromSourcePath(string) (schema.Route, error)       { return schema.Route{}, nil }
func (m *mockRepo) GetRouteFromTargetPath(string) (schema.Route, error)       { return schema.Route{}, nil }
func (m *mockRepo) CreateAuthentication(schema.Authentication) error          { return nil }
func (m *mockRepo) GetAuthentication(uuid.UUID) (schema.Authentication, error) { return schema.Authentication{}, nil }
func (m *mockRepo) GetAuthenticationWithPlainToken(uuid.UUID) (schema.Authentication, error) {
	return schema.Authentication{}, nil
}
func (m *mockRepo) UpdateAuthentication(schema.Authentication) error         { return nil }
func (m *mockRepo) DeleteAuthentication(uuid.UUID) error                     { return nil }
func (m *mockRepo) ListAuthentications() ([]schema.Authentication, error)    { return nil, nil }
func (m *mockRepo) ListSourceAuthsForRoute(uuid.UUID) ([]schema.RouteSourceAuth, error) { return nil, nil }
func (m *mockRepo) SetSourceAuthsForRoute(uuid.UUID, []uuid.UUID) error       { return nil }
func (m *mockRepo) GetTargetAuthForRoute(uuid.UUID) (uuid.UUID, bool, error)  { return uuid.Nil, false, nil }
func (m *mockRepo) SetTargetAuthForRoute(uuid.UUID, *uuid.UUID) error         { return nil }

var _ repo.Repository = (*mockRepo)(nil)

func Test_Service_handler(t *testing.T) {
	sourceUUID := uuid.MustParse("11111111-1111-1111-1111-111111111111")
	routeUUID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	targetUUID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	t.Run("404 when route not found", func(t *testing.T) {
		mock := &mockRepo{
			findRoute: func(_ uuid.UUID, _, _ string) (schema.Route, error) {
				return schema.Route{}, errors.New("no route")
			},
		}
		svc := NewService(mock)
		h := svc.handler(sourceUUID)
		req := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Errorf("status = %d, want 404", rec.Code)
		}
	})

	t.Run("502 when target server not found", func(t *testing.T) {
		mock := &mockRepo{
			findRoute: func(_ uuid.UUID, _, _ string) (schema.Route, error) {
				return schema.Route{
					RouteUUID:        routeUUID,
					TargetServerUUID: targetUUID,
					Method:           "GET",
					SourcePath:       "/foo",
					TargetPath:       "/bar",
				}, nil
			},
			getTarget: func(u uuid.UUID) (schema.TargetServer, error) {
				return schema.TargetServer{}, errors.New("target not found")
			},
		}
		svc := NewService(mock)
		h := svc.handler(sourceUUID)
		req := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want 502", rec.Code)
		}
		if body := rec.Body.String(); body != "target server not found\n" {
			t.Errorf("body = %q", body)
		}
	})

	t.Run("proxies to backend and returns response", func(t *testing.T) {
		backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/api/bar" {
				t.Errorf("backend path = %q, want /api/bar", r.URL.Path)
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("ok"))
		}))
		defer backend.Close()
		backendURL, _ := url.Parse(backend.URL)
		mock := &mockRepo{
			findRoute: func(_ uuid.UUID, _, _ string) (schema.Route, error) {
				return schema.Route{
					RouteUUID:        routeUUID,
					TargetServerUUID: targetUUID,
					Method:           "GET",
					SourcePath:       "/foo",
					TargetPath:       "/bar",
				}, nil
			},
			getTarget: func(u uuid.UUID) (schema.TargetServer, error) {
				return schema.TargetServer{
					TargetServerUUID: targetUUID,
					Protocol:         backendURL.Scheme,
					Host:             backendURL.Host,
					Port:             0,
					BasePath:         "/api",
				}, nil
			},
		}
		svc := NewService(mock)
		h := svc.handler(sourceUUID)
		req := httptest.NewRequest(http.MethodGet, "http://localhost/foo", nil)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want 200", rec.Code)
		}
		if rec.Body.String() != "ok" {
			t.Errorf("body = %q", rec.Body.String())
		}
	})
}

func Test_NewService(t *testing.T) {
	mock := &mockRepo{}
	svc := NewService(mock)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.repo != mock {
		t.Error("repo not set")
	}
}

func Test_Service_Run_no_sources_waits_for_ctx(t *testing.T) {
	mock := &mockRepo{
		listSources: func() ([]schema.SourceServer, error) { return nil, nil },
	}
	svc := NewService(mock)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- svc.Run(ctx) }()
	cancel()
	err := <-done
	if err != nil {
		t.Errorf("Run() = %v", err)
	}
}
