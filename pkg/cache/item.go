package cache

import "time"

var (
	// Never represents a zero time that can be used for LastRefreshed
	// and NextRefresh values.
	Never time.Time = time.Time{}
)

// Metadata represents an information about cache state
type Metadata interface {
	LastRefreshed() time.Time

	// Use NextRefresh().IsZero() to check if the item will be refreshed at
	// some point in the future.
	NextRefresh() time.Time

	// IsStale method returns `true` if the record returend by the cache
	// should have been refreshed but wasn't in proper time.
	IsStale() bool
}

type metadata struct {
	lastRefresh time.Time
	nextRefresh time.Time
}

func (m metadata) LastRefreshed() time.Time {
	return m.lastRefresh
}

func (m metadata) NextRefresh() time.Time {
	return m.nextRefresh
}

func (m metadata) IsStale() bool {
	if m.nextRefresh.Equal(Never) {
		return false
	}

	// If the current time is after the time when the item should be refreshed
	// report a stale item.
	if time.Now().After(m.nextRefresh) {
		return true
	}

	return false
}

// NewMetadata creates metadata instance
func NewMetadata(lastRefresh time.Time, nextRefresh time.Time) metadata {
	return metadata{
		lastRefresh: lastRefresh,
		nextRefresh: nextRefresh,
	}
}

// Item represents a value fetched from cache with additional metadata about
// the item.
type Item interface {
	Metadata

	Value() string
}

type item struct {
	metadata
	value string
}

func (i *item) Value() string {
	return i.value
}

// NewItem creates a cache item with value and metadata
func NewItem(
	value string,
	lastRefreshed time.Time,
	nextRefresh time.Time,
) Item {
	return &item{
		value:    value,
		metadata: NewMetadata(lastRefreshed, nextRefresh),
	}
}
