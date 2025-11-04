package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	auth "cube-castle/internal/auth"
	pkglogger "cube-castle/pkg/logger"
)

// TemporalMonitor 时态数据监控服务
type TemporalMonitor struct {
	db     *sql.DB
	logger pkglogger.Logger
}

// NewTemporalMonitor 创建时态监控服务
func NewTemporalMonitor(db *sql.DB, baseLogger pkglogger.Logger) *TemporalMonitor {
	return &TemporalMonitor{
		db:     db,
		logger: scopedLogger(baseLogger, "temporalMonitor", nil),
	}
}

// MonitoringMetrics 监控指标
type MonitoringMetrics struct {
	TotalOrganizations    int       `json:"totalOrganizations"`
	CurrentRecords        int       `json:"currentRecords"`
	FutureRecords         int       `json:"futureRecords"`
	HistoricalRecords     int       `json:"historicalRecords"`
	DuplicateCurrentCount int       `json:"duplicateCurrentCount"`
	MissingCurrentCount   int       `json:"missingCurrentCount"`
	TimelineOverlapCount  int       `json:"timelineOverlapCount"`
	InconsistentFlagCount int       `json:"inconsistentFlagCount"`
	OrphanRecordCount     int       `json:"orphanRecordCount"`
	HealthScore           float64   `json:"healthScore"` // 0-100
	LastCheckTime         time.Time `json:"lastCheckTime"`
	AlertLevel            string    `json:"alertLevel"` // HEALTHY, WARNING, CRITICAL
}

// AlertRule 告警规则
type AlertRule struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Threshold   int    `json:"threshold"`
	AlertLevel  string `json:"alertLevel"`
}

// GetDefaultAlertRules 获取默认告警规则
func (m *TemporalMonitor) GetDefaultAlertRules() []AlertRule {
	return []AlertRule{
		{
			Name:        "DUPLICATE_CURRENT_RECORDS",
			Description: "重复的当前记录数量超过阈值",
			Threshold:   0, // 任何重复都是严重问题
			AlertLevel:  "CRITICAL",
		},
		{
			Name:        "MISSING_CURRENT_RECORDS",
			Description: "缺失当前记录的组织数量超过阈值",
			Threshold:   0, // 任何缺失都是严重问题
			AlertLevel:  "CRITICAL",
		},
		{
			Name:        "TIMELINE_OVERLAPS",
			Description: "时间线重叠记录数量超过阈值",
			Threshold:   0, // 任何重叠都是严重问题
			AlertLevel:  "CRITICAL",
		},
		{
			Name:        "INCONSISTENT_FLAGS",
			Description: "is_current/is_future标志不一致记录数量超过阈值",
			Threshold:   5, // 少量不一致可能是时间差导致
			AlertLevel:  "WARNING",
		},
		{
			Name:        "ORPHAN_RECORDS",
			Description: "孤立记录（父级不存在）数量超过阈值",
			Threshold:   10, // 少量孤立记录可以接受
			AlertLevel:  "WARNING",
		},
		{
			Name:        "HEALTH_SCORE",
			Description: "系统健康分数低于阈值",
			Threshold:   85, // 健康分数低于85%告警
			AlertLevel:  "WARNING",
		},
	}
}

// CollectMetrics 收集监控指标
func (m *TemporalMonitor) CollectMetrics(ctx context.Context) (*MonitoringMetrics, error) {
	// 多租户隔离：默认按请求上下文租户计算；若无租户（例如后台周期任务），则计算全局汇总，仅用于内部日志
	tenantID := auth.GetTenantID(ctx)
	metrics := &MonitoringMetrics{
		LastCheckTime: time.Now(),
		AlertLevel:    "HEALTHY",
	}

	// 1. 基础统计
	err := m.collectBasicStats(ctx, metrics, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to collect basic stats: %w", err)
	}

	// 2. 数据一致性检查
	err = m.collectConsistencyStats(ctx, metrics, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to collect consistency stats: %w", err)
	}

	// 3. 计算健康分数和告警级别
	m.calculateHealthScore(metrics)

	return metrics, nil
}

