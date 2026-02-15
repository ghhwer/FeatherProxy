//go:build integration

package impl

import (
	"testing"
	"time"

	"FeatherProxy/app/internal/database/cache"
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/schema"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRepositoryIntegration_CRUD_and_FindRoute(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&objects.SourceServer{},
		&objects.ServerOptions{},
		&objects.TargetServer{},
		&objects.Route{},
		&objects.Authentication{},
		&objects.RouteSourceAuth{},
		&objects.RouteTargetAuth{},
	); err != nil {
		t.Fatal(err)
	}

	mem := cache.NewMemory(5 * time.Minute)
	defer mem.Close()
	r := NewWithCache(db, mem, time.Minute)

	sourceID := uuid.New()
	targetID := uuid.New()
	routeID := uuid.New()

	if err := r.CreateSourceServer(schema.SourceServer{
		SourceServerUUID: sourceID,
		Name:             "source1",
		Protocol:         "http",
		Host:             "localhost",
		Port:             8080,
	}); err != nil {
		t.Fatalf("CreateSourceServer: %v", err)
	}
	if err := r.CreateTargetServer(schema.TargetServer{
		TargetServerUUID: targetID,
		Name:             "target1",
		Protocol:         "http",
		Host:             "backend",
		Port:             9090,
		BasePath:         "/api",
	}); err != nil {
		t.Fatalf("CreateTargetServer: %v", err)
	}
	if err := r.CreateRoute(schema.Route{
		RouteUUID:        routeID,
		SourceServerUUID: sourceID,
		TargetServerUUID: targetID,
		Method:           "GET",
		SourcePath:       "/foo",
		TargetPath:       "/bar",
	}); err != nil {
		t.Fatalf("CreateRoute: %v", err)
	}

	route, err := r.FindRouteBySourceMethodPath(sourceID, "GET", "/foo")
	if err != nil {
		t.Fatalf("FindRouteBySourceMethodPath: %v", err)
	}
	if route.RouteUUID != routeID || route.TargetPath != "/bar" {
		t.Errorf("FindRouteBySourceMethodPath: got %+v", route)
	}

	target, err := r.GetTargetServer(targetID)
	if err != nil {
		t.Fatalf("GetTargetServer: %v", err)
	}
	if target.TargetServerUUID != targetID || target.Host != "backend" {
		t.Errorf("GetTargetServer: got %+v", target)
	}

	// Second call exercises cache for GetTargetServer
	target2, err := r.GetTargetServer(targetID)
	if err != nil {
		t.Fatalf("GetTargetServer (cached): %v", err)
	}
	if target2.TargetServerUUID != targetID {
		t.Errorf("GetTargetServer (cached): got %+v", target2)
	}
}
