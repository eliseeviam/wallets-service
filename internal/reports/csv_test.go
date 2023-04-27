package reports_test

import (
	"bytes"
	"github.com/eliseeviam/wallets-service/internal/reports"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"testing"
	"time"
)

func TestCSV_WriteInto(t *testing.T) {
	type args struct {
		history []repository.Transfer
	}
	tests := []struct {
		name    string
		args    args
		wantW   string
		wantErr bool
	}{
		{
			name: "ok",
			args: args{history: []repository.Transfer{
				{
					ID:        1,
					Direction: repository.Direction("deposit"),
					Amount:    100,
					Meta:      map[string]interface{}{"a": "b"},
					Time:      time.Date(2020, 1, 1, 0, 0, 1, 0, time.UTC),
				},
				{
					ID:        2,
					Direction: repository.Direction("deposit"),
					Amount:    200,
					Meta:      map[string]interface{}{"a": "b"},
					Time:      time.Date(2020, 1, 1, 1, 0, 1, 0, time.UTC),
				},
				{
					ID:        3,
					Direction: repository.Direction("deposit"),
					Amount:    300,
					Meta:      map[string]interface{}{"a": "b"},
					Time:      time.Date(2020, 1, 1, 2, 0, 1, 0, time.UTC),
				},
			}},
			wantW: `id,amount,direction,meta,time
1,100,deposit,"{""a"":""b""}",2020-01-01 00:00:01 +0000 UTC
2,200,deposit,"{""a"":""b""}",2020-01-01 01:00:01 +0000 UTC
3,300,deposit,"{""a"":""b""}",2020-01-01 02:00:01 +0000 UTC
`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &reports.CSV{
				WithHeader: true,
			}
			w := &bytes.Buffer{}
			err := r.WriteInto(w, tt.args.history)
			if (err != nil) != tt.wantErr {
				t.Errorf("WriteInto() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("WriteInto() gotW = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
