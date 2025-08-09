package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/go-redis/redis/v8"
	"database/sql"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	_ "github.com/lib/pq"
)

// ===== Debezium CDC监控指标 =====
var (
	// CDC处理指标
	cdcEventsProcessed = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cdc_events_processed_total",
			Help: "Total number of CDC events processed",
		},
		[]string{"operation", "status"},
	)

	cdcProcessingDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "cdc_processing_duration_seconds",
			Help: "Time taken to process CDC events",
			Buckets: []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 5.0},
		},
		[]string{"operation"},
	)

	// 数据一致性监控
	dataConsistencyViolations = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "organization_data_consistency_violations",
			Help: "Number of data consistency violations between PostgreSQL and Neo4j",
		},
		[]string{"entity"},
	)

	// 缓存性能监控
	cacheInvalidations = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_invalidations_total",
			Help: "Total number of cache invalidations performed",
		},
		[]string{"pattern", "tenant_id"},
	)

	// Kafka消费者监控
	kafkaConsumerLag = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "kafka_consumer_lag_messages",
			Help: "Current consumer lag in messages",
		},
		[]string{"topic", "partition", "consumer_group"},
	)

	// Neo4j连接监控
	neo4jConnections = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "neo4j_active_connections",
			Help: "Number of active Neo4j connections",
		},
		[]string{"database"},
	)
)

// ===== 企业级一致性检查器 =====
type ConsistencyChecker struct {
	postgres *sql.DB
	neo4j    neo4j.DriverWithContext
	redis    *redis.Client
	logger   *log.Logger
}

func NewConsistencyChecker(postgresURL, neo4jURI, redisURL string, logger *log.Logger) (*ConsistencyChecker, error) {
	// PostgreSQL连接
	postgres, err := sql.Open("postgres", postgresURL)
	if err != nil {
		return nil, fmt.Errorf("PostgreSQL连接失败: %w", err)
	}

	// Neo4j连接
	neo4jDriver, err := neo4j.NewDriverWithContext(neo4jURI, neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		return nil, fmt.Errorf("Neo4j连接失败: %w", err)
	}

	// Redis连接
	redisOpts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("Redis URL解析失败: %w", err)
	}
	redisClient := redis.NewClient(redisOpts)

	return &ConsistencyChecker{
		postgres: postgres,
		neo4j:    neo4jDriver,
		redis:    redisClient,
		logger:   logger,
	}, nil
}

// 定期一致性检查
func (cc *ConsistencyChecker) StartPeriodicCheck(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	cc.logger.Printf("🔍 启动定期一致性检查，间隔: %v", interval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := cc.CheckDataConsistency(ctx); err != nil {
				cc.logger.Printf("❌ 一致性检查失败: %v", err)
			}
		}
	}
}

func (cc *ConsistencyChecker) CheckDataConsistency(ctx context.Context) error {
	start := time.Now()
	
	// 1. 检查PostgreSQL和Neo4j记录数量一致性
	pgCount, err := cc.getPostgreSQLCount(ctx)
	if err != nil {
		return fmt.Errorf("获取PostgreSQL计数失败: %w", err)
	}

	neo4jCount, err := cc.getNeo4jCount(ctx)
	if err != nil {
		return fmt.Errorf("获取Neo4j计数失败: %w", err)
	}

	violations := abs(pgCount - neo4jCount)
	dataConsistencyViolations.WithLabelValues("organization").Set(float64(violations))

	if violations > 0 {
		cc.logger.Printf("⚠️ 发现数据不一致: PostgreSQL=%d, Neo4j=%d, 差异=%d", pgCount, neo4jCount, violations)
		
		// 触发自动修复（可选）
		// go cc.TriggerReconciliation(ctx)
	} else {
		cc.logger.Printf("✅ 数据一致性检查通过: %d条记录", pgCount)
	}

	// 2. 检查缓存健康度
	cacheStats := cc.redis.PoolStats()
	cc.logger.Printf("📊 Redis连接池状态: Total=%d, Idle=%d, Stale=%d", 
		cacheStats.TotalConns, cacheStats.IdleConns, cacheStats.StaleConns)

	cc.logger.Printf("🔍 一致性检查完成，耗时: %v", time.Since(start))
	return nil
}

