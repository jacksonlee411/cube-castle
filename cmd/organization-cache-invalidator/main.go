package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/redis/go-redis/v9"
)

// CDC事件模型
type CDCOrganizationEvent struct {
	Before *CDCOrganizationData `json:"before"`
	After  *CDCOrganizationData `json:"after"`
	Source CDCSource            `json:"source"`
	Op     string               `json:"op"` // c, u, d, r
	TsMs   int64                `json:"ts_ms"`
}

type CDCOrganizationData struct {
	ID         *string `json:"id"`
	TenantID   *string `json:"tenant_id"`
	Code       *string `json:"code"`
	ParentCode *string `json:"parent_code"`
	Name       *string `json:"name"`
	UnitType   *string `json:"unit_type"`
	Status     *string `json:"status"`
	Level      *int    `json:"level"`
	Path       *string `json:"path"`
	SortOrder  *int    `json:"sort_order"`
	Description *string `json:"description"`
	CreatedAt  *time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type CDCSource struct {
	Version   string `json:"version"`
	Connector string `json:"connector"`
	Name      string `json:"name"`
	TsMs      int64  `json:"ts_ms"`
	Snapshot  string `json:"snapshot"`
	DB        string `json:"db"`
	Schema    string `json:"schema"`
	Table     string `json:"table"`
	TxID      int64  `json:"txId"`
	LSN       int64  `json:"lsn"`
}

// 缓存失效服务
type CacheInvalidator struct {
	redisClient *redis.Client
	consumer    *kafka.Consumer
	logger      *log.Logger
}

func NewCacheInvalidator(redisAddr, redisPassword string, kafkaBrokers []string, groupID string, logger *log.Logger) (*CacheInvalidator, error) {
	// Redis连接
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	// 测试Redis连接
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	// Kafka消费者配置
	config := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(kafkaBrokers, ","),
		"group.id":          groupID,
		"auto.offset.reset": "latest",
		"enable.auto.commit": true,
		"auto.commit.interval.ms": 1000,
	}

	consumer, err := kafka.NewConsumer(config)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	return &CacheInvalidator{
		redisClient: redisClient,
		consumer:    consumer,
		logger:      logger,
	}, nil
}

// 生成缓存键 - 与GraphQL服务保持一致
func (c *CacheInvalidator) getCacheKey(operation string, params ...interface{}) string {
	h := md5.New()
	h.Write([]byte(fmt.Sprintf("org:%s:%v", operation, params)))
	return fmt.Sprintf("cache:%x", h.Sum(nil))
}

// 失效相关缓存
func (c *CacheInvalidator) invalidateOrganizationCaches(ctx context.Context, tenantID string, affectedCode string) error {
	// 需要失效的缓存模式
	patterns := []string{
		"cache:*", // 失效所有组织相关缓存，确保数据一致性
	}

	totalInvalidated := 0
	for _, pattern := range patterns {
		keys, err := c.redisClient.Keys(ctx, pattern).Result()
		if err != nil {
			c.logger.Printf("获取缓存键失败，模式: %s, 错误: %v", pattern, err)
			continue
		}

		if len(keys) > 0 {
			deleted, err := c.redisClient.Del(ctx, keys...).Result()
			if err != nil {
				c.logger.Printf("删除缓存失败，键数量: %d, 错误: %v", len(keys), err)
				continue
			}
			totalInvalidated += int(deleted)
			c.logger.Printf("缓存失效成功 - 模式: %s, 删除: %d 个缓存项", pattern, deleted)
		}
	}

	if totalInvalidated > 0 {
		c.logger.Printf("✅ 缓存失效完成 - 租户: %s, 影响组织: %s, 总计失效: %d 个缓存项", 
			tenantID, affectedCode, totalInvalidated)
	} else {
		c.logger.Printf("ℹ️ 未找到需要失效的缓存 - 租户: %s, 影响组织: %s", tenantID, affectedCode)
	}

	return nil
}

// 处理CDC事件
func (c *CacheInvalidator) processCDCEvent(ctx context.Context, event CDCOrganizationEvent) error {
	var tenantID, code string
	
	// 根据操作类型获取租户ID和组织代码
	switch event.Op {
	case "c", "u": // CREATE, UPDATE
		if event.After == nil {
			return fmt.Errorf("CDC %s事件缺少after数据", event.Op)
		}
		if event.After.TenantID != nil {
			tenantID = *event.After.TenantID
		}
		if event.After.Code != nil {
			code = *event.After.Code
		}
	case "d": // DELETE
		if event.Before == nil {
			return fmt.Errorf("CDC DELETE事件缺少before数据")
		}
		if event.Before.TenantID != nil {
			tenantID = *event.Before.TenantID
		}
		if event.Before.Code != nil {
			code = *event.Before.Code
		}
	default:
		c.logger.Printf("⚠️ 未知的CDC操作类型: %s", event.Op)
		return nil
	}

	if tenantID == "" || code == "" {
		c.logger.Printf("⚠️ CDC事件缺少必要信息 - 租户ID: %s, 组织代码: %s", tenantID, code)
		return nil
	}

	c.logger.Printf("🔄 处理CDC事件 - 操作: %s, 租户: %s, 组织: %s", event.Op, tenantID, code)
	
	// 失效相关缓存
	return c.invalidateOrganizationCaches(ctx, tenantID, code)
}

