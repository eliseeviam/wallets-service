package parsers

import (
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

type HistoryParams struct {
	WalletName string
	StartDate  time.Time
	EndDate    time.Time
	Direction  string
	Limit      int
	OffsetByID int64
}

type HistoryParamsParser struct{}

const (
	layout = "2006-01-02"
)

func parseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.Parse(layout, s)
}

func (p *HistoryParamsParser) Parse(req *http.Request) (*HistoryParams, error) {

	err := req.ParseForm()
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse request",
		}
	}

	startDate, err := parseDate(req.FormValue("start_date"))
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse `start_date`; `" + layout + "` expected",
		}
	}

	endDate, err := parseDate(req.FormValue("end_date"))
	if err != nil {
		return nil, &params.Error{
			Code:    http.StatusBadRequest,
			Message: "cannot parse `end_date`; `" + layout + "` expected",
		}
	}

	var limit int
	if req.FormValue("limit") != "" {
		limit, err = strconv.Atoi(req.FormValue("limit"))
		if err != nil {
			return nil, &params.Error{
				Code:    http.StatusBadRequest,
				Message: "mailformed `limit`",
			}
		}
	}

	var offsetByID int64
	if req.FormValue("offset_by_id") != "" {
		offsetByID, err = strconv.ParseInt(req.FormValue("offset_by_id"), 10, 64)
		if err != nil {
			return nil, &params.Error{
				Code:    http.StatusBadRequest,
				Message: "mailformed `offset_by_id`",
			}
		}
	}

	ps := &HistoryParams{
		WalletName: mux.Vars(req)["wallet_name"],
		Direction:  req.FormValue("direction"),
		StartDate:  startDate,
		EndDate:    endDate,
		Limit:      limit,
		OffsetByID: offsetByID,
	}

	return ps, nil

}
