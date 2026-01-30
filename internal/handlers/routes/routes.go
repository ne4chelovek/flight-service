package routes

import (
	"flight-service/internal/handlers"
	"flight-service/internal/middleware"
	"github.com/gin-gonic/gin"
)

// SetupRoutes настраивает маршруты для обработчика
func SetupRoutes(handler *handlers.FlightHandler) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.MetricsMiddleware())

	r.POST("/api/flights", handler.CreateFlightHandler)
	r.GET("/api/flights", handler.GetFlightHandler)
	r.GET("/api/flights/:flight_number/meta", handler.GetFlightMetaHandler)

	return r
}
