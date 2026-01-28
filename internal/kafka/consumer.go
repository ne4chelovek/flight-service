package kafka

import (
	"context"
	"encoding/json"
	"flight-service/internal/model"
	"fmt"
	"log"
	"time"

	"github.com/IBM/sarama"
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
			return c.consumerGroup.Close()
		default:
			err := c.consumerGroup.Consume(ctx, []string{c.topic}, c)
			if err != nil {
				log.Printf("Ошибка при потреблении сообщений: %v", err)
			}
		}
	}
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
			log.Printf("Ошибка при разборе JSON сообщения: %v", err)
			continue
		}

		// Извлечение metaID из ключа сообщения
		var metaID int
		_, err = fmt.Sscanf(string(message.Key), "%d", &metaID)
		if err != nil {
			log.Printf("Ошибка при извлечении metaID: %v", err)
			continue
		}

		// Обработка сообщения с retry логикой
		err = c.processWithRetry(session.Context(), metaID, &request)
		if err != nil {
			log.Printf("Ошибка при обработке сообщения после всех попыток: %v", err)
			continue
		}

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
			return nil
		}

		log.Printf("Ошибка при обработке сообщения (попытка %d): %v", attempt+1, lastErr)
	}

	return lastErr
}
