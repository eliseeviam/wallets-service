package parsers

import (
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/params"
	"net/http"
)

type WalletCreateParams struct {
	WalletName string
}

type WalletCreateParamsParser struct{}

func (parser *WalletCreateParamsParser) Parse(req *http.Request) (*WalletCreateParams, error) {
	raw := map[string]string{}
	err := json.NewDecoder(req.Body).Decode(&raw)
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse request",
		}
	}
	p := &WalletCreateParams{
		WalletName: raw["wallet_name"],
	}
	return p, nil
}
