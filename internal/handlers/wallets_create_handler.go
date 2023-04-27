package handlers

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

type WalletsCreateHandler struct {
	Parser     *parsers.WalletCreateParamsParser
	Validator  *validators.WalletCreateParamsValidator
	Repository repository.WalletsCreatorRepository
}

func (s *WalletsCreateHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

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

	if wal, _ := s.Repository.Get(p.WalletName); wal != nil {
		HandlerError(repository.ErrWalletAlreadyExists, w)
		return
	}

	wallet, err := s.Repository.Create(p.WalletName)
	if HandlerError(err, w) {
		return
	}

	_ = wallet
}
