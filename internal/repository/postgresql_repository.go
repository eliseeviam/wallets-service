package repository

import (
	"context"
	"fmt"
	"github.com/eliseeviam/wallets-service/internal/wallet"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"strconv"
	"time"
)

type PSQLWalletsRepository struct {
	db *pgxpool.Pool
}

func PostgreSQLConnectionURL(c RepositoryConfig) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%v/%s", c.User(), c.Password(), c.Host(), c.Port(), c.DBName())
}

func newPSQLWalletsRepository(config RepositoryConfig) (commonRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	db, err := pgxpool.Connect(ctx, PostgreSQLConnectionURL(config))
	if err != nil {
		return nil, fmt.Errorf("connection error: %w", err)
	}
	s := &PSQLWalletsRepository{db: db}
	return s, nil
}

func (pgs *PSQLWalletsRepository) Create(name string) (wallet.Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tag, err := pgs.db.Exec(ctx, "INSERT INTO wallets (name, amount) VALUES ($1, $2)", name, 0)
	_ = tag

	if e, ok := err.(*pgconn.PgError); ok {
		switch e.Code {
		case "23505":
			return nil, ErrWalletAlreadyExists
		}
	}
	if err != nil {
		return nil, fmt.Errorf("cannot create wallet: %w", err)
	}
	return wallet.NewDefaultWallet(name), nil
}

func (pgs *PSQLWalletsRepository) Get(name string) (wallet.Wallet, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var exists bool
	err := pgs.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM wallets WHERE name = $1)", name).Scan(&exists)
	if err != nil {
		return nil, fmt.Errorf("cannot fetch wallet: %w", err)
	}
	if !exists {
		return nil, ErrWalletNotFound
	}
	return wallet.NewDefaultWallet(name), nil
}

func (pgs *PSQLWalletsRepository) Deposit(wallet wallet.Wallet, amount int64) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := pgs.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return -1, fmt.Errorf("cannot begin tx: %w", err)
	}

	var a int64
	err = tx.QueryRow(ctx, "UPDATE wallets SET amount = amount + $1  WHERE name = $2  RETURNING amount AS new_amount;", amount, wallet.Name()).Scan(&a)
	if err != nil {
		return -1, fmt.Errorf("cannot fetch updated amount: %w", err)
	}

	meta := map[string]interface{}{
		"source": "unknown",
	}
	err = pgs.writeHistory(ctx, tx, wallet, directionDeposit, amount, meta)
	if err != nil {
		return -1, fmt.Errorf("cannot write history: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return -1, fmt.Errorf("cannot commit tx: %w", err)
	}
	return a, err
}

func (pgs *PSQLWalletsRepository) Balance(wallet wallet.Wallet) (int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	var a int64
	err := pgs.db.QueryRow(ctx, "SELECT amount FROM wallets WHERE name = $1", wallet.Name()).Scan(&a)
	if err == pgx.ErrNoRows {
		return -1, ErrWalletNotFound
	} else if err != nil {
		return -1, err
	}
	return a, nil
}

