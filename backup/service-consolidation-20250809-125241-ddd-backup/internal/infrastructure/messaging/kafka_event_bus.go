package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/cube-castle/cmd/organization-command-server/internal/domain/entities"
	"github.com/cube-castle/cmd/organization-command-server/internal/infrastructure/logging"
)

// KafkaEventBus implements EventBus interface using Kafka
type KafkaEventBus struct {
	producer *kafka.Producer
	logger   logging.Logger
	topic    string
}

// NewKafkaEventBus creates a new Kafka event bus
func NewKafkaEventBus(brokers []string, topic, clientID string, logger logging.Logger) (*KafkaEventBus, error) {
	// Kafka producer configuration
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(brokers, ","),
		"client.id":         clientID,
		"acks":             "all",
		"retries":          3,
		"batch.size":       16384,
		"linger.ms":        10,
		"compression.type": "snappy",
		"idempotent":       true,
		"max.in.flight.requests.per.connection": 5,
	}

	producer, err := kafka.NewProducer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %w", err)
	}

	bus := &KafkaEventBus{
		producer: producer,
		logger:   logger,
		topic:    topic,
	}

	// Start delivery report handler goroutine
	go bus.handleDeliveryReports()

	return bus, nil
}

// Publish publishes a domain event to Kafka
func (b *KafkaEventBus) Publish(ctx context.Context, event entities.DomainEvent) error {
	// Serialize event
	eventData, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Create Kafka message
	message := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &b.topic,
			Partition: kafka.PartitionAny,
		},
		Key:   []byte(event.GetAggregateID()),
		Value: eventData,
		Headers: []kafka.Header{
			{Key: "event-type", Value: []byte(event.GetEventType())},
			{Key: "tenant-id", Value: []byte(event.GetTenantID().String())},
			{Key: "event-id", Value: []byte(event.GetEventID().String())},
			{Key: "event-time", Value: []byte(event.GetEventTime().Format(time.RFC3339))},
			{Key: "aggregate-id", Value: []byte(event.GetAggregateID())},
		},
		Timestamp: event.GetEventTime(),
	}

	// Produce message with delivery channel
	deliveryChan := make(chan kafka.Event, 1)
	defer close(deliveryChan)

	err = b.producer.Produce(message, deliveryChan)
	if err != nil {
		return fmt.Errorf("failed to produce event: %w", err)
	}

	// Wait for delivery confirmation with timeout
	select {
	case e := <-deliveryChan:
		m := e.(*kafka.Message)
		if m.TopicPartition.Error != nil {
			b.logger.Error("event delivery failed",
				"event_type", event.GetEventType(),
				"event_id", event.GetEventID(),
				"error", m.TopicPartition.Error,
			)
			return fmt.Errorf("event delivery failed: %w", m.TopicPartition.Error)
		}

		b.logger.Debug("event delivered successfully",
			"event_type", event.GetEventType(),
			"event_id", event.GetEventID(),
			"topic", b.topic,
			"partition", m.TopicPartition.Partition,
			"offset", m.TopicPartition.Offset,
		)
		return nil

	case <-time.After(10 * time.Second):
		b.logger.Error("event delivery timeout",
			"event_type", event.GetEventType(),
			"event_id", event.GetEventID(),
			"timeout", "10s",
		)
		return fmt.Errorf("event delivery timeout")

	case <-ctx.Done():
		b.logger.Error("event delivery cancelled",
			"event_type", event.GetEventType(),
			"event_id", event.GetEventID(),
			"error", ctx.Err(),
		)
		return fmt.Errorf("event delivery cancelled: %w", ctx.Err())
	}
}

// Close closes the Kafka producer
func (b *KafkaEventBus) Close() {
	if b.producer != nil {
		// Flush any pending messages
		b.producer.Flush(15 * 1000) // 15 seconds timeout
		b.producer.Close()
		b.logger.Info("Kafka event bus closed")
	}
}

// handleDeliveryReports handles delivery reports in a separate goroutine
func (b *KafkaEventBus) handleDeliveryReports() {
	for e := range b.producer.Events() {
		switch ev := e.(type) {
		case *kafka.Message:
			if ev.TopicPartition.Error != nil {
				b.logger.Error("message delivery failed",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"error", ev.TopicPartition.Error,
				)
			} else {
				b.logger.Debug("message delivered",
					"topic", *ev.TopicPartition.Topic,
					"partition", ev.TopicPartition.Partition,
					"offset", ev.TopicPartition.Offset,
				)
			}
		}
	}
}

// RetryableEventBus wraps an event bus with retry functionality
type RetryableEventBus struct {
	wrapped    *KafkaEventBus
	logger     logging.Logger
	maxRetries int
	retryDelay time.Duration
}

// NewRetryableEventBus creates an event bus with retry capability
func NewRetryableEventBus(wrapped *KafkaEventBus, logger logging.Logger, maxRetries int, retryDelay time.Duration) *RetryableEventBus {
	return &RetryableEventBus{
		wrapped:    wrapped,
		logger:     logger,
		maxRetries: maxRetries,
		retryDelay: retryDelay,
	}
}

// Publish publishes an event with retry logic
func (r *RetryableEventBus) Publish(ctx context.Context, event entities.DomainEvent) error {
	var lastErr error
	
	for attempt := 0; attempt <= r.maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry
			select {
			case <-time.After(r.retryDelay * time.Duration(attempt)):
			case <-ctx.Done():
				return fmt.Errorf("publish cancelled during retry wait: %w", ctx.Err())
			}
			
			r.logger.Warn("retrying event publish",
				"event_type", event.GetEventType(),
				"event_id", event.GetEventID(),
				"attempt", attempt+1,
				"max_retries", r.maxRetries,
			)
		}
		
		err := r.wrapped.Publish(ctx, event)
		if err == nil {
			if attempt > 0 {
				r.logger.Info("event publish succeeded after retry",
					"event_type", event.GetEventType(),
					"event_id", event.GetEventID(),
					"attempts", attempt+1,
				)
			}
			return nil
		}
		
		lastErr = err
		r.logger.Warn("event publish attempt failed",
			"event_type", event.GetEventType(),
			"event_id", event.GetEventID(),
			"attempt", attempt+1,
			"error", err,
		)
	}
	
	r.logger.Error("event publish failed after all retries",
		"event_type", event.GetEventType(),
		"event_id", event.GetEventID(),
		"attempts", r.maxRetries+1,
		"final_error", lastErr,
	)
	
	return fmt.Errorf("event publish failed after %d attempts: %w", r.maxRetries+1, lastErr)
}

// Close closes the underlying event bus
func (r *RetryableEventBus) Close() {
	r.wrapped.Close()
}