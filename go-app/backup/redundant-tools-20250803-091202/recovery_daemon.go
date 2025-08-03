package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	_ "github.com/lib/pq"
)

// 自动同步故障恢复守护进程

type RecoveryDaemon struct {
	pgDB        *sql.DB
	neo4jDriver neo4j.DriverWithContext
	ctx         context.Context
	cancel      context.CancelFunc
	config      *RecoveryConfig
	stats       *RecoveryStats
}

type RecoveryConfig struct {
	CheckInterval      time.Duration // 检查间隔
	MaxFailureCount    int           // 最大失败次数阈值
	RecoveryBatchSize  int           // 恢复批次大小
	MaxRecoveryRetries int           // 最大恢复重试次数
	HealthCheckTimeout time.Duration // 健康检查超时
	LogLevel           string        // 日志级别
}

type RecoveryStats struct {
	TotalChecks          int64     // 总检查次数
	FailuresDetected     int64     // 检测到的故障次数
	RecoveriesAttempted  int64     // 尝试恢复次数
	RecoveriesSucceeded  int64     // 恢复成功次数
	RecoveriesFailed     int64     // 恢复失败次数
	LastCheckTime        time.Time // 最后检查时间
	LastRecoveryTime     time.Time // 最后恢复时间
	IsHealthy            bool      // 当前健康状态
}

type SyncIssue struct {
	Type        string    `json:"type"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Timestamp   time.Time `json:"timestamp"`
	Count       int       `json:"count"`
}

func main() {
	log.Println("🛡️ 启动自动同步故障恢复守护进程...")
	
	// 创建恢复守护进程
	daemon, err := NewRecoveryDaemon()
	if err != nil {
		log.Fatal("创建恢复守护进程失败:", err)
	}
	defer daemon.Close()
	
	// 启动守护进程
	if err := daemon.Start(); err != nil {
		log.Fatal("启动恢复守护进程失败:", err)
	}
	
	// 等待中断信号
	daemon.WaitForShutdown()
	
	log.Println("🛑 恢复守护进程已停止")
}

func NewRecoveryDaemon() (*RecoveryDaemon, error) {
	// 连接PostgreSQL
	pgDB, err := sql.Open("postgres", "host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}
	
	// 连接Neo4j
	neo4jDriver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.BasicAuth("neo4j", "password", ""),
	)
	if err != nil {
		return nil, fmt.Errorf("连接Neo4j失败: %w", err)
	}
	
	ctx, cancel := context.WithCancel(context.Background())
	
	config := &RecoveryConfig{
		CheckInterval:      30 * time.Second,
		MaxFailureCount:    3,
		RecoveryBatchSize:  50,
		MaxRecoveryRetries: 3,
		HealthCheckTimeout: 10 * time.Second,
		LogLevel:           "INFO",
	}
	
	stats := &RecoveryStats{
		IsHealthy: true,
	}
	
	return &RecoveryDaemon{
		pgDB:        pgDB,
		neo4jDriver: neo4jDriver,
		ctx:         ctx,
		cancel:      cancel,
		config:      config,
		stats:       stats,
	}, nil
}

func (d *RecoveryDaemon) Start() error {
	log.Printf("🔄 守护进程启动，检查间隔: %v", d.config.CheckInterval)
	
	// 初始健康检查
	if err := d.initialHealthCheck(); err != nil {
		return fmt.Errorf("初始健康检查失败: %w", err)
	}
	
	// 启动监控循环
	go d.monitoringLoop()
	
	// 启动状态报告器
	go d.statusReporter()
	
	// 启动恢复日志清理器
	go d.cleanupOldLogs()
	
	log.Println("✅ 恢复守护进程启动成功")
	return nil
}

func (d *RecoveryDaemon) initialHealthCheck() error {
	log.Println("🔍 执行初始健康检查...")
	
	// 检查PostgreSQL连接
	if err := d.pgDB.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL连接失败: %w", err)
	}
	
	// 检查Neo4j连接
	if err := d.neo4jDriver.VerifyConnectivity(d.ctx); err != nil {
		return fmt.Errorf("Neo4j连接失败: %w", err)
	}
	
	// 检查同步监控表
	var tableExists bool
	checkTableQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sync_monitoring'
		);
	`
	if err := d.pgDB.QueryRow(checkTableQuery).Scan(&tableExists); err != nil {
		return fmt.Errorf("检查同步监控表失败: %w", err)
	}
	
	if !tableExists {
		return fmt.Errorf("同步监控表不存在，请先运行初始化脚本")
	}
	
	log.Println("✅ 初始健康检查通过")
	return nil
}

