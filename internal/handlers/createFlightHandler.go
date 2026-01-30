package handlers

import (
	"flight-service/internal/logger"
	"flight-service/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
)

// CreateFlightHandler обрабатывает POST запрос на /api/flights
func (h *FlightHandler) CreateFlightHandler(c *gin.Context) {
	var flightReq model.FlightRequest

	// Декодируем JSON из тела запроса
	if err := c.ShouldBindJSON(&flightReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Валидация обязательных полей
	if flightReq.FlightNumber == "" || flightReq.DepartureDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight_number and departure_date are required"})
		return
	}

	// Используем сервис для создания полета
	metaID, err := h.flightService.CreateFlight(c.Request.Context(), &flightReq)
	if err != nil {
		logger.Error("Failed to create flight", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flight record"})
		return
	}

	// Возвращаем ответ
	c.JSON(http.StatusOK, gin.H{
		"id":     metaID,
		"status": "pending",
	})
}
