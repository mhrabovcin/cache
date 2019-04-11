package cache

import (
	"time"

	"github.com/go-redis/redis"
)

// NewRedisSource initializes a cache source that fetches data from redis
// provided hashKey using HGETALL command
// See: https://redis.io/commands/hgetall
func NewRedisSource(
	name string,
	hashKey string,
	redisOpts *redis.Options,
	frequency time.Duration,
	opts ...Option,
) StoppableSource {
	opts = append(opts, WithFetchFunc(
		redisFetchFunc(hashKey, redisOpts),
		frequency,
	))
	return NewSource(
		name,
		opts...,
	)
}

func redisFetchFunc(hashKey string, opts *redis.Options) func() (map[string]string, error) {
	c := redis.NewClient(opts)
	return func() (map[string]string, error) {
		return c.HGetAll(hashKey).Result()
	}
}