func (d *RecoveryDaemon) monitoringLoop() {
	ticker := time.NewTicker(d.config.CheckInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-d.ctx.Done():
			log.Println("📊 监控循环已停止")
			return
		case <-ticker.C:
			d.performHealthCheck()
		}
	}
}

func (d *RecoveryDaemon) performHealthCheck() {
	d.stats.TotalChecks++
	d.stats.LastCheckTime = time.Now()
	
	log.Printf("🔍 执行健康检查 #%d...", d.stats.TotalChecks)
	
	// 检测同步问题
	issues, err := d.detectSyncIssues()
	if err != nil {
		log.Printf("❌ 检测同步问题失败: %v", err)
		d.stats.FailuresDetected++
		return
	}
	
	// 如果发现问题，尝试恢复
	if len(issues) > 0 {
		d.stats.FailuresDetected++
		d.stats.IsHealthy = false
		
		log.Printf("⚠️ 检测到 %d 个同步问题:", len(issues))
		for _, issue := range issues {
			log.Printf("   - %s: %s (严重程度: %s)", issue.Type, issue.Description, issue.Severity)
		}
		
		// 尝试自动恢复
		d.attemptRecovery(issues)
	} else {
		d.stats.IsHealthy = true
		log.Println("✅ 健康检查通过")
	}
}

func (d *RecoveryDaemon) detectSyncIssues() ([]SyncIssue, error) {
	var issues []SyncIssue
	
	// 检查待同步的数量
	var pendingCount int
	pendingQuery := `
		SELECT COUNT(*) FROM sync_monitoring 
		WHERE sync_status = 'PENDING' 
		AND created_at < NOW() - INTERVAL '5 minutes'
	`
	if err := d.pgDB.QueryRow(pendingQuery).Scan(&pendingCount); err != nil {
		return nil, fmt.Errorf("检查待同步数量失败: %w", err)
	}
	
	if pendingCount > d.config.MaxFailureCount {
		issues = append(issues, SyncIssue{
			Type:        "PENDING_OVERFLOW",
			Description: fmt.Sprintf("待同步数量过多: %d 个", pendingCount),
			Severity:    "HIGH",
			Timestamp:   time.Now(),
			Count:       pendingCount,
		})
	}
	
	// 检查失败的同步
	var failedCount int
	failedQuery := `
		SELECT COUNT(*) FROM sync_monitoring 
		WHERE sync_status = 'FAILED' 
		AND created_at > NOW() - INTERVAL '1 hour'
	`
	if err := d.pgDB.QueryRow(failedQuery).Scan(&failedCount); err != nil {
		return nil, fmt.Errorf("检查失败同步数量失败: %w", err)
	}
	
	if failedCount > 0 {
		issues = append(issues, SyncIssue{
			Type:        "SYNC_FAILURES",
			Description: fmt.Sprintf("同步失败数量: %d 个", failedCount),
			Severity:    "MEDIUM",
			Timestamp:   time.Now(),
			Count:       failedCount,
		})
	}
	
	// 检查数据一致性
	if err := d.checkDataConsistency(&issues); err != nil {
		log.Printf("⚠️ 数据一致性检查失败: %v", err)
	}
	
	// 检查连接健康状态
	if err := d.checkConnectionHealth(&issues); err != nil {
		log.Printf("⚠️ 连接健康检查失败: %v", err)
	}
	
	return issues, nil
}

