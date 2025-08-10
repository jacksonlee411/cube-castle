package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/redis/go-redis/v9"
	
	"cube-castle-deployment-test/internal/cache"
)

// 新一代缓存管理服务
type NextGenCacheService struct {
	cacheManager  *cache.UnifiedCacheManager
	cdcConsumer   *kafka.Consumer
	eventBus      *cache.CacheEventBus
	logger        *log.Logger
	config        *ServiceConfig
	ctx           context.Context
	cancel        context.CancelFunc
}

// 服务配置
type ServiceConfig struct {
	RedisAddr      string
	RedisPassword  string
	KafkaBrokers   []string
	Neo4jURI       string
	Neo4jUsername  string
	Neo4jPassword  string
	Port           string
	WriteThrough   bool
	ConsistencyMode string
}

// 初始化服务
func NewNextGenCacheService(config *ServiceConfig, logger *log.Logger) (*NextGenCacheService, error) {
	ctx, cancel := context.WithCancel(context.Background())

	// 1. 创建Redis客户端
	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisAddr,
		Password: config.RedisPassword,
		DB:       0,
	})

	// 测试Redis连接
	if _, err := redisClient.Ping(ctx).Result(); err != nil {
		cancel()
		return nil, fmt.Errorf("Redis连接失败: %w", err)
	}

	// 2. 创建Neo4j驱动
	driver, err := neo4j.NewDriverWithContext(config.Neo4jURI, neo4j.BasicAuth(config.Neo4jUsername, config.Neo4jPassword, ""))
	if err != nil {
		cancel()
		return nil, fmt.Errorf("创建Neo4j驱动失败: %w", err)
	}

	if err = driver.VerifyConnectivity(ctx); err != nil {
		cancel()
		return nil, fmt.Errorf("Neo4j连接验证失败: %w", err)
	}

	// 3. 创建L3查询服务
	l3Query := NewNeo4jQueryService(driver, logger)

	// 4. 创建统一缓存管理器
	cacheConfig := &cache.CacheConfig{
		L1TTL:           5 * time.Minute,
		L2TTL:           30 * time.Minute,
		L1MaxSize:       2000,
		WriteThrough:    config.WriteThrough,
		ConsistencyMode: config.ConsistencyMode,
		Namespace:       "org_v2",
	}

	cacheManager := cache.NewUnifiedCacheManager(redisClient, l3Query, cacheConfig, logger)

	// 5. 创建Kafka消费者
	kafkaConfig := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(config.KafkaBrokers, ","),
		"group.id":          "nextgen-cache-service",
		"auto.offset.reset": "latest",
		"enable.auto.commit": true,
	}

	consumer, err := kafka.NewConsumer(kafkaConfig)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("创建Kafka消费者失败: %w", err)
	}

	// 6. 创建事件总线
	eventBus := cache.NewCacheEventBus()

	service := &NextGenCacheService{
		cacheManager: cacheManager,
		cdcConsumer:  consumer,
		eventBus:     eventBus,
		logger:       logger,
		config:       config,
		ctx:          ctx,
		cancel:       cancel,
	}

	return service, nil
}

// 启动服务
func (service *NextGenCacheService) Start() error {
	service.logger.Println("🚀 启动新一代缓存管理服务...")

	// 1. 启动CDC消费者
	go service.startCDCConsumer()

	// 2. 启动HTTP服务器
	go service.startHTTPServer()

	// 3. 启动健康检查
	go service.startHealthMonitor()

	// 4. 等待停止信号
	return service.waitForShutdown()
}

// 启动CDC消费者
func (service *NextGenCacheService) startCDCConsumer() {
	topics := []string{"organization_db.public.organization_units"}
	
	if err := service.cdcConsumer.SubscribeTopics(topics, nil); err != nil {
		service.logger.Printf("❌ 订阅Kafka主题失败: %v", err)
		return
	}

	service.logger.Printf("📡 CDC消费者启动，监听主题: %v", topics)

	for {
		select {
		case <-service.ctx.Done():
			service.logger.Println("CDC消费者收到停止信号")
			return
		default:
			msg, err := service.cdcConsumer.ReadMessage(1000)
			if err != nil {
				if err.(kafka.Error).Code() == kafka.ErrTimedOut {
					continue
				}
				service.logger.Printf("⚠️ 消费消息失败: %v", err)
				continue
			}

			if err := service.processCDCMessage(msg); err != nil {
				service.logger.Printf("❌ 处理CDC消息失败: %v", err)
			}
		}
	}
}

