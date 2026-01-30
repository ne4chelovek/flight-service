package kafka

import (
	"encoding/json"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"flight-service/internal/model"
	"fmt"
	"github.com/IBM/sarama"
	"go.uber.org/zap"
	"sync"
	"time"
)

type Producer struct {
	producer    sarama.SyncProducer
	topic       string
	requestChan chan kafkaRequest
	wg          sync.WaitGroup
	closeChan   chan struct{}
}

type kafkaRequest struct {
	metaID  int
	request *model.FlightRequest
}

// NewProducer создаёт новый экземпляр Producer с асинхронной отправкой
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.Partitioner = sarama.NewRandomPartitioner

	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать SyncProducer: %w", err)
	}

	p := &Producer{
		producer:    syncProducer,
		topic:       topic,
		requestChan: make(chan kafkaRequest, 100), // Буферизированный канал
		closeChan:   make(chan struct{}),
	}

	// Запускаем горутину для асинхронной обработки
	p.wg.Add(2)
	go p.processRequests()
	go p.metricsCollector()

	return p, nil
}

// SendFlightMessage отправляет сообщение FlightRequest в Kafka асинхронно
func (p *Producer) SendFlightMessage(metaID int, request *model.FlightRequest) error {
	// Проверяем, не закрыт ли producer
	select {
	case <-p.closeChan:
		return fmt.Errorf("producer is closed")
	default:
	}

	// Отправляем в канал для асинхронной обработки
	p.requestChan <- kafkaRequest{
		metaID:  metaID,
		request: request,
	}

	logger.Info("Flight request queued for Kafka",
		zap.Int("metaID", metaID),
		zap.String("flightNumber", request.FlightNumber))

	return nil
}

// processRequests обрабатывает запросы из канала
func (p *Producer) processRequests() {
	defer p.wg.Done()

	for {
		select {
		case <-p.closeChan:
			// Обрабатываем оставшиеся сообщения перед выходом
			for req := range p.requestChan {
				p.sendMessage(req.metaID, req.request)
			}
			return
		case req, ok := <-p.requestChan:
			if !ok {
				return
			}
			p.sendMessage(req.metaID, req.request)
		}
	}
}

// sendMessage синхронно отправляет сообщение в Kafka
func (p *Producer) sendMessage(metaID int, request *model.FlightRequest) {
	jsonData, err := json.Marshal(request)
	if err != nil {
		logger.Error("Failed to marshal FlightRequest to JSON",
			zap.Int("metaID", metaID),
			zap.Error(err))
		return
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", metaID)),
		Value: sarama.StringEncoder(jsonData),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		logger.Error("Failed to send message to Kafka",
			zap.Int("metaID", metaID),
			zap.Error(err))
		// Здесь можно добавить retry логику
	} else {
		metrics.KafkaMessagesSent.Inc()
		logger.Info("Message sent to Kafka",
			zap.Int("metaID", metaID),
			zap.String("flightNumber", request.FlightNumber),
			zap.Int32("partition", partition),
			zap.Int64("offset", offset))
	}
}

func (p *Producer) metricsCollector() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-p.closeChan:
			return
		case <-ticker.C:
			// Обновляем метрику размера канала
			metrics.ChannelSize.WithLabelValues("kafka_requests").Set(float64(len(p.requestChan)))
		}
	}
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	close(p.closeChan)   // Сигнализируем горутине о завершении
	close(p.requestChan) // Закрываем канал
	p.wg.Wait()          // Ждем завершения горутины

	// Закрываем соединение с Kafka
	return p.producer.Close()
}
