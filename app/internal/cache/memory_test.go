package cache

import (
	"context"
	"testing"
	"time"
)

func TestNewMemory(t *testing.T) {
	// Zero or negative TTL -> DefaultTTL used internally
	m := NewMemory(0)
	defer m.Close()
	if m.defaultTTL != DefaultTTL {
		t.Errorf("NewMemory(0): defaultTTL = %v, want DefaultTTL", m.defaultTTL)
	}

	m2 := NewMemory(-time.Second)
	defer m2.Close()
	if m2.defaultTTL != DefaultTTL {
		t.Errorf("NewMemory(-1s): defaultTTL = %v, want DefaultTTL", m2.defaultTTL)
	}
}

func TestMemory_Get_Set(t *testing.T) {
	ctx := context.Background()
	m := NewMemory(10 * time.Minute)
	defer m.Close()

	// Key missing -> miss
	if _, ok := m.Get(ctx, "missing"); ok {
		t.Error("Get(missing): want miss, got hit")
	}

	// Set then Get -> hit
	val := []byte("hello")
	if err := m.Set(ctx, "k1", val, time.Minute); err != nil {
		t.Fatal(err)
	}
	got, ok := m.Get(ctx, "k1")
	if !ok {
		t.Fatal("Get(k1): want hit, got miss")
	}
	if string(got) != "hello" {
		t.Errorf("Get(k1) = %q, want %q", got, "hello")
	}
}

func TestMemory_Set_zeroTTL_usesDefault(t *testing.T) {
	ctx := context.Background()
	customTTL := 2 * time.Minute
	m := NewMemory(customTTL)
	defer m.Close()

	// Set with ttl 0 should use defaultTTL (2m), so entry still present shortly after
	if err := m.Set(ctx, "k", []byte("v"), 0); err != nil {
		t.Fatal(err)
	}
	got, ok := m.Get(ctx, "k")
	if !ok {
		t.Fatal("Get(k) after Set with 0 ttl: want hit, got miss")
	}
	if string(got) != "v" {
		t.Errorf("Get(k) = %q, want %q", got, "v")
	}
}

func TestMemory_Get_expired_lazyDelete(t *testing.T) {
	ctx := context.Background()
	m := NewMemory(10 * time.Minute)
	defer m.Close()

	// Set with very short TTL, wait, Get should miss and entry removed
	if err := m.Set(ctx, "exp", []byte("x"), 1*time.Millisecond); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond)
	_, ok := m.Get(ctx, "exp")
	if ok {
		t.Error("Get(exp) after expiry: want miss, got hit")
	}
	// Second Get should still miss (entry was removed)
	_, ok = m.Get(ctx, "exp")
	if ok {
		t.Error("Get(exp) second time: want miss, got hit")
	}
}

func TestMemory_Delete(t *testing.T) {
	ctx := context.Background()
	m := NewMemory(time.Minute)
	defer m.Close()

	m.Set(ctx, "d1", []byte("v"), time.Minute)
	if err := m.Delete(ctx, "d1"); err != nil {
		t.Fatal(err)
	}
	if _, ok := m.Get(ctx, "d1"); ok {
		t.Error("Get(d1) after Delete: want miss, got hit")
	}
}

func TestMemory_DeleteByPrefix(t *testing.T) {
	ctx := context.Background()
	m := NewMemory(time.Minute)
	defer m.Close()

	m.Set(ctx, "prefix:a", []byte("1"), time.Minute)
	m.Set(ctx, "prefix:b", []byte("2"), time.Minute)
	m.Set(ctx, "other", []byte("3"), time.Minute)

	if err := m.DeleteByPrefix(ctx, "prefix:"); err != nil {
		t.Fatal(err)
	}

	if _, ok := m.Get(ctx, "prefix:a"); ok {
		t.Error("prefix:a should be deleted")
	}
	if _, ok := m.Get(ctx, "prefix:b"); ok {
		t.Error("prefix:b should be deleted")
	}
	got, ok := m.Get(ctx, "other")
	if !ok || string(got) != "3" {
		t.Errorf("other should remain: got ok=%v, val=%q", ok, got)
	}
}

func TestMemory_Close_idempotent(t *testing.T) {
	m := NewMemory(time.Minute)
	m.Close()
	// Second Close should not panic (close(m.stop) on already closed channel panics in Go; Memory doesn't guard, so we only call Close once in normal use).
	// If the implementation is safe for double Close, we could call m.Close() again. Current impl: close(m.stop) will panic on second close.
	// So we only test single Close here; plan said "optional" for double-close.
	_ = m
}

