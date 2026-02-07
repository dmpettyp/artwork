package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type MessageBusMetrics struct {
	commands        *prometheus.CounterVec
	commandDuration *prometheus.HistogramVec
	events          *prometheus.CounterVec
	eventDuration   *prometheus.HistogramVec
}

func newMessageBusMetrics(registry *prometheus.Registry) *MessageBusMetrics {
	commands := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "artwork",
		Subsystem: "messagebus",
		Name:      "commands_total",
		Help:      "Total number of commands handled.",
	}, []string{"command", "status"})

	commandDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "artwork",
		Subsystem: "messagebus",
		Name:      "command_duration_seconds",
		Help:      "Command handler latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"command", "status"})

	events := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "artwork",
		Subsystem: "messagebus",
		Name:      "events_total",
		Help:      "Total number of events handled.",
	}, []string{"event", "status"})

	eventDuration := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "artwork",
		Subsystem: "messagebus",
		Name:      "event_duration_seconds",
		Help:      "Event handler latency in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"event", "status"})

	registry.MustRegister(commands, commandDuration, events, eventDuration)

	return &MessageBusMetrics{
		commands:        commands,
		commandDuration: commandDuration,
		events:          events,
		eventDuration:   eventDuration,
	}
}

func (m *MessageBusMetrics) ObserveCommand(commandType string, status string, duration time.Duration) {
	m.commands.WithLabelValues(commandType, status).Inc()
	m.commandDuration.WithLabelValues(commandType, status).Observe(duration.Seconds())
}

func (m *MessageBusMetrics) ObserveEvent(eventType string, status string, duration time.Duration) {
	m.events.WithLabelValues(eventType, status).Inc()
	m.eventDuration.WithLabelValues(eventType, status).Observe(duration.Seconds())
}
