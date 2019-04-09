package cache

import "time"

type staticSource struct {
	data          map[string]string
	lastRefreshed time.Time
	nextRefresh   time.Time
}

func (s *staticSource) Get(key string) (value Item, err error) {
	v, ok := s.data[key]

	if !ok {
		return nil, ErrKeyNotFound
	}

	return NewItem(v, s.lastRefreshed, s.nextRefresh), nil
}

// Options for configuring static source of cached data
type Options struct {
	Data          map[string]string
	LastRefreshed time.Time
	NextRefresh   time.Time
}

// NewStaticSource creates a static source of cached data that never changes
// over the time.
func NewStaticSource(setters ...StaticSourceOpt) Source {
	opts := &Options{
		LastRefreshed: time.Now(),
		NextRefresh:   NeverRefresh,
	}

	for _, setter := range setters {
		setter(opts)
	}

	return &staticSource{
		data:          opts.Data,
		lastRefreshed: opts.LastRefreshed,
		nextRefresh:   opts.NextRefresh,
	}
}

// StaticSourceOpt is an option that modifies default Options struct
type StaticSourceOpt func(*Options)

// WithData sets a custom data for a staticSource
func WithData(data map[string]string) StaticSourceOpt {
	return func(o *Options) {
		o.Data = data
	}
}

// WithLastRefreshTime sets last refresh time for staticSource
func WithLastRefreshTime(t time.Time) StaticSourceOpt {
	return func(o *Options) {
		o.LastRefreshed = t
	}
}
