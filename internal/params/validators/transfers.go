package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"net/http"
)

type TransferValidator struct{}

func (v *TransferValidator) Validate(p *parsers.TransferParams) error {

	if p.WalletFromName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "required `wallet_name_from` not found",
		}
	}

	if p.WalletToName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "required `wallet_name_to` not found",
		}
	}

	if p.WalletFromName == p.WalletToName {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "`wallet_name_from` equals `wallet_name_to`",
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
