package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"database/sql"
	"sync"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

// CDC同步服务 - 实时监听PostgreSQL变更并同步到Neo4j

type CDCSyncService struct {
	pgDB          *sql.DB
	neo4jDriver   neo4j.DriverWithContext
	listener      *pq.Listener
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	syncStats     *SyncStatistics
}

type SyncStatistics struct {
	mu              sync.RWMutex
	TotalProcessed  int64
	SuccessCount    int64
	FailureCount    int64
	LastSyncTime    time.Time
	StartTime       time.Time
}

type OrganizationChange struct {
	Operation string          `json:"operation"`
	TableName string          `json:"table_name"`
	Timestamp time.Time       `json:"timestamp"`
	NewData   json.RawMessage `json:"new_data,omitempty"`
	OldData   json.RawMessage `json:"old_data,omitempty"`
}

type OrganizationData struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenant_id"`
	UnitType       string    `json:"unit_type"`
	Name           string    `json:"name"`
	Description    *string   `json:"description"`
	ParentUnitID   *string   `json:"parent_unit_id"`
	Status         string    `json:"status"`
	Level          int       `json:"level"`
	EmployeeCount  int       `json:"employee_count"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func main() {
	log.Println("🚀 启动CDC组织同步服务...")
	
	// 创建CDC同步服务
	service, err := NewCDCSyncService()
	if err != nil {
		log.Fatal("创建CDC同步服务失败:", err)
	}
	defer service.Close()
	
	// 启动同步服务
	if err := service.Start(); err != nil {
		log.Fatal("启动CDC同步服务失败:", err)
	}
	
	log.Println("✅ CDC同步服务已启动，按Ctrl+C停止...")
	
	// 等待中断信号
	service.Wait()
	
	log.Println("🛑 CDC同步服务已停止")
}

func NewCDCSyncService() (*CDCSyncService, error) {
	// 连接PostgreSQL
	pgDB, err := sql.Open("postgres", "host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("连接PostgreSQL失败: %w", err)
	}
	
	// 测试连接
	if err := pgDB.Ping(); err != nil {
		return nil, fmt.Errorf("PostgreSQL连接测试失败: %w", err)
	}
	
	// 连接Neo4j
	neo4jDriver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.BasicAuth("neo4j", "password", ""),
	)
	if err != nil {
		return nil, fmt.Errorf("连接Neo4j失败: %w", err)
	}
	
	// 创建PostgreSQL监听器
	listener := pq.NewListener("host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable",
		10*time.Second, time.Minute, func(ev pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("⚠️ PostgreSQL监听器事件错误: %v", err)
			}
		})
	
	ctx, cancel := context.WithCancel(context.Background())
	
	return &CDCSyncService{
		pgDB:        pgDB,
		neo4jDriver: neo4jDriver,
		listener:    listener,
		ctx:         ctx,
		cancel:      cancel,
		syncStats: &SyncStatistics{
			StartTime: time.Now(),
		},
	}, nil
}

func (s *CDCSyncService) Start() error {
	// 监听组织变更通知
	if err := s.listener.Listen("organization_change"); err != nil {
		return fmt.Errorf("监听PostgreSQL通知失败: %w", err)
	}
	
	log.Println("📡 开始监听组织变更通知...")
	
	// 启动同步协程
	s.wg.Add(1)
	go s.syncLoop()
	
	// 启动统计报告协程
	s.wg.Add(1)
	go s.statsReporter()
	
	return nil
}

func (s *CDCSyncService) syncLoop() {
	defer s.wg.Done()
	
	for {
		select {
		case <-s.ctx.Done():
			log.Println("📊 同步服务停止中...")
			return
			
		case notification := <-s.listener.Notify:
			if notification != nil {
				s.handleNotification(notification)
			}
			
		case <-time.After(30 * time.Second):
			// 定期检查连接状态
			if err := s.listener.Ping(); err != nil {
				log.Printf("⚠️ PostgreSQL连接检查失败: %v", err)
			}
		}
	}
}

func (s *CDCSyncService) handleNotification(notification *pq.Notification) {
	s.syncStats.mu.Lock()
	s.syncStats.TotalProcessed++
	s.syncStats.mu.Unlock()
	
	log.Printf("📨 收到组织变更通知: %s", notification.Extra)
	
	// 解析变更数据
	var change OrganizationChange
	if err := json.Unmarshal([]byte(notification.Extra), &change); err != nil {
		log.Printf("❌ 解析变更通知失败: %v", err)
		s.updateFailureStats()
		return
	}
	
	// 处理不同类型的变更
	switch change.Operation {
	case "INSERT":
		if err := s.handleInsert(change); err != nil {
			log.Printf("❌ 处理插入操作失败: %v", err)
			s.updateFailureStats()
			return
		}
	case "UPDATE":
		if err := s.handleUpdate(change); err != nil {
			log.Printf("❌ 处理更新操作失败: %v", err)
			s.updateFailureStats()
			return
		}
	case "DELETE":
		if err := s.handleDelete(change); err != nil {
			log.Printf("❌ 处理删除操作失败: %v", err)
			s.updateFailureStats()
			return
		}
	default:
		log.Printf("⚠️ 未知操作类型: %s", change.Operation)
		return
	}
	
	s.updateSuccessStats()
	s.updateSyncLog(change)
}

func (s *CDCSyncService) handleInsert(change OrganizationChange) error {
	// 解析新数据
	var orgData OrganizationData
	if err := json.Unmarshal(change.NewData, &orgData); err != nil {
		return fmt.Errorf("解析组织数据失败: %w", err)
	}
	
	log.Printf("➕ 同步新组织到Neo4j: %s", orgData.Name)
	
	session := s.neo4jDriver.NewSession(s.ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)
	
	_, err := session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
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
			"id":             orgData.ID,
			"tenant_id":      orgData.TenantID,
			"unit_type":      orgData.UnitType,
			"name":           orgData.Name,
			"description":    orgData.Description,
			"status":         orgData.Status,
			"level":          orgData.Level,
			"employee_count": orgData.EmployeeCount,
			"is_active":      orgData.IsActive,
			"created_at":     orgData.CreatedAt.Format(time.RFC3339),
			"updated_at":     orgData.UpdatedAt.Format(time.RFC3339),
			"sync_source":    "cdc_sync_service",
			"last_synced":    time.Now().Format(time.RFC3339),
		}
		
		_, err := tx.Run(s.ctx, cypher, params)
		if err != nil {
			return nil, err
		}
		
		// 处理父子关系
		if orgData.ParentUnitID != nil && *orgData.ParentUnitID != "" {
			relCypher := `
				MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
				MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
				MERGE (parent)-[:PARENT_OF]->(child)
			`
			
			_, err = tx.Run(s.ctx, relCypher, map[string]any{
				"parent_id": *orgData.ParentUnitID,
				"child_id":  orgData.ID,
				"tenant_id": orgData.TenantID,
			})
			if err != nil {
				log.Printf("⚠️ 创建父子关系失败: %v", err)
				// 不返回错误，让节点创建成功
			}
		}
		
		return "success", nil
	})
	
	return err
}

func (s *CDCSyncService) handleUpdate(change OrganizationChange) error {
	// 解析新数据
	var orgData OrganizationData
	if err := json.Unmarshal(change.NewData, &orgData); err != nil {
		return fmt.Errorf("解析组织数据失败: %w", err)
	}
	
	log.Printf("🔄 同步组织更新到Neo4j: %s", orgData.Name)
	
	session := s.neo4jDriver.NewSession(s.ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)
	
	_, err := session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		// 更新节点属性
		updateCypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			SET o.unit_type = $unit_type,
				o.name = $name,
				o.description = $description,
				o.status = $status,
				o.level = $level,
				o.employee_count = $employee_count,
				o.is_active = $is_active,
				o.updated_at = $updated_at,
				o.last_synced = $last_synced
		`
		
		params := map[string]any{
			"id":             orgData.ID,
			"tenant_id":      orgData.TenantID,
			"unit_type":      orgData.UnitType,
			"name":           orgData.Name,
			"description":    orgData.Description,
			"status":         orgData.Status,
			"level":          orgData.Level,
			"employee_count": orgData.EmployeeCount,
			"is_active":      orgData.IsActive,
			"updated_at":     orgData.UpdatedAt.Format(time.RFC3339),
			"last_synced":    time.Now().Format(time.RFC3339),
		}
		
		_, err := tx.Run(s.ctx, updateCypher, params)
		if err != nil {
			return nil, err
		}
		
		// 检查父子关系是否需要更新
		var oldData OrganizationData
		if err := json.Unmarshal(change.OldData, &oldData); err == nil {
			// 检查父组织是否发生变化
			oldParent := ""
			if oldData.ParentUnitID != nil {
				oldParent = *oldData.ParentUnitID
			}
			newParent := ""
			if orgData.ParentUnitID != nil {
				newParent = *orgData.ParentUnitID
			}
			
			if oldParent != newParent {
				// 删除旧关系
				if oldParent != "" {
					deleteCypher := `
						MATCH (parent:Organization {id: $old_parent_id, tenant_id: $tenant_id})-[r:PARENT_OF]->(child:Organization {id: $child_id, tenant_id: $tenant_id})
						DELETE r
					`
					_, err = tx.Run(s.ctx, deleteCypher, map[string]any{
						"old_parent_id": oldParent,
						"child_id":      orgData.ID,
						"tenant_id":     orgData.TenantID,
					})
					if err != nil {
						log.Printf("⚠️ 删除旧父子关系失败: %v", err)
					}
				}
				
				// 创建新关系
				if newParent != "" {
					createCypher := `
						MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
						MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
						MERGE (parent)-[:PARENT_OF]->(child)
					`
					_, err = tx.Run(s.ctx, createCypher, map[string]any{
						"parent_id": newParent,
						"child_id":  orgData.ID,
						"tenant_id": orgData.TenantID,
					})
					if err != nil {
						log.Printf("⚠️ 创建新父子关系失败: %v", err)
					}
				}
			}
		}
		
		return "success", nil
	})
	
	return err
}

