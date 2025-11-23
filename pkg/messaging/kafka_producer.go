package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
	"github.com/stock-market-system/pkg/config"
	"github.com/stock-market-system/pkg/models"
)

// Producer wraps Kafka producer
type Producer struct {
	writer *kafka.Writer
	config *config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(cfg *config.KafkaConfig) (*Producer, error) {
	writer := &kafka.Writer{
		Addr:                   kafka.TCP(cfg.Brokers...),
		Balancer:               &kafka.LeastBytes{},
		MaxAttempts:            cfg.MaxAttempts,
		BatchSize:              cfg.BatchSize,
		BatchTimeout:           cfg.BatchTimeout,
		ReadTimeout:            10 * time.Second,
		WriteTimeout:           10 * time.Second,
		RequiredAcks:           kafka.RequiredAcks(cfg.RequiredAcks),
		Async:                  false,
		Compression:            getCompressionCodec(cfg.CompressionCodec),
		AllowAutoTopicCreation: true,
	}

	return &Producer{
		writer: writer,
		config: cfg,
	}, nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	return p.writer.Close()
}

// PublishDataCollected publishes data collected event
func (p *Producer) PublishDataCollected(ctx context.Context, event models.DataCollectedEvent) error {
	return p.publish(ctx, p.config.TopicDataCollect, "data.collected", event)
}

// PublishIndicatorCalculated publishes indicator calculated event
func (p *Producer) PublishIndicatorCalculated(ctx context.Context, event models.IndicatorCalculatedEvent) error {
	return p.publish(ctx, p.config.TopicIndicators, "indicator.calculated", event)
}

// PublishError publishes error event
func (p *Producer) PublishError(ctx context.Context, errorMsg string, metadata map[string]interface{}) error {
	event := map[string]interface{}{
		"error":    errorMsg,
		"metadata": metadata,
	}
	return p.publish(ctx, p.config.TopicErrors, "error", event)
}

// publish publishes a message to Kafka
func (p *Producer) publish(ctx context.Context, topic string, eventType string, data interface{}) error {
	msg := models.KafkaMessage{
		MessageID: uuid.New().String(),
		Timestamp: time.Now(),
		EventType: eventType,
		Data:      data,
	}

	value, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := kafka.Message{
		Topic:     topic,
		Value:     value,
		Time:      time.Now(),
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(eventType)},
			{Key: "message-id", Value: []byte(msg.MessageID)},
		},
	}

	err = p.writer.WriteMessages(ctx, kafkaMsg)
	if err != nil {
		return fmt.Errorf("failed to write message to Kafka: %w", err)
	}

	return nil
}

// PublishBatch publishes multiple messages in batch
func (p *Producer) PublishBatch(ctx context.Context, topic string, eventType string, dataList []interface{}) error {
	if len(dataList) == 0 {
		return nil
	}

	messages := make([]kafka.Message, 0, len(dataList))

	for _, data := range dataList {
		msg := models.KafkaMessage{
			MessageID: uuid.New().String(),
			Timestamp: time.Now(),
			EventType: eventType,
			Data:      data,
		}

		value, err := json.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		kafkaMsg := kafka.Message{
			Topic:     topic,
			Value:     value,
			Time:      time.Now(),
			Headers: []kafka.Header{
				{Key: "event-type", Value: []byte(eventType)},
				{Key: "message-id", Value: []byte(msg.MessageID)},
			},
		}

		messages = append(messages, kafkaMsg)
	}

	err := p.writer.WriteMessages(ctx, messages...)
	if err != nil {
		return fmt.Errorf("failed to write batch messages to Kafka: %w", err)
	}

	return nil
}

// getCompressionCodec returns Kafka compression codec
func getCompressionCodec(codec string) kafka.Compression {
	switch codec {
	case "gzip":
		return kafka.Gzip
	case "snappy":
		return kafka.Snappy
	case "lz4":
		return kafka.Lz4
	case "zstd":
		return kafka.Zstd
	default:
		return kafka.Snappy
	}
}
