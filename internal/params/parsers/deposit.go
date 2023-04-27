package parsers

import (
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/params"
	"net/http"
	"strconv"
)

type DepositParams struct {
	WalletName string
	Amount     int64
}

type DepositParser struct{}

func (p *DepositParser) Parse(req *http.Request) (*DepositParams, error) {

	raw := map[string]string{}
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

	ps := &DepositParams{
		WalletName: raw["wallet_name"],
		Amount:     amount,
	}

	return ps, nil

}
