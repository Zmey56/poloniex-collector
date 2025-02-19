package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	TradesReceived   prometheus.Counter
	TradesQueued     prometheus.Counter
	TradesProcessed  prometheus.Counter
	TradesDropped    prometheus.Counter
	ProcessingErrors prometheus.Counter
	ProcessingTime   prometheus.Histogram
	QueueLength      prometheus.Gauge
	WSConnections    prometheus.Gauge
	DBConnections    prometheus.Gauge
	DBLatency        prometheus.Histogram
}

func NewMetrics(registry *prometheus.Registry) *Metrics {
	m := &Metrics{
		TradesReceived: promauto.NewCounter(prometheus.CounterOpts{
			Name: "trades_received_total",
			Help: "The total number of received trades",
		}),
		TradesQueued: promauto.NewCounter(prometheus.CounterOpts{
			Name: "trades_queued_total",
			Help: "The total number of trades added to processing queue",
		}),
		TradesProcessed: promauto.NewCounter(prometheus.CounterOpts{
			Name: "trades_processed_total",
			Help: "The total number of successfully processed trades",
		}),
		TradesDropped: promauto.NewCounter(prometheus.CounterOpts{
			Name: "trades_dropped_total",
			Help: "The total number of dropped trades due to queue overflow",
		}),
		ProcessingErrors: promauto.NewCounter(prometheus.CounterOpts{
			Name: "processing_errors_total",
			Help: "The total number of trade processing errors",
		}),
		ProcessingTime: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "trade_processing_duration_seconds",
			Help:    "Time spent processing each trade",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		}),
		QueueLength: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "processing_queue_length",
			Help: "Current length of the processing queue",
		}),
		WSConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "websocket_connections",
			Help: "Number of active WebSocket connections",
		}),
		DBConnections: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "database_connections",
			Help: "Number of active database connections",
		}),
		DBLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "database_operation_duration_seconds",
			Help:    "Time spent on database operations",
			Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
		}),
	}

	registry.MustRegister(
		m.TradesReceived,
		m.TradesQueued,
		m.TradesProcessed,
		m.TradesDropped,
		m.ProcessingErrors,
		m.ProcessingTime,
		m.QueueLength,
		m.WSConnections,
		m.DBConnections,
		m.DBLatency,
	)

	return m
}
