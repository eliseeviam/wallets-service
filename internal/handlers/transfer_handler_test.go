package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
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
	"strconv"
	"sync"
	"testing"
)

func mustPopulateTableForTestTransferServer(t *testing.T, pool *pgxpool.Pool) {
	checkDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", "sourceWallet", 123)
	require.NoError(t, err)

	for i := 0; i < 10; i++ {
		_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", fmt.Sprintf("destWalletNum%v", i), 0)
		require.NoError(t, err)
	}

	require.NoError(t, tx.Commit(context.Background()))
}

func TestTransferServer(t *testing.T) {

	mustInitDatabase(t, repositoryRawPool)
	defer mustClearDatabase(t, repositoryRawPool)

	//var s string
	//err := repositoryRawPool.QueryRow(context.Background(), "SELECT enum_range(NULL::transfer_direction)").Scan(&s)
	//require.NoError(t, err)
	//
	//t.Log("transfer_direction", s)

	mustPopulateTableForTestTransferServer(t, repositoryRawPool)

	transferRepository, err := repository.NewTransferRepository(*repositoryConfig)
	if err != nil {
		panic(err)
	}

	h := http.Handler(&handlers.TransferHandler{
		Parser:     new(parsers.TransferParser),
		Validator:  new(validators.TransferValidator),
		Repository: transferRepository,
	})

	h = handlers.WrapWithMiddlewares(h, middleware.NewStatusWriterOverloader, middleware.NewMetrics, middleware.NewIdempotency(middleware.NewInMemIdempotencyKeysRepository()).Middleware)

	const (
		commonURLPath  = "/transfer"
		newWalletName  = "newWallet"
		idempotencyKey = "idempotencyKey"
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
				"wallet_name_from": "sourceWallet",
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
		t.Run("", func(t *testing.T) {
			b, err := json.Marshal(map[string]interface{}{
				"wallet_name_from": "sourceWallet",
				"wallet_name_to":   "destWalletNum",
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

	t.Run("transferWithoutIdempotencyKey", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name_from": "sourceWallet",
			"wallet_name_to":   "destWalletNum",
			"amount":           "15",
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

	t.Run("transferWithIdempotencyKeyNoSourceWallet", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name_from": "fakeSourceWallet",
			"wallet_name_to":   "destWalletNum0",
			"amount":           "15",
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

	t.Run("transferWithIdempotencyKeyNoDestinationWallet", func(t *testing.T) {

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name_from": "sourceWallet",
			"wallet_name_to":   "fakeDestWalletNum0",
			"amount":           "15",
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

	t.Run("transferWithIdempotencyKey", func(t *testing.T) {

		var (
			total int64
		)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT SUM(amount) FROM wallets").Scan(&total)
		require.NoError(t, err)
		require.Equal(t, int64(123), total)

		b, err := json.Marshal(map[string]interface{}{
			"wallet_name_from": "sourceWallet",
			"wallet_name_to":   "destWalletNum0",
			"amount":           "15",
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

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT SUM(amount) FROM wallets").Scan(&total)
		require.NoError(t, err)
		require.Equal(t, int64(123), total)

		var (
			historyRecordsNumSource int
			historyRecordsNumDest   int
		)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM transfer_history WHERE wallet = $1", "sourceWallet").Scan(&historyRecordsNumSource)
		require.NoError(t, err)
		require.Equal(t, 1, historyRecordsNumSource)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT COUNT(*) FROM transfer_history WHERE wallet = $1", "destWalletNum0").Scan(&historyRecordsNumDest)
		require.NoError(t, err)
		require.Equal(t, 1, historyRecordsNumDest)

	})

	t.Run("transferWithSameIdempotencyKey", func(t *testing.T) {
		b, err := json.Marshal(map[string]interface{}{
			"wallet_name_from": "sourceWallet",
			"wallet_name_to":   "destWalletNum0",
			"amount":           "15",
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

	t.Run("concurrentTransferWithIdempotencyKey", func(t *testing.T) {

		var (
			total int64
		)

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT SUM(amount) FROM wallets").Scan(&total)
		require.NoError(t, err)
		require.Equal(t, int64(123), total)

		mx := new(sync.Mutex)
		wg := new(sync.WaitGroup)
		counts := map[int]int{}

		wg.Add(+10)
		for i := 0; i < 10; i++ {

			go func(i int) {
				defer wg.Done()

				b, err := json.Marshal(map[string]interface{}{
					"wallet_name_from": "sourceWallet",
					"wallet_name_to":   fmt.Sprintf("destWalletNum%v", i),
					"amount":           "15",
				})
				require.NoError(t, err)
				req, err := http.NewRequest(http.MethodPost, commonURLPath, bytes.NewBuffer(b))
				require.NoError(t, err)

				req.Header.Add("Idempotency-Key", idempotencyKey+strconv.Itoa(i)+"noDupSuffix")
				req.Header.Set("Content-Type", "application/json")
				rr := httptest.NewRecorder()
				router := mux.NewRouter()
				router.Handle(commonURLPath, h)
				router.ServeHTTP(rr, req)

				require.True(t, rr.Code == http.StatusOK || rr.Code == http.StatusPaymentRequired)

				mx.Lock()
				counts[rr.Code] += 1
				mx.Unlock()

			}(i)

		}

		wg.Wait()

		require.Equal(t, 7, counts[http.StatusOK])
		require.Equal(t, 3, counts[http.StatusPaymentRequired])

		err = repositoryRawPool.QueryRow(context.Background(), "SELECT SUM(amount) FROM wallets").Scan(&total)
		require.NoError(t, err)
		require.Equal(t, int64(123), total)

	})

}
