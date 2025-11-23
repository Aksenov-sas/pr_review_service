package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	HTTPRequestCount      *prometheus.CounterVec
	HTTPResponseTime      *prometheus.HistogramVec
	ServiceOperationCount *prometheus.CounterVec
}

var GlobalMetrics *Metrics

func init() {
	GlobalMetrics = &Metrics{
		HTTPRequestCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Количество HTTP запросов",
			},
			[]string{"method", "endpoint", "status"},
		),
		HTTPResponseTime: promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Время ответа HTTP запросов",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint"},
		),
		ServiceOperationCount: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "service_operations_total",
				Help: "Количество операций в сервисе",
			},
			[]string{"operation", "result"},
		),
	}
}
