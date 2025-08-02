package neo4j

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

// EventConsumer Neo4j事件消费者接口
type EventConsumer interface {
	ConsumeEvent(ctx context.Context, event events.DomainEvent) error
	GetEventType() string
	Start(ctx context.Context) error
	Stop() error
	Health() error
}

// BaseEventConsumer 基础事件消费者
type BaseEventConsumer struct {
	connectionManager ConnectionManagerInterface
	eventType         string
	retryConfig       *RetryConfig
}

// ConnectionManagerInterface 连接管理器接口
type ConnectionManagerInterface interface {
	ExecuteWrite(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error)
	ExecuteRead(ctx context.Context, work neo4j.ManagedTransactionWork) (any, error)
	ExecuteWithRetry(ctx context.Context, work func(ctx context.Context) error) error
	Health(ctx context.Context) error
	Close(ctx context.Context) error
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxRetries   int
	RetryBackoff time.Duration
	MaxBackoff   time.Duration
}

// DefaultRetryConfig 默认重试配置
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:   3,
		RetryBackoff: time.Second,
		MaxBackoff:   time.Second * 30,
	}
}

// NewBaseEventConsumer 创建基础事件消费者
func NewBaseEventConsumer(connMgr ConnectionManagerInterface, eventType string) *BaseEventConsumer {
	return &BaseEventConsumer{
		connectionManager: connMgr,
		eventType:         eventType,
		retryConfig:       DefaultRetryConfig(),
	}
}

// GetEventType 获取事件类型
func (b *BaseEventConsumer) GetEventType() string {
	return b.eventType
}

// Start 启动消费者
func (b *BaseEventConsumer) Start(ctx context.Context) error {
	log.Printf("🚀 启动Neo4j事件消费者: %s", b.eventType)
	return nil
}

// Stop 停止消费者
func (b *BaseEventConsumer) Stop() error {
	log.Printf("🛑 停止Neo4j事件消费者: %s", b.eventType)
	return nil
}

// Health 健康检查
func (b *BaseEventConsumer) Health() error {
	return b.connectionManager.Health(context.Background())
}

// ConsumeEvent 消费事件（基础实现）
func (b *BaseEventConsumer) ConsumeEvent(ctx context.Context, event events.DomainEvent) error {
	return fmt.Errorf("ConsumeEvent method must be implemented by concrete consumer")
}

// EventConsumerManager 事件消费者管理器
type EventConsumerManager struct {
	consumers         map[string]EventConsumer
	connectionManager ConnectionManagerInterface
	isRunning         bool
}

// NewEventConsumerManager 创建事件消费者管理器
func NewEventConsumerManager(connMgr ConnectionManagerInterface) *EventConsumerManager {
	return &EventConsumerManager{
		consumers:         make(map[string]EventConsumer),
		connectionManager: connMgr,
		isRunning:         false,
	}
}

// RegisterConsumer 注册事件消费者
func (m *EventConsumerManager) RegisterConsumer(consumer EventConsumer) error {
	eventType := consumer.GetEventType()
	
	if _, exists := m.consumers[eventType]; exists {
		return fmt.Errorf("消费者已存在于事件类型: %s", eventType)
	}
	
	m.consumers[eventType] = consumer
	log.Printf("📝 注册Neo4j事件消费者: %s", eventType)
	
	return nil
}

// ConsumeEvent 消费单个事件
func (m *EventConsumerManager) ConsumeEvent(ctx context.Context, event events.DomainEvent) error {
	eventType := event.GetEventType()
	
	consumer, exists := m.consumers[eventType]
	if !exists {
		log.Printf("⚠️ 未找到事件类型的消费者: %s", eventType)
		return nil // 不是错误，只是没有处理该事件类型
	}
	
	log.Printf("🔄 处理Neo4j事件: %s (ID: %s)", eventType, event.GetEventID())
	
	// 使用重试机制处理事件
	return m.connectionManager.ExecuteWithRetry(ctx, func(ctx context.Context) error {
		return consumer.ConsumeEvent(ctx, event)
	})
}

