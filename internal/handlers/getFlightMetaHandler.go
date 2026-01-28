package handlers

import (
	"net/http"
	"strconv"
	"time"

	"flight-service/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetFlightMetaHandler обрабатывает GET запрос на /api/flights/{flight_number}/meta
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

	// Выполнение запроса к таблице flight_meta с фильтрацией и пагинацией
	// Используем смещение 0, так как в ТЗ не указано о смещении
	offset := 0

	metas, total, err := h.metaRepo.GetByFlightNumber(c.Request.Context(), flightNumber, status, limit, offset)
	if err != nil {
		logger.Error("Failed to get flight meta", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get flight meta"})
		return
	}

	// Формирование ответа
	metaList := make([]gin.H, len(metas))
	for i, meta := range metas {
		metaList[i] = gin.H{
			"id":             meta.ID,
			"flight_number":  meta.FlightNumber,
			"departure_date": meta.DepartureDate.Format(time.RFC3339),
			"status":         meta.Status,
			"created_at":     meta.CreatedAt.Format(time.RFC3339),
			"processed_at":   meta.ProcessedAt.Format(time.RFC3339),
		}
	}

	response := gin.H{
		"flight_number": flightNumber,
		"meta":          metaList,
		"pagination": gin.H{
			"total": total,
			"limit": limit,
		},
	}

	c.JSON(http.StatusOK, response)
}
