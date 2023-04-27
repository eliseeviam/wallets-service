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

func mustPopulateTableForTestDepositServer(t *testing.T, pool *pgxpool.Pool) {
	checkDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", "testWallet", 0)
	require.NoError(t, err)

	require.NoError(t, tx.Commit(context.Background()))
}

func TestDepositServer(t *testing.T) {

	mustInitDatabase(t, repositoryRawPool)
	defer mustClearDatabase(t, repositoryRawPool)

	mustPopulateTableForTestDepositServer(t, repositoryRawPool)

	DepositRepository, err := repository.NewDepositRepository(*repositoryConfig)
	if err != nil {
		panic(err)
	}

	h := http.Handler(&handlers.DepositHandler{
		Parser:     new(parsers.DepositParser),
		Validator:  new(validators.DepositValidator),
		Repository: DepositRepository,
	})

	h = handlers.WrapWithMiddlewares(h, middleware.NewStatusWriterOverloader, middleware.NewMetrics, middleware.NewIdempotency(middleware.NewInMemIdempotencyKeysRepository()).Middleware)

	const (
		commonURLPath  = "/deposit"
		walletName     = "testWallet"
		idempotencyKey = "idempotencyKeyXXX"
	)

	t.Run("notAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, commonURLPath, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("noParams", func(t *testing.T) {
		t.Run("", func(t *testing.T) {
			b, err := json.Marshal(map[string]interface{}{})
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonURLPath, h)
			router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusBadRequest, rr.Code)
		})
		t.Run("", func(t *testing.T) {
			b, err := json.Marshal(map[string]interface{}{
				"wallet_name": walletName,
			})
			require.NoError(t, err)
			req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonURLPath, h)
			router.ServeHTTP(rr, req)

			require.Equal(t, http.StatusBadRequest, rr.Code)
		})
	})

	t.Run("badBody", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer([]byte(`{{`)))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("depositWithoutIdempotencyKey", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name": "testWallet",
			"amount":      "15",
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("depositWithIdempotencyKeyNoWallet", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name": "fakeTestWallet",
			"amount":      "15",
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)

	})

	t.Run("depositWithIdempotencyKey", func(t *testing.T) {

		var (
			balance int64
		)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT amount FROM wallets WHERE name = $1", walletName).Scan(&balance)
		require.NoError(t, err)
		require.Equal(t, int64(0), balance)

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name": "testWallet",
			"amount":      "15",
		})
		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT amount FROM wallets WHERE name = $1", walletName).Scan(&balance)
		require.NoError(t, err)
		require.Equal(t, int64(15), balance)

		var historyRecordsNum int
		err = repositoryRawPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM transfer_history WHERE wallet = $1", walletName).Scan(&historyRecordsNum)
		require.NoError(t, err)
		require.Equal(t, 1, historyRecordsNum)

	})

	t.Run("depositWithSameIdempotencyKey", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{
			"wallet_name": "testWallet",
			"amount":      "15",
		})

		require.NoError(t, err)
		req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
		require.NoError(t, err)

		req.Header.Add("Idempotency-Key", idempotencyKey)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonURLPath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusCreated, rr.Code)
	})
}
