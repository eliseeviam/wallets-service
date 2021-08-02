package middleware

import (
	"net/http"
)

type StatusWriter struct {
	http.ResponseWriter
	Status int
}

func (r *StatusWriter) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

func NewStatusWriterOverloader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w = &StatusWriter{
			w,
			http.StatusOK,
		}
		next.ServeHTTP(w, r)
	})
}
