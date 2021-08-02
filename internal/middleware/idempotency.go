package middleware

import (
	"log"
	"net/http"
	"strconv"
)

type Idempotency struct {
	repository IdempotencyKeysRepository
}

func NewIdempotency(repository IdempotencyKeysRepository) *Idempotency {
	return &Idempotency{repository: repository}
}

func (m *Idempotency) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodGet {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		key := r.Header.Get("Idempotency-Key")
		if key == "" {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("`Idempotency-Key` not found"))
			return
		}

		status, err := m.repository.Status(key)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Fatalf("cannot fetch idempotency key status; error: %+v", err)
			return
		}

		switch status {
		case Finished:
			w.WriteHeader(http.StatusCreated)
			return
		case Unknown:
			//
		default:
			panic("unknown status: " + strconv.Itoa(int(status)))
		}

		next.ServeHTTP(w, r)

		if w.(*StatusWriter).Status < 400 {
			err = m.repository.SetStatus(key, Finished)
			if err != nil {
				panic("cannot save idempotency status; error: " + err.Error())
			}
		}

	})
}