// 处理CDC消息
func (service *NextGenCacheService) processCDCMessage(msg *kafka.Message) error {
	topic := *msg.TopicPartition.Topic
	
	if topic != "organization_db.public.organization_units" {
		return nil
	}

	// 解析Debezium消息
	var debeziumMsg struct {
		Payload struct {
			Before *map[string]interface{} `json:"before"`
			After  *map[string]interface{} `json:"after"`
			Op     string                  `json:"op"`
			TsMs   int64                   `json:"ts_ms"`
		} `json:"payload"`
	}

	if err := json.Unmarshal(msg.Value, &debeziumMsg); err != nil {
		return fmt.Errorf("解析Debezium消息失败: %w", err)
	}

	// 转换为统一事件格式
	event := cache.CDCEvent{
		EventID:   uuid.New().String(),
		Timestamp: debeziumMsg.Payload.TsMs,
		Source:    "debezium",
	}

	switch debeziumMsg.Payload.Op {
	case "c":
		event.Operation = "CREATE"
		event.After = *debeziumMsg.Payload.After
	case "u":
		event.Operation = "UPDATE"
		event.After = *debeziumMsg.Payload.After
		if debeziumMsg.Payload.Before != nil {
			event.Before = *debeziumMsg.Payload.Before
		}
	case "d":
		event.Operation = "DELETE"
		if debeziumMsg.Payload.Before != nil {
			event.Before = *debeziumMsg.Payload.Before
		}
	}

	// 提取租户ID和实体ID
	var data map[string]interface{}
	if event.After != nil {
		data = event.After
	} else if event.Before != nil {
		data = event.Before
	}

	if data != nil {
		if tenantID, ok := data["tenant_id"].(string); ok {
			event.TenantID = tenantID
		}
		if code, ok := data["code"].(string); ok {
			event.EntityID = code
		}
	}

	event.EntityType = "organization"

	service.logger.Printf("📨 收到CDC事件: %s %s:%s", event.Operation, event.EntityType, event.EntityID)

	// 处理事件
	return service.cacheManager.HandleCDCEvent(service.ctx, event)
}

// 启动HTTP服务器
func (service *NextGenCacheService) startHTTPServer() {
	r := chi.NewRouter()

	// 中间件
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	// CORS
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// API路由
	r.Route("/api/v2", func(r chi.Router) {
		r.Get("/organizations", service.handleGetOrganizations)
		r.Get("/organizations/{code}", service.handleGetOrganization)
		r.Get("/organizations/stats", service.handleGetStats)
		
		// 缓存管理接口
		r.Delete("/cache/refresh", service.handleRefreshCache)
		r.Get("/cache/stats", service.handleGetCacheStats)
		r.Get("/cache/consistency", service.handleCheckConsistency)
	})

	// 健康检查
	r.Get("/health", service.handleHealth)
	r.Get("/metrics", service.handleMetrics)

	server := &http.Server{
		Addr:    ":" + service.config.Port,
		Handler: r,
	}

	service.logger.Printf("🌐 HTTP服务器启动在端口 %s", service.config.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		service.logger.Printf("❌ HTTP服务器错误: %v", err)
	}
}

