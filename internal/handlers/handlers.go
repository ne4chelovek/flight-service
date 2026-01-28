package handlers

import (
	"flight-service/internal/kafka"
	"flight-service/internal/logger"
	"flight-service/internal/model"
	"flight-service/internal/repository"
	"go.uber.org/zap"
)

type FlightHandler struct {
	metaRepo      repository.MetaRepository
	flightRepo    repository.FlightRepository
	kafkaProducer *kafka.Producer
	requestChan   chan model.FlightRequestData
}

// NewFlightHandler создает новый экземпляр FlightHandler
func NewFlightHandler(metaRepo repository.MetaRepository, flightRepo repository.FlightRepository, kafkaProducer *kafka.Producer) *FlightHandler {
	handler := &FlightHandler{
		metaRepo:      metaRepo,
		flightRepo:    flightRepo,
		kafkaProducer: kafkaProducer,
		requestChan:   make(chan model.FlightRequestData, 100), // буферизированный канал
	}

	// Запускаем горутину для обработки сообщений
	go handler.processKafkaMessages()

	return handler
}

// processKafkaMessages обрабатывает сообщения из канала и отправляет их в Kafka
func (h *FlightHandler) processKafkaMessages() {
	for reqData := range h.requestChan {
		err := h.kafkaProducer.SendFlightMessage(reqData.MetaID, reqData.Request)
		if err != nil {
			logger.Error("Failed to send message to Kafka", zap.Error(err))
		}
	}
}
