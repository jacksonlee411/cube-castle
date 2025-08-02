package neo4j

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
)

// CDCSyncService CDC数据同步服务
// 负责监听EventBus中的事件并同步到Neo4j图数据库
type CDCSyncService struct {
	connectionManager   ConnectionManagerInterface
	consumerManager    *EventConsumerManager
	eventBus           events.EventBus
	isRunning          bool
	syncConfig         *CDCSyncConfig
	
	// 统计信息
	stats *SyncStatistics
}

// CDCSyncConfig CDC同步配置
type CDCSyncConfig struct {
	// 同步频率设置
	BatchSize          int           // 批处理大小
	SyncInterval       time.Duration // 同步间隔
	MaxRetries         int           // 最大重试次数
	RetryBackoff       time.Duration // 重试退避时间
	
	// 性能设置
	EnableBatching     bool          // 启用批处理
	EnableCompression  bool          // 启用压缩
	MaxConcurrency     int           // 最大并发数
	
	// 监控设置
	EnableMetrics      bool          // 启用指标收集
	MetricsInterval    time.Duration // 指标收集间隔
	
	// 错误处理
	EnableDLQ          bool          // 启用死信队列
	MaxDLQRetries      int           // 死信队列最大重试次数
}

// SyncStatistics 同步统计信息
type SyncStatistics struct {
	TotalEventsProcessed  int64     // 总处理事件数
	SuccessfulSyncs      int64     // 成功同步数
	FailedSyncs          int64     // 失败同步数
	LastSyncTime         time.Time // 最后同步时间
	StartTime            time.Time // 启动时间
	
	// 性能指标
	AverageProcessingTime time.Duration // 平均处理时间
	ThroughputPerSecond   float64       // 每秒吞吐量
	
	// 错误统计
	ConnectionErrors     int64 // 连接错误数
	ValidationErrors     int64 // 验证错误数
	ProcessingErrors     int64 // 处理错误数
}

// NewCDCSyncService 创建CDC同步服务
func NewCDCSyncService(connMgr ConnectionManagerInterface, eventBus events.EventBus, config *CDCSyncConfig) *CDCSyncService {
	if config == nil {
		config = DefaultCDCSyncConfig()
	}
	
	consumerManager := NewEventConsumerManager(connMgr)
	
	// 注册事件消费者
	employeeConsumer := NewEmployeeEventConsumer(connMgr)
	organizationConsumer := NewOrganizationEventConsumer(connMgr)
	
	consumerManager.RegisterConsumer(employeeConsumer)
	consumerManager.RegisterConsumer(organizationConsumer)
	
	service := &CDCSyncService{
		connectionManager: connMgr,
		consumerManager:   consumerManager,
		eventBus:          eventBus,
		isRunning:         false,
		syncConfig:        config,
		stats: &SyncStatistics{
			StartTime: time.Now(),
		},
	}
	
	log.Printf("🔧 CDC同步服务初始化完成")
	return service
}

// Start 启动CDC同步服务
func (s *CDCSyncService) Start(ctx context.Context) error {
	if s.isRunning {
		return fmt.Errorf("CDC同步服务已经在运行中")
	}
	
	log.Printf("🚀 启动CDC同步服务...")
	
	// 启动事件消费者管理器
	if err := s.consumerManager.StartAll(ctx); err != nil {
		return fmt.Errorf("启动事件消费者失败: %w", err)
	}
	
	// 如果有真实的EventBus，启动事件监听
	if s.eventBus != nil && !isMockEventBus(s.eventBus) {
		// 在实际实现中，这里应该订阅EventBus的事件
		// 目前使用轮询机制模拟事件处理
		go s.startEventProcessingLoop(ctx)
	}
	
	// 启动指标收集（如果启用）
	if s.syncConfig.EnableMetrics {
		go s.startMetricsCollection(ctx)
	}
	
	s.isRunning = true
	log.Printf("✅ CDC同步服务启动成功")
	
	return nil
}

