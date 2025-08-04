package services

import (
	"context"
	"fmt"
	"time"
	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
)

// Type aliases for easier use
type Logger = logging.StructuredLogger

// MetricsCollector 性能指标收集器接口
type MetricsCollector interface {
	IncrementCounter(name string, tags map[string]string)
	RecordTimer(name string, duration time.Duration, tags map[string]string)
	SetGauge(name string, value float64, tags map[string]string)
}

// DataReconciliationService 数据对账服务
type DataReconciliationService struct {
	postgresRepo repositories.PositionCommandRepository
	neo4jRepo    repositories.PositionQueryRepository
	logger       Logger
	metrics      MetricsCollector
}

// ReconciliationReport 对账报告
type ReconciliationReport struct {
	StartTime          time.Time                    `json:"start_time"`
	EndTime            time.Time                    `json:"end_time"`
	TotalRecords       int                          `json:"total_records"`
	SuccessCount       int                          `json:"success_count"`
	ErrorCount         int                          `json:"error_count"`
	Inconsistencies    []DataInconsistency          `json:"inconsistencies"`
	RepairActions      []RepairAction               `json:"repair_actions"`
	PerformanceMetrics ReconciliationMetrics        `json:"performance_metrics"`
}

// DataInconsistency 数据不一致项
type DataInconsistency struct {
	ID                uuid.UUID                 `json:"id"`
	Type              InconsistencyType         `json:"type"`
	PostgresData      map[string]interface{}    `json:"postgres_data"`
	Neo4jData         map[string]interface{}    `json:"neo4j_data"`
	Severity          Severity                  `json:"severity"`
	Description       string                    `json:"description"`
	SuggestedAction   string                    `json:"suggested_action"`
	DetectedAt        time.Time                 `json:"detected_at"`
}

type InconsistencyType string
const (
	InconsistencyMissingInNeo4j     InconsistencyType = "MISSING_IN_NEO4J"
	InconsistencyMissingInPostgres  InconsistencyType = "MISSING_IN_POSTGRES"
	InconsistencyDataMismatch       InconsistencyType = "DATA_MISMATCH"
	InconsistencyRelationshipError  InconsistencyType = "RELATIONSHIP_ERROR"
)

type Severity string
const (
	SeverityLow      Severity = "LOW"
	SeverityMedium   Severity = "MEDIUM"
	SeverityHigh     Severity = "HIGH"
	SeverityCritical Severity = "CRITICAL"
)

// RepairAction 修复动作
type RepairAction struct {
	ID            uuid.UUID         `json:"id"`
	ActionType    RepairActionType  `json:"action_type"`
	Target        string            `json:"target"`
	Data          map[string]interface{} `json:"data"`
	Status        RepairStatus      `json:"status"`
	ExecutedAt    *time.Time        `json:"executed_at,omitempty"`
	ErrorMessage  *string           `json:"error_message,omitempty"`
}

type RepairActionType string
const (
	RepairCreateInNeo4j    RepairActionType = "CREATE_IN_NEO4J"
	RepairUpdateInNeo4j    RepairActionType = "UPDATE_IN_NEO4J"
	RepairDeleteInNeo4j    RepairActionType = "DELETE_IN_NEO4J"
	RepairCreateInPostgres RepairActionType = "CREATE_IN_POSTGRES"
	RepairUpdateInPostgres RepairActionType = "UPDATE_IN_POSTGRES"
)

type RepairStatus string
const (
	RepairPending   RepairStatus = "PENDING"
	RepairExecuted  RepairStatus = "EXECUTED"
	RepairFailed    RepairStatus = "FAILED"
	RepairSkipped   RepairStatus = "SKIPPED"
)

// ReconciliationMetrics 对账性能指标
type ReconciliationMetrics struct {
	PostgresQueryTime time.Duration `json:"postgres_query_time"`
	Neo4jQueryTime    time.Duration `json:"neo4j_query_time"`
	ComparisonTime    time.Duration `json:"comparison_time"`
	RepairTime        time.Duration `json:"repair_time"`
	MemoryUsage       int64         `json:"memory_usage_bytes"`
}

