package closer

import (
	"context"
	"flight-service/internal/app"
	"flight-service/internal/logger"

	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func WaitForShutdown(ctx context.Context, errChan <-chan error, s *app.Servers) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		logger.Info("Received shutdown signal")
	case err := <-errChan:
		logger.Error("Critical error: ", zap.Error(err))
	case <-ctx.Done():
		logger.Info("Context cancelled")
	}

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	logger.Info("Starting graceful shutdown...")

	// 1. Останавливаем прием новых HTTP запросов
	logger.Info("Stopping HTTP server...")
	if err := s.HTTP.Shutdown(shutdownCtx); err != nil {
		logger.Error("HTTP server shutdown error:", zap.Error(err))
	}

	// 2. Закрываем Kafka consumer
	logger.Info("Closing Kafka consumer...")
	if s.KafkaConsumer != nil {
		if err := s.KafkaConsumer.CloseConsume(); err != nil {
			logger.Error("Kafka consumer close error:", zap.Error(err))
		}
	}

	// 3. Закрываем Kafka producer (он завершит обработку сообщений в канале)
	logger.Info("Closing Kafka producer (waiting for pending messages)...")
	if s.KafkaProducer != nil {
		if err := s.KafkaProducer.Close(); err != nil {
			logger.Error("Kafka producer close error:", zap.Error(err))
		}
	}

	logger.Info("Stopping Prometheus...")
	if err := s.Prometheus.Shutdown(shutdownCtx); err != nil {
		logger.Error("Prometheus shutdown error:", zap.Error(err))
	}

	logger.Info("Closing database connections...")
	s.DB.Close()

	logger.Info("Graceful shutdown completed")
}
