package proxy

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"FeatherProxy/app/internal/cache"
	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"
	"FeatherProxy/app/internal/stats"

	"github.com/google/uuid"
)

// maxDebugPayloadBytes is the maximum request body bytes to log when debug payload is enabled.
const maxDebugPayloadBytes = 2048 << 10 // 2MB

var debugPayloadFileMu sync.Mutex

// debugPayload returns true when FEATHERPROXY_DEBUG_PAYLOAD is set to "1" or "true".
func debugPayload() bool {
	v := os.Getenv("FEATHERPROXY_DEBUG_PAYLOAD")
	return v == "1" || strings.EqualFold(v, "true")
}

// debugPayloadFilePath returns FEATHERPROXY_DEBUG_PAYLOAD_FILE when debug is enabled. If set, payloads are written there instead of stdout.
func debugPayloadFilePath() string {
	return os.Getenv("FEATHERPROXY_DEBUG_PAYLOAD_FILE")
}

// peekAndRestoreBody reads up to maxDebugPayloadBytes from r.Body, writes it to the debug file (or log), then
// replaces r.Body with a reader that yields the same bytes so the proxy forwards the full body.
func peekAndRestoreBody(r *http.Request) {
	prefix, err := io.ReadAll(io.LimitReader(r.Body, maxDebugPayloadBytes))
	if err != nil {
		log.Printf("proxy/debug: read body prefix: %v", err)
	}
	truncated := len(prefix) == maxDebugPayloadBytes
	r.Body = io.NopCloser(io.MultiReader(bytes.NewReader(prefix), r.Body))
	if len(prefix) == 0 {
		return
	}
	msg := fmt.Sprintf("%s %s %d bytes", r.Method, r.URL.Path, len(prefix))
	if truncated {
		msg += " (truncated)"
	}
	line := time.Now().Format(time.RFC3339) + " " + msg + "\n" + string(prefix) + "\n"
	if path := debugPayloadFilePath(); path != "" {
		debugPayloadFileMu.Lock()
		if f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644); err == nil {
			_, _ = f.WriteString(line)
			_ = f.Close()
		} else {
			log.Printf("proxy/debug: write payload file: %v", err)
		}
		debugPayloadFileMu.Unlock()
	} else {
		if truncated {
			log.Printf("proxy/debug: %s body (first %d bytes, truncated): %s", msg, len(prefix), string(prefix))
		} else {
			log.Printf("proxy/debug: %s body: %s", msg, string(prefix))
		}
	}
}

// Service runs one HTTP listener per source server and proxies matching requests to target servers.
type Service struct {
	repo     database.Repository
	resolver HostnameResolver
	recorder stats.Recorder // optional; when set, proxied requests are recorded for stats
}

// NewService returns a proxy service that uses the given repository for route
// and server lookups. It configures a HostnameResolver for ACL hostname and
// wildcard matching, using the same cache instance and TTL as the database
// cache when available. If recorder is non-nil, successfully proxied requests
// are recorded asynchronously for statistics.
func NewService(repo database.Repository, c cache.Cache, cacheTTL time.Duration, recorder stats.Recorder) *Service {
	return &Service{
		repo:     repo,
		resolver: NewResolver(c, cacheTTL),
		recorder: recorder,
	}
}

// Run starts a listener for each source server and blocks until ctx is cancelled.
// On shutdown, all proxy servers are stopped. If there are no source servers, Run returns when ctx is done.
func (s *Service) Run(ctx context.Context) error {
	sources, err := s.repo.ListSourceServers()
	if err != nil {
		return fmt.Errorf("proxy: list source servers: %w", err)
	}

	if len(sources) == 0 {
		log.Println("proxy: no source servers configured, waiting for shutdown")
		<-ctx.Done()
		return nil
	}

	var wg sync.WaitGroup
	shutdown := make(chan struct{})
	defer func() {
		close(shutdown)
		wg.Wait()
	}()

	for i := range sources {
		source := sources[i]
		addr := joinHostPort(source.Host, source.Port)
		server := &http.Server{
			Addr:    addr,
			Handler: s.handler(source.SourceServerUUID),
		}
		opts, _ := s.repo.GetServerOptions(source.SourceServerUUID)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if source.Protocol == "https" {
				if opts.TLSCertPath == "" || opts.TLSKeyPath == "" {
					log.Printf("proxy: HTTPS source %s (%s) missing TLS cert/key paths, skipping", source.Name, addr)
					return
				}
				log.Printf("proxy: listening on https://%s (%s)", addr, source.Name)
				if err := server.ListenAndServeTLS(opts.TLSCertPath, opts.TLSKeyPath); err != nil && err != http.ErrServerClosed {
					log.Printf("proxy: server %s: %v", addr, err)
				}
			} else {
				log.Printf("proxy: listening on http://%s (%s)", addr, source.Name)
				if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					log.Printf("proxy: server %s: %v", addr, err)
				}
			}
		}()
		go func() {
			<-ctx.Done()
			_ = server.Shutdown(context.Background())
		}()
	}

	<-ctx.Done()
	return nil
}

