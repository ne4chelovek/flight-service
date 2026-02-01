package middleware

import (
	"flight-service/internal/metrics"
	"github.com/gin-gonic/gin"
	"strconv"
	"time"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Продолжаем обработку запроса
		c.Next()

		duration := time.Since(start).Seconds()

		// Собираем метрики
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()

		metrics.HttpRequests.WithLabelValues(method, endpoint, status).Inc()
		metrics.HttpDuration.WithLabelValues(method, endpoint).Observe(duration)
	}
}
