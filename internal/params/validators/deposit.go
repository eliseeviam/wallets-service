package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"net/http"
)

type DepositValidator struct{}

func (v *DepositValidator) Validate(p *parsers.DepositParams) error {

	if p.WalletName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "required `wallet_name` not found",
		}
	}

	if p.Amount <= 0 {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "`amount` should be greater than zero",
		}
	}

	return nil
}
