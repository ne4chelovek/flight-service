package handlers

import (
	"net/http"
	"strings"
	"time"

	"flight-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetFlightHandler обрабатывает GET запрос на /api/flights
func (h *FlightHandler) GetFlightHandler(c *gin.Context) {
	// Извлечение параметров flight_number и departure_date из запроса
	flightNumber := c.Query("flight_number")
	departureDateString := c.Query("departure_date")

	if flightNumber == "" || departureDateString == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight_number and departure_date are required"})
		return
	}

	// Преобразование строки даты в time.Time
	departureDate, err := time.Parse(time.RFC3339, departureDateString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid departure_date format, expected RFC3339"})
		return
	}

	// Используем сервис для получения полета
	flight, err := h.flightService.GetFlight(c.Request.Context(), flightNumber, departureDate)
	if err != nil {
		logger.Error("Failed to get flight", zap.Error(err))
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight"})
		return
	}

	// Возврат ВСЕХ данных рейса согласно FlightData
	response := gin.H{
		"aircraft_type":    flight.AircraftType,
		"flight_number":    flight.FlightNumber,
		"departure_date":   flight.DepartureDate.Format(time.RFC3339),
		"arrival_date":     flight.ArrivalDate.Format(time.RFC3339),
		"passengers_count": flight.PassengersCount,
		"updated_at":       flight.UpdatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, response)
}
