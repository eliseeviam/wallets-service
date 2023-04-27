package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"net/http"
)

type HistoryParamsValidator struct{}

func (v *HistoryParamsValidator) Validate(p *parsers.HistoryParams) error {
	if !p.EndDate.IsZero() && p.StartDate.After(p.EndDate) {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "time bounds error; start date greater than end date",
		}
	}

	if p.WalletName == "" {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "wallet_name` not found",
		}
	}

	if p.Limit < 0 {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "limit` less than zero",
		}
	}

	if p.OffsetByID < 0 {
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "offset` less than zero",
		}
	}

	switch p.Direction {
	case "deposit", "transfer", "withdrawal", "":
	default:
		return &params.Error{
			Code:    http.StatusBadRequest,
			Message: "unexpected `direction`",
		}
	}

	return nil
}
