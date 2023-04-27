package middleware_test

import (
	"github.com/eliseeviam/wallets-service/internal/middleware"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type StubResponseWriter struct{}

func (w *StubResponseWriter) Header() http.Header        { return nil }
func (w *StubResponseWriter) Write([]byte) (int, error)  { return 0, nil }
func (w *StubResponseWriter) WriteHeader(statusCode int) {}

type StabHandler struct {
	savedWriter http.ResponseWriter
}

func (h *StabHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	if h.savedWriter != nil {
		panic("writer saving duplication")
	}
	h.savedWriter = w
}

func TestStatusWriter(t *testing.T) {
	stub := http.Handler(&StabHandler{})
	middleware.NewStatusWriterOverloader(stub).ServeHTTP(&StubResponseWriter{}, nil)
	require.IsType(t, &middleware.StatusWriter{}, stub.(*StabHandler).savedWriter)
}
