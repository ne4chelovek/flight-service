package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Технические метрики
	HttpRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.1, 0.3, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"method", "endpoint"},
	)

	// Бизнесовые метрики
	PVZCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "pvz_created_total",
			Help: "Total number of created PVZs",
		},
	)

	ReceptionsCreatedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "receptions_created_total",
			Help: "Total number of created receptions",
		},
	)

	ProductsAddedTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "products_added_total",
			Help: "Total number of added products",
		},
	)
)

func Register() {
	prometheus.MustRegister(HttpRequestTotal)
	prometheus.MustRegister(HttpRequestDuration)
	prometheus.MustRegister(PVZCreatedTotal)
	prometheus.MustRegister(ReceptionsCreatedTotal)
	prometheus.MustRegister(ProductsAddedTotal)
}
