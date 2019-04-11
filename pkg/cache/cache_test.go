package cache_test

import (
	"testing"
	"time"

	"github.com/mhrabovcin/cache/pkg/cache"
)

func TestCacheSource(t *testing.T) {
	lastRefresh := time.Now().Add(-1 * time.Minute)
	s1 := cache.NewStaticSource(
		"s1",
		map[string]string{"key": "value"},
		lastRefresh,
	)
	s2 := cache.NewStaticSource(
		"s2",
		map[string]string{"key2": "value2"},
		lastRefresh,
	)

	c := cache.New(s1, s2)

	var s cache.Source
	s, err := c.Source("s1")

	if err != nil {
		t.Fatal("error should be nil if source exists")
	}

	if s != s1 {
		t.Fatal("wrong source returend")
	}

	item, err := c.Get("s1", "key")
	if item == nil {
		t.Fatal("item `key` should exist in `s1` source")
	}

	if err != nil {
		t.Fatal("err should be nil for existing item")
	}
}

func TestCacheNoSources(t *testing.T) {
	c := cache.New()
	s, err := c.Source("test")
	if s != nil {
		t.Fatal("source should be nil for non-existing source")
	}

	if err != cache.ErrSourceNotFound {
		t.Fatal("error should be source not found")
	}

	item, err := c.Get("test", "test")
	if item != nil {
		t.Fatal("item should be nil if source doesnt exists")
	}

	if err != cache.ErrSourceNotFound {
		t.Fatal("error should be source not found")
	}
}

func TestCacheDupliciteSourceNamesShouldPanic(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("adding 2 cache sources with same names should panic")
		}
	}()

	lastRefresh := time.Now().Add(-1 * time.Minute)
	s1 := cache.NewStaticSource(
		"name",
		map[string]string{"key": "value"},
		lastRefresh,
	)
	s2 := cache.NewStaticSource(
		"name",
		map[string]string{"key2": "value2"},
		lastRefresh,
	)

	cache.New(s1, s2)
}
