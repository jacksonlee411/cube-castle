package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/gaogu/cube-castle/go-app/internal/service"
)

// CDCKafkaConsumer 直接从Kafka消费CDC事件的消费者
type CDCKafkaConsumer struct {
	config         *CDCConsumerConfig
	consumerGroup  sarama.ConsumerGroup
	orgConsumer    *CDCOrganizationConsumer
	neo4jService   *service.Neo4jService
	logger         Logger
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
	isRunning      bool
	runningMutex   sync.RWMutex
}

// CDCConsumerConfig CDC消费者配置
type CDCConsumerConfig struct {
	KafkaBootstrapServers string
	ConsumerGroup         string
	Topics                []string
	SessionTimeout        time.Duration
	HeartbeatInterval     time.Duration
	RebalanceTimeout      time.Duration
	EnableAutoCommit      bool
	AutoCommitInterval    time.Duration
}

// DefaultCDCConsumerConfig 默认CDC消费者配置
func DefaultCDCConsumerConfig() *CDCConsumerConfig {
	return &CDCConsumerConfig{
		KafkaBootstrapServers: "localhost:9092",
		ConsumerGroup:         "cdc-organization-consumers",
		Topics:                []string{"organization_db.public.organization_units"},
		SessionTimeout:        30 * time.Second,
		HeartbeatInterval:     3 * time.Second,
		RebalanceTimeout:      60 * time.Second,
		EnableAutoCommit:      true,
		AutoCommitInterval:    1 * time.Second,
	}
}

// NewCDCKafkaConsumer 创建CDC Kafka消费者
func NewCDCKafkaConsumer(config *CDCConsumerConfig, neo4jService *service.Neo4jService, logger Logger) (*CDCKafkaConsumer, error) {
	if config == nil {
		config = DefaultCDCConsumerConfig()
	}

	// 配置Kafka Consumer
	consumerConfig := sarama.NewConfig()
	consumerConfig.Version = sarama.V2_6_0_0
	consumerConfig.Consumer.Group.Session.Timeout = config.SessionTimeout
	consumerConfig.Consumer.Group.Heartbeat.Interval = config.HeartbeatInterval
	consumerConfig.Consumer.Group.Rebalance.Timeout = config.RebalanceTimeout
	consumerConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerConfig.Consumer.Return.Errors = true

	// 自动提交配置
	if config.EnableAutoCommit {
		consumerConfig.Consumer.Offsets.AutoCommit.Enable = true
		consumerConfig.Consumer.Offsets.AutoCommit.Interval = config.AutoCommitInterval
	} else {
		consumerConfig.Consumer.Offsets.AutoCommit.Enable = false
	}

	// 创建Consumer Group
	consumerGroup, err := sarama.NewConsumerGroup([]string{config.KafkaBootstrapServers}, config.ConsumerGroup, consumerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	// 创建CDC组织消费者
	orgConsumer := NewCDCOrganizationConsumer(neo4jService, logger)

	ctx, cancel := context.WithCancel(context.Background())

	return &CDCKafkaConsumer{
		config:        config,
		consumerGroup: consumerGroup,
		orgConsumer:   orgConsumer,
		neo4jService:  neo4jService,
		logger:        logger,
		ctx:           ctx,
		cancel:        cancel,
		isRunning:     false,
	}, nil
}

// Start 启动CDC消费者
func (c *CDCKafkaConsumer) Start(ctx context.Context) error {
	c.runningMutex.Lock()
	defer c.runningMutex.Unlock()

	if c.isRunning {
		return fmt.Errorf("CDC consumer is already running")
	}

	// 启动错误监听
	c.wg.Add(1)
	go c.monitorErrors()

	// 启动消费循环
	c.wg.Add(1)
	go c.consumeLoop()

	c.isRunning = true
	c.logger.Info("CDC Kafka consumer started successfully", "topics", c.config.Topics)
	return nil
}

// Stop 停止CDC消费者
func (c *CDCKafkaConsumer) Stop() error {
	c.runningMutex.Lock()
	defer c.runningMutex.Unlock()

	if !c.isRunning {
		return nil
	}

	c.logger.Info("Stopping CDC Kafka consumer...")

	// 取消上下文
	c.cancel()

	// 等待所有goroutine结束
	c.wg.Wait()

	// 关闭消费者组
	if err := c.consumerGroup.Close(); err != nil {
		c.logger.Error("Failed to close consumer group", "error", err)
		return fmt.Errorf("failed to close consumer group: %w", err)
	}

	c.isRunning = false
	c.logger.Info("CDC Kafka consumer stopped successfully")
	return nil
}

// Health 健康检查
func (c *CDCKafkaConsumer) Health() error {
	c.runningMutex.RLock()
	defer c.runningMutex.RUnlock()

	if !c.isRunning {
		return fmt.Errorf("CDC consumer is not running")
	}

	select {
	case <-c.ctx.Done():
		return fmt.Errorf("CDC consumer context is cancelled")
	default:
		return nil
	}
}

// monitorErrors 监听消费者错误
func (c *CDCKafkaConsumer) monitorErrors() {
	defer c.wg.Done()

	for {
		select {
		case err := <-c.consumerGroup.Errors():
			if err != nil {
				c.logger.Error("Consumer group error", "error", err)
			}
		case <-c.ctx.Done():
			c.logger.Info("Error monitoring stopped")
			return
		}
	}
}

// consumeLoop 消费循环
func (c *CDCKafkaConsumer) consumeLoop() {
	defer c.wg.Done()

	handler := &cdcConsumerGroupHandler{
		consumer: c,
		logger:   c.logger,
	}

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Info("Consume loop stopped")
			return
		default:
			// 消费消息
			if err := c.consumerGroup.Consume(c.ctx, c.config.Topics, handler); err != nil {
				c.logger.Error("Error from consumer", "error", err)
				// 如果是上下文取消错误，退出循环
				if c.ctx.Err() != nil {
					return
				}
				// 其他错误，短暂休眠后继续
				time.Sleep(time.Second)
			}
		}
	}
}

