package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	// Neo4j connection configuration
	uri := "bolt://localhost:7687"
	username := "neo4j"
	password := "password"

	// Create driver
	driver, err := neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(context.Background())

	// Verify connectivity
	err = driver.VerifyConnectivity(context.Background())
	if err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}

	fmt.Println("✅ Neo4j连接成功")

	// Create session
	session := driver.NewSession(context.Background(), neo4j.SessionConfig{})
	defer session.Close(context.Background())

	ctx := context.Background()

	// Test 1: Basic connectivity - get Neo4j version
	fmt.Println("\n=== 测试1: 基础连接性 ===")
	result, err := session.Run(ctx, "CALL dbms.components() YIELD name, versions, edition", nil)
	if err != nil {
		log.Printf("❌ 获取数据库版本失败: %v", err)
	} else {
		for result.Next(ctx) {
			record := result.Record()
			name, _ := record.Get("name")
			versions, _ := record.Get("versions")
			edition, _ := record.Get("edition")
			fmt.Printf("✅ %s %v (%s)\n", name, versions, edition)
		}
	}

	// Test 2: Database schema and constraints
	fmt.Println("\n=== 测试2: 数据库模式 ===")
	result, err = session.Run(ctx, "SHOW CONSTRAINTS", nil)
	if err != nil {
		log.Printf("❌ 获取约束失败: %v", err)
	} else {
		constraintCount := 0
		for result.Next(ctx) {
			constraintCount++
			record := result.Record()
			name, _ := record.Get("name")
			labelsOrTypes, _ := record.Get("labelsOrTypes")
			properties, _ := record.Get("properties")
			fmt.Printf("✅ 约束: %s on %v(%v)\n", name, labelsOrTypes, properties)
		}
		fmt.Printf("📊 总约束数: %d\n", constraintCount)
	}

	// Test 3: Node counts
	fmt.Println("\n=== 测试3: 节点统计 ===")
	
	// Count Employee nodes
	result, err = session.Run(ctx, "MATCH (e:Employee) RETURN count(e) as count", nil)
	if err != nil {
		log.Printf("❌ 统计Employee节点失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			fmt.Printf("✅ Employee节点数: %v\n", count)
		}
	}

	// Count Position nodes  
	result, err = session.Run(ctx, "MATCH (p:Position) RETURN count(p) as count", nil)
	if err != nil {
		log.Printf("❌ 统计Position节点失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			fmt.Printf("✅ Position节点数: %v\n", count)
		}
	}

	// Count Department nodes
	result, err = session.Run(ctx, "MATCH (d:Department) RETURN count(d) as count", nil)
	if err != nil {
		log.Printf("❌ 统计Department节点失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			fmt.Printf("✅ Department节点数: %v\n", count)
		}
	}

	// Test 4: Relationship analysis
	fmt.Println("\n=== 测试4: 关系分析 ===")
	
	// Count REPORTS_TO relationships
	result, err = session.Run(ctx, "MATCH ()-[r:REPORTS_TO]->() RETURN count(r) as count", nil)
	if err != nil {
		log.Printf("❌ 统计REPORTS_TO关系失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			fmt.Printf("✅ REPORTS_TO关系数: %v\n", count)
		}
	}

	// Count HOLDS_POSITION relationships
	result, err = session.Run(ctx, "MATCH ()-[r:HOLDS_POSITION]->() RETURN count(r) as count", nil)
	if err != nil {
		log.Printf("❌ 统计HOLDS_POSITION关系失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("count")
			fmt.Printf("✅ HOLDS_POSITION关系数: %v\n", count)
		}
	}

	// Test 5: Sample data queries
	fmt.Println("\n=== 测试5: 样本数据查询 ===")
	
	// Get sample employees with their positions
	result, err = session.Run(ctx, `
		MATCH (e:Employee)-[h:HOLDS_POSITION]->(p:Position)
		RETURN e.legal_name as name, p.position_title as title, p.department as dept
		LIMIT 5
	`, nil)
	if err != nil {
		log.Printf("❌ 获取员工职位信息失败: %v", err)
	} else {
		fmt.Println("📋 员工职位样本:")
		for result.Next(ctx) {
			record := result.Record()
			name, _ := record.Get("name")
			title, _ := record.Get("title")
			dept, _ := record.Get("dept")
			fmt.Printf("  • %s - %s (%s)\n", name, title, dept)
		}
	}

	// Test 6: Reporting hierarchy query
	fmt.Println("\n=== 测试6: 汇报层级查询 ===")
	
	result, err = session.Run(ctx, `
		MATCH path = (subordinate:Employee)-[:REPORTS_TO*1..3]->(manager:Employee)
		RETURN subordinate.legal_name as subordinate, 
		       manager.legal_name as manager, 
		       length(path) as levels
		ORDER BY levels DESC
		LIMIT 5
	`, nil)
	if err != nil {
		log.Printf("❌ 查询汇报层级失败: %v", err)
	} else {
		fmt.Println("📊 汇报关系样本:")
		for result.Next(ctx) {
			record := result.Record()
			subordinate, _ := record.Get("subordinate")
			manager, _ := record.Get("manager")
			levels, _ := record.Get("levels")
			fmt.Printf("  • %s → %s (层级: %v)\n", subordinate, manager, levels)
		}
	}

	// Test 7: Department structure
	fmt.Println("\n=== 测试7: 部门结构查询 ===")
	
	result, err = session.Run(ctx, `
		MATCH (d:Department)<-[:BELONGS_TO]-(p:Position)<-[:HOLDS_POSITION]-(e:Employee)
		RETURN d.name as department, count(e) as employee_count
		ORDER BY employee_count DESC
		LIMIT 5
	`, nil)
	if err != nil {
		log.Printf("❌ 查询部门结构失败: %v", err)
	} else {
		fmt.Println("🏢 部门员工统计:")
		for result.Next(ctx) {
			record := result.Record()
			department, _ := record.Get("department")
			count, _ := record.Get("employee_count")
			fmt.Printf("  • %s: %v人\n", department, count)
		}
	}

	// Test 8: Graph algorithms - centrality
	fmt.Println("\n=== 测试8: 图算法测试 ===")
	
	result, err = session.Run(ctx, `
		MATCH (e:Employee)
		OPTIONAL MATCH (e)-[r:REPORTS_TO]->()
		OPTIONAL MATCH ()-[r2:REPORTS_TO]->(e)
		RETURN e.legal_name as name, 
		       count(r) as reports_to_count,
		       count(r2) as direct_reports_count
		ORDER BY direct_reports_count DESC
		LIMIT 5
	`, nil)
	if err != nil {
		log.Printf("❌ 查询员工中心性失败: %v", err)
	} else {
		fmt.Println("👑 管理层分析:")
		for result.Next(ctx) {
			record := result.Record()
			name, _ := record.Get("name")
			reportsTo, _ := record.Get("reports_to_count")
			directReports, _ := record.Get("direct_reports_count")
			fmt.Printf("  • %s: 直接下属 %v人, 汇报给 %v人\n", name, directReports, reportsTo)
		}
	}

	// Test 9: Data freshness check
	fmt.Println("\n=== 测试9: 数据新鲜度检查 ===")
	
	result, err = session.Run(ctx, `
		MATCH (e:Employee)
		WHERE exists(e.created_at)
		RETURN max(e.created_at) as latest_employee, count(e) as total_with_timestamp
	`, nil)
	if err != nil {
		log.Printf("❌ 检查数据时间戳失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			latest, _ := record.Get("latest_employee")
			total, _ := record.Get("total_with_timestamp")
			fmt.Printf("✅ 最新员工记录: %v\n", latest)
			fmt.Printf("✅ 带时间戳的记录数: %v\n", total)
		}
	}

	// Test 10: Performance test
	fmt.Println("\n=== 测试10: 性能测试 ===")
	
	start := time.Now()
	result, err = session.Run(ctx, `
		MATCH (e:Employee)-[:HOLDS_POSITION]->(p:Position)-[:BELONGS_TO]->(d:Department)
		RETURN count(*) as total_employee_position_department_paths
	`, nil)
	duration := time.Since(start)
	
	if err != nil {
		log.Printf("❌ 性能测试查询失败: %v", err)
	} else {
		if result.Next(ctx) {
			record := result.Record()
			count, _ := record.Get("total_employee_position_department_paths")
			fmt.Printf("✅ 复杂查询结果: %v条路径\n", count)
			fmt.Printf("⚡ 查询耗时: %v\n", duration)
		}
	}

	fmt.Println("\n🎉 Neo4j功能验证完成!")
}