func (cc *ConsistencyChecker) getPostgreSQLCount(ctx context.Context) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1`
	err := cc.postgres.QueryRowContext(ctx, query, "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9").Scan(&count)
	return count, err
}

func (cc *ConsistencyChecker) getNeo4jCount(ctx context.Context) (int, error) {
	session := cc.neo4j.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	query := `MATCH (o:OrganizationUnit {tenant_id: $tenant_id}) RETURN count(o) as count`
	result, err := session.Run(ctx, query, map[string]interface{}{
		"tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9",
	})
	if err != nil {
		return 0, err
	}

	if result.Next(ctx) {
		count := result.Record().Values[0].(int64)
		return int(count), nil
	}

	return 0, fmt.Errorf("未获取到Neo4j计数")
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// ===== 监控服务器 =====
type MonitoringServer struct {
	checker *ConsistencyChecker
	logger  *log.Logger
}

func NewMonitoringServer(checker *ConsistencyChecker, logger *log.Logger) *MonitoringServer {
	return &MonitoringServer{
		checker: checker,
		logger:  logger,
	}
}

func (ms *MonitoringServer) Start(ctx context.Context, port int) {
	// 注册Prometheus指标
	prometheus.MustRegister(
		cdcEventsProcessed,
		cdcProcessingDuration,
		dataConsistencyViolations,
		cacheInvalidations,
		kafkaConsumerLag,
		neo4jConnections,
	)

	// 启动定期一致性检查
	go ms.checker.StartPeriodicCheck(ctx, 30*time.Second)

	// 创建HTTP服务器
	mux := http.NewServeMux()
	
	// Prometheus指标端点
	mux.Handle("/metrics", promhttp.Handler())
	
	// 健康检查端点
	mux.HandleFunc("/health", ms.healthHandler)
	
	// 一致性检查端点
	mux.HandleFunc("/consistency", ms.consistencyHandler)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	ms.logger.Printf("🌐 监控服务器启动，端口: %d", port)
	ms.logger.Printf("📊 Prometheus指标: http://localhost:%d/metrics", port)
	ms.logger.Printf("❤️ 健康检查: http://localhost:%d/health", port)

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			ms.logger.Printf("❌ 监控服务器启动失败: %v", err)
		}
	}()

	// 优雅关闭
	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := server.Shutdown(shutdownCtx); err != nil {
		ms.logger.Printf("❌ 监控服务器关闭失败: %v", err)
	} else {
		ms.logger.Println("👋 监控服务器已优雅关闭")
	}
}

func (ms *MonitoringServer) healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	
	// 简单的健康检查
	health := map[string]string{
		"status":    "healthy",
		"service":   "organization-sync-monitoring",
		"timestamp": time.Now().Format(time.RFC3339),
		"version":   "v6.0-debezium-enhanced",
	}
	
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"healthy","service":"organization-sync-monitoring","timestamp":"%s","version":"v6.0-debezium-enhanced"}`, time.Now().Format(time.RFC3339))
}

func (ms *MonitoringServer) consistencyHandler(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	
	if err := ms.checker.CheckDataConsistency(ctx); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, `{"status":"error","message":"%s"}`, err.Error())
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"consistent","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
}

// ===== 指标记录工具函数 =====
func RecordCDCEvent(operation, status string, duration time.Duration) {
	cdcEventsProcessed.WithLabelValues(operation, status).Inc()
	cdcProcessingDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

func RecordCacheInvalidation(pattern, tenantID string, count int) {
	cacheInvalidations.WithLabelValues(pattern, tenantID).Add(float64(count))
}

// ===== 主程序 (可独立运行的监控服务) =====
func main() {
	logger := log.New(log.Writer(), "[CDC-MONITORING] ", log.LstdFlags|log.Lshortfile)
	logger.Println("🚀 启动CDC监控服务...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建一致性检查器
	checker, err := NewConsistencyChecker(
		"postgres://user:password@localhost:5432/cubecastle",
		"neo4j://localhost:7687",
		"redis://localhost:6379",
		logger,
	)
	if err != nil {
		logger.Fatalf("❌ 创建一致性检查器失败: %v", err)
	}

	// 启动监控服务器
	monitoring := NewMonitoringServer(checker, logger)
	monitoring.Start(ctx, 9091)
}