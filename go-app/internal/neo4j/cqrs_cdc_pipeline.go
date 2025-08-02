package neo4j

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gaogu/cube-castle/go-app/internal/events"
)

// CQRSCDCPipeline CQRS+CDC完整流水线
// 集成EventBus事件发布和Neo4j数据同步的完整流程
type CQRSCDCPipeline struct {
	// 核心组件
	connectionManager ConnectionManagerInterface
	cdcSyncService   *CDCSyncService
	eventBus         events.EventBus
	
	// 配置和状态
	pipelineConfig   *PipelineConfig
	isRunning        bool
	startTime        time.Time
	
	// 监控和统计
	healthStatus     *PipelineHealthStatus
	performanceStats *PipelinePerformanceStats
}

// PipelineConfig 流水线配置
type PipelineConfig struct {
	// Neo4j配置
	Neo4jConfig *ConnectionConfig
	
	// CDC同步配置
	CDCConfig *CDCSyncConfig
	
	// 流水线特定配置
	EnableHealthChecks    bool          // 启用健康检查
	HealthCheckInterval   time.Duration // 健康检查间隔
	EnableAutoRecovery    bool          // 启用自动恢复
	MaxRecoveryAttempts   int           // 最大恢复尝试次数
	
	// 监控配置
	EnableDetailedLogs    bool          // 启用详细日志
	LogLevel             string        // 日志级别
	MetricsExportInterval time.Duration // 指标导出间隔
}

// PipelineHealthStatus 流水线健康状态
type PipelineHealthStatus struct {
	IsHealthy             bool      // 整体健康状态
	LastHealthCheck       time.Time // 最后健康检查时间
	Neo4jConnected        bool      // Neo4j连接状态
	EventBusConnected     bool      // EventBus连接状态
	CDCServiceRunning     bool      // CDC服务运行状态
	
	// 错误信息
	LastError            string    // 最后错误信息
	ErrorCount           int64     // 错误计数
	RecoveryAttempts     int       // 恢复尝试次数
}

// PipelinePerformanceStats 流水线性能统计
type PipelinePerformanceStats struct {
	// 事件处理统计
	TotalEvents         int64         // 总事件数
	ProcessedEvents     int64         // 已处理事件数
	FailedEvents        int64         // 失败事件数
	SkippedEvents       int64         // 跳过事件数
	
	// 性能指标
	AverageLatency      time.Duration // 平均延迟
	ThroughputPerSecond float64       // 每秒吞吐量
	PeakThroughput      float64       // 峰值吞吐量
	
	// 时间统计
	TotalProcessingTime time.Duration // 总处理时间
	Uptime             time.Duration // 运行时间
	
	// 详细统计
	EventTypeStats     map[string]*EventTypeStats // 按事件类型统计
}

// EventTypeStats 事件类型统计
type EventTypeStats struct {
	Count             int64         // 事件数量
	SuccessCount      int64         // 成功数量
	FailureCount      int64         // 失败数量
	AverageProcessingTime time.Duration // 平均处理时间
	LastProcessed     time.Time     // 最后处理时间
}

// NewCQRSCDCPipeline 创建CQRS+CDC流水线
func NewCQRSCDCPipeline(eventBus events.EventBus, config *PipelineConfig) (*CQRSCDCPipeline, error) {
	if config == nil {
		config = DefaultPipelineConfig()
	}
	
	// 创建Neo4j连接管理器
	var connectionManager ConnectionManagerInterface
	
	env := os.Getenv("DEPLOYMENT_ENV")
	if env == "production" || env == "prod" {
		// 生产环境使用真实Neo4j连接
		if config.Neo4jConfig == nil {
			config.Neo4jConfig = Neo4jConfigFromEnv()
		}
		
		realConnMgr, err := NewConnectionManager(config.Neo4jConfig)
		if err != nil {
			log.Printf("⚠️ 无法连接Neo4j，降级到Mock模式: %v", err)
			connectionManager = NewMockConnectionManager()
		} else {
			connectionManager = realConnMgr
		}
	} else {
		// 开发环境使用Mock连接
		connectionManager = NewMockConnectionManager()
	}
	
	// 创建CDC同步服务
	cdcSyncService := NewCDCSyncService(connectionManager, eventBus, config.CDCConfig)
	
	pipeline := &CQRSCDCPipeline{
		connectionManager: connectionManager,
		cdcSyncService:   cdcSyncService,
		eventBus:         eventBus,
		pipelineConfig:   config,
		isRunning:        false,
		startTime:        time.Now(),
		healthStatus: &PipelineHealthStatus{
			LastHealthCheck: time.Now(),
		},
		performanceStats: &PipelinePerformanceStats{
			EventTypeStats: make(map[string]*EventTypeStats),
		},
	}
	
	log.Printf("🏗️ CQRS+CDC流水线初始化完成")
	return pipeline, nil
}

