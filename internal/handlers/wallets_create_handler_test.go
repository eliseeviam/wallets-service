package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/handlers"
	"github.com/eliseeviam/wallets-service/internal/middleware"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func mustPopulateTableForTestWalletCreateServer(t *testing.T, pool *pgxpool.Pool) {
	checkDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)

	require.NoError(t, tx.Commit(context.Background()))
}

func TestWalletCreateServer(t *testing.T) {

	mustInitDatabase(t, repositoryRawPool)
	defer mustClearDatabase(t, repositoryRawPool)

	mustPopulateTableForTestWalletCreateServer(t, repositoryRawPool)

	walletsRepository, err := repository.NewWalletsCreatorRepository(*repositoryConfig)
	if err != nil {
		panic(err)
	}

	var h http.Handler

	h = &handlers.WalletsCreateHandler{
		Parser:     new(parsers.WalletCreateParamsParser),
		Validator:  new(validators.WalletCreateParamsValidator),
		Repository: walletsRepository,
	}

	h = handlers.WrapWithMiddlewares(h, middleware.NewStatusWriterOverloader, middleware.NewMetrics, middleware.NewIdempotency(middleware.NewInMemIdempotencyKeysRepository()).Middleware)

	const (
		newWalletName     = "newWallet"
		idempotencyKey    = "idempotencyKey"
		newIdempotencyKey = "secondIdempotencyKey"
	)

	t.Run("notAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/wallet", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("no params", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/wallet", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("createBadBody", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/wallet", bytes.NewBuffer([]byte(`{{`)))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("createWithoutIdempotencyKey", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{"wallet_name": newWalletName})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "/wallet", bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("createWithIdempotencyKey", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{"wallet_name": newWalletName})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "/wallet", bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		var (
			name    string
			balance int64
		)
		err = repositoryRawPool.QueryRow(context.Background(), "SELECT name, amount FROM wallets WHERE name=$1", newWalletName).Scan(&name, &balance)
		require.NoError(t, err)
	})

	t.Run("createWithSameIdempotencyKey", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{"wallet_name": newWalletName})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "/wallet", bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
	})

	t.Run("createWithNewIdempotencyKey", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{"wallet_name": newWalletName})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, "/wallet", bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", newIdempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusConflict, rr.Code)
		require.Equal(t, `{"ok":0,"message":"wallet already exists"}`, rr.Body.String())
	})

}
