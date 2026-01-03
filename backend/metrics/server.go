package metrics

import (
	"log/slog"
	"net/http"
)

func NewMetricsHandler(metrics *AppMetrics) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", metrics.Handler())
	return mux
}

func StartMetricsServer(logger *slog.Logger, addr string, handler http.Handler) *http.Server {
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		logger.Info("starting metrics server", "addr", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("metrics server error", "error", err)
		}
	}()

	return server
}
