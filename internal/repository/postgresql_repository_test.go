package repository_test

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/eliseeviam/wallets-service/internal/wallet"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"
	"testing"
)

func newPostgreSQLRepository(t *testing.T) *repository.PSQLWalletsRepository {
	config := new(repository.RepositoryConfig).
		SetRepositoryType(repository.RepositoryPosgreSQL).
		SetHost("localhost").
		SetPort(15432).
		SetDbName("test_db").
		SetUser("root").
		SetPassword("root")

	r, err := repository.NewWalletsCreatorRepository(*config)
	require.NoError(t, err, "seems that postgresql instance haven't been started")
	return r.(*repository.PSQLWalletsRepository)
}

func TestPostgreSQLWalletsPortgresRepository(t *testing.T) {

	t.Skipf("let's skip that tests because all test cases were implemented through handlers tests")

	r := newPostgreSQLRepository(t)

	t.Run("getNoExists", func(t *testing.T) {
		w, err := r.Get("fake_wallet")
		require.Equal(t, repository.ErrWalletNotFound, err)
		require.Nil(t, w)
	})

	randomWalletName := uuid.NewV4().String()

	t.Run("create", func(t *testing.T) {
		w, err := r.Create(randomWalletName)
		require.NoError(t, err)
		require.Equal(t, randomWalletName, w.Name())
	})

	t.Run("enroll #1", func(t *testing.T) {
		randomWallet, err := r.Get(randomWalletName)
		require.NoError(t, err)
		require.NotNil(t, randomWallet)

		n, err := r.Deposit(randomWallet, 1011)
		require.NoError(t, err)
		require.Equal(t, int64(1011), n)
	})

	t.Run("enroll #2", func(t *testing.T) {
		randomWallet, err := r.Get(randomWalletName)
		require.NoError(t, err)
		require.NotNil(t, randomWallet)

		n, err := r.Deposit(randomWallet, 1022)
		require.NoError(t, err)
		require.Equal(t, int64(2033), n)
	})

	t.Run("getEnrolled", func(t *testing.T) {
		randomWallet, err := r.Get(randomWalletName)
		require.NoError(t, err)
		require.NotNil(t, randomWallet)

		n, err := r.Balance(randomWallet)
		require.NoError(t, err)
		require.Equal(t, int64(2033), n)
	})

	t.Run("transfer", func(t *testing.T) {
		randomWalletFromName := uuid.NewV4().String()

		walletFrom, err := r.Create(randomWalletFromName)
		require.NoError(t, err)
		require.Equal(t, randomWalletFromName, walletFrom.Name())

		randomWalletToName := uuid.NewV4().String()

		walletTo, err := r.Create(randomWalletToName)
		require.NoError(t, err)
		require.Equal(t, randomWalletToName, walletTo.Name())

		secondRandomWalletToName := uuid.NewV4().String()

		secondWalletTo, err := r.Create(secondRandomWalletToName)
		require.NoError(t, err)
		require.Equal(t, secondRandomWalletToName, secondWalletTo.Name())

		n, err := r.Deposit(walletFrom, 102)
		require.NoError(t, err)
		require.Equal(t, int64(102), n)

		g := errgroup.Group{}

		g.Go(func() error {
			err = r.Transfer(walletFrom, walletTo, 100)
			return err
		})

		g.Go(func() error {
			err = r.Transfer(walletFrom, secondWalletTo, 100)
			return err
		})

		err = g.Wait()

		require.Error(t, err)

		n, err = r.Balance(walletFrom)
		require.NoError(t, err)
		require.Equal(t, int64(2), n)

		n1 := func() int64 {
			n, err := r.Balance(walletTo)
			require.NoError(t, err)
			return n
		}()

		n2 := func() int64 {
			n, err := r.Balance(secondWalletTo)
			require.NoError(t, err)
			return n
		}()

		require.True(t, n+n1+n2 == 102)

		t.Run("fetch history for walletFrom", func(t *testing.T) {
			hist, err := r.FetchHistoryForWallet(walletFrom, repository.Filter{})
			require.NoError(t, err)
			require.NotEqual(t, 0, len(hist))
		})

		t.Run("fetch history for walletTo and secondWalletTo", func(t *testing.T) {
			hist, err := r.FetchHistoryForWallet(walletTo, repository.Filter{})
			require.NoError(t, err)
			hist2, err := r.FetchHistoryForWallet(secondWalletTo, repository.Filter{})
			require.NoError(t, err)

			require.NotEqual(t, 0, len(hist)+len(hist2))
		})

		t.Run("fetch history with injection", func(t *testing.T) {
			_, err := r.FetchHistoryForWallet(
				wallet.NewDefaultWallet("1; DROP TABLE transfer_history;"), repository.Filter{})

			if err != nil {
				t.Logf("injection query error: %+v", err)
			}

			hist, err := r.FetchHistoryForWallet(secondWalletTo, repository.Filter{})
			require.NoError(t, err)

			require.NotEqual(t, 0, len(hist))
		})
	})

}