// ReconcilePositions 对账职位数据
func (s *DataReconciliationService) ReconcilePositions(ctx context.Context, tenantID uuid.UUID, opts ReconciliationOptions) (*ReconciliationReport, error) {
	startTime := time.Now()
	report := &ReconciliationReport{
		StartTime:       startTime,
		Inconsistencies: make([]DataInconsistency, 0),
		RepairActions:   make([]RepairAction, 0),
	}

	s.logger.Info("Starting position data reconciliation", "tenant_id", tenantID)

	// 1. 获取PostgreSQL中的职位数据
	pgStart := time.Now()
	pgPositions, err := s.getPostgresPositions(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres positions: %w", err)
	}
	report.PerformanceMetrics.PostgresQueryTime = time.Since(pgStart)

	// 2. 获取Neo4j中的职位数据
	neo4jStart := time.Now()
	neo4jPositions, err := s.getNeo4jPositions(ctx, tenantID, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get neo4j positions: %w", err)
	}
	report.PerformanceMetrics.Neo4jQueryTime = time.Since(neo4jStart)

	// 3. 比较数据并识别不一致
	compStart := time.Now()
	// 将切片转换为映射以便比较
	pgMap := make(map[uuid.UUID]interface{})
	for _, pos := range pgPositions {
		pgMap[pos.ID] = pos
	}
	neo4jMap := make(map[uuid.UUID]interface{})
	for _, pos := range neo4jPositions {
		neo4jMap[pos.ID] = pos
	}
	inconsistencies := s.comparePositionData(pgMap, neo4jMap)
	report.Inconsistencies = inconsistencies
	report.PerformanceMetrics.ComparisonTime = time.Since(compStart)

	// 4. 生成修复动作
	repairActions := s.generateRepairActions(inconsistencies)
	report.RepairActions = repairActions

	// 5. 执行修复（如果启用）
	if opts.AutoRepair {
		repairStart := time.Now()
		s.executeRepairActions(ctx, repairActions)
		report.PerformanceMetrics.RepairTime = time.Since(repairStart)
	}

	// 统计信息
	report.TotalRecords = len(pgPositions) + len(neo4jPositions)
	report.ErrorCount = len(inconsistencies)
	report.SuccessCount = report.TotalRecords - report.ErrorCount
	report.EndTime = time.Now()

	s.logger.Info("Position data reconciliation completed",
		"tenant_id", tenantID,
		"duration", report.EndTime.Sub(report.StartTime),
		"inconsistencies", len(inconsistencies))

	return report, nil
}

// ReconciliationOptions 对账选项
type ReconciliationOptions struct {
	AutoRepair      bool      `json:"auto_repair"`
	RepairSeverity  Severity  `json:"repair_severity"`
	BatchSize       int       `json:"batch_size"`
	MaxRecords      int       `json:"max_records"`
	CreatedAfter    *time.Time `json:"created_after,omitempty"`
	CreatedBefore   *time.Time `json:"created_before,omitempty"`
}

func (s *DataReconciliationService) comparePositionData(pgPositions, neo4jPositions map[uuid.UUID]interface{}) []DataInconsistency {
	var inconsistencies []DataInconsistency

	// 检查PostgreSQL中存在但Neo4j中缺失的记录
	for id, pgData := range pgPositions {
		if _, exists := neo4jPositions[id]; !exists {
			inconsistencies = append(inconsistencies, DataInconsistency{
				ID:              uuid.New(),
				Type:            InconsistencyMissingInNeo4j,
				PostgresData:    pgData.(map[string]interface{}),
				Neo4jData:       nil,
				Severity:        SeverityHigh,
				Description:     fmt.Sprintf("Position %s exists in PostgreSQL but missing in Neo4j", id),
				SuggestedAction: "Create position in Neo4j",
				DetectedAt:      time.Now(),
			})
		}
	}

	// 检查Neo4j中存在但PostgreSQL中缺失的记录
	for id, neo4jData := range neo4jPositions {
		if _, exists := pgPositions[id]; !exists {
			inconsistencies = append(inconsistencies, DataInconsistency{
				ID:              uuid.New(),
				Type:            InconsistencyMissingInPostgres,
				PostgresData:    nil,
				Neo4jData:       neo4jData.(map[string]interface{}),
				Severity:        SeverityCritical,
				Description:     fmt.Sprintf("Position %s exists in Neo4j but missing in PostgreSQL", id),
				SuggestedAction: "Investigate data integrity issue",
				DetectedAt:      time.Now(),
			})
		}
	}

	// 检查字段值不匹配
	for id, pgData := range pgPositions {
		if neo4jData, exists := neo4jPositions[id]; exists {
			if fieldMismatches := s.compareFields(pgData, neo4jData); len(fieldMismatches) > 0 {
				inconsistencies = append(inconsistencies, DataInconsistency{
					ID:              uuid.New(),
					Type:            InconsistencyDataMismatch,
					PostgresData:    pgData.(map[string]interface{}),
					Neo4jData:       neo4jData.(map[string]interface{}),
					Severity:        SeverityMedium,
					Description:     fmt.Sprintf("Position %s has field mismatches: %v", id, fieldMismatches),
					SuggestedAction: "Update Neo4j with PostgreSQL data",
					DetectedAt:      time.Now(),
				})
			}
		}
	}

	return inconsistencies
}

