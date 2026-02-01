package flight

import (
	"context"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"go.uber.org/zap"
)

// UpdateFlightMetaStatusMetrics обновляет метрики статусов рейсов на основе данных из базы данных
func (f *flightService) UpdateFlightMetaStatusMetrics(ctx context.Context) error {
	statusCounts, err := f.metaRepo.GetStatusCounts(ctx)
	if err != nil {
		logger.Error("Failed to get status counts for metrics", zap.Error(err))
		return err
	}

	// Сбрасываем текущие значения метрик
	metrics.FlightMetaStatusCount.Reset()

	// Устанавливаем новые значения на основе данных из базы
	for status, count := range statusCounts {
		metrics.FlightMetaStatusCount.WithLabelValues(status).Set(float64(count))
	}

	logger.Info("Updated flight meta status metrics", zap.Any("status_counts", statusCounts))
	return nil
}
