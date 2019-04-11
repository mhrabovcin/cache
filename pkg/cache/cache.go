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

// Cache represents a global cache object that can be used to access a source
// or cached kedy form a source directly
type Cache interface {
	Source(source string) (Source, error)
	Get(source string, key string) (value Item, err error)
}

// New creates a new global cache instance with provided sources
func New(sources ...Source) Cache {
	s := make(map[string]Source)

	for _, source := range sources {
		if _, ok := s[source.Name()]; ok {
			panic("duplicate source name")
		}
		s[source.Name()] = source
	}

	return &cacheImpl{
		sources: s,
	}
}

type cacheImpl struct {
	sources map[string]Source
}

func (c *cacheImpl) Source(source string) (Source, error) {
	s, ok := c.sources[source]

	if !ok {
		return nil, ErrSourceNotFound
	}

	return s, nil
}

// Get returns a single cache item from a source
func (c *cacheImpl) Get(source string, key string) (value Item, err error) {
	s, err := c.Source(source)

	if err != nil {
		return nil, err
	}

	return s.Get(key)
}

var _ Cache = (*cacheImpl)(nil)
