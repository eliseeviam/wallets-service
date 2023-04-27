package main

import (
	"github.com/eliseeviam/wallets-service/internal/middleware"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/context"
	"log"
	"net/http"
	"time"
)

func main() {

	config := loadConfigFromEnv()
	err := validateConfig(config)

	if err != nil {
		panic("unvalid config; " + err.Error())
	}

	log.Printf("%+v", config)

	var (
		statusWriterOverloader = middleware.NewStatusWriterOverloader
		metricsMiddleware      = middleware.NewMetrics
		idempotencyMiddleware  = middleware.NewIdempotency(
			middleware.NewRedisIdempotencyKeysRepository(
				config.Idempotency.Address, config.Idempotency.Password))
	)

	walletsGetterHandler := mustMakeWalletsGetterHandler(config,
		statusWriterOverloader, metricsMiddleware, idempotencyMiddleware.Middleware)

	walletsCreatorHandler := mustMakeWalletsCreatorHandler(config,
		statusWriterOverloader, metricsMiddleware, idempotencyMiddleware.Middleware)

	depositHandler := mustMakeDepositHandler(config,
		statusWriterOverloader, metricsMiddleware, idempotencyMiddleware.Middleware)

	transferHandler := mustMakeTransferHandler(config,
		statusWriterOverloader, metricsMiddleware, idempotencyMiddleware.Middleware)

	historyHandler := mustMakeHistoryHandler(config,
		statusWriterOverloader, metricsMiddleware, idempotencyMiddleware.Middleware)

	r := mux.NewRouter()
	r.Handle("/wallet/{wallet_name}", walletsGetterHandler)
	r.Handle("/wallet", walletsCreatorHandler)
	r.Handle("/deposit", depositHandler)
	r.Handle("/transfer", transferHandler)
	r.Handle("/withdrawal", http.NotFoundHandler())
	r.Handle("/history/{wallet_name}", historyHandler)
	r.Handle("/metrics", promhttp.Handler())
	addPprofHandler(r)

	s := http.Server{
		Addr:    config.BindAddr,
		Handler: r,
	}

	go func() {
		err := s.ListenAndServe()
		if err == http.ErrServerClosed {
			return
		} else if err != nil {
			panic("listen error: " + err.Error())
		}
	}()

	log.Printf("server started")

	<-context.Background().Done()

	ctx := context.Background()
	if config.GracefulShutdownTimeoutSec >= 0 {
		// -1 is for limitlessness
		ctx, _ = context.WithTimeout(ctx,
			time.Duration(config.GracefulShutdownTimeoutSec)*time.Second)
	}
	_ = s.Shutdown(ctx)
}
