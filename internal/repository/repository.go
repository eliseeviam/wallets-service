package repository

import (
	"errors"
	"github.com/eliseeviam/wallets-service/internal/wallet"
)

var (
	ErrUnknownRepositoryType = errors.New("unknown repository type")
	ErrWalletNotFound        = errors.New("wallet not found")
	ErrWalletAlreadyExists   = errors.New("wallet already exists")
	ErrInsufficientBalance   = errors.New("insufficient balance")
)

type commonRepository interface {
	Create(name string) (wallet.Wallet, error)
	Get(name string) (wallet.Wallet, error)
	Deposit(wallet wallet.Wallet, amount int64) (int64, error)
	Balance(wallet wallet.Wallet) (int64, error)
	Transfer(walletFrom, walletTo wallet.Wallet, amount int64) error
	FetchHistoryForWallet(wallet wallet.Wallet, filter Filter) ([]Transfer, error)
}

type WalletsBasicGetterRepository interface {
	Get(name string) (wallet.Wallet, error)
}

type WalletsGetterRepository interface {
	WalletsBasicGetterRepository
	Balance(wallet wallet.Wallet) (int64, error)
}

type WalletsCreatorRepository interface {
	WalletsBasicGetterRepository
	Create(name string) (wallet.Wallet, error)
}

type DepositRepository interface {
	WalletsBasicGetterRepository
	Deposit(wallet wallet.Wallet, amount int64) (int64, error)
}

type TransferRepository interface {
	WalletsBasicGetterRepository
	Transfer(walletFrom, walletTo wallet.Wallet, amount int64) error
}

type HistoryRepository interface {
	WalletsBasicGetterRepository
	FetchHistoryForWallet(wallet wallet.Wallet, filter Filter) ([]Transfer, error)
}

type RepositoryType string

const (
	RepositoryPosgreSQL RepositoryType = "psql"
	RepositoryMySQL     RepositoryType = "mysql"
)

var repositoryFactories = map[RepositoryType]func(config RepositoryConfig) (commonRepository, error){
	RepositoryPosgreSQL: newPSQLWalletsRepository,
	RepositoryMySQL:     newMySQLWalletsRepository,
}

func ValidRepositoryType(repositoryType RepositoryType) bool {
	return repositoryFactories[repositoryType] != nil
}

func commonRepositoryFactory(repositoryType RepositoryType) func(config RepositoryConfig) (commonRepository, error) {
	return repositoryFactories[repositoryType]
}

func NewWalletsGetterRepository(config RepositoryConfig) (WalletsGetterRepository, error) {
	factory := commonRepositoryFactory(config.RepositoryType())
	if factory == nil {
		return nil, ErrUnknownRepositoryType
	}
	return factory(config)
}

func NewWalletsCreatorRepository(config RepositoryConfig) (WalletsCreatorRepository, error) {
	factory := commonRepositoryFactory(config.RepositoryType())
	if factory == nil {
		return nil, ErrUnknownRepositoryType
	}
	return factory(config)
}

func NewDepositRepository(config RepositoryConfig) (DepositRepository, error) {
	factory := commonRepositoryFactory(config.RepositoryType())
	if factory == nil {
		return nil, ErrUnknownRepositoryType
	}
	return factory(config)
}

func NewTransferRepository(config RepositoryConfig) (TransferRepository, error) {
	factory := commonRepositoryFactory(config.RepositoryType())
	if factory == nil {
		return nil, ErrUnknownRepositoryType
	}
	return factory(config)
}

func NewHistoryRepository(config RepositoryConfig) (HistoryRepository, error) {
	factory := commonRepositoryFactory(config.RepositoryType())
	if factory == nil {
		return nil, ErrUnknownRepositoryType
	}
	return factory(config)
}
