package main

import (
	"github.com/eliseeviam/wallets-service/internal/handlers"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/reports"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"github.com/gorilla/mux"
	"net/http"
	"net/http/pprof"
)

func mustMakeWalletsGetterHandler(config Config, middlewaresChain ...func(next http.Handler) http.Handler) http.Handler {
	walletsRepository, err := repository.NewWalletsGetterRepository(
		newRepositoryConfigFrom(config.Repository))
	if err != nil {
		panic(err)
	}

	h := &handlers.WalletsGetHandler{
		Parser:     new(parsers.WalletGetParamsParser),
		Validator:  new(validators.WalletGetParamsValidator),
		Repository: walletsRepository,
	}
	return handlers.WrapWithMiddlewares(h, middlewaresChain...)
}

func mustMakeWalletsCreatorHandler(config Config, middlewaresChain ...func(next http.Handler) http.Handler) http.Handler {
	walletsRepository, err := repository.NewWalletsCreatorRepository(
		newRepositoryConfigFrom(config.Repository))
	if err != nil {
		panic(err)
	}

	h := &handlers.WalletsCreateHandler{
		Parser:     new(parsers.WalletCreateParamsParser),
		Validator:  new(validators.WalletCreateParamsValidator),
		Repository: walletsRepository,
	}
	return handlers.WrapWithMiddlewares(h, middlewaresChain...)
}

func mustMakeDepositHandler(config Config, middlewaresChain ...func(next http.Handler) http.Handler) http.Handler {
	depositRepository, err := repository.NewDepositRepository(
		newRepositoryConfigFrom(config.Repository))
	if err != nil {
		panic(err)
	}

	h := &handlers.DepositHandler{
		Parser:     new(parsers.DepositParser),
		Validator:  new(validators.DepositValidator),
		Repository: depositRepository,
	}
	return handlers.WrapWithMiddlewares(h, middlewaresChain...)
}

func mustMakeTransferHandler(config Config, middlewaresChain ...func(next http.Handler) http.Handler) http.Handler {
	transferRepository, err := repository.NewTransferRepository(
		newRepositoryConfigFrom(config.Repository))
	if err != nil {
		panic(err)
	}

	h := &handlers.TransferHandler{
		Parser:     new(parsers.TransferParser),
		Validator:  new(validators.TransferValidator),
		Repository: transferRepository,
	}
	return handlers.WrapWithMiddlewares(h, middlewaresChain...)
}

func mustMakeHistoryHandler(config Config, middlewaresChain ...func(next http.Handler) http.Handler) http.Handler {
	historyRepository, err := repository.NewHistoryRepository(
		newRepositoryConfigFrom(config.Repository))
	if err != nil {
		panic(err)
	}

	h := &handlers.HistoryHandler{
		Parser:       new(parsers.HistoryParamsParser),
		Validator:    new(validators.HistoryParamsValidator),
		ReportWriter: &reports.CSV{WithHeader: true},
		Repository:   historyRepository,
	}
	return handlers.WrapWithMiddlewares(h, middlewaresChain...)
}

func addPprofHandler(r *mux.Router) {
	r.HandleFunc("/debug/pprof/", pprof.Index)
	r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	r.HandleFunc("/debug/pprof/profile", pprof.Profile)
	r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	r.HandleFunc("/debug/pprof/trace", pprof.Trace)
}
