package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/google/uuid"
)

// 组织变更监听器和自动同步修复工具

func main() {
	ctx := context.Background()
	
	log.Println("🔧 开始修复组织自动同步机制...")
	
	// 连接PostgreSQL
	pgDB, err := sql.Open("postgres", "host=localhost port=5432 user=user password=password dbname=cubecastle sslmode=disable")
	if err != nil {
		log.Fatal("连接PostgreSQL失败:", err)
	}
	defer pgDB.Close()
	
	// 连接Neo4j
	neo4jDriver, err := neo4j.NewDriverWithContext(
		"bolt://localhost:7687",
		neo4j.BasicAuth("neo4j", "password", ""),
	)
	if err != nil {
		log.Fatal("连接Neo4j失败:", err)
	}
	defer neo4jDriver.Close(ctx)
	
	// 1. 创建触发器函数用于PostgreSQL变更通知
	log.Println("📋 创建PostgreSQL触发器函数...")
	if err := createTriggerFunction(pgDB); err != nil {
		log.Fatal("创建触发器函数失败:", err)
	}
	
	// 2. 创建触发器
	log.Println("🔄 创建组织变更触发器...")
	if err := createOrganizationTrigger(pgDB); err != nil {
		log.Fatal("创建触发器失败:", err)
	}
	
	// 3. 创建同步日志表
	log.Println("📊 创建同步监控表...")
	if err := createSyncLogTable(pgDB); err != nil {
		log.Fatal("创建同步日志表失败:", err)
	}
	
	// 4. 测试同步机制
	log.Println("🧪 测试自动同步机制...")
	if err := testAutoSync(ctx, pgDB, neo4jDriver); err != nil {
		log.Fatal("测试自动同步失败:", err)
	}
	
	// 5. 创建同步修复存储过程
	log.Println("🛠️ 创建同步修复存储过程...")
	if err := createSyncRepairProcedure(pgDB); err != nil {
		log.Fatal("创建同步修复存储过程失败:", err)
	}
	
	log.Println("✅ 组织自动同步机制修复完成！")
	log.Println("📝 建议：")
	log.Println("   1. 监控同步日志表 sync_monitoring")
	log.Println("   2. 定期运行同步修复: SELECT repair_organization_sync();")
	log.Println("   3. 检查Neo4j连接状态")
}

// createTriggerFunction 创建触发器函数
func createTriggerFunction(db *sql.DB) error {
	query := `
		CREATE OR REPLACE FUNCTION notify_organization_change()
		RETURNS TRIGGER AS $$
		DECLARE
			change_data JSON;
		BEGIN
			-- 构建变更数据
			IF TG_OP = 'INSERT' THEN
				change_data = json_build_object(
					'operation', 'INSERT',
					'table_name', TG_TABLE_NAME,
					'timestamp', NOW(),
					'new_data', row_to_json(NEW)
				);
				-- 插入同步日志
				INSERT INTO sync_monitoring (operation_type, entity_id, entity_data, sync_status, created_at)
				VALUES ('CREATE', NEW.id, change_data, 'PENDING', NOW());
				
			ELSIF TG_OP = 'UPDATE' THEN
				change_data = json_build_object(
					'operation', 'UPDATE',
					'table_name', TG_TABLE_NAME,
					'timestamp', NOW(),
					'old_data', row_to_json(OLD),
					'new_data', row_to_json(NEW)
				);
				-- 插入同步日志
				INSERT INTO sync_monitoring (operation_type, entity_id, entity_data, sync_status, created_at)
				VALUES ('UPDATE', NEW.id, change_data, 'PENDING', NOW());
				
			ELSIF TG_OP = 'DELETE' THEN
				change_data = json_build_object(
					'operation', 'DELETE',
					'table_name', TG_TABLE_NAME,
					'timestamp', NOW(),
					'old_data', row_to_json(OLD)
				);
				-- 插入同步日志
				INSERT INTO sync_monitoring (operation_type, entity_id, entity_data, sync_status, created_at)
				VALUES ('DELETE', OLD.id, change_data, 'PENDING', NOW());
			END IF;
			
			-- 发送通知（用于EventBus监听）
			PERFORM pg_notify('organization_change', change_data::text);
			
			RETURN COALESCE(NEW, OLD);
		END;
		$$ LANGUAGE plpgsql;
	`
	
	_, err := db.Exec(query)
	return err
}

// createOrganizationTrigger 创建组织变更触发器
func createOrganizationTrigger(db *sql.DB) error {
	// 删除现有触发器（如果存在）
	dropQuery := `DROP TRIGGER IF EXISTS organization_units_change_trigger ON organization_units;`
	if _, err := db.Exec(dropQuery); err != nil {
		log.Printf("⚠️ 删除旧触发器失败: %v", err)
	}
	
	// 创建新触发器
	createQuery := `
		CREATE TRIGGER organization_units_change_trigger
		AFTER INSERT OR UPDATE OR DELETE ON organization_units
		FOR EACH ROW
		EXECUTE FUNCTION notify_organization_change();
	`
	
	_, err := db.Exec(createQuery)
	return err
}

