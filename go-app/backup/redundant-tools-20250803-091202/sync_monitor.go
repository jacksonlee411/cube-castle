package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	_ "github.com/lib/pq"
)

// 组织同步监控和状态检查工具

type SyncMonitor struct {
	pgDB        *sql.DB
	neo4jDriver neo4j.DriverWithContext
}

type SyncStatus struct {
	PostgreSQLStatus bool      `json:"postgresql_status"`
	Neo4jStatus      bool      `json:"neo4j_status"`
	TotalPending     int       `json:"total_pending"`
	TotalSuccess     int       `json:"total_success"`
	TotalFailed      int       `json:"total_failed"`
	LastSyncTime     time.Time `json:"last_sync_time"`
	SyncLagSeconds   int       `json:"sync_lag_seconds"`
	DataConsistency  bool      `json:"data_consistency"`
	Issues           []string  `json:"issues"`
}

type OrganizationStats struct {
	PostgreSQLCount  int    `json:"postgresql_count"`
	Neo4jCount       int    `json:"neo4j_count"`
	CountDifference  int    `json:"count_difference"`
	MissingInNeo4j   []string `json:"missing_in_neo4j"`
	ExtraInNeo4j     []string `json:"extra_in_neo4j"`
}

func main() {
	log.Println("📊 启动组织同步监控工具...")
	
	monitor, err := NewSyncMonitor()
	if err != nil {
		log.Fatal("创建监控工具失败:", err)
	}
	defer monitor.Close()
	
	// 执行监控检查
	log.Println("🔍 执行同步状态检查...")
	status, err := monitor.CheckSyncStatus()
	if err != nil {
		log.Fatal("检查同步状态失败:", err)
	}
	
	// 显示监控结果
	monitor.DisplayStatus(status)
	
	// 执行数据一致性检查
	log.Println("🔍 执行数据一致性检查...")
	stats, err := monitor.CheckDataConsistency()
	if err != nil {
		log.Fatal("检查数据一致性失败:", err)
	}
	
	// 显示一致性结果
	monitor.DisplayConsistencyStats(stats)
	
	// 如果发现问题，提供修复建议
	if len(status.Issues) > 0 {
		log.Println("⚠️ 发现问题，提供修复建议...")
		monitor.ProvideRecommendations(status, stats)
	}
	
	// 执行自动修复（如果需要）
	if stats.CountDifference > 0 {
		log.Println("🔧 执行自动数据修复...")
		if err := monitor.AutoRepair(stats); err != nil {
			log.Printf("❌ 自动修复失败: %v", err)
		} else {
			log.Println("✅ 自动修复完成")
		}
	}
	
	log.Println("📊 监控检查完成")
}

func NewSyncMonitor() (*SyncMonitor, error) {
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
	
	return &SyncMonitor{
		pgDB:        pgDB,
		neo4jDriver: neo4jDriver,
	}, nil
}

func (m *SyncMonitor) CheckSyncStatus() (*SyncStatus, error) {
	ctx := context.Background()
	status := &SyncStatus{
		Issues: []string{},
	}
	
	// 检查PostgreSQL连接
	if err := m.pgDB.Ping(); err != nil {
		status.PostgreSQLStatus = false
		status.Issues = append(status.Issues, "PostgreSQL连接失败")
	} else {
		status.PostgreSQLStatus = true
	}
	
	// 检查Neo4j连接
	if err := m.neo4jDriver.VerifyConnectivity(ctx); err != nil {
		status.Neo4jStatus = false
		status.Issues = append(status.Issues, "Neo4j连接失败")
	} else {
		status.Neo4jStatus = true
	}
	
	// 检查同步日志表是否存在
	var tableExists bool
	checkTableQuery := `
		SELECT EXISTS (
			SELECT FROM information_schema.tables 
			WHERE table_schema = 'public' 
			AND table_name = 'sync_monitoring'
		);
	`
	if err := m.pgDB.QueryRow(checkTableQuery).Scan(&tableExists); err != nil {
		status.Issues = append(status.Issues, "无法检查同步监控表")
	} else if !tableExists {
		status.Issues = append(status.Issues, "同步监控表不存在")
	} else {
		// 获取同步统计
		if err := m.getSyncStats(status); err != nil {
			status.Issues = append(status.Issues, fmt.Sprintf("获取同步统计失败: %v", err))
		}
	}
	
	// 计算同步延迟
	if !status.LastSyncTime.IsZero() {
		status.SyncLagSeconds = int(time.Since(status.LastSyncTime).Seconds())
		if status.SyncLagSeconds > 300 { // 5分钟
			status.Issues = append(status.Issues, "同步延迟过高 (>5分钟)")
		}
	}
	
	return status, nil
}

