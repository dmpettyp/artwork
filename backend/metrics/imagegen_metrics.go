package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type ImageGenMetrics struct {
	previewRequests *prometheus.CounterVec
	outputRequests  *prometheus.CounterVec
	duration        *prometheus.HistogramVec
}

func newImageGenMetrics(registry *prometheus.Registry) *ImageGenMetrics {
	previewRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "artwork",
		Subsystem: "imagegen",
		Name:      "preview_requests_total",
		Help:      "Total number of preview generation attempts.",
	}, []string{"node_type", "status"})

	outputRequests := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "artwork",
		Subsystem: "imagegen",
		Name:      "output_requests_total",
		Help:      "Total number of output generation attempts.",
	}, []string{"node_type", "status"})

	duration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "artwork",
		Subsystem: "imagegen",
		Name:      "duration_seconds",
		Help:      "Total image generation latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"node_type", "status"})

	registry.MustRegister(previewRequests, outputRequests, duration)

	return &ImageGenMetrics{
		previewRequests: previewRequests,
		outputRequests:  outputRequests,
		duration:        duration,
	}
}

func (m *ImageGenMetrics) ObservePreview(nodeType, status string) {
	m.previewRequests.WithLabelValues(nodeType, status).Inc()
}

func (m *ImageGenMetrics) ObserveOutput(nodeType, status string) {
	m.outputRequests.WithLabelValues(nodeType, status).Inc()
}

func (m *ImageGenMetrics) ObserveTotal(nodeType, status string, duration time.Duration) {
	m.duration.WithLabelValues(nodeType, status).Observe(duration.Seconds())
}
