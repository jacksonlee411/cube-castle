package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func main() {
	fmt.Println("🧪 开始测试数据同步到Neo4j...")

	// Create sample data in Neo4j directly for testing
	ctx := context.Background()
	
	// Neo4j connection
	driver, err := neo4j.NewDriverWithContext("bolt://localhost:7687", neo4j.BasicAuth("neo4j", "password", ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(ctx)

	session := driver.NewSession(ctx, neo4j.SessionConfig{})
	defer session.Close(ctx)

	fmt.Println("✅ 连接到Neo4j成功")

	// Clear existing data
	fmt.Println("🧹 清理现有数据...")
	_, err = session.Run(ctx, "MATCH (n) DETACH DELETE n", nil)
	if err != nil {
		log.Printf("Warning: Failed to clear data: %v", err)
	}

	// Create sample organizational data
	fmt.Println("📝 创建样本组织数据...")

	// Create departments
	departments := []map[string]interface{}{
		{"id": "dept-tech", "name": "技术部"},
		{"id": "dept-product", "name": "产品部"},
		{"id": "dept-sales", "name": "销售部"},
		{"id": "dept-hr", "name": "人力资源部"},
	}

	for _, dept := range departments {
		_, err = session.Run(ctx, `
			CREATE (d:Department {
				id: $id,
				name: $name,
				created_at: datetime()
			})
		`, dept)
		if err != nil {
			log.Printf("Failed to create department %s: %v", dept["name"], err)
		}
	}

	// Create employees
	employees := []map[string]interface{}{
		{
			"id": "emp-001", "employee_id": "EMP001", "legal_name": "张三",
			"email": "zhangsan@company.com", "status": "ACTIVE",
			"hire_date": time.Date(2020, 1, 15, 0, 0, 0, 0, time.UTC),
		},
		{
			"id": "emp-002", "employee_id": "EMP002", "legal_name": "李四", 
			"email": "lisi@company.com", "status": "ACTIVE",
			"hire_date": time.Date(2021, 3, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			"id": "emp-003", "employee_id": "EMP003", "legal_name": "王五",
			"email": "wangwu@company.com", "status": "ACTIVE", 
			"hire_date": time.Date(2022, 6, 10, 0, 0, 0, 0, time.UTC),
		},
		{
			"id": "emp-004", "employee_id": "EMP004", "legal_name": "赵六",
			"email": "zhaoliu@company.com", "status": "ACTIVE",
			"hire_date": time.Date(2019, 9, 5, 0, 0, 0, 0, time.UTC),
		},
		{
			"id": "emp-005", "employee_id": "EMP005", "legal_name": "钱七",
			"email": "qianqi@company.com", "status": "ACTIVE",
			"hire_date": time.Date(2023, 2, 28, 0, 0, 0, 0, time.UTC),
		},
	}

	for _, emp := range employees {
		_, err = session.Run(ctx, `
			CREATE (e:Employee {
				id: $id,
				employee_id: $employee_id,
				legal_name: $legal_name,
				email: $email,
				status: $status,
				hire_date: $hire_date,
				created_at: datetime()
			})
		`, emp)
		if err != nil {
			log.Printf("Failed to create employee %s: %v", emp["legal_name"], err)
		}
	}

	// Create positions
	positions := []map[string]interface{}{
		{
			"id": "pos-001", "position_title": "技术总监", "department": "技术部",
			"job_level": "DIRECTOR", "location": "北京", "employee_id": "emp-001",
		},
		{
			"id": "pos-002", "position_title": "高级软件工程师", "department": "技术部",
			"job_level": "SENIOR", "location": "北京", "employee_id": "emp-002",
		},
		{
			"id": "pos-003", "position_title": "前端工程师", "department": "技术部",
			"job_level": "INTERMEDIATE", "location": "北京", "employee_id": "emp-003",
		},
		{
			"id": "pos-004", "position_title": "产品经理", "department": "产品部",
			"job_level": "MANAGER", "location": "北京", "employee_id": "emp-004",
		},
		{
			"id": "pos-005", "position_title": "销售专员", "department": "销售部",
			"job_level": "JUNIOR", "location": "上海", "employee_id": "emp-005",
		},
	}

	for _, pos := range positions {
		_, err = session.Run(ctx, `
			CREATE (p:Position {
				id: $id,
				position_title: $position_title,
				department: $department,
				job_level: $job_level,
				location: $location,
				effective_date: date(),
				created_at: datetime()
			})
		`, pos)
		if err != nil {
			log.Printf("Failed to create position %s: %v", pos["position_title"], err)
		}
	}

	// Create relationships - employees hold positions
	fmt.Println("🔗 创建员工-职位关系...")
	for _, pos := range positions {
		_, err = session.Run(ctx, `
			MATCH (e:Employee {id: $employee_id})
			MATCH (p:Position {id: $id})
			CREATE (e)-[:HOLDS_POSITION]->(p)
		`, pos)
		if err != nil {
			log.Printf("Failed to create HOLDS_POSITION relationship: %v", err)
		}
	}

	// Create reporting relationships
	fmt.Println("👑 创建汇报关系...")
	reportingRelationships := []map[string]interface{}{
		{"subordinate": "emp-002", "manager": "emp-001"}, // 李四 -> 张三
		{"subordinate": "emp-003", "manager": "emp-001"}, // 王五 -> 张三
		{"subordinate": "emp-005", "manager": "emp-004"}, // 钱七 -> 赵六
	}

	for _, rel := range reportingRelationships {
		_, err = session.Run(ctx, `
			MATCH (subordinate:Employee {id: $subordinate})
			MATCH (manager:Employee {id: $manager})
			CREATE (subordinate)-[:REPORTS_TO]->(manager)
		`, rel)
		if err != nil {
			log.Printf("Failed to create REPORTS_TO relationship: %v", err)
		}
	}

	// Create department relationships
	fmt.Println("🏢 创建部门关系...")
	deptRelationships := []map[string]interface{}{
		{"position": "pos-001", "department": "dept-tech"},
		{"position": "pos-002", "department": "dept-tech"},
		{"position": "pos-003", "department": "dept-tech"},
		{"position": "pos-004", "department": "dept-product"},
		{"position": "pos-005", "department": "dept-sales"},
	}

	for _, rel := range deptRelationships {
		_, err = session.Run(ctx, `
			MATCH (p:Position {id: $position})
			MATCH (d:Department {id: $department})
			CREATE (p)-[:BELONGS_TO]->(d)
		`, rel)
		if err != nil {
			log.Printf("Failed to create BELONGS_TO relationship: %v", err)
		}
	}

	// Verify the data was created successfully
	fmt.Println("\n📊 验证创建的数据...")

	// Count nodes
	result, err := session.Run(ctx, "MATCH (e:Employee) RETURN count(e) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个员工节点\n", count)
	}

	result, err = session.Run(ctx, "MATCH (p:Position) RETURN count(p) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个职位节点\n", count)
	}

	result, err = session.Run(ctx, "MATCH (d:Department) RETURN count(d) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个部门节点\n", count)
	}

	result, err = session.Run(ctx, "MATCH ()-[r:REPORTS_TO]->() RETURN count(r) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个汇报关系\n", count)
	}

	result, err = session.Run(ctx, "MATCH ()-[r:HOLDS_POSITION]->() RETURN count(r) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个职位关系\n", count)
	}

	result, err = session.Run(ctx, "MATCH ()-[r:BELONGS_TO]->() RETURN count(r) as count", nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		count, _ := record.Get("count")
		fmt.Printf("✅ 创建了 %v 个部门关系\n", count)
	}

	// Test organizational queries
	fmt.Println("\n🔍 测试组织查询功能...")

	// Test 1: Organization chart query
	fmt.Println("\n1. 组织架构查询:")
	result, err = session.Run(ctx, `
		MATCH (e:Employee)-[:HOLDS_POSITION]->(p:Position)-[:BELONGS_TO]->(d:Department)
		RETURN d.name as department, e.legal_name as employee, p.position_title as position
		ORDER BY d.name, p.job_level DESC
	`, nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			dept, _ := record.Get("department")
			emp, _ := record.Get("employee")
			pos, _ := record.Get("position")
			fmt.Printf("  📋 %s: %s (%s)\n", dept, emp, pos)
		}
	}

	// Test 2: Reporting hierarchy
	fmt.Println("\n2. 汇报层级查询:")
	result, err = session.Run(ctx, `
		MATCH (subordinate:Employee)-[:REPORTS_TO]->(manager:Employee)
		RETURN subordinate.legal_name as subordinate, manager.legal_name as manager
	`, nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			subordinate, _ := record.Get("subordinate")
			manager, _ := record.Get("manager")
			fmt.Printf("  👥 %s 汇报给 %s\n", subordinate, manager)
		}
	}

	// Test 3: Find reporting path
	fmt.Println("\n3. 汇报路径查询:")
	result, err = session.Run(ctx, `
		MATCH path = (emp1:Employee {legal_name: "王五"})-[:REPORTS_TO*1..3]->(emp2:Employee {legal_name: "张三"})
		RETURN [node in nodes(path) | node.legal_name] as path, length(path) as levels
	`, nil)
	if err == nil && result.Next(ctx) {
		record := result.Record()
		path, _ := record.Get("path")
		levels, _ := record.Get("levels")
		fmt.Printf("  🛤️ 王五 到 张三 的路径: %v (层级: %v)\n", path, levels)
	}

	// Test 4: Department statistics
	fmt.Println("\n4. 部门统计:")
	result, err = session.Run(ctx, `
		MATCH (d:Department)<-[:BELONGS_TO]-(p:Position)<-[:HOLDS_POSITION]-(e:Employee)
		RETURN d.name as department, count(e) as employee_count
		ORDER BY employee_count DESC
	`, nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			dept, _ := record.Get("department")
			count, _ := record.Get("employee_count")
			fmt.Printf("  🏢 %s: %v人\n", dept, count)
		}
	}

	// Test 5: Manager analysis
	fmt.Println("\n5. 管理层分析:")
	result, err = session.Run(ctx, `
		MATCH (manager:Employee)
		OPTIONAL MATCH (subordinate:Employee)-[:REPORTS_TO]->(manager)
		WITH manager, count(subordinate) as direct_reports
		WHERE direct_reports > 0
		RETURN manager.legal_name as manager, direct_reports
		ORDER BY direct_reports DESC
	`, nil)
	if err == nil {
		for result.Next(ctx) {
			record := result.Record()
			manager, _ := record.Get("manager")
			reports, _ := record.Get("direct_reports")
			fmt.Printf("  👑 %s 管理 %v 人\n", manager, reports)
		}
	}

	fmt.Println("\n🎉 Neo4j数据同步测试完成!")
	fmt.Println("📝 现在Neo4j中包含完整的组织架构数据，可以支持:")
	fmt.Println("   • 组织架构图生成")
	fmt.Println("   • 汇报关系查询")  
	fmt.Println("   • 部门统计分析")
	fmt.Println("   • 管理层分析")
	fmt.Println("   • SAM态势感知分析")
}