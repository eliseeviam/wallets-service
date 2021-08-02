package repository

import (
	"github.com/eliseeviam/wallets-service/internal/wallet"
)

type MySQLWalletsRepository struct {
	//db *sql.DB
}

func newMySQLWalletsRepository(config RepositoryConfig) (commonRepository, error) {
	panic("unimplemented")
}

func (ms *MySQLWalletsRepository) Create(name string) (wallet.Wallet, error) {
	panic("unimplemented")
}

func (pgs *MySQLWalletsRepository) Get(name string) (wallet.Wallet, error) {
	panic("unimplemented")
}

func (ms *MySQLWalletsRepository) Deposit(wallet wallet.Wallet, amount int64) (int64, error) {
	panic("unimplemented")
}

func (ms *MySQLWalletsRepository) Transfer(walletFrom, walletTo wallet.Wallet, amount int64) error {
	panic("unimplemented")
}

func (ms *MySQLWalletsRepository) Balance(wallet wallet.Wallet) (int64, error) {
	panic("unimplemented")
}

func (ms *MySQLWalletsRepository) FetchHistoryForWallet(wallet wallet.Wallet, filter Filter) ([]Transfer, error) {
	panic("unimplemented")
}