func (m *TemporalMonitor) collectBasicStats(ctx context.Context, metrics *MonitoringMetrics, tenantID string) error {
	// 统计总组织数
	var err error
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(DISTINCT code) FROM organization_units WHERE status <> 'DELETED' AND tenant_id = $1",
			tenantID,
		).Scan(&metrics.TotalOrganizations)
	} else {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(DISTINCT code) FROM organization_units WHERE status <> 'DELETED'",
		).Scan(&metrics.TotalOrganizations)
	}
	if err != nil {
		return fmt.Errorf("failed to count total organizations: %w", err)
	}

	// 统计当前记录数
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE is_current = true AND status <> 'DELETED' AND tenant_id = $1",
			tenantID,
		).Scan(&metrics.CurrentRecords)
	} else {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE is_current = true AND status <> 'DELETED'",
		).Scan(&metrics.CurrentRecords)
	}
	if err != nil {
		return fmt.Errorf("failed to count current records: %w", err)
	}

	// 统计未来记录数（派生条件）
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE effective_date > CURRENT_DATE AND status <> 'DELETED' AND tenant_id = $1",
			tenantID,
		).Scan(&metrics.FutureRecords)
	} else {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE effective_date > CURRENT_DATE AND status <> 'DELETED'",
		).Scan(&metrics.FutureRecords)
	}
	if err != nil {
		return fmt.Errorf("failed to count future records: %w", err)
	}

	// 统计历史记录数（派生条件：已结束）
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE end_date IS NOT NULL AND end_date <= CURRENT_DATE AND status <> 'DELETED' AND tenant_id = $1",
			tenantID,
		).Scan(&metrics.HistoricalRecords)
	} else {
		err = m.db.QueryRowContext(ctx,
			"SELECT COUNT(*) FROM organization_units WHERE end_date IS NOT NULL AND end_date <= CURRENT_DATE AND status <> 'DELETED'",
		).Scan(&metrics.HistoricalRecords)
	}
	if err != nil {
		return fmt.Errorf("failed to count historical records: %w", err)
	}

	return nil
}

