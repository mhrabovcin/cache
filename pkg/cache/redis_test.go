package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/mhrabovcin/cache/pkg/cache"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"github.com/docker/go-connections/nat"
	"github.com/go-redis/redis"
)

const RedisImageVersion = "5.0-alpine"

func startRedis() (func(), error) {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	image := fmt.Sprintf("redis:%s", RedisImageVersion)

	ctx := context.Background()

	_, err = cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}

	containerConfig := &container.Config{
		Image: image,
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: map[nat.Port][]nat.PortBinding{
			"6379/tcp": []nat.PortBinding{nat.PortBinding{HostPort: "6379"}},
		},
	}
	resp, err := cli.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		"cache-test-redis",
	)
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	return func() {
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
	}, nil
}

func TestRedisSource(t *testing.T) {
	cleanup, err := startRedis()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer cleanup()

	redisOpts := &redis.Options{
		Addr: "localhost:6379",
	}
	cli := redis.NewClient(redisOpts)
	if _, err := cli.Ping().Result(); err != nil {
		t.Fatal(err)
	}

	hashKey := "test"

	_, err = cli.HMSet(hashKey, map[string]interface{}{
		"key": "value",
	}).Result()
	if err != nil {
		t.Fatal("failed to hset", hashKey, "", err)
	}

	s := cache.NewRedisSource(
		"test",
		hashKey,
		redisOpts,
		100*time.Millisecond,
	)
	<-time.After(200 * time.Millisecond)

	item, err := s.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatal("no item returned")
	}
	if item.Value() != "value" {
		t.Fatal("wrong value was returned for `key`")
	}

	cli.HSet(hashKey, "key", "value2")
	cli.HSet(hashKey, "new-key", "new-value")

	<-time.After(200 * time.Millisecond)

	item, err = s.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatal("no item returned")
	}
	if item.Value() != "value2" {
		t.Fatal("wrong value was returned for `key`")
	}

	item, err = s.Get("new-key")
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatal("no item returned")
	}
	if item.Value() != "new-value" {
		t.Fatal("wrong value was returned for `new-key`")
	}

	s.Stop()
}
