package reports

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"io"
	"log"
	"strconv"
)

type CSV struct {
	WithHeader bool
}

func transferTostring(t repository.Transfer) []string {
	return []string{
		strconv.FormatInt(t.ID, 10),
		strconv.FormatInt(t.Amount, 10),
		string(t.Direction),
		func() string {
			b, err := json.Marshal(t.Meta)
			if err != nil {
				log.Printf("cannot marshal transfer meta: %v", err)
			}
			return string(b)
		}(),
		t.Time.String(),
	}
}

func (r *CSV) WriteInto(w io.Writer, history []repository.Transfer) error {
	csvw := csv.NewWriter(w)
	if r.WithHeader {
		err := csvw.Write([]string{
			"id", "amount", "direction", "meta", "time",
		})
		if err != nil {
			return fmt.Errorf("cannot write header: %w", err)
		}
	}

	for _, item := range history {
		err := csvw.Write(transferTostring(item))
		if err != nil {
			return fmt.Errorf("cannot write row: %w", err)
		}
	}
	csvw.Flush()
	return nil
}