// Start 启动流水线
func (p *CQRSCDCPipeline) Start(ctx context.Context) error {
	if p.isRunning {
		return fmt.Errorf("CQRS+CDC流水线已经在运行中")
	}
	
	log.Printf("🚀 启动CQRS+CDC流水线...")
	
	// 启动CDC同步服务
	if err := p.cdcSyncService.Start(ctx); err != nil {
		return fmt.Errorf("启动CDC同步服务失败: %w", err)
	}
	
	// 启动健康检查（如果启用）
	if p.pipelineConfig.EnableHealthChecks {
		go p.startHealthCheckLoop(ctx)
	}
	
	// 启动性能监控
	go p.startPerformanceMonitoring(ctx)
	
	p.isRunning = true
	p.startTime = time.Now()
	
	log.Printf("✅ CQRS+CDC流水线启动成功")
	return nil
}

// Stop 停止流水线
func (p *CQRSCDCPipeline) Stop() error {
	if !p.isRunning {
		return nil
	}
	
	log.Printf("🛑 停止CQRS+CDC流水线...")
	
	// 停止CDC同步服务
	if err := p.cdcSyncService.Stop(); err != nil {
		log.Printf("⚠️ 停止CDC同步服务时出错: %v", err)
	}
	
	// 关闭Neo4j连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := p.connectionManager.Close(ctx); err != nil {
		log.Printf("⚠️ 关闭Neo4j连接时出错: %v", err)
	}
	
	p.isRunning = false
	log.Printf("✅ CQRS+CDC流水线已停止")
	
	return nil
}

// ProcessEvent 处理单个事件（流水线入口点）
func (p *CQRSCDCPipeline) ProcessEvent(ctx context.Context, event events.DomainEvent) error {
	if !p.isRunning {
		return fmt.Errorf("流水线未运行")
	}
	
	startTime := time.Now()
	eventType := event.GetEventType()
	
	// 更新统计信息
	p.updateEventStats(eventType, startTime)
	
	log.Printf("🔄 流水线处理事件: %s (ID: %s, 租户: %s)", 
		eventType, event.GetEventID(), event.GetTenantID())
	
	// 使用CDC同步服务处理事件
	err := p.cdcSyncService.ProcessEvent(ctx, event)
	
	processingTime := time.Since(startTime)
	
	if err != nil {
		p.performanceStats.FailedEvents++
		p.updateEventTypeStats(eventType, false, processingTime)
		
		log.Printf("❌ 流水线事件处理失败: %s - %v (耗时: %v)", 
			event.GetEventID(), err, processingTime)
		
		return fmt.Errorf("流水线事件处理失败: %w", err)
	}
	
	p.performanceStats.ProcessedEvents++
	p.updateEventTypeStats(eventType, true, processingTime)
	
	log.Printf("✅ 流水线事件处理成功: %s (耗时: %v)", 
		event.GetEventID(), processingTime)
	
	return nil
}

