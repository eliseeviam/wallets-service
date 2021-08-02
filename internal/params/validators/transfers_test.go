package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"testing"
)

func TestTransferValidator_Validate(t *testing.T) {
	type args struct {
		p *parsers.TransferParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "any_name",
				WalletToName:   "another_name",
				Amount:         1,
			}},
			wantErr: false,
		},
		{
			name: "NotOkWithoutNames",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "",
				WalletToName:   "",
				Amount:         1,
			}},
			wantErr: true,
		},
		{
			name: "NotOkWithoutNames",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "any_name",
				WalletToName:   "",
				Amount:         1,
			}},
			wantErr: true,
		},
		{
			name: "NotOkEqualNames",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "any_name",
				WalletToName:   "any_name",
				Amount:         1,
			}},
			wantErr: true,
		},
		{
			name: "NotOkZeroAmount",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "any_name",
				WalletToName:   "another_name",
				Amount:         0,
			}},
			wantErr: true,
		},
		{
			name: "NotOkBelowZeroAmount",
			args: args{p: &parsers.TransferParams{
				WalletFromName: "any_name",
				WalletToName:   "another_name",
				Amount:         -1,
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &TransferValidator{}
			if err := v.Validate(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
