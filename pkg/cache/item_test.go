package cache_test

import (
	"testing"
	"time"

	"github.com/mhrabovcin/cache/pkg/cache"
)

func TestMetadata(t *testing.T) {
	refreshed := time.Now()
	nextRefresh := refreshed.Add(time.Duration(5 * time.Second))
	m := cache.NewMetadata(refreshed, nextRefresh)
	if !m.LastRefreshed().Equal(refreshed) {
		t.Fatal("LastRefreshed should be equal to refreshed")
	}

	if !m.NextRefresh().Equal(nextRefresh) {
		t.Fatal("NextRefresh should be equal in 5 seconds")
	}

	if m.IsStale() {
		t.Fatal("IsStale should not be true")
	}
}

func TestMetadataIsStale(t *testing.T) {
	refreshed := time.Now()
	nextRefresh := refreshed.Add(time.Duration(20 * time.Millisecond))
	m := cache.NewMetadata(refreshed, nextRefresh)

	// The record isn't stale as it should have been refreshed in 20 ms
	if m.IsStale() {
		t.Fatal("IsStale should not be true")
	}

	// Make time pass 30ms
	<-time.After(30 * time.Millisecond)

	if !m.IsStale() {
		t.Fatal("Metadata should report stale for old record")
	}
}

func TestMetadataIsStaleNeverRefresh(t *testing.T) {
	m := cache.NewMetadata(time.Now(), cache.NeverRefresh)
	if m.IsStale() {
		t.Fatal("Never refreshed record shouldn't be stale")
	}
}

func TestItem(t *testing.T) {
	item := cache.NewItem("value", time.Now(), cache.NeverRefresh)
	if item.Value() != "value" {
		t.Fatal("Value of the cached item is incorrect")
	}
}