// 处理Kafka消息
func (c *CacheInvalidator) processMessage(ctx context.Context, msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic

	// 只处理组织单元CDC事件
	if topic != "organization_db.public.organization_units" {
		return nil
	}

	c.logger.Printf("📨 收到CDC事件消息 - Topic: %s, Partition: %d, Offset: %d", 
		topic, msg.TopicPartition.Partition, msg.TopicPartition.Offset)

	// 解析Debezium消息格式
	var debeziumMsg struct {
		Payload CDCOrganizationEvent `json:"payload"`
	}
	if err := json.Unmarshal(msg.Value, &debeziumMsg); err != nil {
		return fmt.Errorf("反序列化Debezium消息失败: %w", err)
	}

	return c.processCDCEvent(ctx, debeziumMsg.Payload)
}

// 开始消费
func (c *CacheInvalidator) StartConsuming(ctx context.Context) error {
	// 订阅CDC主题
	topics := []string{"organization_db.public.organization_units"}
	if err := c.consumer.SubscribeTopics(topics, nil); err != nil {
		return fmt.Errorf("订阅Kafka主题失败: %w", err)
	}

	c.logger.Printf("🚀 缓存失效服务开始运行...")
	c.logger.Printf("监听主题: %v", topics)

	for {
		select {
		case <-ctx.Done():
			c.logger.Println("收到停止信号，停止消费...")
			return nil
		default:
			msg, err := c.consumer.ReadMessage(1000)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				c.logger.Printf("消费消息失败: %v", err)
				continue
			}

			if err := c.processMessage(ctx, msg); err != nil {
				c.logger.Printf("处理消息失败: %v", err)
			}
		}
	}
}

// 关闭资源
func (c *CacheInvalidator) Close() error {
	var errs []error
	
	if c.consumer != nil {
		if err := c.consumer.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Kafka消费者失败: %w", err))
		}
	}
	
	if c.redisClient != nil {
		if err := c.redisClient.Close(); err != nil {
			errs = append(errs, fmt.Errorf("关闭Redis连接失败: %w", err))
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("关闭资源时发生错误: %v", errs)
	}
	
	return nil
}

func main() {
	logger := log.New(os.Stdout, "[CACHE-INVALIDATOR] ", log.LstdFlags)

	// 配置参数
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	
	redisPassword := os.Getenv("REDIS_PASSWORD")
	
	kafkaBrokers := []string{"localhost:9092"}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		kafkaBrokers = strings.Split(brokers, ",")
	}
	
	groupID := "cache-invalidator-group"

	// 创建缓存失效服务
	invalidator, err := NewCacheInvalidator(redisAddr, redisPassword, kafkaBrokers, groupID, logger)
	if err != nil {
		log.Fatalf("创建缓存失效服务失败: %v", err)
	}
	defer invalidator.Close()

	// 创建上下文处理优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	
	// 优雅关闭
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		
		logger.Println("正在关闭缓存失效服务...")
		cancel()
	}()

	logger.Println("🚀 组织缓存失效服务启动成功")
	
	// 启动健康检查服务器
	go startHealthServer(logger)
	
	// 开始消费
	if err := invalidator.StartConsuming(ctx); err != nil {
		log.Fatalf("消费失败: %v", err)
	}
	
	logger.Println("缓存失效服务已关闭")
}

// 健康检查服务器
func startHealthServer(logger *log.Logger) {
	mux := http.NewServeMux()
	
	// 健康检查端点
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"service": "organization-cache-invalidator",
			"status": "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
			"features": []string{
				"精确缓存失效",
				"Redis集成", 
				"Kafka消息消费",
				"CDC事件处理",
			},
		}
		json.NewEncoder(w).Encode(response)
	})
	
	// 指标端点
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("# Cache invalidator metrics\ncache_invalidator_status 1\n"))
	})
	
	server := &http.Server{
		Addr:    ":8082",
		Handler: mux,
	}
	
	logger.Printf("🔍 健康检查服务器启动 - 端口 8082")
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Printf("❌ 健康检查服务器错误: %v", err)
	}
}