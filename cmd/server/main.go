package main

import (
	"context"
	"flight-service/internal/app"
	"flight-service/internal/app/closer"
	"flight-service/internal/config"
	"flight-service/internal/kafka"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Configuration loaded - Kafka GroupID: '%s' (empty: %v)",
		cfg.Kafka.GroupID, cfg.Kafka.GroupID == "")

	servers, err := app.SetupServer(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to setup servers: %v", err)
	}

	metrics.Register()

	errChan := make(chan error, 2)

	go runHTTPServer(servers.HTTP, "HTTP", errChan)
	go runHTTPServer(servers.Prometheus, "Prometheus", errChan)

	go runKafkaConsumer(ctx, servers.KafkaConsumer, "Kafka Consumer", errChan)

	closer.WaitForShutdown(ctx, errChan, servers)
}

func runHTTPServer(s *http.Server, name string, errChan chan<- error) {
	logger.Info("Starting server",
		zap.String("name", name),
		zap.String("address", s.Addr),
		zap.Time("started_at", time.Now()),
	)
	if err := s.ListenAndServe(); err != nil {
		errChan <- fmt.Errorf("failed to start %s server: %w", name, err)
	}
}

func runKafkaConsumer(ctx context.Context, consumer *kafka.Consumer, name string, errChan chan<- error) {
	logger.Info("Starting Kafka consumer",
		zap.String("name", name),
		zap.Time("started_at", time.Now()),
	)
	if err := consumer.Consume(ctx); err != nil {
		errChan <- fmt.Errorf("failed to start Kafka consumer: %w", err)
	}
}
