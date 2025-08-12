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
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
)

// ===== 时态CDC同步服务配置 =====

type TemporalSyncConfig struct {
	KafkaBootstrapServers string
	KafkaTopic           string
	Neo4jURI             string
	Neo4jUsername        string 
	Neo4jPassword        string
	RedisAddr            string
	RedisPassword        string
	TenantID             string
	LogLevel             string
}

func loadConfig() *TemporalSyncConfig {
	return &TemporalSyncConfig{
		KafkaBootstrapServers: getEnv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
		KafkaTopic:           getEnv("KAFKA_TOPIC", "organization_db.public.organization_units"),
		Neo4jURI:             getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:        getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:        getEnv("NEO4J_PASSWORD", "password"),
		RedisAddr:            getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:        getEnv("REDIS_PASSWORD", ""),
		TenantID:             getEnv("DEFAULT_TENANT_ID", "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"),
		LogLevel:             getEnv("LOG_LEVEL", "INFO"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ===== Debezium事件模型 =====

type DebeziumEvent struct {
	Schema  json.RawMessage `json:"schema"`
	Payload struct {
		Before json.RawMessage `json:"before"`
		After  json.RawMessage `json:"after"`
		Source struct {
			Version   string `json:"version"`
			Connector string `json:"connector"`
			Name      string `json:"name"`
			TsMs      int64  `json:"ts_ms"`
			Snapshot  string `json:"snapshot"`
			Db        string `json:"db"`
			Table     string `json:"table"`
		} `json:"source"`
		Op     string `json:"op"` // c=create, u=update, d=delete, r=read
		TsMs   int64  `json:"ts_ms"`
		Transaction interface{} `json:"transaction"`
	} `json:"payload"`
}

type TemporalOrganization struct {
	TenantID      string      `json:"tenant_id"`
	Code          string      `json:"code"`
	ParentCode    *string     `json:"parent_code"`
	Name          string      `json:"name"`
	UnitType      string      `json:"unit_type"`
	Status        string      `json:"status"`
	EffectiveDate interface{} `json:"effective_date"` // 可以是string或int64
	EndDate       interface{} `json:"end_date"`       // 可以是string或int64或nil
	IsCurrent     bool        `json:"is_current"`
	ChangeReason  *string     `json:"change_reason"`
	IsTemporal    bool        `json:"is_temporal"`
	CreatedAt     interface{} `json:"created_at"`     // 可以是string或int64
	UpdatedAt     interface{} `json:"updated_at"`     // 可以是string或int64
}

// ===== 辅助函数 =====

// 将interface{}时间值转换为字符串
func formatTimeValue(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int64:
		// Debezium时间戳转换为ISO日期格式
		if v == 0 {
			return ""
		}
		return time.Unix(v/1000, (v%1000)*1000000).Format("2006-01-02")
	case float64:
		if v == 0 {
			return ""
		}
		return time.Unix(int64(v)/1000, (int64(v)%1000)*1000000).Format("2006-01-02")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// 将interface{}时间值转换为Neo4j datetime格式
func formatDateTimeValue(value interface{}) string {
	if value == nil {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case int64:
		if v == 0 {
			return ""
		}
		return time.Unix(v/1000, (v%1000)*1000000).Format(time.RFC3339)
	case float64:
		if v == 0 {
			return ""
		}
		return time.Unix(int64(v)/1000, (int64(v)%1000)*1000000).Format(time.RFC3339)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ===== 时态同步服务 =====

type TemporalSyncService struct {
	config       *TemporalSyncConfig
	kafka        *kafka.Consumer
	neo4j        neo4j.DriverWithContext
	redis        *redis.Client
	logger       *log.Logger
	tenantID     uuid.UUID
}

func NewTemporalSyncService(config *TemporalSyncConfig) (*TemporalSyncService, error) {
	logger := log.New(os.Stdout, "[TEMPORAL-SYNC] ", log.LstdFlags|log.Lshortfile)
	
	// 解析租户ID
	tenantID, err := uuid.Parse(config.TenantID)
	if err != nil {
		return nil, fmt.Errorf("无效的租户ID: %w", err)
	}
	
	// 初始化Kafka消费者
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": config.KafkaBootstrapServers,
		"group.id":          "temporal-sync-service",
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
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})
	
	// 验证Redis连接
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}
	
	service := &TemporalSyncService{
		config:   config,
		kafka:    consumer,
		neo4j:    neo4jDriver,
		redis:    redisClient,
		logger:   logger,
		tenantID: tenantID,
	}
	
	logger.Printf("🚀 时态同步服务初始化完成 - 租户: %s", config.TenantID)
	return service, nil
}

func (s *TemporalSyncService) Start(ctx context.Context) error {
	// 订阅Kafka主题
	if err := s.kafka.SubscribeTopics([]string{s.config.KafkaTopic}, nil); err != nil {
		return fmt.Errorf("订阅Kafka主题失败: %w", err)
	}
	
	s.logger.Printf("📡 开始监听Kafka主题: %s", s.config.KafkaTopic)
	
	// 主消费循环
	for {
		select {
		case <-ctx.Done():
			s.logger.Printf("⏹️ 收到停止信号，正在关闭时态同步服务")
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
			
			if err := s.processTemporalEvent(ctx, msg); err != nil {
				s.logger.Printf("❌ 处理时态事件失败: %v", err)
			}
		}
	}
}

func (s *TemporalSyncService) processTemporalEvent(ctx context.Context, msg *kafka.Message) error {
	var event DebeziumEvent
	if err := json.Unmarshal(msg.Value, &event); err != nil {
		return fmt.Errorf("解析Debezium事件失败: %w", err)
	}
	
	// 解析组织数据
	var org *TemporalOrganization
	var err error
	
	switch event.Payload.Op {
	case "c": // 创建
		org, err = s.parseTemporalOrganization(event.Payload.After)
		if err != nil {
			return fmt.Errorf("解析创建事件失败: %w", err)
		}
		return s.handleOrganizationCreated(ctx, org)
		
	case "u": // 更新
		org, err = s.parseTemporalOrganization(event.Payload.After)
		if err != nil {
			return fmt.Errorf("解析更新事件失败: %w", err)
		}
		return s.handleOrganizationUpdated(ctx, org)
		
	case "d": // 删除
		org, err = s.parseTemporalOrganization(event.Payload.Before)
		if err != nil {
			return fmt.Errorf("解析删除事件失败: %w", err)
		}
		return s.handleOrganizationDeleted(ctx, org)
		
	case "r": // 读取（快照）
		org, err = s.parseTemporalOrganization(event.Payload.After)
		if err != nil {
			return fmt.Errorf("解析读取事件失败: %w", err)
		}
		return s.handleOrganizationSnapshot(ctx, org)
		
	default:
		s.logger.Printf("⚠️ 未知操作类型: %s", event.Payload.Op)
		return nil
	}
}

func (s *TemporalSyncService) parseTemporalOrganization(data json.RawMessage) (*TemporalOrganization, error) {
	// 先解析为map以便处理不同类型的字段
	var rawData map[string]interface{}
	if err := json.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("解析原始数据失败: %w", err)
	}
	
	org := &TemporalOrganization{
		TenantID:    getStringValue(rawData, "tenant_id"),
		Code:        getStringValue(rawData, "code"),
		Name:        getStringValue(rawData, "name"),
		UnitType:    getStringValue(rawData, "unit_type"),
		Status:      getStringValue(rawData, "status"),
		IsCurrent:   getBoolValue(rawData, "is_current"),
		IsTemporal:  true, // 默认为时态记录
	}
	
	// 处理可为空的字符串字段
	if parentCode := getStringValue(rawData, "parent_code"); parentCode != "" {
		org.ParentCode = &parentCode
	}
	if changeReason := getStringValue(rawData, "change_reason"); changeReason != "" {
		org.ChangeReason = &changeReason
	}
	
	// 处理时间字段
	org.EffectiveDate = rawData["effective_date"]
	org.EndDate = rawData["end_date"]
	org.CreatedAt = rawData["created_at"]
	org.UpdatedAt = rawData["updated_at"]
	
	// 验证必需字段
	if org.Code == "" || org.Name == "" || org.TenantID == "" {
		return nil, fmt.Errorf("缺少必需字段: code=%s, name=%s, tenant_id=%s", 
			org.Code, org.Name, org.TenantID)
	}
	
	return org, nil
}

// 辅助函数：安全地提取字符串值
func getStringValue(data map[string]interface{}, key string) string {
	if value, exists := data[key]; exists && value != nil {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// 辅助函数：安全地提取布尔值
func getBoolValue(data map[string]interface{}, key string) bool {
	if value, exists := data[key]; exists && value != nil {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

func (s *TemporalSyncService) handleOrganizationCreated(ctx context.Context, org *TemporalOrganization) error {
	s.logger.Printf("🆕 处理组织创建: %s - %s", org.Code, org.Name)
	
	session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
		CREATE (org:TemporalOrganization {
			tenant_id: $tenant_id,
			code: $code,
			parent_code: $parent_code,
			name: $name,
			unit_type: $unit_type,
			status: $status,
			effective_date: date($effective_date),
			end_date: CASE WHEN $end_date IS NOT NULL THEN date($end_date) ELSE null END,
			is_current: $is_current,
			change_reason: $change_reason,
			is_temporal: $is_temporal,
			created_at: datetime($created_at),
			updated_at: datetime($updated_at),
			synced_from_pg: datetime()
		})
		RETURN org.code as code`
		
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"tenant_id":      org.TenantID,
			"code":           org.Code,
			"parent_code":    org.ParentCode,
			"name":           org.Name,
			"unit_type":      org.UnitType,
			"status":         org.Status,
			"effective_date": formatTimeValue(org.EffectiveDate),
			"end_date":       formatTimeValue(org.EndDate),
			"is_current":     org.IsCurrent,
			"change_reason":  org.ChangeReason,
			"is_temporal":    org.IsTemporal,
			"created_at":     formatDateTimeValue(org.CreatedAt),
			"updated_at":     formatDateTimeValue(org.UpdatedAt),
		})
		
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			return result.Record().Values[0], nil
		}
		
		return nil, fmt.Errorf("创建组织节点失败")
	})
	
	if err != nil {
		return fmt.Errorf("Neo4j创建组织失败: %w", err)
	}
	
	// 重新计算层级结构
	if err := s.recalculateHierarchy(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 层级计算失败: %v", err)
	}
	
	// 清除相关缓存
	if err := s.invalidateCache(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 缓存清除失败: %v", err)
	}
	
	s.logger.Printf("✅ 组织创建同步完成: %s", org.Code)
	return nil
}

func (s *TemporalSyncService) handleOrganizationUpdated(ctx context.Context, org *TemporalOrganization) error {
	s.logger.Printf("🔄 处理组织更新: %s - %s", org.Code, org.Name)
	
	session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 查找现有节点
		findQuery := `
		MATCH (org:TemporalOrganization {tenant_id: $tenant_id, code: $code, effective_date: date($effective_date)})
		RETURN org`
		
		findResult, err := tx.Run(ctx, findQuery, map[string]interface{}{
			"tenant_id":      org.TenantID,
			"code":           org.Code,
			"effective_date": formatTimeValue(org.EffectiveDate),
		})
		
		if err != nil {
			return nil, err
		}
		
		if findResult.Next(ctx) {
			// 更新现有节点
			updateQuery := `
			MATCH (org:TemporalOrganization {tenant_id: $tenant_id, code: $code, effective_date: date($effective_date)})
			SET org.name = $name,
				org.parent_code = $parent_code,
				org.unit_type = $unit_type,
				org.status = $status,
				org.end_date = CASE WHEN $end_date IS NOT NULL THEN date($end_date) ELSE null END,
				org.is_current = $is_current,
				org.change_reason = $change_reason,
				org.updated_at = datetime($updated_at),
				org.synced_from_pg = datetime()
			RETURN org.code as code`
			
			_, err = tx.Run(ctx, updateQuery, map[string]interface{}{
				"tenant_id":      org.TenantID,
				"code":           org.Code,
				"effective_date": formatTimeValue(org.EffectiveDate),
				"name":           org.Name,
				"parent_code":    org.ParentCode,
				"unit_type":      org.UnitType,
				"status":         org.Status,
				"end_date":       formatTimeValue(org.EndDate),
				"is_current":     org.IsCurrent,
				"change_reason":  org.ChangeReason,
				"updated_at":     formatDateTimeValue(org.UpdatedAt),
			})
			
			if err != nil {
				return nil, err
			}
			return nil, nil
		} else {
			// 创建新的时态节点
			createQuery := `
			CREATE (org:TemporalOrganization {
				tenant_id: $tenant_id,
				code: $code,
				parent_code: $parent_code,
				name: $name,
				unit_type: $unit_type,
				status: $status,
				effective_date: date($effective_date),
				end_date: CASE WHEN $end_date IS NOT NULL THEN date($end_date) ELSE null END,
				is_current: $is_current,
				change_reason: $change_reason,
				is_temporal: $is_temporal,
				created_at: datetime($created_at),
				updated_at: datetime($updated_at),
				synced_from_pg: datetime()
			})
			RETURN org.code as code`
			
			_, err = tx.Run(ctx, createQuery, map[string]interface{}{
				"tenant_id":      org.TenantID,
				"code":           org.Code,
				"parent_code":    org.ParentCode,
				"name":           org.Name,
				"unit_type":      org.UnitType,
				"status":         org.Status,
				"effective_date": formatTimeValue(org.EffectiveDate),
				"end_date":       formatTimeValue(org.EndDate),
				"is_current":     org.IsCurrent,
				"change_reason":  org.ChangeReason,
				"is_temporal":    org.IsTemporal,
				"created_at":     formatDateTimeValue(org.CreatedAt),
				"updated_at":     formatDateTimeValue(org.UpdatedAt),
			})
			
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
	})
	
	if err != nil {
		return fmt.Errorf("Neo4j更新组织失败: %w", err)
	}
	
	// 重新计算层级结构
	if err := s.recalculateHierarchy(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 层级计算失败: %v", err)
	}
	
	// 清除相关缓存
	if err := s.invalidateCache(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 缓存清除失败: %v", err)
	}
	
	s.logger.Printf("✅ 组织更新同步完成: %s", org.Code)
	return nil
}

func (s *TemporalSyncService) handleOrganizationDeleted(ctx context.Context, org *TemporalOrganization) error {
	s.logger.Printf("🗑️ 处理组织删除: %s - %s", org.Code, org.Name)
	
	session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		query := `
		MATCH (org:TemporalOrganization {tenant_id: $tenant_id, code: $code, effective_date: date($effective_date)})
		DETACH DELETE org
		RETURN count(*) as deleted_count`
		
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"tenant_id":      org.TenantID,
			"code":           org.Code,
			"effective_date": formatTimeValue(org.EffectiveDate),
		})
		
		if err != nil {
			return nil, err
		}
		
		return result.Consume(ctx)
	})
	
	if err != nil {
		return fmt.Errorf("Neo4j删除组织失败: %w", err)
	}
	
	// 清除相关缓存
	if err := s.invalidateCache(ctx, org.Code); err != nil {
		s.logger.Printf("⚠️ 缓存清除失败: %v", err)
	}
	
	s.logger.Printf("✅ 组织删除同步完成: %s", org.Code)
	return nil
}

func (s *TemporalSyncService) handleOrganizationSnapshot(ctx context.Context, org *TemporalOrganization) error {
	s.logger.Printf("📸 处理组织快照: %s - %s", org.Code, org.Name)
	// 快照处理与创建类似，但不重新计算层级
	return s.handleOrganizationCreated(ctx, org)
}

func (s *TemporalSyncService) recalculateHierarchy(ctx context.Context, orgCode string) error {
	session := s.neo4j.NewSession(ctx, neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (interface{}, error) {
		// 调用时态层级计算函数
		query := `
		CALL temporal.calculateHierarchy($tenant_id, $code, date()) 
		YIELD updated_code, new_level, new_path
		RETURN updated_code, new_level, new_path`
		
		result, err := tx.Run(ctx, query, map[string]interface{}{
			"tenant_id": s.config.TenantID,
			"code":      orgCode,
		})
		
		if err != nil {
			return nil, err
		}
		
		return result.Consume(ctx)
	})
	
	return err
}

func (s *TemporalSyncService) invalidateCache(ctx context.Context, orgCode string) error {
	// 清除组织相关的所有缓存
	cacheKeys := []string{
		fmt.Sprintf("temporal:org:%s:*", orgCode),
		fmt.Sprintf("temporal:hierarchy:%s:*", orgCode),
		fmt.Sprintf("temporal:path:%s", orgCode),
		"temporal:stats:*",
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
	
	s.logger.Printf("🧹 缓存清除完成: %s", orgCode)
	return nil
}

func (s *TemporalSyncService) Close() error {
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

// ===== 健康检查和监控 =====

func (s *TemporalSyncService) setupHealthCheck() {
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		
		health := map[string]interface{}{
			"service": "temporal-sync-service",
			"status":  "healthy",
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
	
	http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		// 返回Prometheus格式指标
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# HELP temporal_sync_processed_total Total processed temporal events\n")
		fmt.Fprintf(w, "# TYPE temporal_sync_processed_total counter\n")
		fmt.Fprintf(w, "temporal_sync_processed_total 0\n")
	})
}

// ===== 主函数 =====

func main() {
	config := loadConfig()
	
	service, err := NewTemporalSyncService(config)
	if err != nil {
		log.Fatalf("初始化时态同步服务失败: %v", err)
	}
	defer service.Close()
	
	// 设置健康检查
	service.setupHealthCheck()
	go http.ListenAndServe(":8092", nil)
	
	// 处理优雅关闭
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		service.logger.Printf("📡 收到关闭信号，正在停止时态同步服务...")
		cancel()
	}()
	
	// 启动服务
	if err := service.Start(ctx); err != nil {
		log.Fatalf("时态同步服务启动失败: %v", err)
	}
	
	service.logger.Printf("👋 时态同步服务已停止")
}