func (pgs *PSQLWalletsRepository) Transfer(walletFrom, walletTo wallet.Wallet, amount int64) error {

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	tx, err := pgs.db.BeginTx(ctx, pgx.TxOptions{
		IsoLevel: pgx.ReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("cannot begin tx: %w", err)
	}

	//Передавая объекты walletFrom и walletTo мы гарантируем, что кошельки созданы.
	//Так как системе нет механизма удаления кошельков, можно проигнорировать проверку
	//на существование кошельков во время выполнения транзакции.
	//
	//Варианты:
	//1. Можно выполнить транзакцию с уровнем serializable. Это гарантирует отсутстсвие ошибок конкурентного доступа, но является ресурсозатратной операцией, так как заблокирует любые читающие и пишущие запросы. Потому вариант неподходящий.
	//2. SELECT FOR UPDATE заблокирует строки на обновление/удаление. Но так как никакой сложной логики принятия решений нет, то вариант избыточный.
	//3. Уровень изоляции RepeatableRead в PostgreSQL гарантирует отсутствие потерянных обновлений при выполении циклов "чтение-обновление-запись" через прерывание транзакций(и их повтор при необходимости). Такой вариант также кажется неподходящим из-за присутствия возможности решить задачу через атомарные обновления.
	//4. Не делаем заключений о возможности списать исходя из состояния счета до. Полагаемся на ограничение поля на стороне СУБД. При попытке уйти в отрицательный баланс, получим ошибку и прервем транзакцию. В этом случае нужно выделить ошибку ограничений и обработать её отдельно.
	tags, err := tx.Exec(ctx, "UPDATE wallets SET amount = amount - $1 WHERE name = $2;", amount, walletFrom.Name())
	_ = tags

	if e, ok := err.(*pgconn.PgError); ok {
		switch e.Code {
		case "23514":
			return ErrInsufficientBalance
		}
	}

	if err != nil {
		return fmt.Errorf("cannot withdraw: %w", err)
	}

	_, err = tx.Exec(ctx, "UPDATE wallets SET amount = amount + $1 WHERE name = $2;", amount, walletTo.Name())
	if err != nil {
		return fmt.Errorf("cannot deposit: %w", err)
	}

	sourceMeta := map[string]interface{}{
		"destination": walletTo.Name(),
	}
	err = pgs.writeHistory(ctx, tx, walletFrom, directionTransfer, -amount, sourceMeta)
	if err != nil {
		return fmt.Errorf("cannot write history: %w", err)
	}

	destinationMeta := map[string]interface{}{
		"source": walletFrom.Name(),
	}
	err = pgs.writeHistory(ctx, tx, walletTo, directionTransfer, amount, destinationMeta)
	if err != nil {
		return fmt.Errorf("cannot write history: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("cannot commit tx: %w", err)
	}
	return nil
}

type Direction string

const (
	directionDeposit    = "deposit"
	directionWithdrawal = "withdrawal"
	directionTransfer   = "transfer"
)

var _ = directionWithdrawal

func (pgs *PSQLWalletsRepository) writeHistory(ctx context.Context, tx pgx.Tx, wallet wallet.Wallet, direction Direction, amount int64, meta map[string]interface{}) error {
	_, err := tx.Exec(ctx, `INSERT INTO transfer_history (wallet, direction, amount, meta) VALUES ($1, $2, $3, $4);`, wallet.Name(), direction, amount, meta)
	return err
}

type Filter struct {
	Direction  string
	StartDate  time.Time
	EndDate    time.Time
	Limit      int
	OffsetByID int64
}

type Transfer struct {
	ID        int64                  `json:"id"`
	Direction Direction              `json:"direction"`
	Amount    int64                  `json:"amount"`
	Meta      map[string]interface{} `json:"meta"`
	Time      time.Time              `json:"time"`
}

func makeQuery(wallet wallet.Wallet, filter Filter) (string, []interface{}) {
	const (
		queryPrefix = "SELECT id, direction, amount, meta, time FROM transfer_history WHERE wallet = $1"
	)
	nextArgIdx := 2
	query := queryPrefix
	args := []interface{}{wallet.Name()}
	if filter.Direction != "" {
		query += " AND direction = $" + strconv.Itoa(nextArgIdx)
		nextArgIdx++
		args = append(args, filter.Direction)
	}
	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND DATE(time) >= DATE($%v)", nextArgIdx)
		nextArgIdx++
		args = append(args, filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND DATE(time) <= DATE($%v)", nextArgIdx)
		nextArgIdx++
		args = append(args, filter.EndDate)
	}

	if filter.OffsetByID > 0 {
		query += " AND id > $" + strconv.Itoa(nextArgIdx)
		nextArgIdx++
		args = append(args, filter.OffsetByID)
	}

	if filter.Limit > 0 {
		query += " LIMIT $" + strconv.Itoa(nextArgIdx)
		args = append(args, filter.Limit)
	}

	return query, args
}

func (pgs *PSQLWalletsRepository) FetchHistoryForWallet(wallet wallet.Wallet, filter Filter) ([]Transfer, error) {

	query, args := makeQuery(wallet, filter)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	rows, err := pgs.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("cannot request history: %w", err)
	}
	defer rows.Close()
	var txs []Transfer
	for rows.Next() {
		tx := Transfer{}
		err := rows.Scan(&tx.ID, &tx.Direction, &tx.Amount, &tx.Meta, &tx.Time)
		if err != nil {
			return nil, fmt.Errorf("cannot scan history row: %w", err)
		}
		txs = append(txs, tx)
	}
	return txs, nil
}
