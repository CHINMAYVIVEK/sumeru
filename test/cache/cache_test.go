package cache_test

import (
	"testing"
	"time"

	"sumeru/core/cache"
)

func TestGetSetDelete(t *testing.T) {
	cache.Clear()
	t.Cleanup(cache.Clear)

	if _, ok := cache.Get("k"); ok {
		t.Fatal("expected miss")
	}
	cache.Set("k", 42, 0)
	v, ok := cache.Get("k")
	if !ok || v.(int) != 42 {
		t.Fatalf("got %v %v", v, ok)
	}
	cache.Delete("k")
	if _, ok := cache.Get("k"); ok {
		t.Fatal("expected miss after delete")
	}
}

func TestTTLExpiry(t *testing.T) {
	cache.Clear()
	t.Cleanup(cache.Clear)

	cache.Set("exp", "x", 20*time.Millisecond)
	if _, ok := cache.Get("exp"); !ok {
		t.Fatal("expected hit before TTL")
	}
	time.Sleep(40 * time.Millisecond)
	if _, ok := cache.Get("exp"); ok {
		t.Fatal("expected miss after TTL")
	}
}

func TestClear(t *testing.T) {
	cache.Clear()
	cache.Set("a", 1, 0)
	cache.Set("b", 2, 0)
	cache.Clear()
	if _, ok := cache.Get("a"); ok {
		t.Fatal("clear should empty store")
	}
}
