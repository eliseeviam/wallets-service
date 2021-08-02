package reports

import (
	"github.com/eliseeviam/wallets-service/internal/repository"
	"io"
)

type ReportWriter interface {
	WriteInto(w io.Writer, history []repository.Transfer) error
}
