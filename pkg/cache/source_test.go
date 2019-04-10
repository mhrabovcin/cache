package cache_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/mhrabovcin/cache/pkg/cache"
)

func TestStaticSource(t *testing.T) {
	lastRefresh := time.Now().Add(-1 * time.Minute)
	s := cache.NewStaticSource(
		map[string]string{
			"key":  "value",
			"key2": "value2",
		},
		lastRefresh,
	)

	item, err := s.Get("non-existing")
	if item != nil {
		t.Fatal("item should be nil for non-existing cache item")
	}

	if err != cache.ErrKeyNotFound {
		t.Fatal("err should be key not found for missing cache item")
	}

	item, err = s.Get("key")
	if err != nil {
		t.Fatal("err should be nil for existing cache item")
	}

	if item.Value() != "value" {
		t.Fatal("value for accessed `key` is incorrect")
	}

	if !item.LastRefreshed().Equal(lastRefresh) {
		t.Fatal("last refresh time for item is incorrect")
	}

	if !item.NextRefresh().Equal(cache.Never) {
		t.Fatal("Static cache shouldnt get next refresh time")
	}
}

func BenchmarkStaticSourceGet(b *testing.B) {
	lastRefresh := time.Now().Add(-1 * time.Minute)
	s := cache.NewStaticSource(
		map[string]string{
			"key":  "value",
			"key2": "value2",
		},
		lastRefresh,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Get("key")
	}
}

func BenchmarkStaticSourceGetParallel(b *testing.B) {
	lastRefresh := time.Now().Add(-1 * time.Minute)
	s := cache.NewStaticSource(
		map[string]string{
			"key":  "value",
			"key2": "value2",
		},
		lastRefresh,
	)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			s.Get("key")
		}
	})
}

func TestRefresh(t *testing.T) {
	refreshRan := false
	fetchFunc := func() (map[string]string, error) {
		refreshRan = true
		m := map[string]string{
			"key": "refreshed",
		}
		return m, nil
	}

	s := cache.NewSource(
		cache.WithDefaultData(map[string]string{
			"key": "default",
		}),
		cache.WithFetchFunc(fetchFunc, 15*time.Millisecond),
	)

	item, _ := s.Get("key")
	if item.Value() != "default" {
		t.Fatal("default value should be returned")
	}

	<-time.After(20 * time.Millisecond)

	item, _ = s.Get("key")
	if item.Value() != "refreshed" {
		t.Fatal("cached value should be refreshed")
	}

	if item.LastRefreshed().Equal(cache.Never) {
		t.Fatal("refresh time should be recoded")
	}

	if !refreshRan {
		t.Fatal("refresh function should have been triggered")
	}

	s.Stop()
}

func TestRefreshError(t *testing.T) {
	refreshCount := 0
	fetchFunc := func() (map[string]string, error) {
		refreshCount += 1
		return nil, fmt.Errorf("error")
	}

	s := cache.NewSource(
		cache.WithFetchFunc(fetchFunc, 10*time.Millisecond),
		cache.WithRetryWait(10*time.Millisecond),
	)
	defer s.Stop()

	<-time.After(5 * time.Millisecond)

	if refreshCount != 3 {
		t.Fatal("refresh should be retried at least 3 time. Actual retries: ", refreshCount)
	}
}