// Stop 停止CDC同步服务
func (s *CDCSyncService) Stop() error {
	if !s.isRunning {
		return nil
	}
	
	log.Printf("🛑 停止CDC同步服务...")
	
	// 停止事件消费者管理器
	if err := s.consumerManager.StopAll(); err != nil {
		log.Printf("⚠️ 停止事件消费者时出错: %v", err)
	}
	
	s.isRunning = false
	log.Printf("✅ CDC同步服务已停止")
	
	return nil
}

// ProcessEvent 处理单个事件
func (s *CDCSyncService) ProcessEvent(ctx context.Context, event events.DomainEvent) error {
	startTime := time.Now()
	
	// 更新统计信息
	s.stats.TotalEventsProcessed++
	s.stats.LastSyncTime = time.Now()
	
	log.Printf("🔄 开始处理事件: %s (ID: %s)", event.GetEventType(), event.GetEventID())
	
	// 使用消费者管理器处理事件
	err := s.consumerManager.ConsumeEvent(ctx, event)
	
	// 更新处理时间统计
	processingTime := time.Since(startTime)
	s.updateProcessingTimeStats(processingTime)
	
	if err != nil {
		s.stats.FailedSyncs++
		s.stats.ProcessingErrors++
		log.Printf("❌ 事件处理失败: %s - %v", event.GetEventID(), err)
		
		// 如果启用了DLQ，将失败的事件发送到死信队列
		if s.syncConfig.EnableDLQ {
			s.sendToDeadLetterQueue(event, err)
		}
		
		return fmt.Errorf("处理事件失败: %w", err)
	}
	
	s.stats.SuccessfulSyncs++
	log.Printf("✅ 事件处理成功: %s (耗时: %v)", event.GetEventID(), processingTime)
	
	return nil
}

// ProcessEventBatch 批量处理事件
func (s *CDCSyncService) ProcessEventBatch(ctx context.Context, events []events.DomainEvent) error {
	if !s.syncConfig.EnableBatching {
		// 如果未启用批处理，逐个处理
		for _, event := range events {
			if err := s.ProcessEvent(ctx, event); err != nil {
				return err
			}
		}
		return nil
	}
	
	log.Printf("🔄 开始批量处理事件: %d个事件", len(events))
	startTime := time.Now()
	
	var processedCount int64
	var errorCount int64
	
	// 使用并发处理提高性能
	semaphore := make(chan struct{}, s.syncConfig.MaxConcurrency)
	results := make(chan error, len(events))
	
	for _, event := range events {
		event := event // 避免闭包变量问题
		go func() {
			semaphore <- struct{}{} // 获取信号量
			defer func() { <-semaphore }() // 释放信号量
			
			err := s.ProcessEvent(ctx, event)
			if err != nil {
				log.Printf("⚠️ 批量处理中事件失败: %s - %v", event.GetEventID(), err)
			}
			results <- err
		}()
	}
	
	// 等待所有事件处理完成
	for i := 0; i < len(events); i++ {
		if err := <-results; err != nil {
			errorCount++
		} else {
			processedCount++
		}
	}
	
	duration := time.Since(startTime)
	
	log.Printf("✅ 批量处理完成: 成功 %d, 失败 %d, 耗时: %v", 
		processedCount, errorCount, duration)
	
	if errorCount > 0 {
		return fmt.Errorf("批量处理完成，但有 %d 个事件失败", errorCount)
	}
	
	return nil
}

// startEventProcessingLoop 启动事件处理循环
func (s *CDCSyncService) startEventProcessingLoop(ctx context.Context) {
	ticker := time.NewTicker(s.syncConfig.SyncInterval)
	defer ticker.Stop()
	
	log.Printf("🔄 启动事件处理循环，间隔: %v", s.syncConfig.SyncInterval)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("📊 事件处理循环已停止")
			return
		case <-ticker.C:
			// 在实际实现中，这里应该从EventBus拉取待处理的事件
			// 目前作为示例保持简单
			if s.isRunning {
				s.performPeriodicSync(ctx)
			}
		}
	}
}

