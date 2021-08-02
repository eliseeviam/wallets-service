package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ResponseTime *prometheus.HistogramVec
)

func init() {
	ResponseTime = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "svc",
		Subsystem: "wallets_processing",
		Name:      "response_time_seconds",
		Help:      "Response time separated by route and status cote",
		Buckets:   prometheus.DefBuckets,
	}, []string{"route", "status_code"})

	prometheus.MustRegister(ResponseTime)
}