func (m *TemporalMonitor) collectConsistencyStats(ctx context.Context, metrics *MonitoringMetrics, tenantID string) error {
	// 检查重复当前记录
	var err error
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT tenant_id, code
            FROM organization_units 
            WHERE is_current = true AND status <> 'DELETED' AND tenant_id = $1
            GROUP BY tenant_id, code
            HAVING COUNT(*) > 1
        ) duplicates
    `, tenantID).Scan(&metrics.DuplicateCurrentCount)
	} else {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT tenant_id, code
            FROM organization_units 
            WHERE is_current = true AND status <> 'DELETED'
            GROUP BY tenant_id, code
            HAVING COUNT(*) > 1
        ) duplicates
    `).Scan(&metrics.DuplicateCurrentCount)
	}
	if err != nil {
		return fmt.Errorf("failed to count duplicate current records: %w", err)
	}

	// 检查缺失当前记录
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT tenant_id, code
            FROM organization_units
            WHERE tenant_id = $1
              AND (tenant_id, code) NOT IN (
                SELECT tenant_id, code 
                FROM organization_units 
                WHERE is_current = true AND status <> 'DELETED' AND tenant_id = $1
              )
              AND (tenant_id, code) NOT IN (
                SELECT tenant_id, code
                FROM organization_units
                WHERE tenant_id = $1
                GROUP BY tenant_id, code
                HAVING MIN(CASE WHEN status <> 'DELETED' THEN effective_date ELSE NULL END) > CURRENT_DATE
              )
              AND EXISTS (
                SELECT 1 FROM organization_units u
                WHERE u.tenant_id = organization_units.tenant_id
                  AND u.code = organization_units.code
                  AND u.status <> 'DELETED'
              )
        ) missing
    `, tenantID).Scan(&metrics.MissingCurrentCount)
	} else {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT tenant_id, code
            FROM organization_units
            WHERE (tenant_id, code) NOT IN (
                SELECT tenant_id, code 
                FROM organization_units 
                WHERE is_current = true AND status <> 'DELETED'
            )
            AND (tenant_id, code) NOT IN (
                SELECT tenant_id, code
                FROM organization_units
                GROUP BY tenant_id, code
                HAVING MIN(CASE WHEN status <> 'DELETED' THEN effective_date ELSE NULL END) > CURRENT_DATE
            )
            AND EXISTS (
                SELECT 1 FROM organization_units u
                WHERE u.tenant_id = organization_units.tenant_id
                  AND u.code = organization_units.code
                  AND u.status <> 'DELETED'
            )
        ) missing
    `).Scan(&metrics.MissingCurrentCount)
	}
	if err != nil {
		return fmt.Errorf("failed to count missing current records: %w", err)
	}

	// 检查时间线重叠
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT o1.tenant_id, o1.code
            FROM organization_units o1
            JOIN organization_units o2 ON (
                o1.tenant_id = o2.tenant_id 
                AND o1.code = o2.code 
                AND o1.record_id != o2.record_id
            )
            WHERE 
                o1.status <> 'DELETED'
                AND o2.status <> 'DELETED'
                AND o1.tenant_id = $1
                AND o1.effective_date < COALESCE(o2.end_date, '9999-12-31'::date)
                AND o2.effective_date < COALESCE(o1.end_date, '9999-12-31'::date)
        ) AS timeline_overlaps
    `, tenantID).Scan(&metrics.TimelineOverlapCount)
	} else {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM (
            SELECT DISTINCT o1.tenant_id, o1.code
            FROM organization_units o1
            JOIN organization_units o2 ON (
                o1.tenant_id = o2.tenant_id 
                AND o1.code = o2.code 
                AND o1.record_id != o2.record_id
            )
            WHERE 
                o1.status <> 'DELETED'
                AND o2.status <> 'DELETED'
                AND o1.effective_date < COALESCE(o2.end_date, '9999-12-31'::date)
                AND o2.effective_date < COALESCE(o1.end_date, '9999-12-31'::date)
        ) AS timeline_overlaps
    `).Scan(&metrics.TimelineOverlapCount)
	}
	if err != nil {
		return fmt.Errorf("failed to count timeline overlaps: %w", err)
	}

	// 检查标志不一致记录（仅校验 is_current；is_future 已移除，使用派生值但不与列比较）
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM organization_units
        WHERE is_current != (
            effective_date <= CURRENT_DATE 
            AND (end_date IS NULL OR end_date > CURRENT_DATE)
        )
        AND status <> 'DELETED'
        AND tenant_id = $1
    `, tenantID).Scan(&metrics.InconsistentFlagCount)
	} else {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM organization_units
        WHERE is_current != (
            effective_date <= CURRENT_DATE 
            AND (end_date IS NULL OR end_date > CURRENT_DATE)
        )
        AND status <> 'DELETED'
    `).Scan(&metrics.InconsistentFlagCount)
	}
	if err != nil {
		return fmt.Errorf("failed to count inconsistent flags: %w", err)
	}

	// 检查孤立记录
	if tenantID != "" {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM organization_units o1
        WHERE 
            parent_code IS NOT NULL
            AND o1.status <> 'DELETED'
            AND o1.tenant_id = $1
            AND NOT EXISTS (
                SELECT 1 FROM organization_units o2 
                WHERE o2.tenant_id = o1.tenant_id 
                    AND o2.code = o1.parent_code 
                    AND o2.is_current = true
                    AND o2.status <> 'DELETED'
            )
    `, tenantID).Scan(&metrics.OrphanRecordCount)
	} else {
		err = m.db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM organization_units o1
        WHERE 
            parent_code IS NOT NULL
            AND o1.status <> 'DELETED'
            AND NOT EXISTS (
                SELECT 1 FROM organization_units o2 
                WHERE o2.tenant_id = o1.tenant_id 
                    AND o2.code = o1.parent_code 
                    AND o2.is_current = true
                    AND o2.status <> 'DELETED'
            )
    `).Scan(&metrics.OrphanRecordCount)
	}
	if err != nil {
		return fmt.Errorf("failed to count orphan records: %w", err)
	}

	return nil
}