func (m *SyncMonitor) getSyncStats(status *SyncStatus) error {
	query := `
		SELECT 
			COUNT(CASE WHEN sync_status = 'PENDING' THEN 1 END) as pending,
			COUNT(CASE WHEN sync_status = 'SUCCESS' THEN 1 END) as success,
			COUNT(CASE WHEN sync_status = 'FAILED' THEN 1 END) as failed,
			MAX(synced_at) as last_sync
		FROM sync_monitoring
		WHERE created_at > NOW() - INTERVAL '24 hours'
	`
	
	var lastSync sql.NullTime
	err := m.pgDB.QueryRow(query).Scan(
		&status.TotalPending,
		&status.TotalSuccess,
		&status.TotalFailed,
		&lastSync,
	)
	
	if err != nil {
		return err
	}
	
	if lastSync.Valid {
		status.LastSyncTime = lastSync.Time
	}
	
	return nil
}

func (m *SyncMonitor) CheckDataConsistency() (*OrganizationStats, error) {
	ctx := context.Background()
	stats := &OrganizationStats{
		MissingInNeo4j: []string{},
		ExtraInNeo4j:   []string{},
	}
	
	// 获取PostgreSQL中的组织数量
	pgQuery := "SELECT COUNT(*) FROM organization_units WHERE status = 'ACTIVE'"
	if err := m.pgDB.QueryRow(pgQuery).Scan(&stats.PostgreSQLCount); err != nil {
		return nil, fmt.Errorf("获取PostgreSQL组织数量失败: %w", err)
	}
	
	// 获取Neo4j中的组织数量
	session := m.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := "MATCH (o:Organization {status: 'ACTIVE'}) RETURN count(o) as count"
		result, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			return count, nil
		}
		
		return 0, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("获取Neo4j组织数量失败: %w", err)
	}
	
	stats.Neo4jCount = int(result.(int64))
	stats.CountDifference = stats.PostgreSQLCount - stats.Neo4jCount
	
	// 查找缺失的组织
	if stats.CountDifference != 0 {
		if err := m.findMissingOrganizations(ctx, stats); err != nil {
			return nil, fmt.Errorf("查找缺失组织失败: %w", err)
		}
	}
	
	return stats, nil
}

func (m *SyncMonitor) findMissingOrganizations(ctx context.Context, stats *OrganizationStats) error {
	// 获取PostgreSQL中的所有组织ID
	pgQuery := "SELECT id, name FROM organization_units WHERE status = 'ACTIVE'"
	rows, err := m.pgDB.Query(pgQuery)
	if err != nil {
		return err
	}
	defer rows.Close()
	
	pgOrgs := make(map[string]string) // id -> name
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return err
		}
		pgOrgs[id] = name
	}
	
	// 获取Neo4j中的所有组织ID
	session := m.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	neo4jResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := "MATCH (o:Organization {status: 'ACTIVE'}) RETURN o.id as id, o.name as name"
		result, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		
		neo4jOrgs := make(map[string]string)
		for result.Next(ctx) {
			record := result.Record()
			id, _ := record.Get("id")
			name, _ := record.Get("name")
			neo4jOrgs[id.(string)] = name.(string)
		}
		
		return neo4jOrgs, nil
	})
	
	if err != nil {
		return err
	}
	
	neo4jOrgs := neo4jResult.(map[string]string)
	
	// 查找缺失的组织
	for id, name := range pgOrgs {
		if _, exists := neo4jOrgs[id]; !exists {
			stats.MissingInNeo4j = append(stats.MissingInNeo4j, fmt.Sprintf("%s (%s)", name, id))
		}
	}
	
	// 查找多余的组织
	for id, name := range neo4jOrgs {
		if _, exists := pgOrgs[id]; !exists {
			stats.ExtraInNeo4j = append(stats.ExtraInNeo4j, fmt.Sprintf("%s (%s)", name, id))
		}
	}
	
	return nil
}

