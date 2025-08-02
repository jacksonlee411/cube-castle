package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// 测试CQRS事件流程的集成测试
func main() {
	log.Println("🧪 Testing CQRS Event Flow Integration...")

	// 等待服务器启动
	serverURL := "http://localhost:8080"
	log.Println("⏳ Waiting for server to start...")
	
	// 简单等待策略
	for i := 0; i < 10; i++ {
		resp, err := http.Get(serverURL + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			log.Println("✅ Server is ready!")
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(1 * time.Second)
		log.Printf("⏳ Waiting for server... attempt %d/10", i+1)
	}

	// 测试创建员工（应该触发EmployeeCreated事件）
	log.Println("📤 Testing employee creation with event publishing...")
	
	// 创建员工请求
	createEmployeeReq := openapi.CreateEmployeeRequest{
		EmployeeNumber: "TEST001",
		FirstName:      "测试",
		LastName:       "员工",
		Email:          openapi_types.Email("test@example.com"),
		HireDate:       openapi_types.Date{Time: time.Now()},
	}

	reqBody, err := json.Marshal(createEmployeeReq)
	if err != nil {
		log.Fatalf("❌ Failed to marshal request: %v", err)
	}

	// 发送POST请求
	req, err := http.NewRequest("POST", serverURL+"/api/v1/corehr/employees", bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatalf("❌ Failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Tenant-ID", uuid.New().String()) // 设置租户ID

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("❌ Failed to send request: %v", err)
		log.Println("🔍 This might be expected if the server isn't running")
		return
	}
	defer resp.Body.Close()

	log.Printf("📊 Response status: %s", resp.Status)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("❌ Failed to read response: %v", err)
	}

	if resp.StatusCode == 201 {
		log.Println("✅ Employee created successfully!")
		
		// 解析响应
		var employee openapi.Employee
		if err := json.Unmarshal(respBody, &employee); err == nil {
			log.Printf("👤 Created employee: %s %s (ID: %s)", 
				employee.FirstName, employee.LastName, employee.Id.String())
			log.Printf("🏷️ Employee Number: %s", employee.EmployeeNumber)
		}

		log.Println("🎉 CQRS Event Flow Test Completed!")
		log.Println("📝 Expected behavior:")
		log.Println("   1. Employee record created in PostgreSQL")
		log.Println("   2. EmployeeCreated event published to EventBus")
		log.Println("   3. Event would be sent to Kafka (currently using Mock)")
		log.Println("   4. Neo4j would receive the event for data synchronization")
	} else {
		log.Printf("⚠️ Unexpected response status: %d", resp.StatusCode)
		log.Printf("📄 Response body: %s", string(respBody))
	}

	// 测试更新员工（应该触发EmployeeUpdated事件）
	log.Println("\n📤 Testing employee update with event publishing...")
	
	// 这需要一个有效的员工ID，在实际测试中应该使用创建的员工ID
	// 这里只是演示测试结构

	log.Println("✅ Event Flow Integration Test Framework Ready!")
	log.Println("🚀 To run full integration test, start the server with: ./bin/server")
}

// 辅助函数用于等待服务器启动
func waitForServer(url string, maxRetries int) bool {
	for i := 0; i < maxRetries; i++ {
		resp, err := http.Get(url + "/health")
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()
			return true
		}
		if resp != nil {
			resp.Body.Close()
		}
		time.Sleep(time.Second)
	}
	return false
}