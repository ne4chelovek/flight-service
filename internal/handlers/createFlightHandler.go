package handlers

import (
	"encoding/json"
	"flight-service/internal/logger"
	"flight-service/internal/model"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
)

// CreateFlightHandler обрабатывает POST запрос на /api/flights
func (h *FlightHandler) CreateFlightHandler(c *gin.Context) {
	var flightReq model.FlightRequest

	// Декодируем JSON из тела запроса
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields() // Валидация: не допускаем неизвестные поля

	if err := decoder.Decode(&flightReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON: " + err.Error()})
		return
	}

	// Валидация обязательных полей
	if flightReq.FlightNumber == "" || flightReq.DepartureDate.IsZero() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight_number and departure_date are required"})
		return
	}

	// Создаем запись в таблице meta со статусом "pending"
	meta := &model.FlightMeta{
		FlightNumber:  flightReq.FlightNumber,
		DepartureDate: flightReq.DepartureDate,
		Status:        "pending",
		CreatedAt:     time.Now(),
	}

	err := h.metaRepo.Create(c.Request.Context(), meta)
	if err != nil {
		logger.Error("Failed to create flight meta", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create flight record"})
		return
	}

	// Отправляем сообщение в Kafka асинхронно через буферизированный канал
	reqData := model.FlightRequestData{
		Request: flightReq,
		MetaID:  meta.ID,
	}

	h.requestChan <- reqData

	// Возвращаем ответ
	c.JSON(http.StatusOK, gin.H{
		"id":     meta.ID,
		"status": "pending",
	})
}
