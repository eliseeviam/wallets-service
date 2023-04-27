package repository_test

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/eliseeviam/wallets-service/internal/wallet"
	"github.com/stretchr/testify/require"
	"testing"
)

func newMySQLRepository(t *testing.T) *repository.MySQLWalletsRepository {
	config := new(repository.RepositoryConfig).
		SetRepositoryType(repository.RepositoryMySQL).
		SetHost("localhost").
		SetPort(13306).
		SetDbName("test_db").
		SetUser("root").
		SetPassword("root")

	r, err := repository.NewWalletsCreatorRepository(*config)
	require.NoError(t, err)
	return r.(*repository.MySQLWalletsRepository)
}

func TestNewMySQLWalletsRepository(t *testing.T) {

	var repo *repository.MySQLWalletsRepository
	require.Panics(t, func() {
		repo = newMySQLRepository(t)
	})

	require.Panics(t, func() {
		_, _ = repo.Create("wallet_name")
	})

	require.Panics(t, func() {
		_, _ = repo.Get("wallet_name")
	})

	require.Panics(t, func() {
		_, _ = repo.Deposit(wallet.NewDefaultWallet("wallet_name"), 1.)
	})

	require.Panics(t, func() {
		_ = repo.Transfer(wallet.NewDefaultWallet("from_wallet_name"), wallet.NewDefaultWallet("to_wallet_name"), 1)
	})

	require.Panics(t, func() {
		_, _ = repo.Balance(wallet.NewDefaultWallet("wallet_name"))
	})

	require.Panics(t, func() {
		_, _ = repo.FetchHistoryForWallet(wallet.NewDefaultWallet("wallet_name"), repository.Filter{})
	})

}
