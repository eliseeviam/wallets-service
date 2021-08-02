package validators

import (
	"github.com/eliseeviam/wallets-service/internal/params/parsers"
	"testing"
)

func TestCreateWalletValidator_Validate(t *testing.T) {
	type args struct {
		p *parsers.WalletCreateParams
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Ok",
			args: args{
				p: &parsers.WalletCreateParams{WalletName: "111"},
			},
			wantErr: false,
		},
		{
			name: "NotOk",
			args: args{
				p: &parsers.WalletCreateParams{},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &WalletCreateParamsValidator{}
			if err := v.Validate(tt.args.p); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
