package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type AppMetrics struct {
	registry *prometheus.Registry
	HTTP     *HTTPMetrics
	ImageGen *ImageGenMetrics
}

func NewAppMetrics() *AppMetrics {
	registry := prometheus.NewRegistry()

	httpMetrics := newHTTPMetrics(registry)
	imageGenMetrics := newImageGenMetrics(registry)

	return &AppMetrics{
		registry: registry,
		HTTP:     httpMetrics,
		ImageGen: imageGenMetrics,
	}
}

func (m *AppMetrics) Registry() *prometheus.Registry {
	return m.registry
}

func (m *AppMetrics) Handler() http.Handler {
	return promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})
}