// responseRecorder wraps http.ResponseWriter to capture status code and duration.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
	start      time.Time
}

func (rw *responseRecorder) WriteHeader(code int) {
	if !rw.written {
		rw.statusCode = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseRecorder) Write(p []byte) (int, error) {
	if !rw.written {
		rw.statusCode = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(p)
}

// clientIPString returns the client IP string for stats, using the same logic as ACL.
func clientIPString(r *http.Request, acl *schema.ACLOptions) string {
	ip := clientIPFromRequest(r, acl)
	if ip != nil {
		return ip.String()
	}
	// Fallback: strip port from RemoteAddr if present
	addr := r.RemoteAddr
	if host, _, err := net.SplitHostPort(addr); err == nil {
		return host
	}
	return addr
}

// handler returns an http.Handler that routes requests for the given source server.
func (s *Service) handler(sourceServerUUID uuid.UUID) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if debugPayload() && r.Body != nil {
			peekAndRestoreBody(r)
		}
		acl, err := s.repo.GetACLOptions(sourceServerUUID)
		if err != nil {
			log.Printf("proxy: get ACL options: %v", err)
		}
		if err == nil && aclDeny(r.Context(), r, &acl, s.resolver) {
			log.Printf("proxy/acl: %s %s denied by ACL", r.Method, r.URL.Path)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		route, err := s.repo.FindRouteBySourceMethodPath(sourceServerUUID, r.Method, r.URL.Path)
		if err != nil {
			log.Printf("proxy/auth: %s %s no route match", r.Method, r.URL.Path)
			http.NotFound(w, r)
			return
		}
		log.Printf("proxy/auth: %s %s route=%s target_server=%s", r.Method, r.URL.Path, route.RouteUUID, route.TargetServerUUID)

		// Enforce source authentication (client auth) if configured for this route.
		authorized, err := s.isSourceAuthorized(r, route.RouteUUID)
		if err != nil {
			log.Printf("proxy/auth: route=%s source auth error: %v", route.RouteUUID, err)
			http.Error(w, "source auth error", http.StatusInternalServerError)
			return
		}
		if !authorized {
			log.Printf("proxy/auth: route=%s source auth denied", route.RouteUUID)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		target, err := s.repo.GetTargetServer(route.TargetServerUUID)
		if err != nil {
			log.Printf("proxy/auth: target server not found: %v", err)
			http.Error(w, "target server not found", http.StatusBadGateway)
			return
		}
		var targetAuth *schema.Authentication
		if auth, ok, err := s.repo.GetTargetAuthenticationWithPlainToken(route.RouteUUID); err == nil && ok {
			targetAuth = &auth
			log.Printf("proxy/auth: route=%s target auth enabled name=%s type=%s", route.RouteUUID, auth.Name, auth.TokenType)
		} else {
			if err != nil {
				log.Printf("proxy/auth: route=%s get target auth error: %v", route.RouteUUID, err)
			} else {
				log.Printf("proxy/auth: route=%s no target auth, forwarding Authorization as-is", route.RouteUUID)
			}
		}
		targetURL := buildTargetURL(&target, &route, r.URL.RawQuery)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Director = director(targetURL, r, targetAuth)

		start := time.Now()
		var rec *responseRecorder
		if s.recorder != nil {
			rec = &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK, start: start}
			w = rec
		}
		proxy.ServeHTTP(w, r)

		if s.recorder != nil && rec != nil {
			dur := time.Since(rec.start).Milliseconds()
			stat := schema.ProxyStat{
				Timestamp:        start,
				SourceServerUUID: sourceServerUUID,
				RouteUUID:        route.RouteUUID,
				TargetServerUUID: route.TargetServerUUID,
				Method:           r.Method,
				Path:             r.URL.Path,
				StatusCode:       intPtr(rec.statusCode),
				DurationMs:       int64Ptr(dur),
				ClientIP:         clientIPString(r, &acl),
			}
			s.recorder.Record(stat)
		}
	})
}

