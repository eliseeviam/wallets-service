package handlers

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

type TransferHandler struct {
	Parser     *parsers.TransferParser
	Validator  *validators.TransferValidator
	Repository repository.TransferRepository
}

func (s *TransferHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	if req.Method != http.MethodPost {
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

	walletFrom, err := s.Repository.Get(p.WalletFromName)
	if HandlerError(err, w) {
		return
	}

	walletTo, err := s.Repository.Get(p.WalletToName)
	if HandlerError(err, w) {
		return
	}

	err = s.Repository.Transfer(walletFrom, walletTo, p.Amount)
	if HandlerError(err, w) {
		return
	}
}