func (d *RecoveryDaemon) checkDataConsistency(issues *[]SyncIssue) error {
	// 获取PostgreSQL中的组织数量
	var pgCount int
	pgQuery := "SELECT COUNT(*) FROM organization_units WHERE status = 'ACTIVE'"
	if err := d.pgDB.QueryRow(pgQuery).Scan(&pgCount); err != nil {
		return fmt.Errorf("获取PostgreSQL组织数量失败: %w", err)
	}
	
	// 获取Neo4j中的组织数量
	session := d.neo4jDriver.NewSession(d.ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(d.ctx)
	
	result, err := session.ExecuteRead(d.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := "MATCH (o:Organization {status: 'ACTIVE'}) RETURN count(o) as count"
		result, err := tx.Run(d.ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		
		if result.Next(d.ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			return count, nil
		}
		
		return 0, nil
	})
	
	if err != nil {
		return fmt.Errorf("获取Neo4j组织数量失败: %w", err)
	}
	
	neo4jCount := int(result.(int64))
	difference := pgCount - neo4jCount
	
	if difference != 0 {
		severity := "LOW"
		if difference > 5 {
			severity = "HIGH"
		} else if difference > 1 {
			severity = "MEDIUM"
		}
		
		*issues = append(*issues, SyncIssue{
			Type:        "DATA_INCONSISTENCY",
			Description: fmt.Sprintf("数据不一致: PostgreSQL %d vs Neo4j %d", pgCount, neo4jCount),
			Severity:    severity,
			Timestamp:   time.Now(),
			Count:       difference,
		})
	}
	
	return nil
}

func (d *RecoveryDaemon) checkConnectionHealth(issues *[]SyncIssue) error {
	// 检查PostgreSQL连接
	if err := d.pgDB.Ping(); err != nil {
		*issues = append(*issues, SyncIssue{
			Type:        "PG_CONNECTION",
			Description: "PostgreSQL连接失败",
			Severity:    "CRITICAL",
			Timestamp:   time.Now(),
			Count:       1,
		})
	}
	
	// 检查Neo4j连接
	if err := d.neo4jDriver.VerifyConnectivity(d.ctx); err != nil {
		*issues = append(*issues, SyncIssue{
			Type:        "NEO4J_CONNECTION",
			Description: "Neo4j连接失败",
			Severity:    "CRITICAL",
			Timestamp:   time.Now(),
			Count:       1,
		})
	}
	
	return nil
}

func (d *RecoveryDaemon) attemptRecovery(issues []SyncIssue) {
	d.stats.RecoveriesAttempted++
	d.stats.LastRecoveryTime = time.Now()
	
	log.Println("🔧 开始自动恢复...")
	
	recoverySuccess := true
	
	for _, issue := range issues {
		switch issue.Type {
		case "PENDING_OVERFLOW":
			if err := d.recoverPendingSync(); err != nil {
				log.Printf("❌ 恢复待同步数据失败: %v", err)
				recoverySuccess = false
			} else {
				log.Println("✅ 待同步数据恢复成功")
			}
			
		case "SYNC_FAILURES":
			if err := d.recoverFailedSync(); err != nil {
				log.Printf("❌ 恢复失败同步失败: %v", err)
				recoverySuccess = false
			} else {
				log.Println("✅ 失败同步恢复成功")
			}
			
		case "DATA_INCONSISTENCY":
			if err := d.recoverDataInconsistency(); err != nil {
				log.Printf("❌ 数据一致性恢复失败: %v", err)
				recoverySuccess = false
			} else {
				log.Println("✅ 数据一致性恢复成功")
			}
			
		case "PG_CONNECTION", "NEO4J_CONNECTION":
			log.Printf("⚠️ 连接问题需要手动干预: %s", issue.Description)
			recoverySuccess = false
			
		default:
			log.Printf("⚠️ 未知问题类型: %s", issue.Type)
		}
	}
	
	if recoverySuccess {
		d.stats.RecoveriesSucceeded++
		d.stats.IsHealthy = true
		log.Println("🎉 自动恢复完成")
	} else {
		d.stats.RecoveriesFailed++
		log.Println("❌ 部分恢复失败，需要人工干预")
	}
}

func (d *RecoveryDaemon) recoverPendingSync() error {
	// 重置超时的待同步记录为失败状态
	updateQuery := `
		UPDATE sync_monitoring 
		SET sync_status = 'FAILED',
			error_message = 'Timeout after 1 hour',
			updated_at = NOW()
		WHERE sync_status = 'PENDING' 
		AND created_at < NOW() - INTERVAL '1 hour'
	`
	
	result, err := d.pgDB.Exec(updateQuery)
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	log.Printf("📊 重置了 %d 个超时的待同步记录", rowsAffected)
	
	return nil
}

func (d *RecoveryDaemon) recoverFailedSync() error {
	// 调用修复存储过程
	repairQuery := "SELECT repair_organization_sync();"
	if _, err := d.pgDB.Exec(repairQuery); err != nil {
		return fmt.Errorf("调用修复存储过程失败: %w", err)
	}
	
	log.Println("📊 已调用同步修复存储过程")
	return nil
}

func (d *RecoveryDaemon) recoverDataInconsistency() error {
	// 触发全量数据同步
	log.Println("📊 开始数据一致性修复...")
	
	// 获取PostgreSQL中缺失的组织
	missingQuery := `
		SELECT ou.id, ou.name 
		FROM organization_units ou
		WHERE ou.status = 'ACTIVE'
		AND NOT EXISTS (
			SELECT 1 FROM sync_monitoring sm 
			WHERE sm.entity_id = ou.id 
			AND sm.sync_status = 'SUCCESS'
			AND sm.created_at > NOW() - INTERVAL '1 day'
		)
		LIMIT $1
	`
	
	rows, err := d.pgDB.Query(missingQuery, d.config.RecoveryBatchSize)
	if err != nil {
		return fmt.Errorf("查询缺失组织失败: %w", err)
	}
	defer rows.Close()
	
	count := 0
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			log.Printf("⚠️ 扫描组织记录失败: %v", err)
			continue
		}
		
		// 创建同步记录
		insertQuery := `
			INSERT INTO sync_monitoring (operation_type, entity_id, entity_data, sync_status, created_at)
			VALUES ('REPAIR', $1, '{"repair_type": "data_consistency", "name": "' || $2 || '"}', 'PENDING', NOW())
		`
		
		if _, err := d.pgDB.Exec(insertQuery, id, name); err != nil {
			log.Printf("⚠️ 创建修复记录失败 (%s): %v", name, err)
		} else {
			count++
		}
	}
	
	log.Printf("📊 创建了 %d 个数据一致性修复记录", count)
	return nil
}

