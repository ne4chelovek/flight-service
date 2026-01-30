package kafka

import (
	"context"
	"encoding/json"
	"flight-service/internal/logger"
	"flight-service/internal/metrics"
	"flight-service/internal/model"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"
)

// MessageHandler интерфейс для обработки сообщений
type MessageHandler interface {
	ProcessFlightMessage(ctx context.Context, metaID int, request *model.FlightRequest) error
}

type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	handler       MessageHandler
	retryAttempts int
	retryDelay    time.Duration
}

// NewConsumer создаёт новый экземпляр Consumer
func NewConsumer(brokers []string, groupID, topic string, handler MessageHandler) (*Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	config.Consumer.Return.Errors = true

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать ConsumerGroup: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         topic,
		handler:       handler,
		retryAttempts: 3,
		retryDelay:    5 * time.Second,
	}, nil
}

// Consume запускает потребление сообщений
func (c *Consumer) Consume(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, stopping consumer...")
			return c.consumerGroup.Close()
		default:
			err := c.consumerGroup.Consume(ctx, []string{c.topic}, c)
			if err != nil {
				logger.Info("Ошибка при потреблении сообщений: %v", zap.Error(err))
			}
		}
	}
}

// Close закрывает соединение с Kafka
func (c *Consumer) Close() error {
	return c.consumerGroup.Close()
}

// Setup вызывается при инициализации сессии
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup вызывается при завершении сессии
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim обрабатывает сообщения из конкретного раздела
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		var request model.FlightRequest
		err := json.Unmarshal(message.Value, &request)
		if err != nil {
			logger.Error("Ошибка при разборе JSON сообщения", zap.Error(err))
			metrics.KafkaProcessingErrors.Inc()
			continue
		}

		// Извлечение metaID из ключа сообщения
		var metaID int
		_, err = fmt.Sscanf(string(message.Key), "%d", &metaID)
		if err != nil {
			logger.Error("Ошибка при извлечении metaID", zap.Error(err))
			metrics.KafkaProcessingErrors.Inc()
			continue
		}

		// Обработка сообщения с retry логикой
		err = c.processWithRetry(session.Context(), metaID, &request)
		if err != nil {
			logger.Error("Ошибка при обработке сообщения после всех попыток", zap.Error(err))
			metrics.KafkaProcessingErrors.Inc()
			continue
		}

		metrics.KafkaMessagesProcessed.Inc()

		// Подтверждение обработки сообщения только при успешной транзакции
		session.MarkMessage(message, "")
	}

	return nil
}

// processWithRetry выполняет обработку сообщения с retry логикой
func (c *Consumer) processWithRetry(ctx context.Context, metaID int, request *model.FlightRequest) error {
	var lastErr error

	for attempt := 0; attempt < c.retryAttempts; attempt++ {
		if attempt > 0 {
			time.Sleep(c.retryDelay)
		}

		lastErr = c.handler.ProcessFlightMessage(ctx, metaID, request)
		if lastErr == nil {
			metrics.KafkaMessagesProcessed.Inc()
			return nil
		}

		logger.Error("Ошибка при обработке сообщения",
			zap.Int("attempt", attempt+1),
			zap.Int("meta_id", metaID),
			zap.Error(lastErr))
	}

	return lastErr
}
