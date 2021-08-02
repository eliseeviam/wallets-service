package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"net/http"
)

type WalletGetParamsValidator struct{}

func (v *WalletGetParamsValidator) Validate(p *parsers.WalletGetParams) error {

	if p.WalletName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "`wallet_name` not found",
		}
	}

	return nil
}