// cdcConsumerGroupHandler 实现sarama.ConsumerGroupHandler接口
type cdcConsumerGroupHandler struct {
	consumer *CDCKafkaConsumer
	logger   Logger
}

// Setup 消费者组设置
func (h *cdcConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	h.logger.Info("CDC consumer group session setup")
	return nil
}

// Cleanup 消费者组清理
func (h *cdcConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	h.logger.Info("CDC consumer group session cleanup")
	return nil
}

// ConsumeClaim 消费消息
func (h *cdcConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}

			h.logger.Info("Received CDC message", 
				"topic", message.Topic, 
				"partition", message.Partition, 
				"offset", message.Offset,
				"timestamp", message.Timestamp,
			)

			// 处理消息
			if err := h.handleMessage(session.Context(), message); err != nil {
				h.logger.Error("Failed to handle CDC message", 
					"error", err,
					"topic", message.Topic,
					"offset", message.Offset,
				)
				// 继续处理下一条消息，不中断整个消费流程
				continue
			}

			// 标记消息已处理
			session.MarkMessage(message, "")

		case <-session.Context().Done():
			h.logger.Info("CDC consumer session context done")
			return nil
		}
	}
}

// handleMessage 处理单个消息
func (h *cdcConsumerGroupHandler) handleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	// 解析Debezium CDC事件
	var cdcEvent DebeziumEvent
	if err := json.Unmarshal(message.Value, &cdcEvent); err != nil {
		return fmt.Errorf("failed to unmarshal CDC event: %w", err)
	}

	h.logger.Info("Processing CDC event", 
		"topic", message.Topic,
		"op", cdcEvent.Op,
		"table", h.getTableName(cdcEvent),
		"timestamp", cdcEvent.TsMs,
	)

	// 根据topic路由到相应的处理器
	switch message.Topic {
	case "organization_db.public.organization_units":
		return h.consumer.orgConsumer.ConsumeRawEvent(ctx, message.Value)
	default:
		h.logger.Warn("Unknown CDC topic", "topic", message.Topic)
		return nil
	}
}

// getTableName 获取表名
func (h *cdcConsumerGroupHandler) getTableName(event DebeziumEvent) string {
	if source, ok := event.Source["table"]; ok {
		return source.(string)
	}
	return "unknown"
}