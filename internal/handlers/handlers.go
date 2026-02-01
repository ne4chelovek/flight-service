package handlers

import (
	"context"
	"flight-service/internal/model"
	"flight-service/internal/service"
)

type FlightHandler struct {
	flightService service.FlightService
}

func NewFlightHandler(flightService service.FlightService) *FlightHandler {
	return &FlightHandler{
		flightService: flightService,
	}
}

func (h *FlightHandler) ProcessFlightMessage(ctx context.Context, metaID int, request *model.FlightRequest) error {
	return h.flightService.ProcessFlightFromKafka(ctx, metaID, request)
}
