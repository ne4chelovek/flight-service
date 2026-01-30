package flight

import (
	"context"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"flight-service/internal/model"
	"go.uber.org/zap"
	"time"
)

func (f *flightService) CreateFlight(ctx context.Context, request *model.FlightRequest) (int, error) {
	// Метрики по типу самолета
	metrics.AircraftTypeCount.WithLabelValues(request.AircraftType).Inc()

	// Метрики по пассажирам
	metrics.Passengers.Observe(float64(request.PassengersCount))

	// Создаем запись в таблице meta со статусом "pending"
	meta := &model.FlightMeta{
		FlightNumber:  request.FlightNumber,
		DepartureDate: request.DepartureDate,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	err := f.metaRepo.Create(ctx, meta)
	if err != nil {
		logger.Error("Failed to create flight meta", zap.Error(err))
		return 0, err
	}

	metrics.FlightMetaStatusCount.WithLabelValues("pending").Inc()

	// Асинхронная отправка в Kafka через producer с каналом
	err = f.kafkaProducer.SendFlightMessage(meta.ID, request)
	if err != nil {
		logger.Error("Failed to queue message for Kafka", zap.Error(err))
		// Не возвращаем ошибку пользователю - сообщение в meta уже создано
		// Можно обновить статус на "error" или оставить "pending" для повторной отправки
	}

	return meta.ID, nil
}
