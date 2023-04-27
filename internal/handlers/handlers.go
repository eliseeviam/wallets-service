package handlers

import (
	"encoding/json"
	"errors"
	"github.com/eliseeviam/wallets-service/internal/params"
	"github.com/eliseeviam/wallets-service/internal/repository"
	"net/http"
)

func WrapWithMiddlewares(handler http.Handler, chain ...func(next http.Handler) http.Handler) http.Handler {
	for i := len(chain) - 1; i >= 0; i -= 1 {
		handler = chain[i](handler)
	}
	return handler
}

func HandlerError(err error, w http.ResponseWriter) (abort bool) {
	if err == nil {
		return false
	}

	var e *params.Error

	var (
		statusCode int
		ok         int
		message    string
	)

	if errors.As(err, &e) {

		statusCode = e.Code
		message = e.Message

	} else if errors.Is(err, repository.ErrWalletNotFound) {

		statusCode = http.StatusNotFound
		message = err.Error()

	} else if errors.Is(err, repository.ErrInsufficientBalance) {

		statusCode = http.StatusPaymentRequired
		message = err.Error()

	} else if errors.Is(err, repository.ErrWalletAlreadyExists) {

		statusCode = http.StatusConflict
		message = err.Error()

	} else {

		statusCode = http.StatusInternalServerError
		message = err.Error()

	}

	b, _ := json.Marshal(struct {
		Ok      int    `json:"ok"`
		Message string `json:"message"`
	}{
		Ok:      ok,
		Message: message,
	})

	w.WriteHeader(statusCode)
	_, _ = w.Write(b)

	return true
}
