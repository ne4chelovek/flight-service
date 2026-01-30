package service

import (
	"context"
	"flight-service/internal/model"
	"time"
)

type FlightService interface {
	CreateFlight(ctx context.Context, request *model.FlightRequest) (int, error)
	ProcessFlightMessage(ctx context.Context, metaID int, request *model.FlightRequest) error
	ProcessFlightMessageAsync(ctx context.Context, metaID int, request *model.FlightRequest) error
	GetFlight(ctx context.Context, flightNumber string, departureDate time.Time) (*model.FlightData, error)
	GetFlightMeta(ctx context.Context, flightNumber string, status string, limit int) (*model.FlightMetaResponse, error)
}
