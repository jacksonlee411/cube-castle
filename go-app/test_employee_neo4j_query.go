package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/gaogu/cube-castle/go-app/internal/cqrs/queries"
	"github.com/gaogu/cube-castle/go-app/internal/repositories"
	"github.com/gaogu/cube-castle/go-app/internal/service"
	"github.com/google/uuid"
)

// TestNeo4jEmployeeQuery 测试Neo4j员工查询功能
func main() {
	// Neo4j连接配置
	uri := "bolt://localhost:7687"
	username := "neo4j"
	password := "password"

	// 创建服务和仓储
	logger := log.New(os.Stdout, "[NEO4J_TEST] ", log.LstdFlags)
	
	config := service.Neo4jConfig{
		URI:      uri,
		Username: username,
		Password: password,
		Database: "neo4j",
	}
	
	neo4jService, err := service.NewNeo4jService(config, logger)
	if err != nil {
		log.Fatalf("Failed to create Neo4j service: %v", err)
	}
	
	// 创建简单的Logger实现
	repoLogger := &SimpleLogger{logger: logger}
	
	// 创建员工查询仓储
	employeeQueryRepo := repositories.NewNeo4jEmployeeQueryRepository(neo4jService, repoLogger)

	ctx := context.Background()

	// 测试1: 搜索所有员工
	fmt.Println("🔍 Testing SearchEmployees...")
	searchQuery := queries.SearchEmployeesQuery{
		TenantID: uuid.New(), // 使用测试租户ID
		Limit:    10,
		Offset:   0,
	}

	employees, err := employeeQueryRepo.SearchEmployees(ctx, searchQuery)
	if err != nil {
		log.Printf("❌ SearchEmployees failed: %v", err)
	} else {
		fmt.Printf("✅ Found %d employees (total: %d)\n", len(employees.Employees), employees.TotalCount)
		if len(employees.Employees) > 0 {
			// 显示第一个员工的信息
			emp := employees.Employees[0]
			empJson, _ := json.MarshalIndent(emp, "", "  ")
			fmt.Printf("📋 First employee: %s\n", empJson)
			
			// 调试信息：显示实际的ID值
			fmt.Printf("🔍 Debug - Employee ID: %s (is zero: %t)\n", emp.ID.String(), emp.ID == uuid.Nil)
		}
	}

	// 测试2: 使用email进行搜索测试
	if len(employees.Employees) > 0 {
		fmt.Println("\n🔍 Testing SearchEmployees by email...")
		firstEmp := employees.Employees[0]
		
		// 使用第一个员工的email进行搜索
		emailSearchQuery := queries.SearchEmployeesQuery{
			TenantID: searchQuery.TenantID,
			Email:    &firstEmp.Email,
			Limit:    5,
			Offset:   0,
		}

		emailResults, err := employeeQueryRepo.SearchEmployees(ctx, emailSearchQuery)
		if err != nil {
			log.Printf("❌ Email search failed: %v", err)
		} else {
			fmt.Printf("✅ Found %d employees by email '%s'\n", len(emailResults.Employees), firstEmp.Email)
			if len(emailResults.Employees) > 0 {
				foundEmp := emailResults.Employees[0]
				fmt.Printf("📋 Found employee: %s %s (%s)\n", foundEmp.FirstName, foundEmp.LastName, foundEmp.Email)
			}
		}
	} else {
		fmt.Println("\n⏭️ Skipping email search test - no employees found")
	}

	fmt.Println("\n🎉 Neo4j employee query test completed!")
}

// SimpleLogger 简单的Logger实现用于测试
type SimpleLogger struct {
	logger *log.Logger
}

func (l *SimpleLogger) Info(msg string, fields ...interface{}) {
	l.logger.Printf("[INFO] %s: %v", msg, fields)
}

func (l *SimpleLogger) Error(msg string, fields ...interface{}) {
	l.logger.Printf("[ERROR] %s: %v", msg, fields)
}

func (l *SimpleLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Printf("[WARN] %s: %v", msg, fields)
}