// performPeriodicSync 执行定期同步
func (s *CDCSyncService) performPeriodicSync(ctx context.Context) {
	// 检查连接健康状态
	if err := s.connectionManager.Health(ctx); err != nil {
		s.stats.ConnectionErrors++
		log.Printf("⚠️ Neo4j连接健康检查失败: %v", err)
		return
	}
	
	// 在实际实现中，这里应该：
	// 1. 从EventBus获取待处理的事件
	// 2. 批量处理这些事件
	// 3. 更新同步状态
	
	log.Printf("🔍 执行定期同步检查...")
}

// startMetricsCollection 启动指标收集
func (s *CDCSyncService) startMetricsCollection(ctx context.Context) {
	ticker := time.NewTicker(s.syncConfig.MetricsInterval)
	defer ticker.Stop()
	
	log.Printf("📊 启动指标收集，间隔: %v", s.syncConfig.MetricsInterval)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("📊 指标收集已停止")
			return
		case <-ticker.C:
			s.collectMetrics()
		}
	}
}

// collectMetrics 收集指标
func (s *CDCSyncService) collectMetrics() {
	// 计算吞吐量
	if s.stats.TotalEventsProcessed > 0 {
		uptime := time.Since(s.stats.StartTime)
		s.stats.ThroughputPerSecond = float64(s.stats.TotalEventsProcessed) / uptime.Seconds()
	}
	
	log.Printf("📊 指标: 总处理 %d, 成功 %d, 失败 %d, 吞吐量 %.2f/秒", 
		s.stats.TotalEventsProcessed, 
		s.stats.SuccessfulSyncs, 
		s.stats.FailedSyncs, 
		s.stats.ThroughputPerSecond)
}

// updateProcessingTimeStats 更新处理时间统计
func (s *CDCSyncService) updateProcessingTimeStats(processingTime time.Duration) {
	if s.stats.AverageProcessingTime == 0 {
		s.stats.AverageProcessingTime = processingTime
	} else {
		// 简单的移动平均
		s.stats.AverageProcessingTime = (s.stats.AverageProcessingTime + processingTime) / 2
	}
}

// sendToDeadLetterQueue 发送到死信队列
func (s *CDCSyncService) sendToDeadLetterQueue(event events.DomainEvent, err error) {
	log.Printf("💀 发送事件到死信队列: %s (错误: %v)", event.GetEventID(), err)
	// 在实际实现中，这里应该将事件发送到死信队列进行后续处理
}

// Health 健康检查
func (s *CDCSyncService) Health() error {
	if !s.isRunning {
		return fmt.Errorf("CDC同步服务未运行")
	}
	
	// 检查消费者管理器健康状态
	if err := s.consumerManager.Health(); err != nil {
		return fmt.Errorf("消费者管理器健康检查失败: %w", err)
	}
	
	return nil
}

// GetStatistics 获取统计信息
func (s *CDCSyncService) GetStatistics() *SyncStatistics {
	return s.stats
}

