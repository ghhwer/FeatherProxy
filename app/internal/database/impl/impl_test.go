package impl

import (
	"testing"
	"time"

	"FeatherProxy/app/internal/database/cache"
	"FeatherProxy/app/internal/database/objects"
	"FeatherProxy/app/internal/database/repo"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func migrateDB(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(
		&objects.SourceServer{},
		&objects.ServerOptions{},
		&objects.ACLOptions{},
		&objects.TargetServer{},
		&objects.Route{},
		&objects.Authentication{},
		&objects.RouteSourceAuth{},
		&objects.RouteTargetAuth{},
	); err != nil {
		t.Fatal(err)
	}
}

func TestProtocolsCompatible(t *testing.T) {
	tests := []struct {
		source, target string
		want           bool
	}{
		{"http", "http", true},
		{"https", "https", true},
		{"http", "https", true},
		{"https", "http", true},
		{"grpc", "grpc", true},
		{"http", "grpc", false},
		{"grpc", "https", false},
	}
	for _, tt := range tests {
		got := protocolsCompatible(tt.source, tt.target)
		if got != tt.want {
			t.Errorf("protocolsCompatible(%q, %q) = %v, want %v", tt.source, tt.target, got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migrateDB(t, db)
	r := New(db)
	if r == nil {
		t.Fatal("New returned nil")
	}
	// Verify it implements Repository by calling a method
	list, err := r.ListSourceServers()
	if err != nil {
		t.Errorf("ListSourceServers: %v", err)
	}
	if list == nil {
		t.Error("ListSourceServers returned nil slice")
	}
}

func TestNewWithCache(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migrateDB(t, db)
	mem := cache.NewMemory(5 * time.Minute)
	defer mem.Close()
	r := NewWithCache(db, mem, time.Minute)
	if r == nil {
		t.Fatal("NewWithCache returned nil")
	}
	var _ repo.Repository = r
	list, err := r.ListSourceServers()
	if err != nil {
		t.Errorf("ListSourceServers: %v", err)
	}
	if list == nil {
		t.Error("ListSourceServers returned nil slice")
	}
}

func TestNewWithCache_nilCache_usesNoOp(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	migrateDB(t, db)
	r := NewWithCache(db, nil, 0)
	if r == nil {
		t.Fatal("NewWithCache(nil cache) returned nil")
	}
	list, err := r.ListSourceServers()
	if err != nil {
		t.Errorf("ListSourceServers: %v", err)
	}
	if list != nil && len(list) != 0 {
		t.Logf("list length = %d", len(list))
	}
}
