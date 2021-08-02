package handlers

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/reports"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

type HistoryHandler struct {
	Parser       *parsers.HistoryParamsParser
	Validator    *validators.HistoryParamsValidator
	ReportWriter reports.ReportWriter
	Repository   repository.HistoryRepository
}

func (s *HistoryHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	p, err := s.Parser.Parse(req)
	if HandlerError(err, w) {
		return
	}

	err = s.Validator.Validate(p)
	if HandlerError(err, w) {
		return
	}

	wallet, err := s.Repository.Get(p.WalletName)
	if HandlerError(err, w) {
		return
	}

	filter := repository.Filter{
		Direction:  p.Direction,
		StartDate:  p.StartDate,
		EndDate:    p.EndDate,
		Limit:      p.Limit,
		OffsetByID: p.OffsetByID,
	}

	history, err := s.Repository.FetchHistoryForWallet(wallet, filter)
	if HandlerError(err, w) {
		return
	}

	_ = s.ReportWriter.WriteInto(w, history)
}
