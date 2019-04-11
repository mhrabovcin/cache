package cache

import (
	"fmt"
	"sync"
	"time"
)

// Source represents a single source of cached data from a distinct data source.
type Source interface {
	// Name of the source
	Name() string

	// Get either returns an item form the cache or an error
	Get(key string) (value Item, err error)
}

// StoppableSource is a cache source that automatically fetches data based on
// configured schedule and can be stopped.
type StoppableSource interface {
	Source

	// Stop is used to clean up gorotines that try to fetch new data
	Stop()
}

// FetchFunc is a function that should be used by dynamic source to refresh
// data.
type FetchFunc func() (map[string]string, error)

// Option is a function that sets an option for a source
type Option func(*Options)

// Options for configuring source of cached data. The options shouldn't be
// used directly but through With... functions.
type Options struct {
	DefaultData      map[string]string
	LastRefreshed    time.Time
	FetchFunc        FetchFunc
	RefreshFrequency time.Duration
	RetryWait        time.Duration
}

type source struct {
	metadata

	name string

	data map[string]string
	lock sync.RWMutex

	fetchFunc        FetchFunc
	refreshFrequency time.Duration
	retryWait        time.Duration
	stopCh           chan struct{}
	stoppedCh        chan struct{}
}

func (s *source) start() {
	defer close(s.stoppedCh)

	// Start automated data refresh only if fetch function is provided
	if s.fetchFunc == nil {
		return
	}

	// Default data hasn't been provided, use initial refresh
	if len(s.data) == 0 {
		fmt.Println("No data provided, initial fetch")
		s.refresh()
	}

	select {
	case <-s.stopCh:
		return
	case <-time.After(s.refreshFrequency):
		fmt.Println("Refreshing data")
		s.refresh()
	}
}

func (s *source) refresh() {
	var data map[string]string
	var err error
	var refreshTime time.Time

	// Super simple retry that should be converted to a configurable
	// retry with backoff?
	for i := 0; i < 3; i++ {
		data, err = s.fetchFunc()
		refreshTime = time.Now()
		if err != nil {
			// Log error
			// s.logger.Error()
			fmt.Println("Failed to fetch data source. Error:", err)

			// We've reached end of retry
			if i == 2 {
				return
			}

			<-time.After(s.retryWait)
			continue
		}
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.data = data
	s.lastRefresh = refreshTime
	s.nextRefresh = refreshTime.Add(s.refreshFrequency)
}

func (s *source) Stop() {
	close(s.stopCh)
	<-s.stoppedCh
}

func (s *source) Name() string {
	return s.name
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
func NewSource(name string, opts ...Option) StoppableSource {
	o := &Options{
		DefaultData:   map[string]string{},
		LastRefreshed: Never,
		RetryWait:     1 * time.Second,
	}

	for _, opt := range opts {
		opt(o)
	}

	s := &source{
		metadata: metadata{
			lastRefresh: o.LastRefreshed,
		},
		name:             name,
		data:             o.DefaultData,
		fetchFunc:        o.FetchFunc,
		refreshFrequency: o.RefreshFrequency,
		stopCh:           make(chan struct{}),
		stoppedCh:        make(chan struct{}),
	}

	go s.start()
	return s
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

// WithFetchFunc sets a refresh function and frequency in which should be
// function invoked.
func WithFetchFunc(f FetchFunc, freq time.Duration) Option {
	return func(o *Options) {
		o.FetchFunc = f
		o.RefreshFrequency = freq
	}
}

// WithRetryWait overrides default 1 second retry wait period when fetch
// function returns an error
func WithRetryWait(d time.Duration) Option {
	return func(o *Options) {
		o.RetryWait = d
	}
}

// NewStaticSource returns a cache source that never refresheshes and always
// serves static data
func NewStaticSource(
	name string,
	data map[string]string,
	lastRefresh time.Time,
) Source {
	return NewSource(
		name,
		WithDefaultData(data),
		WithLastRefreshTime(lastRefresh),
	)
}