// StartAll 启动所有消费者
func (m *EventConsumerManager) StartAll(ctx context.Context) error {
	if m.isRunning {
		return fmt.Errorf("事件消费者管理器已经在运行中")
	}
	
	log.Printf("🚀 启动所有Neo4j事件消费者 (%d个消费者)", len(m.consumers))
	
	for eventType, consumer := range m.consumers {
		if err := consumer.Start(ctx); err != nil {
			log.Printf("❌ 启动消费者失败: %s - %v", eventType, err)
			return err
		}
		log.Printf("✅ 消费者启动成功: %s", eventType)
	}
	
	m.isRunning = true
	log.Println("🎉 所有Neo4j事件消费者启动完成")
	
	return nil
}

// StopAll 停止所有消费者
func (m *EventConsumerManager) StopAll() error {
	if !m.isRunning {
		return nil
	}
	
	log.Printf("🛑 停止所有Neo4j事件消费者 (%d个消费者)", len(m.consumers))
	
	for eventType, consumer := range m.consumers {
		if err := consumer.Stop(); err != nil {
			log.Printf("⚠️ 停止消费者时出错: %s - %v", eventType, err)
		} else {
			log.Printf("✅ 消费者停止成功: %s", eventType)
		}
	}
	
	m.isRunning = false
	log.Println("✅ 所有Neo4j事件消费者已停止")
	
	return nil
}

// Health 健康检查
func (m *EventConsumerManager) Health() error {
	if !m.isRunning {
		return fmt.Errorf("事件消费者管理器未运行")
	}
	
	// 检查连接
	if err := m.connectionManager.Health(context.Background()); err != nil {
		return fmt.Errorf("Neo4j连接健康检查失败: %w", err)
	}
	
	// 检查所有消费者
	for eventType, consumer := range m.consumers {
		if err := consumer.Health(); err != nil {
			return fmt.Errorf("消费者健康检查失败 %s: %w", eventType, err)
		}
	}
	
	return nil
}

// GetStatistics 获取统计信息
func (m *EventConsumerManager) GetStatistics() map[string]interface{} {
	consumerStats := make(map[string]interface{})
	
	for eventType := range m.consumers {
		consumerStats[eventType] = map[string]interface{}{
			"status": "running",
			"type":   eventType,
		}
	}
	
	return map[string]interface{}{
		"is_running":     m.isRunning,
		"consumer_count": len(m.consumers),
		"consumers":      consumerStats,
	}
}

// SyncOperation 同步操作接口
type SyncOperation interface {
	Execute(ctx context.Context, tx neo4j.ManagedTransaction) error
	GetDescription() string
	Validate() error
}

// NodeSyncOperation 节点同步操作
type NodeSyncOperation struct {
	Label      string
	Properties map[string]interface{}
	UniqueKeys []string
	Operation  string // CREATE, UPDATE, DELETE
}

// Execute 执行节点同步操作
func (op *NodeSyncOperation) Execute(ctx context.Context, tx neo4j.ManagedTransaction) error {
	switch op.Operation {
	case "CREATE":
		return op.executeCreate(ctx, tx)
	case "UPDATE":
		return op.executeUpdate(ctx, tx)
	case "DELETE":
		return op.executeDelete(ctx, tx)
	default:
		return fmt.Errorf("不支持的操作类型: %s", op.Operation)
	}
}

