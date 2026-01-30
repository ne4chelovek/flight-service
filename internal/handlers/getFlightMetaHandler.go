package handlers

import (
	"net/http"
	"strconv"
	"time"

	"flight-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func (h *FlightHandler) GetFlightMetaHandler(c *gin.Context) {
	// Извлечение параметра flight_number из пути
	flightNumber := c.Param("flight_number")
	if flightNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "flight_number is required"})
		return
	}

	// Извлечение опциональных параметров status и limit
	status := c.Query("status")
	limitStr := c.Query("limit")

	// Установка значений по умолчанию
	limit := 50
	if limitStr != "" {
		parsedLimit, err := strconv.Atoi(limitStr)
		if err != nil || parsedLimit <= 0 || parsedLimit > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be a positive integer not exceeding 100"})
			return
		}
		limit = parsedLimit
	}

	// Используем сервис для получения метаданных
	response, err := h.flightService.GetFlightMeta(c.Request.Context(), flightNumber, status, limit)
	if err != nil {
		logger.Error("Failed to get flight meta", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight meta"})
		return
	}

	// Формирование ответа
	metaList := make([]gin.H, len(response.Meta))
	for i, meta := range response.Meta {
		processedAt := ""
		if !meta.ProcessedAt.IsZero() {
			processedAt = meta.ProcessedAt.Format(time.RFC3339)
		}

		metaList[i] = gin.H{
			"id":             meta.ID,
			"flight_number":  meta.FlightNumber,
			"departure_date": meta.DepartureDate.Format(time.RFC3339),
			"status":         meta.Status,
			"created_at":     meta.CreatedAt.Format(time.RFC3339),
			"processed_at":   processedAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"flight_number": response.FlightNumber,
		"meta":          metaList,
		"pagination": gin.H{
			"total": response.Pagination.Total,
			"limit": response.Pagination.Limit,
		},
	})
}