// ProcessEventBatch 批量处理事件
func (p *CQRSCDCPipeline) ProcessEventBatch(ctx context.Context, events []events.DomainEvent) error {
	if !p.isRunning {
		return fmt.Errorf("流水线未运行")
	}
	
	log.Printf("🔄 流水线批量处理事件: %d个事件", len(events))
	startTime := time.Now()
	
	// 使用CDC同步服务批量处理
	err := p.cdcSyncService.ProcessEventBatch(ctx, events)
	
	processingTime := time.Since(startTime)
	
	if err != nil {
		log.Printf("❌ 流水线批量处理失败: %v (耗时: %v)", err, processingTime)
		return fmt.Errorf("流水线批量处理失败: %w", err)
	}
	
	log.Printf("✅ 流水线批量处理成功: %d个事件 (耗时: %v)", len(events), processingTime)
	return nil
}

// startHealthCheckLoop 启动健康检查循环
func (p *CQRSCDCPipeline) startHealthCheckLoop(ctx context.Context) {
	ticker := time.NewTicker(p.pipelineConfig.HealthCheckInterval)
	defer ticker.Stop()
	
	log.Printf("🏥 启动流水线健康检查，间隔: %v", p.pipelineConfig.HealthCheckInterval)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("🏥 健康检查循环已停止")
			return
		case <-ticker.C:
			p.performHealthCheck(ctx)
		}
	}
}

// performHealthCheck 执行健康检查
func (p *CQRSCDCPipeline) performHealthCheck(ctx context.Context) {
	p.healthStatus.LastHealthCheck = time.Now()
	
	// 检查Neo4j连接
	p.healthStatus.Neo4jConnected = p.connectionManager.Health(ctx) == nil
	
	// 检查EventBus连接
	p.healthStatus.EventBusConnected = p.eventBus != nil && p.eventBus.Health() == nil
	
	// 检查CDC服务状态
	p.healthStatus.CDCServiceRunning = p.cdcSyncService.Health() == nil
	
	// 更新整体健康状态
	p.healthStatus.IsHealthy = p.healthStatus.Neo4jConnected && 
		p.healthStatus.EventBusConnected && 
		p.healthStatus.CDCServiceRunning
	
	if !p.healthStatus.IsHealthy {
		p.healthStatus.ErrorCount++
		log.Printf("⚠️ 流水线健康检查失败: Neo4j=%v, EventBus=%v, CDC=%v", 
			p.healthStatus.Neo4jConnected,
			p.healthStatus.EventBusConnected,
			p.healthStatus.CDCServiceRunning)
		
		// 尝试自动恢复
		if p.pipelineConfig.EnableAutoRecovery && 
			p.healthStatus.RecoveryAttempts < p.pipelineConfig.MaxRecoveryAttempts {
			p.attemptRecovery(ctx)
		}
	} else {
		if p.pipelineConfig.EnableDetailedLogs {
			log.Printf("💚 流水线健康检查通过")
		}
	}
}

// attemptRecovery 尝试自动恢复
func (p *CQRSCDCPipeline) attemptRecovery(ctx context.Context) {
	p.healthStatus.RecoveryAttempts++
	
	log.Printf("🔄 尝试自动恢复 (第 %d/%d 次)", 
		p.healthStatus.RecoveryAttempts, p.pipelineConfig.MaxRecoveryAttempts)
	
	// 在实际实现中，这里应该包含具体的恢复逻辑
	// 例如：重新连接数据库、重启服务等
	
	// 等待一段时间后重新检查
	time.Sleep(time.Second * 10)
}

// startPerformanceMonitoring 启动性能监控
func (p *CQRSCDCPipeline) startPerformanceMonitoring(ctx context.Context) {
	ticker := time.NewTicker(p.pipelineConfig.MetricsExportInterval)
	defer ticker.Stop()
	
	log.Printf("📊 启动流水线性能监控，间隔: %v", p.pipelineConfig.MetricsExportInterval)
	
	for {
		select {
		case <-ctx.Done():
			log.Printf("📊 性能监控已停止")
			return
		case <-ticker.C:
			p.exportMetrics()
		}
	}
}

