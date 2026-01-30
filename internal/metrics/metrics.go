package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HttpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{0.1, 0.3, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"method", "endpoint"},
	)

	KafkaMessagesSent = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_sent_total",
			Help: "Total number of messages sent to Kafka",
		},
	)

	KafkaMessagesProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_messages_processed_total",
			Help: "Total number of messages processed from Kafka",
		},
	)

	KafkaProcessingErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_processing_errors_total",
			Help: "Total number of errors during Kafka message processing",
		},
	)

	FlightsProcessed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "flights_processed_total",
			Help: "Total number of flights processed",
		},
	)

	Passengers = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "passengers_per_flight",
			Help:    "Distribution of passengers per flight",
			Buckets: []float64{50, 100, 150, 200, 300, 400, 500},
		},
	)

	AircraftTypeCount = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "aircraft_type_count_total",
			Help: "Total count of flights by aircraft type",
		},
		[]string{"type"},
	)

	ChannelSize = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "channel_size",
			Help: "Current size of channels",
		},
		[]string{"channel"},
	)

	FlightMetaStatusCount = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "flight_meta_status_count",
			Help: "Count of flight meta records by status",
		},
		[]string{"status"},
	)

	KafkaConsumerLag = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag_current",
			Help: "Current lag of Kafka consumer",
		},
	)
)

func Register() {
	prometheus.MustRegister(HttpRequests)
	prometheus.MustRegister(HttpDuration)
	prometheus.MustRegister(KafkaMessagesSent)
	prometheus.MustRegister(KafkaMessagesProcessed)
	prometheus.MustRegister(KafkaProcessingErrors)
	prometheus.MustRegister(FlightsProcessed)
	prometheus.MustRegister(Passengers)
	prometheus.MustRegister(AircraftTypeCount)
	prometheus.MustRegister(ChannelSize)
	prometheus.MustRegister(FlightMetaStatusCount)
	prometheus.MustRegister(KafkaConsumerLag)
}