func (s *DataReconciliationService) compareFields(pgData, neo4jData interface{}) []string {
	var mismatches []string
	
	pg := pgData.(map[string]interface{})
	neo4j := neo4jData.(map[string]interface{})
	
	// 比较关键字段
	keyFields := []string{"status", "position_type", "department_id", "budgeted_fte"}
	
	for _, field := range keyFields {
		if pgVal, pgOk := pg[field]; pgOk {
			if neo4jVal, neo4jOk := neo4j[field]; neo4jOk {
				if pgVal != neo4jVal {
					mismatches = append(mismatches, fmt.Sprintf("%s: pg=%v, neo4j=%v", field, pgVal, neo4jVal))
				}
			} else {
				mismatches = append(mismatches, fmt.Sprintf("%s: missing in neo4j", field))
			}
		}
	}
	
	return mismatches
}

// SyncHealthMonitor 同步健康监控器
type SyncHealthMonitor struct {
	reconciliationService *DataReconciliationService
	alertManager         AlertManager
	logger               Logger
}

type AlertManager interface {
	SendAlert(ctx context.Context, alert Alert) error
}

type Alert struct {
	Level       AlertLevel `json:"level"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Timestamp   time.Time  `json:"timestamp"`
}

type AlertLevel string
const (
	AlertInfo     AlertLevel = "INFO"
	AlertWarning  AlertLevel = "WARNING"
	AlertError    AlertLevel = "ERROR"
	AlertCritical AlertLevel = "CRITICAL"
)

// MonitorSyncHealth 监控同步健康状态
func (m *SyncHealthMonitor) MonitorSyncHealth(ctx context.Context) error {
	// 定期对账检查
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := m.performHealthCheck(ctx); err != nil {
				m.logger.Error("Sync health check failed", "error", err)
			}
		}
	}
}

func (m *SyncHealthMonitor) performHealthCheck(ctx context.Context) error {
	// 获取所有租户并执行快速对账检查
	tenants := []uuid.UUID{} // 从配置或数据库获取租户列表
	
	for _, tenantID := range tenants {
		opts := ReconciliationOptions{
			AutoRepair:    false, // 监控模式下不自动修复
			BatchSize:     100,
			MaxRecords:    1000,
			CreatedAfter:  timePtr(time.Now().Add(-24 * time.Hour)), // 只检查最近24小时的数据
		}
		
		report, err := m.reconciliationService.ReconcilePositions(ctx, tenantID, opts)
		if err != nil {
			m.alertManager.SendAlert(ctx, Alert{
				Level:       AlertError,
				Title:       "Data Reconciliation Failed",
				Description: fmt.Sprintf("Failed to reconcile data for tenant %s: %v", tenantID, err),
				Metadata:    map[string]interface{}{"tenant_id": tenantID},
				Timestamp:   time.Now(),
			})
			continue
		}
		
		// 检查是否有严重的不一致
		criticalInconsistencies := 0
		for _, inc := range report.Inconsistencies {
			if inc.Severity == SeverityCritical {
				criticalInconsistencies++
			}
		}
		
		if criticalInconsistencies > 0 {
			m.alertManager.SendAlert(ctx, Alert{
				Level:       AlertCritical,
				Title:       "Critical Data Inconsistencies Detected",
				Description: fmt.Sprintf("Found %d critical inconsistencies for tenant %s", criticalInconsistencies, tenantID),
				Metadata: map[string]interface{}{
					"tenant_id":               tenantID,
					"critical_inconsistencies": criticalInconsistencies,
					"total_inconsistencies":   len(report.Inconsistencies),
				},
				Timestamp: time.Now(),
			})
		}
	}
	
	return nil
}

// getPostgresPositions 获取PostgreSQL中的职位数据
func (s *DataReconciliationService) getPostgresPositions(ctx context.Context, tenantID uuid.UUID, opts ReconciliationOptions) ([]repositories.Position, error) {
	// 简化实现：返回空列表
	// 实际实现中需要调用PostgreSQL查询
	return []repositories.Position{}, nil
}

// getNeo4jPositions 获取Neo4j中的职位数据
func (s *DataReconciliationService) getNeo4jPositions(ctx context.Context, tenantID uuid.UUID, opts ReconciliationOptions) ([]repositories.Position, error) {
	// 简化实现：返回空列表
	// 实际实现中需要调用Neo4j查询
	return []repositories.Position{}, nil
}

// generateRepairActions 生成修复动作
func (s *DataReconciliationService) generateRepairActions(inconsistencies []DataInconsistency) []RepairAction {
	// 简化实现：返回空列表
	// 实际实现中根据不一致类型生成相应的修复动作
	return []RepairAction{}
}


// executeRepairActions 执行修复动作
func (s *DataReconciliationService) executeRepairActions(ctx context.Context, actions []RepairAction) {
	// 简化实现：什么都不做
	// 实际实现中执行具体的修复操作
}

// timePtr 辅助函数，返回时间指针
func timePtr(t time.Time) *time.Time {
	return &t
}