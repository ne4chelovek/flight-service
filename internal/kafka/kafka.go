package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"log"

	"flight-service/internal/config"
)

// Consumer handles Kafka message consumption
type Consumer struct {
	consumerGroup sarama.ConsumerGroup
	topic         string
	done          chan bool
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.Config) (*Consumer, error) {
	consumerGroup, err := sarama.NewConsumerGroup(cfg.KafkaBrokers, cfg.KafkaGroupID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return &Consumer{
		consumerGroup: consumerGroup,
		topic:         cfg.KafkaTopic,
		done:          make(chan bool),
	}, nil
}

// ConsumeMessages starts consuming messages from Kafka
func (c *Consumer) ConsumeMessages(ctx context.Context) error {
	log.Printf("Starting to consume messages from topic: %s", c.topic)

	// Start consuming messages
	go func() {
		for {
			select {
			case <-c.done:
				return
			default:
				// Handle session
				err := c.consumerGroup.Consume(ctx, []string{c.topic}, c)
				if err != nil {
					log.Printf("Error consuming messages: %v", err)
				}
			}
		}
	}()

	return nil
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages()
func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)

		// Process the message here
		c.processMessage(message)

		// Mark message as processed
		session.MarkMessage(message, "")
	}

	return nil
}

// processMessage processes individual Kafka messages
func (c *Consumer) processMessage(msg *sarama.ConsumerMessage) {
	// Placeholder for message processing logic
	log.Printf("Processing message from partition %d with offset %d", msg.Partition, msg.Offset)
	log.Printf("Message content: %s", string(msg.Value))
}

// Close closes the consumer
func (c *Consumer) Close() {
	c.done <- true
	c.consumerGroup.Close()
}
