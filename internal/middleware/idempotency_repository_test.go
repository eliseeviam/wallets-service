package middleware_test

import (
	"context"
	"github.com/eliseeviam/wallets-service/internal/middleware"
	"github.com/go-redis/redis/v8"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
	"time"
)

func TestNewInMemIdempotencyKeysRepository(t *testing.T) {

	r := middleware.NewInMemIdempotencyKeysRepository()

	t.Run("setFinished", func(t *testing.T) {
		err := r.SetStatus("finishedKey", middleware.Finished)
		require.NoError(t, err)
	})

	t.Run("getFinished", func(t *testing.T) {
		status, err := r.Status("finishedKey")
		require.NoError(t, err)
		require.Equal(t, middleware.Finished, status)
	})

	t.Run("getUnfinished", func(t *testing.T) {
		status, err := r.Status("unfinishedKey")
		require.NoError(t, err)
		require.Equal(t, middleware.Unknown, status)
	})

}

var (
	redisPort string
)

func TestNewRedisIdempotencyKeysRepository(t *testing.T) {

	r := middleware.NewRedisIdempotencyKeysRepository("localhost:"+redisPort, "")

	require.NoError(t, r.SetStatus("11", middleware.Finished), "seems that redis instance haven't been started")

	t.Run("setFinished", func(t *testing.T) {
		err := r.SetStatus("finishedKey", middleware.Finished)
		require.NoError(t, err)
	})

	t.Run("getFinished", func(t *testing.T) {
		status, err := r.Status("finishedKey")
		require.NoError(t, err)
		require.Equal(t, middleware.Finished, status)
	})

	t.Run("getUnfinished", func(t *testing.T) {
		status, err := r.Status("unfinishedKey")
		require.NoError(t, err)
		require.Equal(t, middleware.Unknown, status)
	})

}

func TestMain(m *testing.M) {

	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "redis",
		Tag:        "5.0.9",
		ExposedPorts: []string{
			"6379",
		},
	})

	if err != nil {
		panic(err)
	}

	if err := pool.Retry(func() error {

		redisPort = resource.GetPort("6379/tcp")
		cli := redis.NewClient(&redis.Options{
			Addr: "localhost:" + redisPort,
		})

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		return cli.Ping(ctx).Err()

	}); err != nil {
		panic(err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		panic(err)
	}

	os.Exit(code)

}
