package handlers

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

type DepositHandler struct {
	Parser     *parsers.DepositParser
	Validator  *validators.DepositValidator
	Repository repository.DepositRepository
}

func (s *DepositHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

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

	wallet, err := s.Repository.Get(p.WalletName)
	if HandlerError(err, w) {
		return
	}

	_, err = s.Repository.Deposit(wallet, p.Amount)
	if HandlerError(err, w) {
		return
	}
}
