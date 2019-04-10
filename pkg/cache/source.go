package cache

import (
	"sync"
	"time"
)

// RefreshFunc is a function that should be used by dynamic source to refresh
// data.
type RefreshFunc func() (map[string]string, error)

// Option is a function that sets an option for a source
type Option func(*Options)

// Options for configuring source of cached data. The options shouldn't be
// used directly but through With... functions.
type Options struct {
	DefaultData      map[string]string
	LastRefreshed    time.Time
	RefreshFunc      RefreshFunc
	RefreshFrequency time.Duration
}

type source struct {
	MetadataImpl

	data map[string]string
	lock sync.RWMutex

	refreshFunc      RefreshFunc
	refreshFrequency time.Duration
}

func (s *source) Get(key string) (value Item, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	v, ok := s.data[key]

	if !ok {
		return nil, ErrKeyNotFound
	}

	var nextRefresh time.Time
	if s.refreshFrequency > 0 {
		nextRefresh = s.NextRefresh()
	} else {
		nextRefresh = Never
	}

	return NewItem(v, s.LastRefreshed(), nextRefresh), nil
}

func (s *source) NextRefresh() time.Time {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.LastRefreshed().Add(s.refreshFrequency)
}

// NewSource creates a cache source
func NewSource(opts ...Option) *source {
	o := &Options{
		DefaultData:   map[string]string{},
		LastRefreshed: Never,
	}

	for _, opt := range opts {
		opt(o)
	}

	return &source{
		MetadataImpl: MetadataImpl{
			lastRefresh: o.LastRefreshed,
		},
		data:             o.DefaultData,
		refreshFunc:      o.RefreshFunc,
		refreshFrequency: o.RefreshFrequency,
	}
}

// WithDefaultData provides a default cached data for a source
func WithDefaultData(data map[string]string) Option {
	return func(o *Options) {
		o.DefaultData = data
	}
}

// WithLastRefreshTime sets last refresh time and can be used with
// default data setter
func WithLastRefreshTime(t time.Time) Option {
	return func(o *Options) {
		o.LastRefreshed = t
	}
}

// WithLogger provides a custom logger interface
// func WithLogger()

// WithRefreshFunc sets a refresh function and frequency in which should be
// function invoked.
func WithRefreshFunc(f RefreshFunc, freq time.Duration) Option {
	return func(o *Options) {
		o.RefreshFunc = f
		o.RefreshFrequency = freq
	}
}

// NewStaticSource returns a cache source that never refresheshes and always
// serves static data
func NewStaticSource(data map[string]string, lastRefresh time.Time) Source {
	return NewSource(
		WithDefaultData(data),
		WithLastRefreshTime(lastRefresh),
	)
}