func (d *RecoveryDaemon) statusReporter() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.reportStatus()
		}
	}
}

func (d *RecoveryDaemon) reportStatus() {
	healthStatus := "健康"
	if !d.stats.IsHealthy {
		healthStatus = "异常"
	}
	
	uptime := time.Since(d.stats.LastCheckTime).Round(time.Second)
	successRate := float64(0)
	if d.stats.RecoveriesAttempted > 0 {
		successRate = float64(d.stats.RecoveriesSucceeded) / float64(d.stats.RecoveriesAttempted) * 100
	}
	
	log.Printf("📊 恢复守护进程状态报告:")
	log.Printf("   状态: %s", healthStatus)
	log.Printf("   总检查: %d, 检测故障: %d", d.stats.TotalChecks, d.stats.FailuresDetected)
	log.Printf("   尝试恢复: %d, 成功: %d, 失败: %d", 
		d.stats.RecoveriesAttempted, d.stats.RecoveriesSucceeded, d.stats.RecoveriesFailed)
	log.Printf("   恢复成功率: %.1f%%", successRate)
	log.Printf("   运行时间: %v", uptime)
}

func (d *RecoveryDaemon) cleanupOldLogs() {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	
	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			d.performLogCleanup()
		}
	}
}

func (d *RecoveryDaemon) performLogCleanup() {
	// 清理7天前的同步日志
	cleanupQuery := `
		DELETE FROM sync_monitoring 
		WHERE created_at < NOW() - INTERVAL '7 days'
		AND sync_status IN ('SUCCESS', 'FAILED')
	`
	
	result, err := d.pgDB.Exec(cleanupQuery)
	if err != nil {
		log.Printf("⚠️ 清理旧日志失败: %v", err)
		return
	}
	
	rowsDeleted, _ := result.RowsAffected()
	if rowsDeleted > 0 {
		log.Printf("🧹 清理了 %d 条7天前的同步日志", rowsDeleted)
	}
}

func (d *RecoveryDaemon) WaitForShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	<-sigChan
	log.Println("🛑 收到停止信号，开始优雅关闭...")
	
	d.cancel()
	
	// 等待清理完成
	time.Sleep(2 * time.Second)
}

func (d *RecoveryDaemon) Close() {
	if d.pgDB != nil {
		d.pgDB.Close()
	}
	if d.neo4jDriver != nil {
		d.neo4jDriver.Close(d.ctx)
	}
}