// 处理获取组织列表
func (service *NextGenCacheService) handleGetOrganizations(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9") // 默认租户
	
	first := 50
	if f := r.URL.Query().Get("first"); f != "" {
		if parsed, err := strconv.Atoi(f); err == nil {
			first = parsed
		}
	}
	
	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}
	
	searchText := r.URL.Query().Get("searchText")
	
	params := cache.QueryParams{
		First:      first,
		Offset:     offset,
		SearchText: searchText,
	}

	// 通过缓存管理器获取数据
	startTime := time.Now()
	orgs, err := service.cacheManager.GetOrganizations(r.Context(), tenantID, params)
	duration := time.Since(startTime)

	if err != nil {
		service.logger.Printf("❌ 获取组织列表失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 返回结果
	result := map[string]interface{}{
		"data":          orgs,
		"total":         len(orgs),
		"query_time_ms": duration.Milliseconds(),
		"cached":        duration < 10*time.Millisecond, // 简单的缓存命中判断
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	service.logger.Printf("✅ 组织列表查询完成: %d条记录, 耗时: %v", len(orgs), duration)
}

// 处理获取单个组织
func (service *NextGenCacheService) handleGetOrganization(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")

	startTime := time.Now()
	org, err := service.cacheManager.GetOrganization(r.Context(), tenantID, code)
	duration := time.Since(startTime)

	if err != nil {
		service.logger.Printf("❌ 获取组织失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if org == nil {
		http.Error(w, "组织不存在", http.StatusNotFound)
		return
	}

	result := map[string]interface{}{
		"data":          org,
		"query_time_ms": duration.Milliseconds(),
		"cached":        duration < 10*time.Millisecond,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	service.logger.Printf("✅ 组织查询完成: %s, 耗时: %v", code, duration)
}

// 处理获取统计信息
func (service *NextGenCacheService) handleGetStats(w http.ResponseWriter, r *http.Request) {
	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")

	startTime := time.Now()
	stats, err := service.cacheManager.GetOrganizationStats(r.Context(), tenantID)
	duration := time.Since(startTime)

	if err != nil {
		service.logger.Printf("❌ 获取统计信息失败: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{
		"data":          stats,
		"query_time_ms": duration.Milliseconds(),
		"cached":        duration < 10*time.Millisecond,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)

	service.logger.Printf("✅ 统计查询完成, 耗时: %v", duration)
}

// 处理缓存刷新
func (service *NextGenCacheService) handleRefreshCache(w http.ResponseWriter, r *http.Request) {
	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
	entityType := r.URL.Query().Get("type")
	entityID := r.URL.Query().Get("id")

	if entityType == "" {
		entityType = "organizations"
	}

	err := service.cacheManager.RefreshCache(r.Context(), tenantID, entityType, entityID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	result := map[string]string{
		"message": fmt.Sprintf("缓存已刷新: %s:%s", entityType, entityID),
		"status":  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// 处理获取缓存统计
func (service *NextGenCacheService) handleGetCacheStats(w http.ResponseWriter, r *http.Request) {
	stats := service.cacheManager.GetCacheStats(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// 处理一致性检查
func (service *NextGenCacheService) handleCheckConsistency(w http.ResponseWriter, r *http.Request) {
	// 简化版一致性检查
	result := map[string]interface{}{
		"status":     "healthy",
		"checked_at": time.Now(),
		"message":    "缓存一致性检查功能待实现",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// 健康检查
func (service *NextGenCacheService) handleHealth(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"service":   "nextgen-cache-service",
		"status":    "healthy",
		"timestamp": time.Now().Format(time.RFC3339),
		"features": []string{
			"三层缓存架构",
			"写时更新策略",
			"智能列表更新",
			"一致性保障",
			"CDC实时同步",
		},
		"cache_stats": service.cacheManager.GetCacheStats(r.Context()),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(health)
}

// 指标端点
func (service *NextGenCacheService) handleMetrics(w http.ResponseWriter, r *http.Request) {
	stats := service.cacheManager.GetCacheStats(r.Context())

	metrics := fmt.Sprintf(`# NextGen Cache Service Metrics
cache_l1_hits_total %d
cache_l1_misses_total %d  
cache_l1_size_current %d
cache_l1_hit_rate %.2f
cache_l2_connected %s
cache_write_through_enabled %s
`,
		stats.L1Stats.HitCount,
		stats.L1Stats.MissCount,
		stats.L1Stats.Size,
		stats.L1Stats.HitRate,
		boolToMetric(stats.L2Connected),
		boolToMetric(stats.WriteThrough),
	)

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(metrics))
}

func boolToMetric(b bool) string {
	if b {
		return "1"
	}
	return "0"
}

// 启动健康监控
func (service *NextGenCacheService) startHealthMonitor() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-service.ctx.Done():
			return
		case <-ticker.C:
			stats := service.cacheManager.GetCacheStats(service.ctx)
			service.logger.Printf("📊 缓存统计: L1命中率=%.2f%%, L1大小=%d, L2连接=%t, 写时更新=%t",
				stats.L1Stats.HitRate*100,
				stats.L1Stats.Size,
				stats.L2Connected,
				stats.WriteThrough,
			)
		}
	}
}

// 等待关闭信号
func (service *NextGenCacheService) waitForShutdown() error {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	service.logger.Println("📡 收到关闭信号，开始优雅关闭...")

	// 设置关闭超时
	shutdownTimeout := 15 * time.Second
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()

	// 关闭服务
	service.cancel()

	// 等待所有协程结束或超时
	done := make(chan struct{})
	go func() {
		defer close(done)
		
		if service.cdcConsumer != nil {
			service.cdcConsumer.Close()
		}
		
		if service.cacheManager != nil {
			service.cacheManager.Close()
		}
		
		if service.eventBus != nil {
			service.eventBus.Close()
		}
	}()

	select {
	case <-done:
		service.logger.Println("✅ 服务优雅关闭完成")
		return nil
	case <-shutdownCtx.Done():
		service.logger.Println("⏰ 关闭超时，强制退出")
		return shutdownCtx.Err()
	}
}

// 主函数
func main() {
	logger := log.New(os.Stdout, "[NEXTGEN-CACHE] ", log.LstdFlags)

	// 配置参数
	config := &ServiceConfig{
		RedisAddr:       getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:   getEnv("REDIS_PASSWORD", ""),
		KafkaBrokers:    strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		Neo4jURI:        getEnv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUsername:   getEnv("NEO4J_USERNAME", "neo4j"),
		Neo4jPassword:   getEnv("NEO4J_PASSWORD", "password"),
		Port:            getEnv("PORT", "8088"),
		WriteThrough:    getEnv("WRITE_THROUGH", "true") == "true",
		ConsistencyMode: getEnv("CONSISTENCY_MODE", "STRONG"),
	}

	// 创建服务
	service, err := NewNextGenCacheService(config, logger)
	if err != nil {
		log.Fatalf("❌ 创建服务失败: %v", err)
	}

	// 启动服务
	logger.Println("🚀 新一代缓存管理服务启动...")
	if err := service.Start(); err != nil {
		log.Fatalf("❌ 服务运行失败: %v", err)
	}

	logger.Println("👋 新一代缓存管理服务已退出")
}

// 获取环境变量
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}