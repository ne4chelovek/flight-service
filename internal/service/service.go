package service

import (
	"context"
	"flight-service/internal/model"
	"time"
)

type FlightService interface {
	CreateFlight(ctx context.Context, request *model.FlightRequest) (int, error)
	GetFlight(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error)
	GetFlightMeta(ctx context.Context, flightNumber string, status string, limit int) (*model.FlightMetaResponse, error)
	ProcessFlightFromKafka(ctx context.Context, metaID int, request *model.FlightRequest) error
	UpdateFlightMetaStatusMetrics(ctx context.Context) error
}