func (s *CDCSyncService) handleDelete(change OrganizationChange) error {
	// 解析旧数据
	var orgData OrganizationData
	if err := json.Unmarshal(change.OldData, &orgData); err != nil {
		return fmt.Errorf("解析组织数据失败: %w", err)
	}
	
	log.Printf("🗑️ 从Neo4j删除组织: %s", orgData.Name)
	
	session := s.neo4jDriver.NewSession(s.ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(s.ctx)
	
	_, err := session.ExecuteWrite(s.ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (o:Organization {id: $id, tenant_id: $tenant_id})
			DETACH DELETE o
		`
		
		_, err := tx.Run(s.ctx, cypher, map[string]any{
			"id":        orgData.ID,
			"tenant_id": orgData.TenantID,
		})
		
		return "success", err
	})
	
	return err
}

func (s *CDCSyncService) updateSuccessStats() {
	s.syncStats.mu.Lock()
	s.syncStats.SuccessCount++
	s.syncStats.LastSyncTime = time.Now()
	s.syncStats.mu.Unlock()
}

func (s *CDCSyncService) updateFailureStats() {
	s.syncStats.mu.Lock()
	s.syncStats.FailureCount++
	s.syncStats.mu.Unlock()
}

func (s *CDCSyncService) updateSyncLog(change OrganizationChange) {
	// 提取组织ID
	var orgID string
	if change.Operation == "DELETE" {
		var orgData OrganizationData
		if err := json.Unmarshal(change.OldData, &orgData); err == nil {
			orgID = orgData.ID
		}
	} else {
		var orgData OrganizationData
		if err := json.Unmarshal(change.NewData, &orgData); err == nil {
			orgID = orgData.ID
		}
	}
	
	if orgID != "" {
		updateQuery := `
			UPDATE sync_monitoring 
			SET sync_status = 'SUCCESS', synced_at = NOW() 
			WHERE entity_id = $1 AND operation_type = $2 AND sync_status = 'PENDING'
		`
		
		operationType := ""
		switch change.Operation {
		case "INSERT":
			operationType = "CREATE"
		case "UPDATE":
			operationType = "UPDATE"
		case "DELETE":
			operationType = "DELETE"
		}
		
		if _, err := s.pgDB.Exec(updateQuery, orgID, operationType); err != nil {
			log.Printf("⚠️ 更新同步日志失败: %v", err)
		}
	}
}

func (s *CDCSyncService) statsReporter() {
	defer s.wg.Done()
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.reportStats()
		}
	}
}

func (s *CDCSyncService) reportStats() {
	s.syncStats.mu.RLock()
	stats := *s.syncStats
	s.syncStats.mu.RUnlock()
	
	uptime := time.Since(stats.StartTime).Round(time.Second)
	successRate := float64(0)
	if stats.TotalProcessed > 0 {
		successRate = float64(stats.SuccessCount) / float64(stats.TotalProcessed) * 100
	}
	
	log.Printf("📊 CDC同步统计 - 运行时间: %v, 总处理: %d, 成功: %d, 失败: %d, 成功率: %.1f%%",
		uptime, stats.TotalProcessed, stats.SuccessCount, stats.FailureCount, successRate)
}

func (s *CDCSyncService) Wait() {
	// 等待中断信号
	// 这里简化处理，实际应该监听SIGINT/SIGTERM
	time.Sleep(24 * time.Hour) // 运行24小时
	s.Stop()
}

func (s *CDCSyncService) Stop() {
	log.Println("🛑 停止CDC同步服务...")
	s.cancel()
	s.wg.Wait()
}

func (s *CDCSyncService) Close() {
	s.Stop()
	
	if s.listener != nil {
		s.listener.Close()
	}
	
	if s.pgDB != nil {
		s.pgDB.Close()
	}
	
	if s.neo4jDriver != nil {
		s.neo4jDriver.Close(s.ctx)
	}
}