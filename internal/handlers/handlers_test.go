package handlers_test

import (
	"context"
	"errors"
	"github.com/eliseeviam/wallets-service/internal/handlers"
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"os"
	"strconv"
	"sync/atomic"
	"testing"
	"time"
)

func TestHandlerParamsError(t *testing.T) {

	t.Run("params", func(t *testing.T) {
		e := &params.Error{
			Code:    500,
			Message: "internal",
		}
		recorder := httptest.NewRecorder()
		ok := handlers.HandlerError(e, recorder)
		require.True(t, ok)
		require.Equal(t, 500, recorder.Code)
		require.Equal(t, `{"ok":0,"message":"internal"}`, recorder.Body.String())
	})

	t.Run("unspecified", func(t *testing.T) {
		e := errors.New("custom error")
		recorder := httptest.NewRecorder()
		ok := handlers.HandlerError(e, recorder)
		require.True(t, ok)
		require.Equal(t, 500, recorder.Code)
		require.Equal(t, `{"ok":0,"message":"custom error"}`, recorder.Body.String())
	})

	t.Run("nil", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		ok := handlers.HandlerError(nil, recorder)
		require.False(t, ok)
		require.Equal(t, 200, recorder.Code)
		require.Equal(t, ``, recorder.Body.String())
	})
}

var (
	repositoryConfig  *repository.RepositoryConfig
	repositoryRawPool *pgxpool.Pool
)

const (
	PostgreSQLUser     = "user"
	PostgreSQLPassword = "pwd"
	PostgreSQLDatabase = "test_db"
)

var databaseInited int64

func mustInitDatabase(t *testing.T, pool *pgxpool.Pool) {

	if atomic.SwapInt64(&databaseInited, 1) == 1 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `CREATE TABLE "public"."wallets" (
    "name" text NOT NULL,
    "amount" int8 NOT NULL DEFAULT 0 CHECK (amount >= (0)::bigint),
    "create_time" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("name")
);`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `CREATE SEQUENCE transfer_history_id_seq;`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `CREATE TYPE "public"."transfer_direction" AS ENUM ('deposit', 'transfer', 'withdrawal');`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `CREATE TABLE "public"."transfer_history" (
    "id" int4 NOT NULL DEFAULT nextval('transfer_history_id_seq'::regclass),
    "wallet" text NOT NULL,
    "direction" "transfer_direction" NOT NULL,
    "amount" int8 NOT NULL,
	"meta" jsonb,
    "time" timestamp NOT NULL DEFAULT now(),
    PRIMARY KEY ("id")
);`)
	require.NoError(t, err)

	require.NoError(t, tx.Commit(ctx))

}

func checkDatabase(t *testing.T, pool *pgxpool.Pool) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var walletsNumber int
	err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM wallets").Scan(&walletsNumber)
	require.NoError(t, err)
	require.Equal(t, 0, walletsNumber)

	var transferNumber int
	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM transfer_history").Scan(&transferNumber)
	require.NoError(t, err)
	require.Equal(t, 0, walletsNumber)

	var transferCurrentID int
	err = pool.QueryRow(ctx, "SELECT CURRVAL('transfer_history_id_seq');").Scan(&transferCurrentID)
	require.True(t, err != nil || transferCurrentID > 0)
}

func mustClearDatabase(t *testing.T, pool *pgxpool.Pool) {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `TRUNCATE TABLE "wallets";`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `TRUNCATE TABLE "transfer_history";`)
	require.NoError(t, err)

	_, err = tx.Exec(ctx, `ALTER SEQUENCE transfer_history_id_seq RESTART;`)
	require.NoError(t, err)

	require.NoError(t, tx.Commit(ctx))

}

func TestMain(m *testing.M) {

	pool, err := dockertest.NewPool("")
	if err != nil {
		panic(err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "latest", // Не лучший выбор. Используйте "13.3",
		Env: []string{
			"POSTGRES_USER=" + PostgreSQLUser,
			"POSTGRES_PASSWORD=" + PostgreSQLPassword,
			"POSTGRES_DB=" + PostgreSQLDatabase,
		},
		ExposedPorts: []string{
			"5432",
		},
	})

	if err != nil {
		panic(err)
	}

	if err := pool.Retry(func() error {
		var err error

		repositoryConfig = new(repository.RepositoryConfig).
			SetRepositoryType(repository.RepositoryPosgreSQL).
			SetHost("localhost").
			SetPort(func() int {
				port, err := strconv.Atoi(resource.GetPort("5432/tcp"))
				if err != nil {
					panic(err)
				}
				return port
			}()).
			SetDbName(PostgreSQLDatabase).
			SetUser(PostgreSQLUser).
			SetPassword(PostgreSQLPassword)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		repositoryRawPool, err = pgxpool.Connect(ctx, repository.PostgreSQLConnectionURL(*repositoryConfig))
		if err != nil {
			return err
		}
		return repositoryRawPool.Ping(ctx)

	}); err != nil {
		panic(err)
	}

	code := m.Run()

	if err := pool.Purge(resource); err != nil {
		panic(err)
	}

	os.Exit(code)

}
