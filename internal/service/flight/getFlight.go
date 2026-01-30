package flight

import (
	"context"
	"flight-service/internal/logger"
	"flight-service/internal/model"
	"go.uber.org/zap"
	"time"
)

func (f *flightService) GetFlight(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error) {
	flight, err := f.flightRepo.Get(ctx, flightNumber, departureDate)
	if err != nil {
		logger.Error("Failed to get flight", zap.Error(err))
		return nil, err
	}

	return flight, nil
}
