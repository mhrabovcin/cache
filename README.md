# Auto refreshing cache

A golang package that provides auto refreshing cache interface for various
data sources.

To use library:

```go
import "github.com/mhrabovcin/cache/pkg/cache"
```

## Concepts

A library work around a concept of a global `Cache` that consists of multiple
`Source`s. Each `Source` is independent cache that can be configured with
different source and refresh frequency.

## Interface

The library provides basic interface for accessing cached data:

```go
c := cache.New(source1, source2, ...)
item, err := c.Get("source_name", "key")
item.Value()
```

Each source is independent and can be configured with own data source
and refresh frequency. A general `Source` usage:

```go
source1 := cache.NewSource(
    "source_name",
    cache.WithDefaultData(map[string]string{
        "intial_key": "inital_value",
    }),
)
// Required to stop automatic fetching goroutine
defer source1.Stop()

item, err := source1.Get("cache_key")
item.Value()
```

Source always returns a `Item` that includes `Metadata`. If cache refresh
fails the cache source will serve stale data. To check when data was refreshed
use `Item` methods:

```go
item, err := source1.Get("cache_key")

// When the item was last refreshed
item.LastRefreshed()

// If the item is stale
item.IsStale()

// When the chace will be refreshed again
item.NextRefresh()
```

### Data sources

Redis and Database sources are supported.

```go
cache.NewRedisSource(...)
cache.NewDbSource(...)
cache.NewStaticSource(...)
```

### Custom data sources

It is possible to provide custom data fetcher to general cache `Source`. A data
fetcher must implement `FetchFunc` interface.

```go
func myFetcher() (map[string]string, error) {
    return map[string]string{}, nil
}

customSource := cache.NewSource(
    "custom_source",
    // Refresh every 1 hour
    cache.WithFetchFunc(myFetcher, 1*time.Hour),
)
```

## Development

A library was built with Go `1.12` version and uses go modules.

To run tests:

```
make coverage
```

To run benchmark

```
make bench
```