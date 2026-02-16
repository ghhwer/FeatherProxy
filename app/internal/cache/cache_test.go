package cache

import (
	"os"
	"testing"
	"time"
)

func TestFromEnv(t *testing.T) {
	tests := []struct {
		name           string
		strategy       string
		ttl            string
		wantCacheNil   bool
		wantErr        bool
		wantTTLDefault bool // true if we only care that ttl > 0 for memory, or default for none
	}{
		{"empty strategy", "", "", true, false, true},
		{"none", "none", "", true, false, true},
		{"NONE upper", "NONE", "1m", true, false, false},
		{"memory", "memory", "", false, false, true},
		{"memory with ttl", "memory", "5m", false, false, false},
		{"redis stub", "redis", "1h", false, false, false},
		{"unknown", "unknown", "", true, true, false},
		{"invalid strategy", "memcached", "", true, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("CACHING_STRATEGY")
			os.Unsetenv("CACHE_TTL")
			if tt.strategy != "" {
				os.Setenv("CACHING_STRATEGY", tt.strategy)
			}
			if tt.ttl != "" {
				os.Setenv("CACHE_TTL", tt.ttl)
			}
			defer func() {
				os.Unsetenv("CACHING_STRATEGY")
				os.Unsetenv("CACHE_TTL")
			}()

			c, ttl, err := FromEnv()
			if (err != nil) != tt.wantErr {
				t.Errorf("FromEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if (c == nil) != tt.wantCacheNil {
				t.Errorf("FromEnv() cache nil = %v, want nil %v", c == nil, tt.wantCacheNil)
			}
			if c != nil {
				defer c.Close()
			}
			if tt.wantTTLDefault && ttl != DefaultTTL && ttl <= 0 {
				if ttl <= 0 {
					t.Errorf("FromEnv() ttl = %v, want positive or DefaultTTL", ttl)
				}
			}
			if tt.ttl == "5m" && ttl != 5*time.Minute {
				t.Errorf("FromEnv() ttl = %v, want 5m", ttl)
			}
			if tt.ttl == "1h" && ttl != time.Hour {
				t.Errorf("FromEnv() ttl = %v, want 1h", ttl)
			}
		})
	}
}

func TestFromEnv_parseTTL_coverage(t *testing.T) {
	// Cover parseTTL via FromEnv: empty, valid, invalid, zero/negative
	os.Unsetenv("CACHING_STRATEGY")
	os.Unsetenv("CACHE_TTL")
	defer func() {
		os.Unsetenv("CACHING_STRATEGY")
		os.Unsetenv("CACHE_TTL")
	}()

	// Empty TTL -> DefaultTTL
	os.Setenv("CACHING_STRATEGY", "memory")
	_, ttl, err := FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if ttl != DefaultTTL {
		t.Errorf("empty CACHE_TTL: got ttl %v, want DefaultTTL", ttl)
	}

	// Invalid duration -> DefaultTTL
	os.Setenv("CACHE_TTL", "invalid")
	_, ttl, err = FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if ttl != DefaultTTL {
		t.Errorf("invalid CACHE_TTL: got ttl %v, want DefaultTTL", ttl)
	}

	// Zero/negative -> DefaultTTL (parseTTL returns DefaultTTL when d <= 0)
	os.Setenv("CACHE_TTL", "0")
	_, ttl, err = FromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if ttl != DefaultTTL {
		t.Errorf("CACHE_TTL=0: got ttl %v, want DefaultTTL", ttl)
	}
}