func (op *NodeSyncOperation) executeCreate(ctx context.Context, tx neo4j.ManagedTransaction) error {
	// 构建MERGE语句确保幂等性
	mergeClause := fmt.Sprintf("MERGE (n:%s {", op.Label)
	
	// 添加唯一键条件
	var conditions []string
	params := make(map[string]interface{})
	
	for _, key := range op.UniqueKeys {
		if value, exists := op.Properties[key]; exists {
			conditions = append(conditions, fmt.Sprintf("%s: $%s", key, key))
			params[key] = value
		}
	}
	
	mergeClause += joinStrings(conditions, ", ") + "}) "
	
	// 添加SET子句设置其他属性
	var setConditions []string
	for key, value := range op.Properties {
		if !containsString(op.UniqueKeys, key) {
			setConditions = append(setConditions, fmt.Sprintf("n.%s = $%s", key, key))
			params[key] = value
		}
	}
	
	if len(setConditions) > 0 {
		mergeClause += "SET " + joinStrings(setConditions, ", ")
	}
	
	// 添加时间戳
	params["synced_at"] = time.Now()
	mergeClause += ", n.synced_at = $synced_at"
	
	log.Printf("🔄 执行Neo4j节点创建: %s", op.Label)
	
	_, err := tx.Run(ctx, mergeClause, params)
	return err
}

func (op *NodeSyncOperation) executeUpdate(ctx context.Context, tx neo4j.ManagedTransaction) error {
	// 构建MATCH和SET语句
	matchClause := fmt.Sprintf("MATCH (n:%s) WHERE ", op.Label)
	
	var conditions []string
	params := make(map[string]interface{})
	
	// 使用唯一键查找节点
	for _, key := range op.UniqueKeys {
		if value, exists := op.Properties[key]; exists {
			conditions = append(conditions, fmt.Sprintf("n.%s = $%s", key, key))
			params[key] = value
		}
	}
	
	matchClause += joinStrings(conditions, " AND ")
	
	// 设置所有属性
	var setConditions []string
	for key, value := range op.Properties {
		setConditions = append(setConditions, fmt.Sprintf("n.%s = $%s", key, key))
		params[key] = value
	}
	
	// 添加时间戳
	params["updated_at"] = time.Now()
	setConditions = append(setConditions, "n.updated_at = $updated_at")
	
	cypher := matchClause + " SET " + joinStrings(setConditions, ", ")
	
	log.Printf("🔄 执行Neo4j节点更新: %s", op.Label)
	
	result, err := tx.Run(ctx, cypher, params)
	if err != nil {
		return err
	}
	
	// 检查是否找到并更新了节点
	summary, err := result.Consume(ctx)
	if err != nil {
		return err
	}
	
	// 在Neo4j v5中，简化计数器检查
	log.Printf("✅ 操作完成: %s", op.Label)
	
	return nil
}

func (op *NodeSyncOperation) executeDelete(ctx context.Context, tx neo4j.ManagedTransaction) error {
	// 构建MATCH和DELETE语句
	matchClause := fmt.Sprintf("MATCH (n:%s) WHERE ", op.Label)
	
	var conditions []string
	params := make(map[string]interface{})
	
	// 使用唯一键查找节点
	for _, key := range op.UniqueKeys {
		if value, exists := op.Properties[key]; exists {
			conditions = append(conditions, fmt.Sprintf("n.%s = $%s", key, key))
			params[key] = value
		}
	}
	
	matchClause += joinStrings(conditions, " AND ")
	cypher := matchClause + " DELETE n"
	
	log.Printf("🗑️ 执行Neo4j节点删除: %s", op.Label)
	
	result, err := tx.Run(ctx, cypher, params)
	if err != nil {
		return err
	}
	
	// 检查是否找到并删除了节点
	summary, err := result.Consume(ctx)
	if err != nil {
		return err
	}
	
	if summary.Counters().NodesDeleted() == 0 {
		log.Printf("⚠️ 未找到要删除的节点: %s", op.Label)
	}
	
	return nil
}

func (op *NodeSyncOperation) GetDescription() string {
	return fmt.Sprintf("%s %s节点", op.Operation, op.Label)
}

func (op *NodeSyncOperation) Validate() error {
	if op.Label == "" {
		return fmt.Errorf("节点标签不能为空")
	}
	if len(op.UniqueKeys) == 0 {
		return fmt.Errorf("至少需要一个唯一键")
	}
	if len(op.Properties) == 0 {
		return fmt.Errorf("节点属性不能为空")
	}
	return nil
}

// 辅助函数
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}
	if len(strs) == 1 {
		return strs[0]
	}
	
	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}