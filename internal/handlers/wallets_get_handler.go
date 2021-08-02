package handlers

import (
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"github.com/eliseeviam/wallets-service/internal/params/validators"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

type WalletsGetHandler struct {
	Parser     *parsers.WalletGetParamsParser
	Validator  *validators.WalletGetParamsValidator
	Repository repository.WalletsGetterRepository
}

func (s *WalletsGetHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

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

	type Wallet struct {
		Name    string `json:"name"`
		Balance int64  `json:"balance"`
	}

	marshal := func(w *Wallet) []byte {
		j, err := json.Marshal(w)
		if err != nil {
			panic(err)
		}
		return j
	}

	wallet, err := s.Repository.Get(p.WalletName)
	if HandlerError(err, w) {
		return
	}

	balance, err := s.Repository.Balance(wallet)
	if HandlerError(err, w) {
		return
	}

	_, _ = w.Write(marshal(&Wallet{
		Name:    wallet.Name(),
		Balance: balance,
	}))

}
