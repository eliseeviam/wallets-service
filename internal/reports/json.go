package reports

import (
	"encoding/json"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"io"
)

type JSON struct {
	WithIndent bool
}

func (r *JSON) WriteInto(w io.Writer, history []repository.Transfer) error {
	e := json.NewEncoder(w)
	if r.WithIndent {
		e.SetIndent("\t", "\t")
	}
	return e.Encode(history)
}
