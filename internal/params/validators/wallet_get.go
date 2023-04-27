package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"net/http"
)

type WalletCreateParamsValidator struct{}

func (v *WalletCreateParamsValidator) Validate(p *parsers.WalletCreateParams) error {

	if p.WalletName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "`wallet_name` not found",
		}
	}

	return nil
}
