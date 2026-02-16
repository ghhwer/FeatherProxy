package proxy

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"FeatherProxy/app/internal/database/schema"
)

func TestJoinHostPort(t *testing.T) {
	if got := joinHostPort("example.com", 0); got != "example.com" {
		t.Fatalf("joinHostPort host only = %q, want %q", got, "example.com")
	}
	if got := joinHostPort("example.com", 8080); got != "example.com:8080" {
		t.Fatalf("joinHostPort with port = %q, want %q", got, "example.com:8080")
	}
}

func TestJoinPath(t *testing.T) {
	tests := []struct {
		base string
		path string
		want string
	}{
		{"", "/foo", "/foo"},
		{"/api", "v1", "/api/v1"},
		{"/api/", "/v1/", "/api/v1/"},
	}
	for _, tt := range tests {
		if got := joinPath(tt.base, tt.path); got != tt.want {
			t.Errorf("joinPath(%q, %q) = %q, want %q", tt.base, tt.path, got, tt.want)
		}
	}
}

func TestBuildTargetURL(t *testing.T) {
	target := &schema.TargetServer{
		Name:     "backend",
		Protocol: "https",
		Host:     "backend.internal",
		Port:     8443,
		BasePath: "/api",
	}
	route := &schema.Route{
		TargetPath: "/v1/resource",
	}
	u := buildTargetURL(target, route, "q=1")
	if got := u.Scheme; got != "https" {
		t.Errorf("Scheme = %q, want %q", got, "https")
	}
	if got := u.Host; got != "backend.internal:8443" {
		t.Errorf("Host = %q, want %q", got, "backend.internal:8443")
	}
	if got := u.Path; got != "/api/v1/resource" {
		t.Errorf("Path = %q, want %q", got, "/api/v1/resource")
	}
	if got := u.RawQuery; got != "q=1" {
		t.Errorf("RawQuery = %q, want %q", got, "q=1")
	}
}

func TestBuildAuthHeaderValue(t *testing.T) {
	if got := buildAuthHeaderValue(nil); got != "" {
		t.Errorf("nil auth = %q, want empty", got)
	}
	if got := buildAuthHeaderValue(&schema.Authentication{}); got != "" {
		t.Errorf("empty token = %q, want empty", got)
	}
	if got := buildAuthHeaderValue(&schema.Authentication{TokenType: "bearer", Token: "abc"}); got != "Bearer abc" {
		t.Errorf("bearer lower = %q, want %q", got, "Bearer abc")
	}
	if got := buildAuthHeaderValue(&schema.Authentication{TokenType: "Bearer", Token: "abc"}); got != "Bearer abc" {
		t.Errorf("Bearer upper = %q, want %q", got, "Bearer abc")
	}
	if got := buildAuthHeaderValue(&schema.Authentication{TokenType: "basic", Token: "xyz"}); got != "xyz" {
		t.Errorf("other type = %q, want %q", got, "xyz")
	}
}

func TestDirectorSetsForwardHeadersAndAuth(t *testing.T) {
	target, _ := url.Parse("https://backend.internal/api/v1")

	// Incoming request with TLS and Authorization header.
	incoming := httptest.NewRequest(http.MethodGet, "https://proxy.internal/foo", nil)
	incoming.RemoteAddr = "192.0.2.1:1234"
	incoming.TLS = &tls.ConnectionState{}
	incoming.Header.Set("Authorization", "Bearer incoming")

	// Outgoing request that the director will modify.
	out, _ := http.NewRequest(http.MethodGet, "/", nil)

	// Case 1: no target auth -> incoming Authorization is forwarded.
	d1 := director(target, incoming, nil)
	d1(out)
	if got := out.Header.Get("Authorization"); got != "Bearer incoming" {
		t.Fatalf("Authorization forwarded = %q, want %q", got, "Bearer incoming")
	}
	if got := out.Header.Get("X-Forwarded-For"); got != "192.0.2.1:1234" {
		t.Errorf("X-Forwarded-For = %q, want %q", got, "192.0.2.1:1234")
	}
	if got := out.Header.Get("X-Forwarded-Proto"); got != "https" {
		t.Errorf("X-Forwarded-Proto = %q, want %q", got, "https")
	}

	// Case 2: target auth present -> override Authorization.
	out2, _ := http.NewRequest(http.MethodGet, "/", nil)
	targetAuth := &schema.Authentication{TokenType: "bearer", Token: "secret"}
	d2 := director(target, incoming, targetAuth)
	d2(out2)
	if got := out2.Header.Get("Authorization"); got != "Bearer secret" {
		t.Fatalf("Authorization with target auth = %q, want %q", got, "Bearer secret")
	}
}
