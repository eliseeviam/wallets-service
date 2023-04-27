package parsers

import (
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/params"
	"net/http"
	"strconv"
)

type TransferParams struct {
	WalletFromName string
	WalletToName   string
	Amount         int64
}

type TransferParser struct{}

func (p *TransferParser) Parse(req *http.Request) (*TransferParams, error) {

	raw := map[string]string{}
	if req.Body == nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse request",
		}
	}
	err := json.NewDecoder(req.Body).Decode(&raw)
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse request",
		}
	}

	rawAmount := raw["amount"]
	if rawAmount == "" {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "required `amount` not found",
		}
	}

	amount, err := strconv.ParseInt(rawAmount, 10, 64)
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "mailformed `amount`",
		}
	}

	ps := &TransferParams{
		WalletFromName: raw["wallet_name_from"],
		WalletToName:   raw["wallet_name_to"],
		Amount:         amount,
	}

	return ps, nil

}
