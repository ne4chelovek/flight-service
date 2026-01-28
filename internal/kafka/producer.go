package kafka

import (
	"encoding/json"
	"flight-service/internal/model"
	"fmt"
	"github.com/IBM/sarama"
)

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

// NewProducer создаёт новый экземпляр Producer
func NewProducer(brokers []string, topic string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll          // Ожидание подтверждения от всех реплик
	config.Producer.Retry.Max = 3                             // Количество попыток при неудаче
	config.Producer.Return.Successes = true                   // Возвращать успешные отправки
	config.Producer.Return.Errors = true                      // Возвращать ошибки отправки
	config.Producer.Partitioner = sarama.NewRandomPartitioner // Использовать случайное распределение по партициям

	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать SyncProducer: %w", err)
	}

	return &Producer{
		producer: syncProducer,
		topic:    topic,
	}, nil
}

// SendFlightMessage отправляет сообщение FlightRequest в Kafka
func (p *Producer) SendFlightMessage(metaID int, request model.FlightRequest) error {
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("ошибка маршалинга FlightRequest в JSON: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", metaID)), // Используем metaID как ключ
		Value: sarama.StringEncoder(jsonData),
	}

	// Отправляем сообщение
	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("ошибка при отправке сообщения в Kafka: %w", err)
	}

	fmt.Printf("Сообщение успешно отправлено в топик '%s', партиция: %d, смещение: %d\n", p.topic, partition, offset)
	return nil
}

// Close закрывает соединение с Kafka
func (p *Producer) Close() error {
	return p.producer.Close()
}
