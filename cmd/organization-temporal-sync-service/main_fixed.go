package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
)

// ===== 配置 =====

type Config struct {
	KafkaBootstrapServers string
	KafkaTopic           string
	Neo4jURI             string
	Neo4jUsername        string 
	Neo4jPassword        string
	RedisAddr            string
	TenantID             string
	Port                 string
}

func loadConfig() *Config {
	return &Config{
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaTopic:           getEnv("KAFKA_TOPIC", "organization_db.public.organization_units"),
		Neo4jURI:             getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:        getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:        getEnv("NEO4J_PASSWORD", "password"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		TenantID:             getEnv("DEFAULT_TENANT_ID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"),
		Port:                 getEnv("PORT", "8092"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ===== 数据模型 =====

type OrganizationUnit struct {
	Code        string  `json:"code"`
	ParentCode  *string `json:"parent_code"`
	TenantID    string  `json:"tenant_id"`
	Name        string  `json:"name"`
	UnitType    string  `json:"unit_type"`
	Status      string  `json:"status"`
	Level       int     `json:"level"`
	Path        string  `json:"path"`
	SortOrder   int     `json:"sort_order"`
	Description string  `json:"description"`
	Profile     string  `json:"profile"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
	Operation   string  `json:"operation"`
}

// ===== 服务 =====

type SyncService struct {
	config *Config
	kafka  *kafka.Consumer
	neo4j  neo4j.DriverWithContext
	redis  *redis.Client
	logger *log.Logger
}

func NewSyncService(config *Config) (*SyncService, error) {
	logger := log.New(os.Stdout, "[SYNC] ", log.LstdFlags|log.Lshortfile)
	
	// 初始化Kafka消费者
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": config.KafkaBootstrapServers,
		"group.id":          "temporal-sync-fixed",
		"auto.offset.reset": "latest",
		"enable.auto.commit": true,
	}
	
	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}
	
	// 初始化Neo4j驱动
	neo4jDriver, err := neo4j.NewDriverWithContext(
		config.Neo4jURI, 
		neo4j.BasicAuth(config.Neo4jUsername, config.Neo4jPassword, ""),
	)
	if err != nil {
		return nil, fmt.Errorf("创建Neo4j驱动失败: %w", err)
	}
	
	// 验证Neo4j连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := neo4jDriver.VerifyConnectivity(ctx); err != nil {
		return nil, fmt.Errorf("Neo4j连接验证失败: %w", err)
	}
	
	// 初始化Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr: config.RedisAddr,
		DB:   0,
	})
	
	// 验证Redis连接
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}
	
	service := &SyncService{
		config: config,
		kafka:  consumer,
		neo4j:  neo4jDriver,
		redis:  redisClient,
		logger: logger,
	}
	
	logger.Printf("🚀 时态同步服务初始化完成")
	return service, nil
}

func (s *SyncService) Start(ctx context.Context) error {
	// 订阅Kafka主题
	if err := s.kafka.SubscribeTopics([]string{s.config.KafkaTopic}, nil); err != nil {
		return fmt.Errorf("订阅Kafka主题失败: %w", err)
	}
	
	s.logger.Printf("📡 开始监听Kafka主题: %s", s.config.KafkaTopic)
	
	// 主消费循环
	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("⏹️ 收到停止信号")
			return nil
		default:
			msg, err := s.kafka.ReadMessage(1 * time.Second)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				s.logger.Printf("❌ 读取Kafka消息失败: %v", err)
				continue
			}
			
			if err := s.processMessage(ctx, msg); err != nil {
				s.logger.Printf("❌ 处理消息失败: %v", err)
			}
		}
	}
}

func (s *SyncService) processMessage(ctx context.Context, msg *kafka.Message) error {
	// 解析Debezium消息
	org, err := parseDebeziumMessage(string(msg.Value))
	if err != nil {
		return fmt.Errorf("解析Debezium消息失败: %w", err)
	}

	s.logger.Printf("🔄 处理组织事件: %s [%s] %s", org.Operation, org.Code, org.Name)
	
	// 同步到Neo4j
	if err := s.syncToNeo4j(ctx, org); err != nil {
		return fmt.Errorf("同步到Neo4j失败: %w", err)
	}
	
	// 清除缓存
	if err := s.invalidateCache(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 缓存清除失败: %v", err)
	}
	
	s.logger.Printf("✅ 同步完成: %s", org.Code)
	return nil
}

// 解析Debezium消息
func parseDebeziumMessage(message string) (*OrganizationUnit, error) {
	var debeziumEvent map[string]interface{}
	if err := json.Unmarshal([]byte(message), &debeziumEvent); err != nil {
		return nil, fmt.Errorf("解析Debezium事件失败: %v", err)
	}

	// 提取payload
	payload, ok := debeziumEvent["payload"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("无效的payload格式")
	}

	// 提取操作类型
	op, _ := payload["op"].(string)
	
	var after map[string]interface{}
	
	// 根据操作类型提取数据
	switch op {
	case "c", "r", "u": // CREATE, READ, UPDATE
		after, ok = payload["after"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("无效的after数据格式，操作: %s", op)
		}
	case "d": // DELETE
		after, ok = payload["before"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("无效的before数据格式，操作: %s", op)
		}
	default:
		return nil, fmt.Errorf("不支持的操作类型: %s", op)
	}

	// 转换为组织单元结构
	org := &OrganizationUnit{
		Code:        getString(after, "code"),
		TenantID:    getString(after, "tenant_id"),
		Name:        getString(after, "name"),
		UnitType:    getString(after, "unit_type"),
		Status:      getString(after, "status"),
		Level:       int(getFloat64(after, "level")),
		Path:        getString(after, "path"),
		SortOrder:   int(getFloat64(after, "sort_order")),
		Description: getString(after, "description"),
		Profile:     getString(after, "profile"),
		CreatedAt:   getString(after, "created_at"),
		UpdatedAt:   getString(after, "updated_at"),
		Operation:   op,
	}
	
	// 处理可为空的parent_code
	if parentCode := getString(after, "parent_code"); parentCode != "" {
		org.ParentCode = &parentCode
	}

	return org, nil
}

// 辅助函数：安全地提取字符串值
func getString(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：安全地提取数字值
func getFloat64(data map[string]interface{}, key string) float64 {
	if value, exists := data[key]; exists && value != nil {
		switch v := value.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int32:
			return float64(v)
		}
	}
	return 0
}

func (s *SyncService) syncToNeo4j(ctx context.Context, org *OrganizationUnit) error {
	session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		switch org.Operation {
		case "c", "r", "u": // CREATE, READ, UPDATE
			query := `
			MERGE (org:Organization {tenant_id: $tenant_id, code: $code})
			SET org.name = $name,
				org.parent_code = $parent_code,
				org.unit_type = $unit_type,
				org.status = $status,
				org.level = $level,
				org.path = $path,
				org.sort_order = $sort_order,
				org.description = $description,
				org.profile = $profile,
				org.created_at = datetime($created_at),
				org.updated_at = datetime($updated_at),
				org.synced_at = datetime()
			RETURN org.code as code`
			
			_, err := tx.Run(ctx, query, map[string]interface{}{
				"tenant_id":    org.TenantID,
				"code":         org.Code,
				"name":         org.Name,
				"parent_code":  org.ParentCode,
				"unit_type":    org.UnitType,
				"status":       org.Status,
				"level":        org.Level,
				"path":         org.Path,
				"sort_order":   org.SortOrder,
				"description":  org.Description,
				"profile":      org.Profile,
				"created_at":   org.CreatedAt,
				"updated_at":   org.UpdatedAt,
			})
			return nil, err
			
		case "d": // DELETE
			query := `
			MATCH (org:Organization {tenant_id: $tenant_id, code: $code})
			DETACH DELETE org
			RETURN count(*) as deleted_count`
			
			_, err := tx.Run(ctx, query, map[string]interface{}{
				"tenant_id": org.TenantID,
				"code":      org.Code,
			})
			return nil, err
			
		default:
			return nil, fmt.Errorf("不支持的操作: %s", org.Operation)
		}
	})
	
	return err
}

func (s *SyncService) invalidateCache(ctx context.Context, orgCode string) error {
	// 清除组织相关的所有缓存
	cacheKeys := []string{
		fmt.Sprintf("org:%s:*", orgCode),
		fmt.Sprintf("hierarchy:%s:*", orgCode),
		"org:stats:*",
	}
	
	for _, pattern := range cacheKeys {
		keys, err := s.redis.Keys(ctx, pattern).Result()
		if err != nil {
			continue
		}
		
		if len(keys) > 0 {
			s.redis.Del(ctx, keys...)
		}
	}
	
	return nil
}

func (s *SyncService) Close() error {
	if s.kafka != nil {
		s.kafka.Close()
	}
	if s.neo4j != nil {
		s.neo4j.Close(context.Background())
	}
	if s.redis != nil {
		s.redis.Close()
	}
	return nil
}

func (s *SyncService) setupHealthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		health := map[string]interface{}{
			"service":   "temporal-sync-service-fixed",
			"status":    "healthy",
			"timestamp": time.Now().Format(time.RFC3339),
		}
		
		// 检查Neo4j连接
		if err := s.neo4j.VerifyConnectivity(ctx); err != nil {
			health["neo4j"] = "unhealthy: " + err.Error()
			health["status"] = "unhealthy"
		} else {
			health["neo4j"] = "healthy"
		}
		
		// 检查Redis连接
		if _, err := s.redis.Ping(ctx).Result(); err != nil {
			health["redis"] = "unhealthy: " + err.Error() 
			health["status"] = "unhealthy"
		} else {
			health["redis"] = "healthy"
		}
		
		w.Header().Set("Content-Type", "application/json")
		if health["status"] == "unhealthy" {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(health)
	})
}

func main() {
	config := loadConfig()
	
	service, err := NewSyncService(config)
	if err != nil {
		log.Fatalf("初始化同步服务失败: %v", err)
	}
	defer service.Close()
	
	// 设置健康检查
	service.setupHealthCheck()
	go http.ListenAndServe(":"+config.Port, nil)
	
	// 处理优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		service.logger.Printf("📡 收到关闭信号")
		cancel()
	}()
	
	// 启动服务
	if err := service.Start(ctx); err != nil {
		log.Fatalf("同步服务启动失败: %v", err)
	}
	
	service.logger.Printf("👋 同步服务已停止")
}