func intPtr(n int) *int       { return &n }
func int64Ptr(n int64) *int64 { return &n }

// director returns a Director that rewrites the outgoing request to the backend URL and sets forwarding headers.
// If targetAuth is set, the outgoing Authorization header is set to that credential (e.g. Bearer token) and the
// incoming Authorization is not forwarded, so the backend sees only the configured credential.
func director(target *url.URL, incoming *http.Request, targetAuth *schema.Authentication) func(*http.Request) {
	return func(out *http.Request) {
		out.URL.Scheme = target.Scheme
		out.URL.Host = target.Host
		out.URL.Path = target.Path
		out.URL.RawQuery = target.RawQuery
		out.Host = target.Host
		// Check if incoming contains X-Forwarded-For or X-Forwarded-Proto, if not, set them.
		if incoming.Header.Get("X-Forwarded-For") == "" {
			out.Header.Set("X-Forwarded-For", incoming.RemoteAddr)
		}
		if incoming.Header.Get("X-Forwarded-Proto") == "" {
			out.Header.Set("X-Forwarded-Proto", "https")
		}
		if incoming.Header.Get("X-Forwarded-Proto") == "http" {
			out.Header.Set("X-Forwarded-Proto", "http")
		}
		if targetAuth != nil && targetAuth.Token != "" {
			switch targetAuth.TokenType {
			case "bearer", "Bearer":
				out.Header.Set("Authorization", "Bearer "+targetAuth.Token)
				log.Printf("proxy/auth: director set Authorization Bearer (len=%d)", len(targetAuth.Token))
			default:
				out.Header.Set("Authorization", targetAuth.Token)
				log.Printf("proxy/auth: director set Authorization raw type=%s", targetAuth.TokenType)
			}
		} else {
			// No target auth: forward incoming Authorization as-is
			if v := incoming.Header.Get("Authorization"); v != "" {
				out.Header.Set("Authorization", v)
				log.Printf("proxy/auth: director forwarded incoming Authorization")
			}
		}
	}
}

// buildTargetURL constructs the backend URL from target server, route, and query string.
func buildTargetURL(target *schema.TargetServer, route *schema.Route, rawQuery string) *url.URL {
	path := joinPath(target.BasePath, route.TargetPath)
	u := &url.URL{
		Scheme:   target.Protocol,
		Host:     joinHostPort(target.Host, target.Port),
		Path:     path,
		RawQuery: rawQuery,
	}
	return u
}

func joinHostPort(host string, port int) string {
	if port == 0 {
		return host
	}
	return fmt.Sprintf("%s:%d", host, port)
}

func joinPath(base, p string) string {
	base = strings.TrimSuffix(base, "/")
	p = strings.TrimPrefix(p, "/")
	if base == "" {
		return "/" + p
	}
	return base + "/" + p
}

// isSourceAuthorized returns true if the incoming request is allowed by the
// route's configured source authentications. If no source authentications are
// configured for the route, it returns true (no auth required).
//
// If one or more source authentications are configured, the incoming
// Authorization header must match at least one of the configured credentials,
// formatted according to its TokenType (e.g. "Bearer <token>" for bearer).
func (s *Service) isSourceAuthorized(r *http.Request, routeUUID uuid.UUID) (bool, error) {
	list, err := s.repo.ListSourceAuthsForRoute(routeUUID)
	if err != nil {
		return false, err
	}
	if len(list) == 0 {
		// No source auth configured for this route.
		return true, nil
	}
	incoming := strings.TrimSpace(r.Header.Get("Authorization"))
	if incoming == "" {
		// Auth required but no credentials provided.
		return false, nil
	}
	for _, mapping := range list {
		auth, err := s.repo.GetAuthenticationWithPlainToken(mapping.AuthenticationUUID)
		if err != nil {
			return false, err
		}
		if expected := buildAuthHeaderValue(&auth); expected != "" && incoming == expected {
			return true, nil
		}
	}
	// No match found among allowed source authentications.
	return false, nil
}

// buildAuthHeaderValue formats an Authentication as an Authorization header
// value, mirroring the behavior used for target auth in director().
func buildAuthHeaderValue(a *schema.Authentication) string {
	if a == nil || a.Token == "" {
		return ""
	}
	switch a.TokenType {
	case "bearer", "Bearer":
		return "Bearer " + a.Token
	default:
		return a.Token
	}
}
