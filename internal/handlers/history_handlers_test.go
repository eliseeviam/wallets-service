package handlers_test

import (
	"context"
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/handlers"
	"github.com/eliseeviam/wallets-service/internal/middleware"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/reports"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"
)

func mustPopulateTableForTestHistoryServer(t *testing.T, pool *pgxpool.Pool) {
	checkDatabase(t, pool)

	tx, err := pool.Begin(context.Background())
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", "walletName", 0)
	require.NoError(t, err)

	_, err = tx.Exec(context.Background(), "INSERT INTO wallets (name, amount) VALUES ($1, $2)", "walletNameN2", 0)
	require.NoError(t, err)

	times := []time.Time{
		time.Date(2000, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2001, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2002, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2003, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2004, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2005, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2006, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2007, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2008, 01, 01, 15, 10, 0, 0, time.UTC),
		time.Date(2009, 01, 01, 15, 10, 0, 0, time.UTC),
	}

	q := "INSERT INTO transfer_history (wallet, direction, amount, meta, time) VALUES ($1, $2, $3, $4, $5)"
	for _, tranferTime := range times {

		_, err = tx.Exec(context.Background(), q,
			"walletName", "deposit", 100, nil, tranferTime)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), q,
			"walletName", "transfer", 100, nil, tranferTime)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), q,
			"walletNameN2", "deposit", 100, nil, tranferTime)
		require.NoError(t, err)

		_, err = tx.Exec(context.Background(), q,
			"walletNameN2", "transfer", 100, nil, tranferTime)
		require.NoError(t, err)

	}

	require.NoError(t, tx.Commit(context.Background()))
}

func TestHistoryServer(t *testing.T) {

	mustInitDatabase(t, repositoryRawPool)
	defer mustClearDatabase(t, repositoryRawPool)

	mustPopulateTableForTestHistoryServer(t, repositoryRawPool)

	historyRepository, err := repository.NewHistoryRepository(*repositoryConfig)
	if err != nil {
		panic(err)
	}

	var h http.Handler

	h = &handlers.HistoryHandler{
		Parser:       new(parsers.HistoryParamsParser),
		Validator:    new(validators.HistoryParamsValidator),
		ReportWriter: new(reports.JSON),
		Repository:   historyRepository,
	}

	h = handlers.WrapWithMiddlewares(h, middleware.NewStatusWriterOverloader, middleware.NewMetrics)

	const (
		commonURLPath         = "/history/walletName"
		commonRoutePath       = "/history/{wallet_name}"
		existentWalletName    = "existentWallet"
		nonExistentWalletName = "nonExistentWallet"
	)

	t.Run("notAllowed", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, commonURLPath, nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonRoutePath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusMethodNotAllowed, rr.Code)
	})

	t.Run("noParams", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/history", nil)
		require.NoError(t, err)

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.Handle(commonRoutePath, h)
		router.ServeHTTP(rr, req)

		require.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("fetch", func(t *testing.T) {

		t.Run("all", func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, commonURLPath, nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code)

			var hist []repository.Transfer
			err = json.Unmarshal(rr.Body.Bytes(), &hist)
			require.NoError(t, err)
			require.Len(t, hist, 20)
		})

		t.Run("byFakeDirection", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("direction", "fakeDirection")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})

		t.Run("byDirection", func(t *testing.T) {
			const fullSetSize = 10
			t.Run("byDirection", func(t *testing.T) {
				vals := url.Values{}
				vals.Add("direction", "transfer")
				req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
				require.NoError(t, err)

				rr := httptest.NewRecorder()
				router := mux.NewRouter()
				router.Handle(commonRoutePath, h)
				router.ServeHTTP(rr, req)
				require.Equal(t, http.StatusOK, rr.Code)

				var hist []repository.Transfer
				err = json.Unmarshal(rr.Body.Bytes(), &hist)
				require.NoError(t, err)
				require.Len(t, hist, fullSetSize)
			})

			var lastRecordID int64
			t.Run("byDirectionAndLimit", func(t *testing.T) {
				vals := url.Values{}
				vals.Add("direction", "transfer")
				vals.Add("limit", "5")
				req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
				require.NoError(t, err)

				rr := httptest.NewRecorder()
				router := mux.NewRouter()
				router.Handle(commonRoutePath, h)
				router.ServeHTTP(rr, req)
				require.Equal(t, http.StatusOK, rr.Code)

				var hist []repository.Transfer
				err = json.Unmarshal(rr.Body.Bytes(), &hist)
				require.NoError(t, err)
				require.Len(t, hist, 5)

				lastRecordID = hist[4].ID
			})

			t.Run("byDirectionAndOffset", func(t *testing.T) {
				vals := url.Values{}
				vals.Add("direction", "transfer")
				vals.Add("offset_by_id", strconv.FormatInt(lastRecordID, 10))
				req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
				require.NoError(t, err)

				rr := httptest.NewRecorder()
				router := mux.NewRouter()
				router.Handle(commonRoutePath, h)
				router.ServeHTTP(rr, req)
				require.Equal(t, http.StatusOK, rr.Code)

				var hist []repository.Transfer
				err = json.Unmarshal(rr.Body.Bytes(), &hist)
				require.NoError(t, err)
				require.Len(t, hist, 5)
				require.True(t, hist[0].ID > lastRecordID)
			})
		})

		t.Run("byDateOnlyStartDate", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2005-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code)

			var hist []repository.Transfer
			err = json.Unmarshal(rr.Body.Bytes(), &hist)
			require.NoError(t, err)
			require.Len(t, hist, 10)
		})

		t.Run("byDateOnlyEndDate", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("end_date", "2004-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code)
			var hist []repository.Transfer
			err = json.Unmarshal(rr.Body.Bytes(), &hist)
			require.NoError(t, err)
			require.Len(t, hist, 10)
		})

		t.Run("byDateStartAndEndDateAreSame", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2004-01-01")
			vals.Add("end_date", "2004-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code)
			var hist []repository.Transfer
			err = json.Unmarshal(rr.Body.Bytes(), &hist)
			require.NoError(t, err)
			require.Len(t, hist, 2)
		})

		t.Run("byDateStartAndEndDateAreDifferent", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2002-01-01")
			vals.Add("end_date", "2008-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusOK, rr.Code)
			var hist []repository.Transfer
			err = json.Unmarshal(rr.Body.Bytes(), &hist)
			require.NoError(t, err)
			require.Len(t, hist, 14)
		})

		t.Run("byDateStartAndEndDateAreMisplaced", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2005-01-01")
			vals.Add("end_date", "2004-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})

		t.Run("byDateStartAndEndDateAreMisplaced", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2005-01-01")
			vals.Add("end_date", "2004-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})

		t.Run("byDateCorruptedStartDate", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "---2005-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})

		t.Run("byDateCorruptedEndDate", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("end_date", "---2005-01-01")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})

		t.Run("byDateStartDateIsTime", func(t *testing.T) {
			vals := url.Values{}
			vals.Add("start_date", "2005-01-01 10:01:13")
			req, err := http.NewRequest(http.MethodGet, commonURLPath+"?"+vals.Encode(), nil)
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			router := mux.NewRouter()
			router.Handle(commonRoutePath, h)
			router.ServeHTTP(rr, req)
			require.Equal(t, http.StatusBadRequest, rr.Code)
		})
	})

}
