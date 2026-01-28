package handlers

import (
	"net/http"
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

	// Выполнение запроса к таблице flights
	flight, err := h.flightRepo.Get(c.Request.Context(), flightNumber, departureDate)
	if err != nil {
		logger.Error("Failed to get flight", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight"})
		return
	}

	// Возврат данных рейса в нужном формате
	response := gin.H{
		"flight_number":    flight.FlightNumber,
		"departure_date":   flight.DepartureDate.Format(time.RFC3339),
		"passengers_count": flight.PassengersCount,
	}

	c.JSON(http.StatusOK, response)
}