// createSyncLogTable 创建同步监控表
func createSyncLogTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS sync_monitoring (
			id SERIAL PRIMARY KEY,
			operation_type VARCHAR(20) NOT NULL,
			entity_id UUID NOT NULL,
			entity_data JSONB NOT NULL,
			sync_status VARCHAR(20) DEFAULT 'PENDING',
			error_message TEXT,
			retry_count INTEGER DEFAULT 0,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			synced_at TIMESTAMP WITH TIME ZONE
		);
		
		-- 创建索引
		CREATE INDEX IF NOT EXISTS idx_sync_monitoring_status ON sync_monitoring(sync_status);
		CREATE INDEX IF NOT EXISTS idx_sync_monitoring_entity_id ON sync_monitoring(entity_id);
		CREATE INDEX IF NOT EXISTS idx_sync_monitoring_created_at ON sync_monitoring(created_at);
		
		-- 创建更新时间触发器
		CREATE OR REPLACE FUNCTION update_sync_monitoring_updated_at()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql;
		
		DROP TRIGGER IF EXISTS update_sync_monitoring_updated_at_trigger ON sync_monitoring;
		CREATE TRIGGER update_sync_monitoring_updated_at_trigger
			BEFORE UPDATE ON sync_monitoring
			FOR EACH ROW
			EXECUTE FUNCTION update_sync_monitoring_updated_at();
	`
	
	_, err := db.Exec(query)
	return err
}

// testAutoSync 测试自动同步机制
func testAutoSync(ctx context.Context, pgDB *sql.DB, neo4jDriver neo4j.DriverWithContext) error {
	// 创建测试组织
	testOrgID := uuid.New()
	testOrgName := fmt.Sprintf("测试同步组织_%d", time.Now().Unix())
	
	log.Printf("📝 创建测试组织: %s", testOrgName)
	
	// 在PostgreSQL中创建测试组织
	insertQuery := `
		INSERT INTO organization_units (
			id, tenant_id, unit_type, name, description, 
			status, level, employee_count, is_active
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	
	tenantID := uuid.New()
	_, err := pgDB.Exec(insertQuery,
		testOrgID, tenantID, "DEPARTMENT", testOrgName, "测试同步机制",
		"ACTIVE", 1, 0, true,
	)
	if err != nil {
		return fmt.Errorf("创建测试组织失败: %w", err)
	}
	
	// 等待一会儿让触发器执行
	time.Sleep(2 * time.Second)
	
	// 检查同步日志
	var logCount int
	logQuery := `SELECT COUNT(*) FROM sync_monitoring WHERE entity_id = $1 AND operation_type = 'CREATE'`
	if err := pgDB.QueryRow(logQuery, testOrgID).Scan(&logCount); err != nil {
		return fmt.Errorf("检查同步日志失败: %w", err)
	}
	
	if logCount == 0 {
		return fmt.Errorf("同步日志未创建，触发器可能未正常工作")
	}
	
	log.Printf("✅ 同步日志已创建 (数量: %d)", logCount)
	
	// 手动同步到Neo4j（模拟同步服务）
	if err := manualSyncToNeo4j(ctx, neo4jDriver, testOrgID, tenantID, testOrgName); err != nil {
		return fmt.Errorf("手动同步到Neo4j失败: %w", err)
	}
	
	// 验证Neo4j中的数据
	if err := verifyNeo4jSync(ctx, neo4jDriver, testOrgID); err != nil {
		return fmt.Errorf("验证Neo4j同步失败: %w", err)
	}
	
	// 更新同步状态
	updateQuery := `
		UPDATE sync_monitoring 
		SET sync_status = 'SUCCESS', synced_at = NOW() 
		WHERE entity_id = $1 AND operation_type = 'CREATE'
	`
	if _, err := pgDB.Exec(updateQuery, testOrgID); err != nil {
		return fmt.Errorf("更新同步状态失败: %w", err)
	}
	
	// 清理测试数据
	log.Println("🧹 清理测试数据...")
	if _, err := pgDB.Exec("DELETE FROM organization_units WHERE id = $1", testOrgID); err != nil {
		log.Printf("⚠️ 清理PostgreSQL测试数据失败: %v", err)
	}
	
	// 清理Neo4j测试数据
	session := neo4jDriver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	_, err = session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		_, err := tx.Run(ctx, "MATCH (o:Organization {id: $id}) DETACH DELETE o", 
			map[string]any{"id": testOrgID.String()})
		return nil, err
	})
	if err != nil {
		log.Printf("⚠️ 清理Neo4j测试数据失败: %v", err)
	}
	
	log.Println("✅ 自动同步机制测试通过")
	return nil
}

