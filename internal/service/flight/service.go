package flight

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
}

// NewFlightService создает новый экземпляр FlightService
func NewFlightService(metaRepo repository.MetaRepository, flightRepo repository.FlightRepository, kafkaProducer *kafka.Producer, dbPool *pgxpool.Pool) service.FlightService {
	fs := &flightService{
		metaRepo:      metaRepo,
		flightRepo:    flightRepo,
		kafkaProducer: kafkaProducer,
		dbPool:        dbPool,
	}

	return fs
}
