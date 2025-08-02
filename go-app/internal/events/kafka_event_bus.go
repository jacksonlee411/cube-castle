package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/google/uuid"
)

// KafkaEventBus Kafka事件总线实现
type KafkaEventBus struct {
	config         *EventBusConfig
	producer       sarama.SyncProducer
	asyncProducer  sarama.AsyncProducer
	consumers      map[string]sarama.ConsumerGroup
	handlers       map[string][]EventHandler
	handlersMutex  sync.RWMutex
	consumersMutex sync.RWMutex
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	isRunning      bool
	runningMutex   sync.RWMutex
}

// NewKafkaEventBus 创建Kafka事件总线
func NewKafkaEventBus(config *EventBusConfig) (*KafkaEventBus, error) {
	if config == nil {
		config = DefaultEventBusConfig()
	}

	// 配置Kafka Producer
	producerConfig := sarama.NewConfig()
	producerConfig.Producer.Return.Successes = true
	producerConfig.Producer.Return.Errors = true
	producerConfig.Producer.RequiredAcks = sarama.WaitForAll
	producerConfig.Producer.Retry.Max = config.MaxRetries
	producerConfig.Producer.Retry.Backoff = config.RetryBackoff

	// 性能优化配置
	producerConfig.Producer.Flush.Frequency = config.BatchTimeout
	producerConfig.Producer.Flush.Messages = config.BatchSize
	producerConfig.Producer.Compression = sarama.CompressionSnappy

	// 幂等性配置
	producerConfig.Producer.Idempotent = true
	producerConfig.Net.MaxOpenRequests = 1

	// TLS配置
	if config.EnableTLS && config.TLSConfig != nil {
		tlsConfig, err := newTLSConfig(config.TLSConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create TLS config: %w", err)
		}
		producerConfig.Net.TLS.Enable = true
		producerConfig.Net.TLS.Config = tlsConfig
	}

	// 创建同步Producer
	producer, err := sarama.NewSyncProducer(strings.Split(config.KafkaBootstrapServers, ","), producerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create sync producer: %w", err)
	}

	// 创建异步Producer（用于批量发布）
	asyncProducerConfig := *producerConfig
	asyncProducerConfig.Producer.Return.Successes = false // 异步模式不需要等待成功响应
	asyncProducer, err := sarama.NewAsyncProducer(strings.Split(config.KafkaBootstrapServers, ","), &asyncProducerConfig)
	if err != nil {
		producer.Close()
		return nil, fmt.Errorf("failed to create async producer: %w", err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaEventBus{
		config:        config,
		producer:      producer,
		asyncProducer: asyncProducer,
		consumers:     make(map[string]sarama.ConsumerGroup),
		handlers:      make(map[string][]EventHandler),
		ctx:           ctx,
		cancel:        cancel,
		isRunning:     false,
	}, nil
}

// Publish 发布单个事件
func (k *KafkaEventBus) Publish(ctx context.Context, event DomainEvent) error {
	// 序列化事件
	eventData, err := event.Serialize()
	if err != nil {
		return fmt.Errorf("failed to serialize event: %w", err)
	}

	// 构造Kafka消息
	message := &sarama.ProducerMessage{
		Topic: k.getTopicName(event.GetEventType()),
		Key:   sarama.StringEncoder(event.GetAggregateID().String()),
		Value: sarama.ByteEncoder(eventData),
		Headers: k.buildKafkaHeaders(event),
		Timestamp: event.GetTimestamp(),
	}

	// 发送消息
	partition, offset, err := k.producer.SendMessage(message)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %w", err)
	}

	log.Printf("Event published successfully: topic=%s, partition=%d, offset=%d, eventID=%s",
		message.Topic, partition, offset, event.GetEventID())

	return nil
}

// PublishBatch 批量发布事件
func (k *KafkaEventBus) PublishBatch(ctx context.Context, events []DomainEvent) error {
	if len(events) == 0 {
		return nil
	}

	// 使用异步Producer进行批量发布
	errChan := make(chan error, len(events))
	
	// 监听错误
	go func() {
		for err := range k.asyncProducer.Errors() {
			errChan <- fmt.Errorf("async producer error: %w", err.Err)
		}
	}()

	// 发送所有事件
	for _, event := range events {
		eventData, err := event.Serialize()
		if err != nil {
			return fmt.Errorf("failed to serialize event %s: %w", event.GetEventID(), err)
		}

		message := &sarama.ProducerMessage{
			Topic: k.getTopicName(event.GetEventType()),
			Key:   sarama.StringEncoder(event.GetAggregateID().String()),
			Value: sarama.ByteEncoder(eventData),
			Headers: k.buildKafkaHeaders(event),
			Timestamp: event.GetTimestamp(),
		}

		select {
		case k.asyncProducer.Input() <- message:
			// 消息已发送到输入通道
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	// 等待一段时间检查是否有错误
	timeout := time.After(5 * time.Second)
	select {
	case err := <-errChan:
		return err
	case <-timeout:
		// 没有错误，认为批量发布成功
		log.Printf("Batch published successfully: %d events", len(events))
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Subscribe 订阅事件
func (k *KafkaEventBus) Subscribe(ctx context.Context, eventType string, handler EventHandler) error {
	k.handlersMutex.Lock()
	defer k.handlersMutex.Unlock()

	// 注册处理器
	if k.handlers[eventType] == nil {
		k.handlers[eventType] = make([]EventHandler, 0)
	}
	k.handlers[eventType] = append(k.handlers[eventType], handler)

	// 如果是第一个处理器，创建Consumer Group
	if len(k.handlers[eventType]) == 1 {
		return k.createConsumerGroup(eventType)
	}

	return nil
}

// Start 启动事件总线
func (k *KafkaEventBus) Start(ctx context.Context) error {
	k.runningMutex.Lock()
	defer k.runningMutex.Unlock()

	if k.isRunning {
		return fmt.Errorf("event bus is already running")
	}

	// 启动所有Consumer Groups
	k.consumersMutex.RLock()
	for eventType, consumer := range k.consumers {
		k.wg.Add(1)
		go k.runConsumer(eventType, consumer)
	}
	k.consumersMutex.RUnlock()

	k.isRunning = true
	log.Println("Kafka EventBus started successfully")
	return nil
}

// Stop 停止事件总线
func (k *KafkaEventBus) Stop() error {
	k.runningMutex.Lock()
	defer k.runningMutex.Unlock()

	if !k.isRunning {
		return nil
	}

	// 取消上下文
	k.cancel()

	// 关闭所有消费者
	k.consumersMutex.Lock()
	for _, consumer := range k.consumers {
		if err := consumer.Close(); err != nil {
			log.Printf("Error closing consumer: %v", err)
		}
	}
	k.consumersMutex.Unlock()

	// 关闭生产者
	if err := k.producer.Close(); err != nil {
		log.Printf("Error closing sync producer: %v", err)
	}
	if err := k.asyncProducer.Close(); err != nil {
		log.Printf("Error closing async producer: %v", err)
	}

	// 等待所有goroutine结束
	k.wg.Wait()

	k.isRunning = false
	log.Println("Kafka EventBus stopped successfully")
	return nil
}

// Health 健康检查
func (k *KafkaEventBus) Health() error {
	k.runningMutex.RLock()
	isRunning := k.isRunning
	k.runningMutex.RUnlock()

	if !isRunning {
		return fmt.Errorf("event bus is not running")
	}

	// 检查生产者连接
	// 通过发送一个测试消息来验证连接
	testEvent := &BaseDomainEvent{
		EventID:       uuid.New(),
		EventType:     "health.check",
		EventVersion:  "v1.0",
		AggregateID:   uuid.New(),
		AggregateType: "health",
		TenantID:      uuid.New(),
		Timestamp:     time.Now(),
		OccurredAt:    time.Now(),
		Metadata:      map[string]interface{}{"test": true},
		CorrelationID: uuid.New().String(),
		CausationID:   "",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := k.Publish(ctx, testEvent); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	return nil
}

// 辅助方法

// getTopicName 获取Topic名称
func (k *KafkaEventBus) getTopicName(eventType string) string {
	return fmt.Sprintf("%s.%s", k.config.KafkaTopicPrefix, eventType)
}

// buildKafkaHeaders 构建Kafka消息头
func (k *KafkaEventBus) buildKafkaHeaders(event DomainEvent) []sarama.RecordHeader {
	headers := make([]sarama.RecordHeader, 0)
	
	eventHeaders := event.GetHeaders()
	for key, value := range eventHeaders {
		headers = append(headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(value),
		})
	}

	return headers
}

// createConsumerGroup 创建消费者组
func (k *KafkaEventBus) createConsumerGroup(eventType string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Return.Errors = true

	// TLS配置
	if k.config.EnableTLS && k.config.TLSConfig != nil {
		tlsConfig, err := newTLSConfig(k.config.TLSConfig)
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		config.Net.TLS.Enable = true
		config.Net.TLS.Config = tlsConfig
	}

	consumerGroup, err := sarama.NewConsumerGroup(
		strings.Split(k.config.KafkaBootstrapServers, ","),
		k.config.KafkaConsumerGroup,
		config,
	)
	if err != nil {
		return fmt.Errorf("failed to create consumer group for event type %s: %w", eventType, err)
	}

	k.consumersMutex.Lock()
	k.consumers[eventType] = consumerGroup
	k.consumersMutex.Unlock()

	return nil
}

// runConsumer 运行消费者
func (k *KafkaEventBus) runConsumer(eventType string, consumer sarama.ConsumerGroup) {
	defer k.wg.Done()

	topic := k.getTopicName(eventType)
	handler := &consumerGroupHandler{
		eventBus:  k,
		eventType: eventType,
	}

	for {
		select {
		case <-k.ctx.Done():
			return
		default:
			if err := consumer.Consume(k.ctx, []string{topic}, handler); err != nil {
				log.Printf("Error consuming from topic %s: %v", topic, err)
				time.Sleep(time.Second) // 防止快速重试
			}
		}
	}
}

// consumerGroupHandler 消费者组处理器
type consumerGroupHandler struct {
	eventBus  *KafkaEventBus
	eventType string
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		if err := h.handleMessage(session.Context(), message); err != nil {
			log.Printf("Error handling message: %v", err)
			// 不标记消息为已处理，让它重试
			continue
		}
		session.MarkMessage(message, "")
	}
	return nil
}

func (h *consumerGroupHandler) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// 反序列化事件 (这里需要根据eventType进行具体的反序列化)
	// 简化版本：直接使用BaseDomainEvent
	var event BaseDomainEvent
	if err := json.Unmarshal(message.Value, &event); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// 获取处理器
	h.eventBus.handlersMutex.RLock()
	handlers := h.eventBus.handlers[h.eventType]
	h.eventBus.handlersMutex.RUnlock()

	// 调用所有处理器
	for _, handler := range handlers {
		if err := handler.Handle(ctx, &event); err != nil {
			log.Printf("Handler %s failed to process event %s: %v", 
				handler.GetHandlerName(), event.EventID, err)
			// 继续处理其他处理器，不因为一个处理器失败而停止
		}
	}

	return nil
}