func (m *TemporalMonitor) calculateHealthScore(metrics *MonitoringMetrics) {
	// 健康分数计算逻辑
	score := 100.0

	// 严重问题直接大幅扣分
	if metrics.DuplicateCurrentCount > 0 {
		score -= 40.0 // 重复当前记录是严重问题
	}
	if metrics.MissingCurrentCount > 0 {
		score -= 40.0 // 缺失当前记录是严重问题
	}
	if metrics.TimelineOverlapCount > 0 {
		score -= 30.0 // 时间线重叠是严重问题
	}

	// 轻微问题按比例扣分
	if metrics.InconsistentFlagCount > 0 {
		score -= float64(metrics.InconsistentFlagCount) * 2.0 // 每个不一致记录扣2分
	}
	if metrics.OrphanRecordCount > 0 {
		score -= float64(metrics.OrphanRecordCount) * 1.0 // 每个孤立记录扣1分
	}

	// 确保分数不低于0
	if score < 0 {
		score = 0
	}

	metrics.HealthScore = score

	// 确定告警级别
	if score < 50 || metrics.DuplicateCurrentCount > 0 || metrics.MissingCurrentCount > 0 || metrics.TimelineOverlapCount > 0 {
		metrics.AlertLevel = "CRITICAL"
	} else if score < 85 || metrics.InconsistentFlagCount > 5 || metrics.OrphanRecordCount > 10 {
		metrics.AlertLevel = "WARNING"
	} else {
		metrics.AlertLevel = "HEALTHY"
	}
}

// CheckAlerts 检查告警条件
func (m *TemporalMonitor) CheckAlerts(ctx context.Context) ([]string, error) {
	metrics, err := m.CollectMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to collect metrics: %w", err)
	}

	var alerts []string
	rules := m.GetDefaultAlertRules()

	for _, rule := range rules {
		var currentValue int
		var triggered bool

		switch rule.Name {
		case "DUPLICATE_CURRENT_RECORDS":
			currentValue = metrics.DuplicateCurrentCount
			triggered = currentValue > rule.Threshold
		case "MISSING_CURRENT_RECORDS":
			currentValue = metrics.MissingCurrentCount
			triggered = currentValue > rule.Threshold
		case "TIMELINE_OVERLAPS":
			currentValue = metrics.TimelineOverlapCount
			triggered = currentValue > rule.Threshold
		case "INCONSISTENT_FLAGS":
			currentValue = metrics.InconsistentFlagCount
			triggered = currentValue > rule.Threshold
		case "ORPHAN_RECORDS":
			currentValue = metrics.OrphanRecordCount
			triggered = currentValue > rule.Threshold
		case "HEALTH_SCORE":
			currentValue = int(metrics.HealthScore)
			triggered = currentValue < rule.Threshold
		}

		if triggered {
			alertMsg := fmt.Sprintf("[%s] %s: 当前值=%d, 阈值=%d",
				rule.AlertLevel, rule.Description, currentValue, rule.Threshold)
			alerts = append(alerts, alertMsg)
		}
	}

	// 记录监控结果到审计日志（使用新的标准化审计系统）
	// 注释掉旧的审计代码，等待统一重构时一起处理
	// metricsJSON, _ := json.Marshal(metrics)
	//
	// 系统健康监控结果可以单独记录，不必强制写入操作审计表
	// 可考虑使用专门的监控日志表或改为应用日志记录
	if len(alerts) > 0 {
		m.logger.Infof("监控结果: 健康分数=%.1f, 告警=%d个", metrics.HealthScore, len(alerts))
	}

	return alerts, nil
}

// GetMetricsHandler 获取监控指标的HTTP处理器函数
func (m *TemporalMonitor) GetMetricsHandler() func(ctx context.Context) (interface{}, error) {
	return func(ctx context.Context) (interface{}, error) {
		return m.CollectMetrics(ctx)
	}
}

// StartPeriodicMonitoring 启动定期监控
func (m *TemporalMonitor) StartPeriodicMonitoring(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				m.logger.Warn("停止时态数据监控服务")
				return
			case <-ticker.C:
				alerts, err := m.CheckAlerts(ctx)
				if err != nil {
					m.logger.Errorf("监控检查失败: %v", err)
					continue
				}

				if len(alerts) > 0 {
					m.logger.Warnf("发现 %d 个告警:", len(alerts))
					for _, alert := range alerts {
						m.logger.Warnf("告警详情: %s", alert)
					}
				} else {
					m.logger.Info("时态数据监控: 系统健康")
				}
			}
		}
	}()

	m.logger.Infof("时态数据监控服务已启动 (检查间隔: %v)", interval)
}