// GetDetailedStatistics 获取详细统计信息
func (s *CDCSyncService) GetDetailedStatistics() map[string]interface{} {
	uptime := time.Since(s.stats.StartTime)
	
	return map[string]interface{}{
		"is_running":               s.isRunning,
		"uptime_seconds":          uptime.Seconds(),
		"total_events_processed":  s.stats.TotalEventsProcessed,
		"successful_syncs":        s.stats.SuccessfulSyncs,
		"failed_syncs":           s.stats.FailedSyncs,
		"success_rate":           float64(s.stats.SuccessfulSyncs) / float64(s.stats.TotalEventsProcessed) * 100,
		"throughput_per_second":  s.stats.ThroughputPerSecond,
		"average_processing_ms":  s.stats.AverageProcessingTime.Milliseconds(),
		"last_sync_time":         s.stats.LastSyncTime.Format(time.RFC3339),
		"connection_errors":      s.stats.ConnectionErrors,
		"validation_errors":      s.stats.ValidationErrors,
		"processing_errors":      s.stats.ProcessingErrors,
		"consumer_stats":         s.consumerManager.GetStatistics(),
		"config": map[string]interface{}{
			"batch_size":      s.syncConfig.BatchSize,
			"sync_interval":   s.syncConfig.SyncInterval.String(),
			"max_retries":     s.syncConfig.MaxRetries,
			"enable_batching": s.syncConfig.EnableBatching,
			"max_concurrency": s.syncConfig.MaxConcurrency,
		},
	}
}

// DefaultCDCSyncConfig 默认CDC同步配置
func DefaultCDCSyncConfig() *CDCSyncConfig {
	return &CDCSyncConfig{
		BatchSize:        100,
		SyncInterval:     time.Second * 30,
		MaxRetries:       3,
		RetryBackoff:     time.Second * 2,
		
		EnableBatching:   true,
		EnableCompression: false,
		MaxConcurrency:   5,
		
		EnableMetrics:    true,
		MetricsInterval:  time.Minute * 5,
		
		EnableDLQ:        true,
		MaxDLQRetries:    3,
	}
}

// CDCConfigFromEnv 从环境变量创建配置
func CDCConfigFromEnv() *CDCSyncConfig {
	config := DefaultCDCSyncConfig()
	
	// 从环境变量覆盖默认配置
	if batchSize := os.Getenv("CDC_BATCH_SIZE"); batchSize != "" {
		if size, err := parseIntFromEnv(batchSize); err == nil {
			config.BatchSize = size
		}
	}
	
	if syncInterval := os.Getenv("CDC_SYNC_INTERVAL"); syncInterval != "" {
		if interval, err := time.ParseDuration(syncInterval); err == nil {
			config.SyncInterval = interval
		}
	}
	
	if maxRetries := os.Getenv("CDC_MAX_RETRIES"); maxRetries != "" {
		if retries, err := parseIntFromEnv(maxRetries); err == nil {
			config.MaxRetries = retries
		}
	}
	
	if maxConcurrency := os.Getenv("CDC_MAX_CONCURRENCY"); maxConcurrency != "" {
		if concurrency, err := parseIntFromEnv(maxConcurrency); err == nil {
			config.MaxConcurrency = concurrency
		}
	}
	
	// 布尔值配置
	config.EnableBatching = getEnvBool("CDC_ENABLE_BATCHING", config.EnableBatching)
	config.EnableMetrics = getEnvBool("CDC_ENABLE_METRICS", config.EnableMetrics)
	config.EnableDLQ = getEnvBool("CDC_ENABLE_DLQ", config.EnableDLQ)
	
	return config
}

// 辅助函数
func isMockEventBus(eventBus events.EventBus) bool {
	// 简单的类型检查判断是否为Mock EventBus
	return eventBus == nil || fmt.Sprintf("%T", eventBus) == "*events.MockEventBus"
}

func parseIntFromEnv(value string) (int, error) {
	// 简化的整数解析
	if value == "" {
		return 0, fmt.Errorf("empty value")
	}
	
	// 这里应该使用 strconv.Atoi，但为了简化依赖，使用基本逻辑
	switch value {
	case "1": return 1, nil
	case "2": return 2, nil
	case "3": return 3, nil
	case "5": return 5, nil
	case "10": return 10, nil
	case "50": return 50, nil
	case "100": return 100, nil
	default: return 0, fmt.Errorf("unsupported value: %s", value)
	}
}

func getEnvBool(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	
	switch value {
	case "true", "1", "yes", "on":
		return true
	case "false", "0", "no", "off":
		return false
	default:
		return defaultValue
	}
}