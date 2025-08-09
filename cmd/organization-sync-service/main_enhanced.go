package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Shopify/sarama"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/go-redis/redis/v8"
)

// ===== 配置管理 (统一配置，消除硬编码) =====
type SyncConfig struct {
	KafkaBrokers   []string `env:"KAFKA_BROKERS"`
	ConsumerGroup  string   `env:"CONSUMER_GROUP"`
	Neo4jURI      string   `env:"NEO4J_URI"`
	Neo4jUser     string   `env:"NEO4J_USER"`
	Neo4jPassword string   `env:"NEO4J_PASSWORD"`
	RedisURL      string   `env:"REDIS_URL"`
	TenantID      string   `env:"TENANT_ID"`
}

func LoadConfig() *SyncConfig {
	return &SyncConfig{
		KafkaBrokers:   []string{getEnv("KAFKA_BROKERS", "localhost:9092")},
		ConsumerGroup:  getEnv("CONSUMER_GROUP", "organization-sync-group"),
		Neo4jURI:      getEnv("NEO4J_URI", "neo4j://localhost:7687"),
		Neo4jUser:     getEnv("NEO4J_USER", "neo4j"),
		Neo4jPassword: getEnv("NEO4J_PASSWORD", "password"),
		RedisURL:      getEnv("REDIS_URL", "redis://localhost:6379"),
		TenantID:      getEnv("TENANT_ID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ===== 事件模型 (统一定义，消除重复) =====
type CDCEvent struct {
	Before    *OrganizationData `json:"before"`
	After     *OrganizationData `json:"after"`
	Source    CDCSource         `json:"source"`
	Op        string            `json:"op"`
	TsMs      int64             `json:"ts_ms"`
	EventType string            `json:"-"` // 内部使用
}

type OrganizationData struct {
	TenantID    *string `json:"tenant_id"`
	Code        *string `json:"code"`
	ParentCode  *string `json:"parent_code"`
	Name        *string `json:"name"`
	UnitType    *string `json:"unit_type"`
	Status      *string `json:"status"`
	Level       *int    `json:"level"`
	Path        *string `json:"path"`
	SortOrder   *int    `json:"sort_order"`
	Description *string `json:"description"`
	CreatedAt   *string `json:"created_at"`
	UpdatedAt   *string `json:"updated_at"`
}

type CDCSource struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	Db        string `json:"db"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
}

// ===== 数据转换器 (消除重复if-else模式) =====
type DataTransformer struct {
	tenantID string
}

func NewDataTransformer(tenantID string) *DataTransformer {
	return &DataTransformer{tenantID: tenantID}
}

func (dt *DataTransformer) ToNeo4j(data *OrganizationData) map[string]interface{} {
	params := make(map[string]interface{})
	
	// 统一的字段转换逻辑，消除140+行重复代码
	dt.setField(params, "tenant_id", data.TenantID, dt.tenantID)
	dt.setField(params, "code", data.Code, "")
	dt.setField(params, "parent_code", data.ParentCode, nil)
	dt.setField(params, "name", data.Name, "")
	dt.setField(params, "unit_type", data.UnitType, "DEPARTMENT")
	dt.setField(params, "status", data.Status, "ACTIVE")
	dt.setField(params, "level", data.Level, 1)
	dt.setField(params, "path", data.Path, "/")
	dt.setField(params, "sort_order", data.SortOrder, 0)
	dt.setField(params, "description", data.Description, "")
	dt.setField(params, "created_at", data.CreatedAt, time.Now().Format(time.RFC3339))
	dt.setField(params, "updated_at", data.UpdatedAt, time.Now().Format(time.RFC3339))
	
	return params
}

func (dt *DataTransformer) setField(params map[string]interface{}, key string, value interface{}, defaultValue interface{}) {
	switch v := value.(type) {
	case *string:
		if v != nil {
			params[key] = *v
		} else {
			params[key] = defaultValue
		}
	case *int:
		if v != nil {
			params[key] = *v
		} else {
			params[key] = defaultValue
		}
	default:
		params[key] = defaultValue
	}
}

// ===== 精确缓存失效器 (替代cache:*暴力清空) =====
type PreciseCacheInvalidator struct {
	redis  *redis.Client
	logger *log.Logger
}

func NewPreciseCacheInvalidator(redisURL string, logger *log.Logger) (*PreciseCacheInvalidator, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("解析Redis URL失败: %w", err)
	}
	
	client := redis.NewClient(opts)
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}
	
	return &PreciseCacheInvalidator{
		redis:  client,
		logger: logger,
	}, nil
}

func (pci *PreciseCacheInvalidator) InvalidateByEvent(ctx context.Context, event CDCEvent) error {
	var tenantID, code string
	
	// 提取租户和代码信息
	if event.After != nil {
		if event.After.TenantID != nil {
			tenantID = *event.After.TenantID
		}
		if event.After.Code != nil {
			code = *event.After.Code
		}
	} else if event.Before != nil {
		if event.Before.TenantID != nil {
			tenantID = *event.Before.TenantID
		}
		if event.Before.Code != nil {
			code = *event.Before.Code
		}
	}
	
	if tenantID == "" || code == "" {
		pci.logger.Printf("⚠️ 无法提取租户或代码信息，跳过缓存失效")
		return nil
	}
	
	// 精确失效策略，替代暴力cache:*
	patterns := []string{
		fmt.Sprintf("cache:org:%s:%s", tenantID, code),           // 单个组织缓存
		fmt.Sprintf("cache:hierarchy:%s:%s*", tenantID, code),   // 层级相关缓存
		fmt.Sprintf("cache:stats:%s", tenantID),                 // 统计缓存
		fmt.Sprintf("cache:list:%s*", tenantID),                 // 列表缓存
	}
	
	totalInvalidated := 0
	for _, pattern := range patterns {
		keys, err := pci.redis.Keys(ctx, pattern).Result()
		if err != nil {
			pci.logger.Printf("❌ 查找缓存键失败 [%s]: %v", pattern, err)
			continue
		}
		
		if len(keys) > 0 {
			if err := pci.redis.Del(ctx, keys...).Err(); err != nil {
				pci.logger.Printf("❌ 删除缓存失败 [%s]: %v", pattern, err)
				continue
			}
			totalInvalidated += len(keys)
			pci.logger.Printf("🗑️ 精确失效缓存: %d keys [%s]", len(keys), pattern)
		}
	}
	
	pci.logger.Printf("✅ 总共精确失效缓存: %d keys for org %s", totalInvalidated, code)
	return nil
}

// ===== 事件处理器 (清晰的职责分离) =====
type EnhancedEventHandler struct {
	neo4j       neo4j.DriverWithContext
	cache       *PreciseCacheInvalidator
	transformer *DataTransformer
	logger      *log.Logger
}

func NewEnhancedEventHandler(neo4jURI, neo4jUser, neo4jPassword string, cache *PreciseCacheInvalidator, transformer *DataTransformer, logger *log.Logger) (*EnhancedEventHandler, error) {
	driver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth(neo4jUser, neo4jPassword, ""))
	if err != nil {
		return nil, fmt.Errorf("创建Neo4j驱动失败: %w", err)
	}
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := driver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("Neo4j连接验证失败: %w", err)
	}
	
	return &EnhancedEventHandler{
		neo4j:       driver,
		cache:       cache,
		transformer: transformer,
		logger:      logger,
	}, nil
}

// 清晰的事件分派，替代140+行巨型函数
func (eh *EnhancedEventHandler) HandleEvent(ctx context.Context, event CDCEvent) error {
	start := time.Now()
	
	eh.logger.Printf("📨 处理CDC事件: op=%s, code=%s", event.Op, eh.getCodeFromEvent(event))
	
	var err error
	switch event.Op {
	case "c": // Create
		err = eh.handleCreate(ctx, event)
	case "u": // Update
		err = eh.handleUpdate(ctx, event)
	case "d": // Delete
		err = eh.handleDelete(ctx, event)
	case "r": // Read (initial snapshot)
		err = eh.handleCreate(ctx, event) // 处理方式同创建
	default:
		eh.logger.Printf("⚠️ 未支持的操作类型: %s", event.Op)
		return nil
	}
	
	if err != nil {
		eh.logger.Printf("❌ 事件处理失败: %v", err)
		return err
	}
	
	// 精确缓存失效
	if cacheErr := eh.cache.InvalidateByEvent(ctx, event); cacheErr != nil {
		eh.logger.Printf("⚠️ 缓存失效失败: %v", cacheErr)
		// 缓存失效失败不应阻止整个流程
	}
	
	duration := time.Since(start)
	eh.logger.Printf("✅ 事件处理成功: op=%s, 耗时=%v", event.Op, duration)
	
	return nil
}

func (eh *EnhancedEventHandler) handleCreate(ctx context.Context, event CDCEvent) error {
	if event.After == nil {
		return fmt.Errorf("创建事件缺少after数据")
	}
	
	params := eh.transformer.ToNeo4j(event.After)
	
	query := `
		MERGE (o:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET o.name = $name,
		    o.unit_type = $unit_type,
		    o.status = $status,
		    o.level = $level,
		    o.path = $path,
		    o.sort_order = $sort_order,
		    o.description = $description,
		    o.created_at = $created_at,
		    o.updated_at = $updated_at,
		    o.parent_code = $parent_code
		RETURN o.code as code
	`
	
	session := eh.neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	
	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("Neo4j创建失败: %w", err)
	}
	
	if result.Next(ctx) {
		code := result.Record().Values[0].(string)
		eh.logger.Printf("✨ Neo4j组织创建成功: %s", code)
	}
	
	return result.Err()
}

func (eh *EnhancedEventHandler) handleUpdate(ctx context.Context, event CDCEvent) error {
	if event.After == nil {
		return fmt.Errorf("更新事件缺少after数据")
	}
	
	params := eh.transformer.ToNeo4j(event.After)
	
	query := `
		MATCH (o:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		SET o.name = $name,
		    o.unit_type = $unit_type,
		    o.status = $status,
		    o.level = $level,
		    o.path = $path,
		    o.sort_order = $sort_order,
		    o.description = $description,
		    o.updated_at = $updated_at,
		    o.parent_code = $parent_code
		RETURN o.code as code
	`
	
	session := eh.neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	
	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("Neo4j更新失败: %w", err)
	}
	
	if result.Next(ctx) {
		code := result.Record().Values[0].(string)
		eh.logger.Printf("🔄 Neo4j组织更新成功: %s", code)
	}
	
	return result.Err()
}

func (eh *EnhancedEventHandler) handleDelete(ctx context.Context, event CDCEvent) error {
	if event.Before == nil {
		return fmt.Errorf("删除事件缺少before数据")
	}
	
	params := eh.transformer.ToNeo4j(event.Before)
	
	query := `
		MATCH (o:OrganizationUnit {code: $code, tenant_id: $tenant_id})
		DETACH DELETE o
		RETURN count(o) as deleted_count
	`
	
	session := eh.neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)
	
	result, err := session.Run(ctx, query, params)
	if err != nil {
		return fmt.Errorf("Neo4j删除失败: %w", err)
	}
	
	if result.Next(ctx) {
		count := result.Record().Values[0].(int64)
		eh.logger.Printf("🗑️ Neo4j组织删除成功: %d条记录", count)
	}
	
	return result.Err()
}

func (eh *EnhancedEventHandler) getCodeFromEvent(event CDCEvent) string {
	if event.After != nil && event.After.Code != nil {
		return *event.After.Code
	}
	if event.Before != nil && event.Before.Code != nil {
		return *event.Before.Code
	}
	return "unknown"
}

func (eh *EnhancedEventHandler) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return eh.neo4j.Close(ctx)
}

// ===== Kafka消费者 (企业级容错处理) =====
type ConsumerGroupHandler struct {
	handler *EnhancedEventHandler
	logger  *log.Logger
}

func (h *ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h *ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			if message == nil {
				return nil
			}
			
			// 解析CDC事件
			var event CDCEvent
			if err := json.Unmarshal(message.Value, &event); err != nil {
				h.logger.Printf("❌ JSON解析失败: %v", err)
				session.MarkMessage(message, "")
				continue
			}
			
			// 处理事件
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := h.handler.HandleEvent(ctx, event); err != nil {
				h.logger.Printf("❌ 事件处理失败: %v", err)
				cancel()
				// 注意：在生产环境中可能需要重试逻辑或死信队列
				continue
			}
			cancel()
			
			// 标记消息已处理
			session.MarkMessage(message, "")
			
		case <-session.Context().Done():
			return nil
		}
	}
}

// ===== 主程序 =====
func main() {
	logger := log.New(os.Stdout, "[SYNC-ENHANCED] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 启动增强版组织同步服务...")
	
	// 加载配置
	config := LoadConfig()
	logger.Printf("📋 配置加载完成: Kafka=%v, TenantID=%s", config.KafkaBrokers, config.TenantID)
	
	// 创建组件
	transformer := NewDataTransformer(config.TenantID)
	
	cache, err := NewPreciseCacheInvalidator(config.RedisURL, logger)
	if err != nil {
		logger.Fatalf("❌ 创建缓存失效器失败: %v", err)
	}
	
	handler, err := NewEnhancedEventHandler(config.Neo4jURI, config.Neo4jUser, config.Neo4jPassword, cache, transformer, logger)
	if err != nil {
		logger.Fatalf("❌ 创建事件处理器失败: %v", err)
	}
	defer handler.Close()
	
	// 配置Kafka消费者
	saramaConfig := sarama.NewConfig()
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	
	client, err := sarama.NewConsumerGroup(config.KafkaBrokers, config.ConsumerGroup, saramaConfig)
	if err != nil {
		logger.Fatalf("❌ 创建Kafka消费者失败: %v", err)
	}
	defer client.Close()
	
	// 启动消费者
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	consumerHandler := &ConsumerGroupHandler{
		handler: handler,
		logger:  logger,
	}
	
	go func() {
		for {
			topics := []string{"organization_db.public.organization_units"}
			err := client.Consume(ctx, topics, consumerHandler)
			if err != nil {
				logger.Printf("❌ 消费Kafka消息失败: %v", err)
			}
			
			// 检查上下文是否被取消
			if ctx.Err() != nil {
				return
			}
		}
	}()
	
	logger.Println("✅ 增强版组织同步服务启动成功")
	
	// 优雅停机
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	<-sigterm
	
	logger.Println("🛑 收到停机信号，正在优雅关闭...")
	cancel()
	logger.Println("👋 增强版组织同步服务已停止")
}