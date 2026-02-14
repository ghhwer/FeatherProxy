package proxy

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"FeatherProxy/app/internal/database"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
)

// Service runs one HTTP listener per source server and proxies matching requests to target servers.
type Service struct {
	repo database.Repository
}

// NewService returns a proxy service that uses the given repository for route and server lookups.
func NewService(repo database.Repository) *Service {
	return &Service{repo: repo}
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
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Printf("proxy: listening on %s (%s)", addr, source.Name)
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("proxy: server %s: %v", addr, err)
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

// handler returns an http.Handler that routes requests for the given source server.
func (s *Service) handler(sourceServerUUID uuid.UUID) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route, err := s.repo.FindRouteBySourceMethodPath(sourceServerUUID, r.Method, r.URL.Path)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		target, err := s.repo.GetTargetServer(route.TargetServerUUID)
		if err != nil {
			http.Error(w, "target server not found", http.StatusBadGateway)
			return
		}
		targetURL := buildTargetURL(&target, &route, r.URL.RawQuery)
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.Director = director(targetURL, r)
		proxy.ServeHTTP(w, r)
	})
}

// director returns a Director that rewrites the outgoing request to the backend URL and sets forwarding headers.
func director(target *url.URL, incoming *http.Request) func(*http.Request) {
	return func(out *http.Request) {
		out.URL.Scheme = target.Scheme
		out.URL.Host = target.Host
		out.URL.Path = target.Path
		out.URL.RawQuery = target.RawQuery
		out.Host = target.Host
		if incoming.RemoteAddr != "" {
			out.Header.Set("X-Forwarded-For", incoming.RemoteAddr)
		}
		if incoming.TLS != nil {
			out.Header.Set("X-Forwarded-Proto", "https")
		} else {
			out.Header.Set("X-Forwarded-Proto", "http")
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
