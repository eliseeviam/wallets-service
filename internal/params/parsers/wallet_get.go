package parsers

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/gorilla/mux"
	"net/http"
)

type WalletGetParams struct {
	WalletName string
}

type WalletGetParamsParser struct{}

func (parser *WalletGetParamsParser) Parse(req *http.Request) (*WalletGetParams, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse request",
		}
	}
	p := &WalletGetParams{
		WalletName: mux.Vars(req)["wallet_name"],
	}
	return p, nil
}
