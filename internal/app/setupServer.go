package app

import (
	"context"
	"flag"
	"flight-service/internal/config"
	"flight-service/internal/handlers"
	"flight-service/internal/handlers/routes"
	"flight-service/internal/kafka"
	"flight-service/internal/logger"
	"flight-service/internal/repository/flightRepo"
	"flight-service/internal/repository/metaRepo"
	"flight-service/internal/service"
	"flight-service/internal/service/flight"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"net/http"
	"os"
	"time"
)

var logLevel = flag.String("1", "info", "log level")

type Servers struct {
	HTTP          *http.Server
	Prometheus    *http.Server
	DB            *pgxpool.Pool
	KafkaProducer *kafka.Producer
	KafkaConsumer *kafka.Consumer // Добавляем consumer
}

func SetupServer(ctx context.Context, cfg *config.Config) (*Servers, error) {
	logger.Init(getCore(getAtomicLevel()))

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Name,
	)

	pool, err := initDB(ctx, dsn)
	if err != nil {
		logger.Error("Failed to connect to database", zap.Error(err))
		return nil, err
	}

	// Создаем Kafka producer
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka.KafkaBrokers, cfg.Kafka.Topic)
	if err != nil {
		logger.Error("Failed to create Kafka producer", zap.Error(err))
		return nil, err
	}

	flightService := createFlightService(kafkaProducer, pool)

	initHandler := handlers.NewFlightHandler(flightService)

	kafkaConsumer, err := kafka.NewConsumer(
		cfg.Kafka.KafkaBrokers,
		cfg.Kafka.GroupID,
		cfg.Kafka.Topic,
		initHandler,
	)

	if err != nil {
		logger.Error("Failed to create Kafka consumer", zap.Error(err))
		return nil, err
	}

	ginEng := routes.SetupRoutes(initHandler)

	return &Servers{
		HTTP: &http.Server{
			Addr:    cfg.Server.Port,
			Handler: ginEng,
		},
		Prometheus: &http.Server{
			Addr:        ":9000",
			Handler:     promhttp.Handler(),
			ReadTimeout: 15 * time.Second,
		},
		DB:            pool,
		KafkaProducer: kafkaProducer,
		KafkaConsumer: kafkaConsumer,
	}, nil
}

func getCore(level zap.AtomicLevel) zapcore.Core {
	stdout := zapcore.AddSync(os.Stdout)

	file := zapcore.AddSync(&lumberjack.Logger{
		Filename:   "logs/app.log",
		MaxSize:    10,
		MaxBackups: 3,
		MaxAge:     7,
	})

	productionCfg := zap.NewProductionEncoderConfig()
	productionCfg.TimeKey = "timestamp"
	productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	developmentCfg := zap.NewDevelopmentEncoderConfig()
	developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

	consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
	fileEncoder := zapcore.NewJSONEncoder(productionCfg)

	return zapcore.NewTee(
		zapcore.NewCore(consoleEncoder, stdout, level),
		zapcore.NewCore(fileEncoder, file, level),
	)
}

func getAtomicLevel() zap.AtomicLevel {
	var level zapcore.Level
	if err := level.Set(*logLevel); err != nil {
		log.Fatalf("failed to set log level: %v", err)
	}
	return zap.NewAtomicLevelAt(level)
}

func initDB(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		logger.Info("database ping failed:", zap.Error(err))
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func createFlightService(kafkaProducer *kafka.Producer, dbPool *pgxpool.Pool) service.FlightService {
	return flight.NewFlightService(metaRepo.NewMetaRepository(dbPool),
		flightRepo.NewFlightRepository(dbPool),
		kafkaProducer,
		dbPool)
}
