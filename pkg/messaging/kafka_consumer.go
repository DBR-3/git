package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/logger"
	"github.com/stock-market-system/pkg/models"
)

// Consumer wraps Kafka consumer
type Consumer struct {
	reader *kafka.Reader
	config *config.KafkaConfig
	logger *logger.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(cfg *config.KafkaConfig, topic string, logger *logger.Logger) (*Consumer, error) {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:                cfg.Brokers,
		GroupID:                cfg.GroupID,
		Topic:                  topic,
		MinBytes:               1e3,  // 1KB
		MaxBytes:               10e6, // 10MB
		MaxWait:                cfg.BatchTimeout,
		CommitInterval:         cfg.CommitInterval,
		SessionTimeout:         cfg.SessionTimeout,
		RebalanceTimeout:       cfg.RebalanceTimeout,
		StartOffset:            getStartOffset(cfg.StartOffset),
		RetentionTime:          7 * 24 * time.Hour,
		Logger:                 kafka.LoggerFunc(logger.Printf),
		ErrorLogger:            kafka.LoggerFunc(logger.Errorf),
		PartitionWatchInterval: 5 * time.Second,
	})

	return &Consumer{
		reader: reader,
		config: cfg,
		logger: logger,
	}, nil
}

// Close closes the Kafka consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}

// ReadMessage reads a single message
func (c *Consumer) ReadMessage(ctx context.Context) (*models.KafkaMessage, error) {
	msg, err := c.reader.ReadMessage(ctx)
	if err != nil {
		return nil, err
	}

	var kafkaMsg models.KafkaMessage
	if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &kafkaMsg, nil
}

// FetchMessage fetches a message without committing
func (c *Consumer) FetchMessage(ctx context.Context) (*kafka.Message, error) {
	return c.reader.FetchMessage(ctx)
}

// CommitMessages commits messages
func (c *Consumer) CommitMessages(ctx context.Context, msgs ...kafka.Message) error {
	return c.reader.CommitMessages(ctx, msgs...)
}

// ConsumeDataCollected consumes data collected events
func (c *Consumer) ConsumeDataCollected(ctx context.Context, handler func(context.Context, models.DataCollectedEvent) error) error {
	c.logger.Info("Starting to consume data collected events")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Consumer context cancelled, stopping")
			return ctx.Err()
		default:
		}

		msg, err := c.reader.FetchMessage(ctx)
		if err != nil {
			if err == context.Canceled {
				return nil
			}
			c.logger.Errorf("Error fetching message: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		var kafkaMsg models.KafkaMessage
		if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
			c.logger.Errorf("Failed to unmarshal message: %v", err)
			// Commit anyway to skip bad message
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Errorf("Failed to commit message: %v", err)
			}
			continue
		}

		// Parse event data
		eventData, err := json.Marshal(kafkaMsg.Data)
		if err != nil {
			c.logger.Errorf("Failed to marshal event data: %v", err)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Errorf("Failed to commit message: %v", err)
			}
			continue
		}

		var event models.DataCollectedEvent
		if err := json.Unmarshal(eventData, &event); err != nil {
			c.logger.Errorf("Failed to unmarshal data collected event: %v", err)
			if err := c.reader.CommitMessages(ctx, msg); err != nil {
				c.logger.Errorf("Failed to commit message: %v", err)
			}
			continue
		}

		// Process event
		if err := handler(ctx, event); err != nil {
			c.logger.Errorf("Failed to process event for ticker %s: %v", event.Ticker, err)
			// Don't commit on error - will retry
			continue
		}

		// Commit message
		if err := c.reader.CommitMessages(ctx, msg); err != nil {
			c.logger.Errorf("Failed to commit message: %v", err)
		}
	}
}

// ConsumeBatch consumes messages in batch
func (c *Consumer) ConsumeBatch(ctx context.Context, batchSize int, timeout time.Duration) ([]*models.KafkaMessage, error) {
	messages := make([]*models.KafkaMessage, 0, batchSize)

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for i := 0; i < batchSize; i++ {
		msg, err := c.reader.FetchMessage(timeoutCtx)
		if err != nil {
			if err == context.DeadlineExceeded {
				break
			}
			return messages, err
		}

		var kafkaMsg models.KafkaMessage
		if err := json.Unmarshal(msg.Value, &kafkaMsg); err != nil {
			c.logger.Errorf("Failed to unmarshal message: %v", err)
			continue
		}

		messages = append(messages, &kafkaMsg)
	}

	return messages, nil
}

// Stats returns consumer statistics
func (c *Consumer) Stats() kafka.ReaderStats {
	return c.reader.Stats()
}

// Lag returns consumer lag
func (c *Consumer) Lag() int64 {
	stats := c.reader.Stats()
	return stats.Lag
}

// getStartOffset converts string to Kafka offset
func getStartOffset(offset string) int64 {
	switch offset {
	case "earliest":
		return kafka.FirstOffset
	case "latest":
		return kafka.LastOffset
	default:
		return kafka.LastOffset
	}
}
