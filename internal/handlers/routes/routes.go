package routes

import (
	"flight-service/internal/handlers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes настраивает маршруты для обработчика
func SetupRoutes(r *gin.Engine, handler *handlers.FlightHandler) {
	r.POST("/api/flights", handler.CreateFlightHandler)
	r.GET("/api/flights", handler.GetFlightHandler)
	r.GET("/api/flights/:flight_number/meta", handler.GetFlightMetaHandler)
}
