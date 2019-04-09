package cache

import "time"

var (
	// NeverRefresh is used when a source shouldn't be refreshed
	NeverRefresh time.Time = time.Time{}
)

// Metadata represents an information about cache state
type Metadata interface {
	LastRefreshed() time.Time
	// Use NextRefresh().IsZero() to check if the item will be refreshed at
	// some point in the future.
	NextRefresh() time.Time
	IsStale() bool
}

type MetadataImpl struct {
	lastRefresh time.Time
	nextRefresh time.Time
}

func (m MetadataImpl) LastRefreshed() time.Time {
	return m.lastRefresh
}

func (m MetadataImpl) NextRefresh() time.Time {
	return m.nextRefresh
}

func (m MetadataImpl) IsStale() bool {
	if m.nextRefresh.IsZero() {
		return false
	}

	if m.nextRefresh.After(time.Now()) {
		return true
	}

	return false
}

func NewMetadata(lastRefresh time.Time, nextRefresh time.Time) MetadataImpl {
	return MetadataImpl{
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
	MetadataImpl
	value string
}

func (i *item) Value() string {
	return i.value
}

func NewItem(
	value string,
	lastRefreshed time.Time,
	nextRefresh time.Time,
) Item {
	return &item{
		value:        value,
		MetadataImpl: NewMetadata(lastRefreshed, nextRefresh),
	}
}
