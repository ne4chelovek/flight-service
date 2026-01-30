package service

import (
	"flight-service/internal/kafka"
	"flight-service/internal/repository"
	"flight-service/internal/service"
	"github.com/jackc/pgx/v5/pgxpool"
)

type flightService struct {
	metaRepo      repository.MetaRepository
	flightRepo    repository.FlightRepository
	kafkaProducer *kafka.Producer
	dbPool        *pgxpool.Pool
	messageChan   chan ProcessFlightMessageRequest
}

// NewFlightService создает новый экземпляр FlightService
func NewFlightService(metaRepo repository.MetaRepository, flightRepo repository.FlightRepository, kafkaProducer *kafka.Producer, dbPool *pgxpool.Pool) service.FlightService {
	service := &flightService{
		metaRepo:      metaRepo,
		flightRepo:    flightRepo,
		kafkaProducer: kafkaProducer,
		dbPool:        dbPool,
		messageChan:   make(chan ProcessFlightMessageRequest, 100), // буферизированный канал
	}

	// Запускаем горутину для асинхронной обработки сообщений
	go service.processMessagesAsync()

	return service
}
