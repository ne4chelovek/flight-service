package flight

import (
	"context"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"flight-service/internal/model"
	"fmt"
	"go.uber.org/zap"
	"time"
)

func (f *flightService) ProcessFlightFromKafka(ctx context.Context, metaID int, request *model.FlightRequest) error {
	// Начинаем транзакцию
	tx, err := f.dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Гарантируем откат при ошибке
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			metrics.KafkaProcessingErrors.Inc()
			panic(p)
		} else if err != nil {
			tx.Rollback(ctx)
			metrics.KafkaProcessingErrors.Inc()
		}
	}()

	// Создаем репозитории с транзакцией
	metaRepoWithTx := f.metaRepo.WithTx(tx)
	flightRepoWithTx := f.flightRepo.WithTx(tx)

	err = metaRepoWithTx.UpdateStatus(ctx, metaID, "processed")
	if err != nil {
		return fmt.Errorf("failed to update meta status: %w", err)
	}
	// Обновляем метрики статусов
	metrics.FlightMetaStatusCount.WithLabelValues("pending").Dec()
	metrics.FlightMetaStatusCount.WithLabelValues("processed").Inc()

	// 2. Создаем или обновляем данные о полете
	// Создаем FlightData из FlightRequest
	flightData := &model.FlightData{
		AircraftType:    request.AircraftType,
		FlightNumber:    request.FlightNumber,
		DepartureDate:   request.DepartureDate,
		ArrivalDate:     request.ArrivalDate,
		PassengersCount: request.PassengersCount,
		UpdatedAt:       time.Now(),
	}

	// Здесь можно добавить дополнительную логику, если нужно
	// Например, рассчитать ArrivalDate на основе DepartureDate и времени полета

	err = flightRepoWithTx.Upsert(ctx, flightData)
	if err != nil {
		return fmt.Errorf("failed to upsert flight: %w", err)
	}

	// Успешная обработка
	metrics.FlightsProcessed.Inc()

	// 3. Если все успешно - коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Successfully processed Kafka message",
		zap.Int("metaID", metaID),
		zap.String("flightNumber", request.FlightNumber))

	return nil
}
