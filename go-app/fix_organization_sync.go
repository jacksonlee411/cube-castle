package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"database/sql"
	_ "github.com/lib/pq"
)

// OrganizationUnit represents an organization unit from PostgreSQL
type OrganizationUnit struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenant_id"`
	UnitType     string    `json:"unit_type"`
	Name         string    `json:"name"`
	Description  *string   `json:"description"`
	ParentUnitID *string   `json:"parent_unit_id"`
	Status       string    `json:"status"`
	Level        int       `json:"level"`
	Profile      string    `json:"profile"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func main() {
	ctx := context.Background()
	
	log.Println("🚀 开始修复组织单元同步...")
	
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
	
	// 获取所有组织单元
	log.Println("📋 从PostgreSQL获取组织单元数据...")
	orgs, err := getOrganizationUnits(pgDB)
	if err != nil {
		log.Fatal("获取组织单元失败:", err)
	}
	
	log.Printf("📊 找到 %d 个组织单元", len(orgs))
	
	// 清理Neo4j中现有的组织数据
	log.Println("🧹 清理Neo4j中现有的组织数据...")
	if err := cleanupNeo4jOrganizations(ctx, neo4jDriver); err != nil {
		log.Fatal("清理Neo4j数据失败:", err)
	}
	
	// 同步组织单元到Neo4j
	log.Println("🔄 开始同步组织单元到Neo4j...")
	if err := syncOrganizationsToNeo4j(ctx, neo4jDriver, orgs); err != nil {
		log.Fatal("同步组织单元失败:", err)
	}
	
	// 建立层级关系
	log.Println("🔗 建立组织层级关系...")
	if err := createHierarchyRelationships(ctx, neo4jDriver, orgs); err != nil {
		log.Fatal("建立层级关系失败:", err)
	}
	
	// 验证同步结果
	log.Println("✅ 验证同步结果...")
	if err := verifySync(ctx, neo4jDriver, len(orgs)); err != nil {
		log.Fatal("验证失败:", err)
	}
	
	log.Println("🎉 组织单元同步修复完成！")
}

// getOrganizationUnits 从PostgreSQL获取所有组织单元
func getOrganizationUnits(db *sql.DB) ([]OrganizationUnit, error) {
	query := `
		SELECT 
			id, tenant_id, unit_type, name, description, 
			parent_unit_id, status, level, 
			COALESCE(profile::text, '{}') as profile,
			created_at, updated_at
		FROM organization_units 
		ORDER BY level, name
	`
	
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var orgs []OrganizationUnit
	for rows.Next() {
		var org OrganizationUnit
		err := rows.Scan(
			&org.ID, &org.TenantID, &org.UnitType, &org.Name, &org.Description,
			&org.ParentUnitID, &org.Status, &org.Level, &org.Profile,
			&org.CreatedAt, &org.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		orgs = append(orgs, org)
	}
	
	return orgs, nil
}

// cleanupNeo4jOrganizations 清理Neo4j中现有的组织数据
func cleanupNeo4jOrganizations(ctx context.Context, driver neo4j.DriverWithContext) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	// 删除所有Organization节点和关系
	_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `MATCH (o:Organization) DETACH DELETE o`
		_, err := tx.Run(ctx, cypher, nil)
		return nil, err
	})
	
	if err != nil {
		return fmt.Errorf("清理Neo4j组织数据失败: %w", err)
	}
	
	log.Println("✅ Neo4j组织数据清理完成")
	return nil
}

// syncOrganizationsToNeo4j 同步组织单元到Neo4j
func syncOrganizationsToNeo4j(ctx context.Context, driver neo4j.DriverWithContext, orgs []OrganizationUnit) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	for i, org := range orgs {
		_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
			// 解析profile JSON
			var profile map[string]interface{}
			if err := json.Unmarshal([]byte(org.Profile), &profile); err != nil {
				log.Printf("⚠️ 组织 %s 的profile解析失败，使用空对象: %v", org.Name, err)
				profile = make(map[string]interface{})
			}
			
			cypher := `
				CREATE (o:Organization {
					id: $id,
					tenant_id: $tenant_id,
					unit_type: $unit_type,
					name: $name,
					description: $description,
					status: $status,
					level: $level,
					is_active: $is_active,
					profile: $profile,
					created_at: $created_at,
					updated_at: $updated_at,
					synced_at: $synced_at
				})
				RETURN o.id as org_id
			`
			
			params := map[string]interface{}{
				"id":          org.ID,
				"tenant_id":   org.TenantID,
				"unit_type":   org.UnitType,
				"name":        org.Name,
				"description": org.Description,
				"status":      org.Status,
				"level":       org.Level,
				"is_active":   org.Status == "ACTIVE",
				"profile":     org.Profile,
				"created_at":  org.CreatedAt.Format(time.RFC3339),
				"updated_at":  org.UpdatedAt.Format(time.RFC3339),
				"synced_at":   time.Now().Format(time.RFC3339),
			}
			
			result, err := tx.Run(ctx, cypher, params)
			if err != nil {
				return nil, fmt.Errorf("创建组织节点失败 (%s): %w", org.Name, err)
			}
			
			if result.Next(ctx) {
				record := result.Record()
				orgID, _ := record.Get("org_id")
				log.Printf("✅ 创建组织节点: %s (ID: %s) [%d/%d]", org.Name, orgID, i+1, len(orgs))
			}
			
			return nil, nil
		})
		
		if err != nil {
			return err
		}
	}
	
	log.Printf("🎯 成功创建 %d 个组织节点", len(orgs))
	return nil
}

// createHierarchyRelationships 建立组织层级关系
func createHierarchyRelationships(ctx context.Context, driver neo4j.DriverWithContext, orgs []OrganizationUnit) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeWrite,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	relationshipCount := 0
	
	for _, org := range orgs {
		if org.ParentUnitID != nil && *org.ParentUnitID != "" {
			_, err := session.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
				cypher := `
					MATCH (parent:Organization {id: $parent_id, tenant_id: $tenant_id})
					MATCH (child:Organization {id: $child_id, tenant_id: $tenant_id})
					CREATE (parent)-[:PARENT_OF {
						created_at: $created_at,
						sync_source: 'organization_sync_fix'
					}]->(child)
					RETURN parent.name as parent_name, child.name as child_name
				`
				
				params := map[string]interface{}{
					"parent_id":  *org.ParentUnitID,
					"child_id":   org.ID,
					"tenant_id":  org.TenantID,
					"created_at": time.Now().Format(time.RFC3339),
				}
				
				result, err := tx.Run(ctx, cypher, params)
				if err != nil {
					return nil, fmt.Errorf("创建层级关系失败 (%s -> %s): %w", *org.ParentUnitID, org.ID, err)
				}
				
				if result.Next(ctx) {
					record := result.Record()
					parentName, _ := record.Get("parent_name")
					childName, _ := record.Get("child_name")
					log.Printf("🔗 建立关系: %s -> %s", parentName, childName)
					relationshipCount++
				}
				
				return nil, nil
			})
			
			if err != nil {
				return err
			}
		}
	}
	
	log.Printf("🎯 成功建立 %d 个层级关系", relationshipCount)
	return nil
}

// verifySync 验证同步结果
func verifySync(ctx context.Context, driver neo4j.DriverWithContext, expectedCount int) error {
	session := driver.NewSession(ctx, neo4j.SessionConfig{
		AccessMode:   neo4j.AccessModeRead,
		DatabaseName: "neo4j",
	})
	defer session.Close(ctx)
	
	// 验证节点数量
	result, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `
			MATCH (o:Organization) 
			RETURN 
				count(o) as total_nodes,
				count(CASE WHEN o.level = 0 THEN 1 END) as level_0_count,
				count(CASE WHEN o.level = 1 THEN 1 END) as level_1_count,
				count(CASE WHEN o.level = 2 THEN 1 END) as level_2_count
		`
		result, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			record := result.Record()
			totalNodes, _ := record.Get("total_nodes")
			level0Count, _ := record.Get("level_0_count")
			level1Count, _ := record.Get("level_1_count")
			level2Count, _ := record.Get("level_2_count")
			
			return map[string]interface{}{
				"total_nodes":   totalNodes,
				"level_0_count": level0Count,
				"level_1_count": level1Count,
				"level_2_count": level2Count,
			}, nil
		}
		
		return nil, fmt.Errorf("未找到统计数据")
	})
	
	if err != nil {
		return fmt.Errorf("验证节点数量失败: %w", err)
	}
	
	stats := result.(map[string]interface{})
	totalNodes := stats["total_nodes"].(int64)
	
	log.Printf("📊 验证结果:")
	log.Printf("   - 总节点数: %d (期望: %d)", totalNodes, expectedCount)
	log.Printf("   - Level 0 (公司): %d", stats["level_0_count"].(int64))
	log.Printf("   - Level 1 (部门): %d", stats["level_1_count"].(int64))
	log.Printf("   - Level 2 (团队): %d", stats["level_2_count"].(int64))
	
	if int(totalNodes) != expectedCount {
		return fmt.Errorf("节点数量不匹配: 实际 %d, 期望 %d", totalNodes, expectedCount)
	}
	
	// 验证关系数量
	relationshipResult, err := session.ExecuteRead(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		cypher := `MATCH ()-[:PARENT_OF]->() RETURN count(*) as relationship_count`
		result, err := tx.Run(ctx, cypher, nil)
		if err != nil {
			return nil, err
		}
		
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("relationship_count")
			return count, nil
		}
		
		return 0, nil
	})
	
	if err != nil {
		return fmt.Errorf("验证关系数量失败: %w", err)
	}
	
	relationshipCount := relationshipResult.(int64)
	expectedRelationships := expectedCount - 1 // 总节点数 - 1 (根节点没有父节点)
	
	log.Printf("   - 层级关系数: %d (期望: %d)", relationshipCount, expectedRelationships)
	
	if int(relationshipCount) != expectedRelationships {
		return fmt.Errorf("关系数量不匹配: 实际 %d, 期望 %d", relationshipCount, expectedRelationships)
	}
	
	log.Println("✅ 数据验证通过！")
	return nil
}