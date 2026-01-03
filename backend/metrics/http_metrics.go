package metrics

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HTTPMetrics struct {
	requests *prometheus.CounterVec
	duration *prometheus.HistogramVec
}

func newHTTPMetrics(registry *prometheus.Registry) *HTTPMetrics {
	requests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "artwork",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total number of HTTP requests.",
	}, []string{"route", "method", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "artwork",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"route", "method", "status"})

	registry.MustRegister(requests, duration)

	return &HTTPMetrics{
		requests: requests,
		duration: duration,
	}
}

func (m *HTTPMetrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		srw := &statusResponseWriter{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(srw, r)

		route := r.Pattern
		if route == "" {
			route = r.URL.Path
		}

		status := strconv.Itoa(srw.status)
		m.requests.WithLabelValues(route, r.Method, status).Inc()
		m.duration.WithLabelValues(route, r.Method, status).Observe(time.Since(start).Seconds())
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	status int
}

func (srw *statusResponseWriter) WriteHeader(statusCode int) {
	srw.status = statusCode
	srw.ResponseWriter.WriteHeader(statusCode)
}

func (srw *statusResponseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	h, ok := srw.ResponseWriter.(http.Hijacker)
	if !ok {
		return nil, nil, fmt.Errorf("hijacker not supported")
	}
	return h.Hijack()
}

func (srw *statusResponseWriter) Flush() {
	if f, ok := srw.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

func (srw *statusResponseWriter) Push(target string, opts *http.PushOptions) error {
	if p, ok := srw.ResponseWriter.(http.Pusher); ok {
		return p.Push(target, opts)
	}
	return http.ErrNotSupported
}
