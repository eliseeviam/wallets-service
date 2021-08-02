package handlers_test

import (
	"context"
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

func mustPopulateTableForTestWalletGetServer(t *testing.T, pool *pgxpool.Pool) {
	checkDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", "existentWallet", 123)
	require.NoError(t, err)

	require.NoError(t, tx.Commit(context.Background()))
}

func TestWalletGetServer(t *testing.T) {

	mustInitDatabase(t, repositoryRawPool)
	defer mustClearDatabase(t, repositoryRawPool)

	mustPopulateTableForTestWalletGetServer(t, repositoryRawPool)

	walletsRepository, err := repository.NewWalletsGetterRepository(*repositoryConfig)
	if err != nil {
		panic(err)
	}

	var h http.Handler

	h = &handlers.WalletsGetHandler{
		Parser:     new(parsers.WalletGetParamsParser),
		Validator:  new(validators.WalletGetParamsValidator),
		Repository: walletsRepository,
	}

	h = handlers.WrapWithMiddlewares(h, middleware.NewStatusWriterOverloader, middleware.NewMetrics)

	const (
		existentWalletName    = "existentWallet"
		nonExistentWalletName = "nonExistentWallet"
	)

	t.Run("notAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, "/wallet/"+existentWalletName, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet/{wallet_name}", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("noParams", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/wallet/", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet/{wallet_name}", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("getExistent", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/wallet/"+existentWalletName, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet/{wallet_name}", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusOK, rr.Code)
		expected := `{"name":"existentWallet","balance":123}`
		require.Equal(t, expected, rr.Body.String())
	})

	t.Run("getNonExistent", func(t *testing.T) {
		req, err := http.NewRequest("GET", "/wallet/"+nonExistentWalletName, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle("/wallet/{wallet_name}", h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

}
