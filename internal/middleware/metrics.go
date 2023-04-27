package middleware

import (
	"github.com/eliseeviam/wallets-service/internal/metrics"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
	"time"
)

func NewMetrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		route := mux.CurrentRoute(r)
		var path string
		if route != nil {
			path, _ = route.GetPathTemplate()
		}
		startTime := time.Now()
		next.ServeHTTP(w, r)
		metrics.ResponseTime.WithLabelValues(path, strconv.Itoa(w.(*StatusWriter).Status)).
			Observe(time.Since(startTime).Seconds())
	})
}
