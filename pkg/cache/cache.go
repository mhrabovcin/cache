package cache

import (
	"errors"
)

var (
	// ErrSourceNotFound is returned if the source doesn't exists in cache
	ErrSourceNotFound = errors.New("Source not found")

	// ErrKeyNotFound indicates that key wasn't found in the data source
	ErrKeyNotFound = errors.New("Key was not found")
)

// Source represents a single source of cached data from a distinct data source.
type Source interface {
	Get(key string) (value Item, err error)
}

// Cache represents a global cache object that can be used to access a source
// or cached kedy form a source directly
type Cache interface {
	Source(source string) Source
	Get(source string, key string) (value Item, err error)
}

// NewCache creates a new global cache instance with provided sources
func NewCache(sources map[string]Source) Cache {
	return &cacheImpl{
		sources: sources,
	}
}

type cacheImpl struct {
	sources map[string]Source
}

func (c *cacheImpl) Source(source string) Source {
	return c.sources[source]
}

// Get returns a single cache item from a source
func (c *cacheImpl) Get(source string, key string) (value Item, err error) {
	s := c.Source(source)

	if s == nil {
		return nil, ErrSourceNotFound
	}

	return s.Get(key)
}

var _ Cache = (*cacheImpl)(nil)