func (m *SyncMonitor) DisplayStatus(status *SyncStatus) {
	log.Println("📊 同步状态报告:")
	log.Printf("   PostgreSQL连接: %v", status.PostgreSQLStatus)
	log.Printf("   Neo4j连接: %v", status.Neo4jStatus)
	log.Printf("   待同步数量: %d", status.TotalPending)
	log.Printf("   成功同步数量: %d", status.TotalSuccess)
	log.Printf("   失败同步数量: %d", status.TotalFailed)
	
	if !status.LastSyncTime.IsZero() {
		log.Printf("   最后同步时间: %s", status.LastSyncTime.Format("2006-01-02 15:04:05"))
		log.Printf("   同步延迟: %d秒", status.SyncLagSeconds)
	}
	
	if len(status.Issues) > 0 {
		log.Println("⚠️ 发现的问题:")
		for _, issue := range status.Issues {
			log.Printf("   - %s", issue)
		}
	} else {
		log.Println("✅ 没有发现问题")
	}
}

func (m *SyncMonitor) DisplayConsistencyStats(stats *OrganizationStats) {
	log.Println("📊 数据一致性报告:")
	log.Printf("   PostgreSQL组织数量: %d", stats.PostgreSQLCount)
	log.Printf("   Neo4j组织数量: %d", stats.Neo4jCount)
	log.Printf("   数量差异: %d", stats.CountDifference)
	
	if len(stats.MissingInNeo4j) > 0 {
		log.Println("❌ Neo4j中缺失的组织:")
		for _, org := range stats.MissingInNeo4j {
			log.Printf("   - %s", org)
		}
	}
	
	if len(stats.ExtraInNeo4j) > 0 {
		log.Println("❌ Neo4j中多余的组织:")
		for _, org := range stats.ExtraInNeo4j {
			log.Printf("   - %s", org)
		}
	}
	
	if stats.CountDifference == 0 {
		log.Println("✅ 数据一致性正常")
	}
}

func (m *SyncMonitor) ProvideRecommendations(status *SyncStatus, stats *OrganizationStats) {
	log.Println("💡 修复建议:")
	
	if !status.PostgreSQLStatus {
		log.Println("   1. 检查PostgreSQL服务是否运行")
		log.Println("   2. 验证数据库连接配置")
	}
	
	if !status.Neo4jStatus {
		log.Println("   1. 检查Neo4j服务是否运行")
		log.Println("   2. 验证Neo4j连接配置和认证")
	}
	
	if status.TotalFailed > 0 {
		log.Println("   1. 检查失败的同步记录:")
		log.Println("      SELECT * FROM sync_monitoring WHERE sync_status = 'FAILED' ORDER BY created_at DESC LIMIT 10;")
		log.Println("   2. 运行同步修复命令:")
		log.Println("      SELECT repair_organization_sync();")
	}
	
	if status.SyncLagSeconds > 300 {
		log.Println("   1. 检查CDC触发器是否正常工作")
		log.Println("   2. 检查事件总线状态")
		log.Println("   3. 重启同步服务")
	}
	
	if stats.CountDifference > 0 {
		log.Println("   1. 运行数据一致性修复:")
		log.Println("      go run fix_organization_sync.go")
		log.Println("   2. 检查同步日志了解根本原因")
	}
}