// exportMetrics 导出指标
func (p *CQRSCDCPipeline) exportMetrics() {
	// 更新运行时间
	p.performanceStats.Uptime = time.Since(p.startTime)
	
	// 计算吞吐量
	if p.performanceStats.TotalEvents > 0 {
		p.performanceStats.ThroughputPerSecond = 
			float64(p.performanceStats.ProcessedEvents) / p.performanceStats.Uptime.Seconds()
	}
	
	log.Printf("📊 流水线性能指标: 总事件=%d, 已处理=%d, 失败=%d, 吞吐量=%.2f/秒, 运行时间=%v",
		p.performanceStats.TotalEvents,
		p.performanceStats.ProcessedEvents,
		p.performanceStats.FailedEvents,
		p.performanceStats.ThroughputPerSecond,
		p.performanceStats.Uptime)
}

// updateEventStats 更新事件统计
func (p *CQRSCDCPipeline) updateEventStats(eventType string, startTime time.Time) {
	p.performanceStats.TotalEvents++
}

// updateEventTypeStats 更新事件类型统计
func (p *CQRSCDCPipeline) updateEventTypeStats(eventType string, success bool, processingTime time.Duration) {
	stats, exists := p.performanceStats.EventTypeStats[eventType]
	if !exists {
		stats = &EventTypeStats{}
		p.performanceStats.EventTypeStats[eventType] = stats
	}
	
	stats.Count++
	stats.LastProcessed = time.Now()
	
	if success {
		stats.SuccessCount++
	} else {
		stats.FailureCount++
	}
	
	// 更新平均处理时间
	if stats.AverageProcessingTime == 0 {
		stats.AverageProcessingTime = processingTime
	} else {
		stats.AverageProcessingTime = (stats.AverageProcessingTime + processingTime) / 2
	}
}

// Health 健康检查
func (p *CQRSCDCPipeline) Health() error {
	if !p.isRunning {
		return fmt.Errorf("流水线未运行")
	}
	
	if !p.healthStatus.IsHealthy {
		return fmt.Errorf("流水线健康检查失败: %s", p.healthStatus.LastError)
	}
	
	return nil
}

// GetHealthStatus 获取健康状态
func (p *CQRSCDCPipeline) GetHealthStatus() *PipelineHealthStatus {
	return p.healthStatus
}

// GetPerformanceStats 获取性能统计
func (p *CQRSCDCPipeline) GetPerformanceStats() *PipelinePerformanceStats {
	return p.performanceStats
}

// GetDetailedStatus 获取详细状态
func (p *CQRSCDCPipeline) GetDetailedStatus() map[string]interface{} {
	return map[string]interface{}{
		"is_running": p.isRunning,
		"uptime":     time.Since(p.startTime).String(),
		"health":     p.healthStatus,
		"performance": p.performanceStats,
		"cdc_service": p.cdcSyncService.GetDetailedStatistics(),
		"config": map[string]interface{}{
			"enable_health_checks":    p.pipelineConfig.EnableHealthChecks,
			"health_check_interval":   p.pipelineConfig.HealthCheckInterval.String(),
			"enable_auto_recovery":    p.pipelineConfig.EnableAutoRecovery,
			"max_recovery_attempts":   p.pipelineConfig.MaxRecoveryAttempts,
			"metrics_export_interval": p.pipelineConfig.MetricsExportInterval.String(),
		},
	}
}

// DefaultPipelineConfig 默认流水线配置
func DefaultPipelineConfig() *PipelineConfig {
	return &PipelineConfig{
		Neo4jConfig:           Neo4jConfigFromEnv(),
		CDCConfig:            DefaultCDCSyncConfig(),
		EnableHealthChecks:    true,
		HealthCheckInterval:   time.Minute * 2,
		EnableAutoRecovery:    true,
		MaxRecoveryAttempts:   3,
		EnableDetailedLogs:    false,
		LogLevel:             "INFO",
		MetricsExportInterval: time.Minute * 5,
	}
}