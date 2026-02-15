package database

import (
	"os"
	"sync"
	"testing"
)

func TestNewHandler(t *testing.T) {
	// Restore env after test
	saveDriver := os.Getenv("DB_DRIVER")
	saveDSN := os.Getenv("DB_DSN")
	defer func() {
		if saveDriver != "" {
			os.Setenv("DB_DRIVER", saveDriver)
		} else {
			os.Unsetenv("DB_DRIVER")
		}
		if saveDSN != "" {
			os.Setenv("DB_DSN", saveDSN)
		} else {
			os.Unsetenv("DB_DSN")
		}
	}()

	t.Run("success with sqlite memory", func(t *testing.T) {
		os.Setenv("DB_DRIVER", "sqlite")
		os.Setenv("DB_DSN", "file::memory:?cache=shared")
		h, err := NewHandler()
		if err != nil {
			t.Fatalf("NewHandler: %v", err)
		}
		if h == nil {
			t.Fatal("NewHandler returned nil handler")
		}
		if h.DB() == nil {
			t.Fatal("DB() is nil")
		}
		_ = h.Close()
	})

	t.Run("error when DB_DRIVER missing", func(t *testing.T) {
		os.Unsetenv("DB_DRIVER")
		os.Setenv("DB_DSN", "file::memory:?cache=shared")
		h, err := NewHandler()
		if err == nil {
			_ = h.Close()
			t.Fatal("NewHandler: want error when DB_DRIVER missing")
		}
		if h != nil {
			t.Fatal("NewHandler should return nil handler on error")
		}
	})

	t.Run("error when DB_DSN missing", func(t *testing.T) {
		os.Setenv("DB_DRIVER", "sqlite")
		os.Unsetenv("DB_DSN")
		h, err := NewHandler()
		if err == nil {
			if h != nil {
				_ = h.Close()
			}
			t.Fatal("NewHandler: want error when DB_DSN missing")
		}
	})

	t.Run("error for unsupported driver", func(t *testing.T) {
		os.Setenv("DB_DRIVER", "mysql")
		os.Setenv("DB_DSN", "user:pass@/db")
		h, err := NewHandler()
		if err == nil {
			if h != nil {
				_ = h.Close()
			}
			t.Fatal("NewHandler: want error for unsupported driver")
		}
	})
}

func TestHandler_AutoMigrate(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DSN", "file::memory:?cache=shared")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_DSN")
	}()

	h, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	if err := h.AutoMigrate(); err != nil {
		t.Errorf("AutoMigrate: %v", err)
	}
}

func TestHandler_Close(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DSN", "file::memory:?cache=shared")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_DSN")
	}()

	h, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	if err := h.Close(); err != nil {
		t.Errorf("Close: %v", err)
	}
	// Second Close should not panic; current impl sets db=nil so sqlDB.Close() not called again
	if err := h.Close(); err != nil {
		t.Errorf("Close second time: %v", err)
	}
}

func TestHandler_DB_concurrent(t *testing.T) {
	os.Setenv("DB_DRIVER", "sqlite")
	os.Setenv("DB_DSN", "file::memory:?cache=shared")
	defer func() {
		os.Unsetenv("DB_DRIVER")
		os.Unsetenv("DB_DSN")
	}()

	h, err := NewHandler()
	if err != nil {
		t.Fatal(err)
	}
	defer h.Close()

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.DB()
		}()
	}
	wg.Wait()
}