func (m *SyncMonitor) AutoRepair(stats *OrganizationStats) error {
	if len(stats.MissingInNeo4j) == 0 {
		return nil
	}
	
	ctx := context.Background()
	log.Printf("🔧 开始自动修复 %d 个缺失的组织...", len(stats.MissingInNeo4j))
	
	// 获取缺失组织的详细信息
	for _, missingInfo := range stats.MissingInNeo4j {
		// 从字符串中提取ID（格式: "name (id)"）
		var orgID string
		fmt.Sscanf(missingInfo, "%*s (%s)", &orgID)
		
		if err := m.syncSingleOrganization(ctx, orgID); err != nil {
			log.Printf("❌ 修复组织 %s 失败: %v", orgID, err)
		} else {
			log.Printf("✅ 修复组织 %s 成功", orgID)
		}
	}
	
	return nil
}

func (m *SyncMonitor) syncSingleOrganization(ctx context.Context, orgID string) error {
	// 从PostgreSQL获取组织信息
	query := `
		SELECT id, tenant_id, unit_type, name, description, 
		       parent_unit_id, status, level, employee_count, 
		       is_active, created_at, updated_at
		FROM organization_units 
		WHERE id = $1
	`
	
	var org struct {
		ID           string
		TenantID     string
		UnitType     string
		Name         string
		Description  *string
		ParentUnitID *string
		Status       string
		Level        int
		EmployeeCount int
		IsActive     bool
		CreatedAt    time.Time
		UpdatedAt    time.Time
	}
	
	err := m.pgDB.QueryRow(query, orgID).Scan(
		&org.ID, &org.TenantID, &org.UnitType, &org.Name, &org.Description,
		&org.ParentUnitID, &org.Status, &org.Level, &org.EmployeeCount,
		&org.IsActive, &org.CreatedAt, &org.UpdatedAt,
	)
	
	if err != nil {
		return fmt.Errorf("获取组织信息失败: %w", err)
	}
	
	// 同步到Neo4j
	session := m.neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MERGE (o:Organization {id: $id, tenant_id: $tenant_id})
			SET o.unit_type = $unit_type,
				o.name = $name,
				o.description = $description,
				o.status = $status,
				o.level = $level,
				o.employee_count = $employee_count,
				o.is_active = $is_active,
				o.created_at = $created_at,
				o.updated_at = $updated_at,
				o.sync_source = $sync_source,
				o.last_synced = $last_synced
		`
		
		params := map[string]any{
			"id":             org.ID,
			"tenant_id":      org.TenantID,
			"unit_type":      org.UnitType,
			"name":           org.Name,
			"description":    org.Description,
			"status":         org.Status,
			"level":          org.Level,
			"employee_count": org.EmployeeCount,
			"is_active":      org.IsActive,
			"created_at":     org.CreatedAt.Format(time.RFC3339),
			"updated_at":     org.UpdatedAt.Format(time.RFC3339),
			"sync_source":    "auto_repair",
			"last_synced":    time.Now().Format(time.RFC3339),
		}
		
		_, err := tx.Run(ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 处理父子关系
		if org.ParentUnitID != nil && *org.ParentUnitID != "" {
			relCypher := `
				MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
				MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
				MERGE (parent)-[:PARENT_OF]->(child)
			`
			
			_, err = tx.Run(ctx, relCypher, map[string]any{
				"parent_id": *org.ParentUnitID,
				"child_id":  org.ID,
				"tenant_id": org.TenantID,
			})
			if err != nil {
				log.Printf("⚠️ 创建父子关系失败: %v", err)
			}
		}
		
		return "success", nil
	})
	
	return err
}

func (m *SyncMonitor) Close() {
	if m.pgDB != nil {
		m.pgDB.Close()
	}
	if m.neo4jDriver != nil {
		ctx := context.Background()
		m.neo4jDriver.Close(ctx)
	}
}