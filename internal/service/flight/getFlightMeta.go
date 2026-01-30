package flight

import (
	"context"
	"flight-service/internal/logger"
	"flight-service/internal/model"
	"go.uber.org/zap"
)

func (f *flightService) GetFlightMeta(ctx context.Context, flightNumber string, status string, limit int) (*model.FlightMetaResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := 0

	metas, total, err := f.metaRepo.GetByFlightNumber(ctx, flightNumber, status, limit, offset)
	if err != nil {
		logger.Error("Failed to get flight meta", zap.Error(err))
		return nil, err
	}

	return &model.FlightMetaResponse{
		FlightNumber: flightNumber,
		Meta:         metas,
		Pagination: model.Pagination{
			Total: total,
			Limit: limit,
		},
	}, nil
}
