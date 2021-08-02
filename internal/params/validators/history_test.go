package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"testing"
	"time"
)

func TestHistoryValidator_Validate(t *testing.T) {
	type args struct {
		p *parsers.HistoryParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{p: &parsers.HistoryParams{
				WalletName: "any_name",
				StartDate:  time.Time{},
				EndDate:    time.Time{},
				Direction:  "",
			}},
			wantErr: false,
		},
		{
			name: "NotOkWrongDates",
			args: args{p: &parsers.HistoryParams{
				WalletName: "any_name",
				StartDate:  time.Now().Add(time.Hour),
				EndDate:    time.Now(),
				Direction:  "",
			}},
			wantErr: true,
		},
		{
			name: "NotOkWrongDirection",
			args: args{p: &parsers.HistoryParams{
				WalletName: "any_name",
				StartDate:  time.Time{},
				EndDate:    time.Time{},
				Direction:  "fake_direction",
			}},
			wantErr: true,
		},
		{
			name: "NotOkWrongLimit",
			args: args{p: &parsers.HistoryParams{
				WalletName: "any_name",
				Limit:      -10,
			}},
			wantErr: true,
		},
		{
			name: "NotOkWrongOffset",
			args: args{p: &parsers.HistoryParams{
				WalletName: "any_name",
				OffsetByID: -10,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &HistoryParamsValidator{}
			if err := v.Validate(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