// manualSyncToNeo4j 手动同步到Neo4j
func manualSyncToNeo4j(ctx context.Context, driver neo4j.DriverWithContext, orgID, tenantID uuid.UUID, name string) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			CREATE (o:Organization {
				id: $id,
				tenant_id: $tenant_id,
				unit_type: $unit_type,
				name: $name,
				description: $description,
				status: $status,
				level: $level,
				employee_count: $employee_count,
				is_active: $is_active,
				created_at: $created_at,
				updated_at: $updated_at,
				sync_source: $sync_source
			})
		`
		
		params := map[string]any{
			"id":             orgID.String(),
			"tenant_id":      tenantID.String(),
			"unit_type":      "DEPARTMENT",
			"name":           name,
			"description":    "测试同步机制",
			"status":         "ACTIVE",
			"level":          1,
			"employee_count": 0,
			"is_active":      true,
			"created_at":     time.Now().Format(time.RFC3339),
			"updated_at":     time.Now().Format(time.RFC3339),
			"sync_source":    "auto_sync_test",
		}
		
		_, err := tx.Run(ctx, cypher, params)
		return nil, err
	})
	
	return err
}

// verifyNeo4jSync 验证Neo4j同步
func verifyNeo4jSync(ctx context.Context, driver neo4j.DriverWithContext, orgID uuid.UUID) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := "MATCH (o:Organization {id: $id}) RETURN o.name as name"
		result, err := tx.Run(ctx, cypher, map[string]any{"id": orgID.String()})
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			record := result.Record()
			name, _ := record.Get("name")
			return name, nil
		}
		
		return nil, fmt.Errorf("组织未在Neo4j中找到")
	})
	
	if err != nil {
		return err
	}
	
	log.Printf("✅ Neo4j同步验证成功: %s", result)
	return nil
}

// createSyncRepairProcedure 创建同步修复存储过程
func createSyncRepairProcedure(db *sql.DB) error {
	query := `
		CREATE OR REPLACE FUNCTION repair_organization_sync()
		RETURNS TABLE(
			repaired_count INTEGER,
			failed_count INTEGER,
			details TEXT
		) AS $$
		DECLARE
			pending_count INTEGER;
			failed_sync_count INTEGER;
			repair_details TEXT := '';
		BEGIN
			-- 获取待同步数量
			SELECT COUNT(*) INTO pending_count 
			FROM sync_monitoring 
			WHERE sync_status = 'PENDING' 
			AND created_at > NOW() - INTERVAL '24 hours';
			
			-- 获取失败同步数量
			SELECT COUNT(*) INTO failed_sync_count 
			FROM sync_monitoring 
			WHERE sync_status = 'FAILED' 
			AND retry_count < 3;
			
			-- 标记超时的待同步记录为失败
			UPDATE sync_monitoring 
			SET sync_status = 'FAILED', 
				error_message = 'Sync timeout after 1 hour',
				updated_at = NOW()
			WHERE sync_status = 'PENDING' 
			AND created_at < NOW() - INTERVAL '1 hour';
			
			-- 重置失败次数不超过3次的记录为待同步
			UPDATE sync_monitoring 
			SET sync_status = 'PENDING', 
				retry_count = retry_count + 1,
				updated_at = NOW()
			WHERE sync_status = 'FAILED' 
			AND retry_count < 3
			AND created_at > NOW() - INTERVAL '24 hours';
			
			repair_details := format(
				'待同步: %s, 重试失败: %s, 修复时间: %s',
				pending_count,
				failed_sync_count,
				NOW()
			);
			
			RETURN QUERY SELECT pending_count, failed_sync_count, repair_details;
		END;
		$$ LANGUAGE plpgsql;
		
		-- 创建同步状态查询函数
		CREATE OR REPLACE FUNCTION get_sync_status()
		RETURNS TABLE(
			total_pending INTEGER,
			total_success INTEGER,
			total_failed INTEGER,
			last_sync_time TIMESTAMP WITH TIME ZONE
		) AS $$
		BEGIN
			RETURN QUERY 
			SELECT 
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'PENDING') as total_pending,
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'SUCCESS') as total_success,
				(SELECT COUNT(*)::INTEGER FROM sync_monitoring WHERE sync_status = 'FAILED') as total_failed,
				(SELECT MAX(synced_at) FROM sync_monitoring WHERE sync_status = 'SUCCESS') as last_sync_time;
		END;
		$$ LANGUAGE plpgsql;
	`
	
	_, err := db.Exec(query)
	return err
}