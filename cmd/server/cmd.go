package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"flight-service/internal/config"
	"flight-service/internal/handlers"
	"flight-service/internal/kafka"
	"flight-service/internal/service"
)

func main() {
	// Load configuration from environment variables
	cfg := config.LoadConfig()

	log.Printf("Starting flight service with configuration: %+v", cfg)

	// Create flight service instance
	svc, err := service.NewFlightService(cfg)
	if err != nil {
		log.Fatalf("Failed to create flight service: %v", err)
	}
	defer svc.Close()

	// Create and register handlers
	handler := handlers.NewHandler(svc, cfg)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Register health and metrics endpoints
	mux.HandleFunc("/health", svc.HealthHandler)
	mux.HandleFunc("/metrics", svc.MetricsHandler)

	// Start HTTP server in a goroutine
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.HTTPPort),
		Handler: mux,
	}

	go func() {
		log.Printf("Starting HTTP server on port %s", cfg.HTTPPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	// Create and start Kafka consumer
	kafkaConsumer, err := kafka.NewConsumer(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka consumer: %v", err)
	}
	defer kafkaConsumer.Close()

	if err := kafkaConsumer.ConsumeMessages(); err != nil {
		log.Fatalf("Failed to start consuming messages: %v", err)
	}

	// Wait for interrupt signal to gracefully shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("Shutting down...")

	// Gracefully shutdown the server
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
