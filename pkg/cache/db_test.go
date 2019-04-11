package cache_test

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/mhrabovcin/cache/pkg/cache"

	docker "docker.io/go-docker"
	"docker.io/go-docker/api/types"
	"docker.io/go-docker/api/types/container"
	"github.com/docker/go-connections/nat"
)

const PgImageVersion = "11.2-alpine"

func startPg() (func(), error) {
	cli, err := docker.NewEnvClient()
	if err != nil {
		return nil, err
	}

	image := fmt.Sprintf("postgres:%s", PgImageVersion)

	ctx := context.Background()

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return nil, err
	}
	io.Copy(os.Stdout, reader)

	containerConfig := &container.Config{
		Image: image,
	}
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		PortBindings: map[nat.Port][]nat.PortBinding{
			"5432/tcp": []nat.PortBinding{nat.PortBinding{HostPort: "5432"}},
		},
	}
	resp, err := cli.ContainerCreate(
		ctx,
		containerConfig,
		hostConfig,
		nil,
		"cache-test-pq",
	)
	if err != nil {
		return nil, err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	return func() {
		cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force: true})
	}, nil
}

func TestDbSource(t *testing.T) {
	cleanup, err := startPg()
	if err != nil {
		t.Fatal(err)
		return
	}
	defer cleanup()

	<-time.After(10 * time.Second)

	connStr := "postgres://postgres:@localhost/?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal("fialed to connect to pg:", err)
	}
	if err := db.Ping(); err != nil {
		t.Fatal(err)
	}

	_, err = db.Exec("CREATE DATABASE cache_test")
	if err != nil {
		t.Fatal("failed to create test database:", err)
	}
	db.Close()

	connStr = "postgres://postgres:@localhost/cache_test?sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		t.Fatal("fialed to connect to pg:", err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE cache (key text, value text)")
	if err != nil {
		t.Fatal("failed to create test table:", err)
	}

	_, err = db.Exec("INSERT INTO cache VALUES ($1, $2)", "key", "value")
	if err != nil {
		t.Fatal("failed to insert cache entry:", err)
	}

	s := cache.NewDbSource(
		"db_cache",
		"postgres",
		connStr,
		&cache.DbQuery{
			Query: "SELECT key, value FROM cache",
		},
		100*time.Millisecond,
	)
	defer s.Stop()

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

	// Refresh data in DB
	if _, err := db.Exec("DELETE FROM cache"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO cache VALUES ($1, $2)", "key", "value-updated"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec("INSERT INTO cache VALUES ($1, $2)", "new-key", "new-value"); err != nil {
		t.Fatal(err)
	}

	<-time.After(200 * time.Millisecond)

	item, err = s.Get("key")
	if err != nil {
		t.Fatal(err)
	}
	if item == nil {
		t.Fatal("no item returned")
	}
	if item.Value() != "value-updated" {
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
}
