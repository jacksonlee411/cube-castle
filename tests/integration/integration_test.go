package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/internal/intelligencegateway"
	"github.com/gaogu/cube-castle/go-app/internal/monitoring"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
)

// TestSystemIntegration 测试系统各组件的集成
func TestSystemIntegration(t *testing.T) {
	// 创建各个组件
	monitor := monitoring.NewMonitor(&monitoring.MonitorConfig{
		ServiceName: "integration-test",
		Version:     "1.0.0",
		Environment: "test",
	})
	
	igService := intelligencegateway.NewService()
	workflowEngine := workflow.NewEngine()
	
	// 注册测试工作流
	testWorkflow := &workflow.WorkflowDefinition{
		ID:          "intelligence-processing",
		Name:        "Intelligence Processing Workflow",
		Description: "Process intelligence queries through workflow",
		Steps:       []string{"validate", "ai_query", "process", "notify"},
	}
	
	err := workflowEngine.RegisterWorkflow(testWorkflow)
	if err != nil {
		t.Fatalf("Failed to register workflow: %v", err)
	}
	
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	
	// 测试Intelligence Gateway处理
	t.Run("Intelligence Gateway Processing", func(t *testing.T) {
		req := &intelligencegateway.InterpretUserQueryRequest{
			Query:    "Test integration query",
			UserID:   userID,
			TenantID: tenantID,
		}
		
		start := time.Now()
		resp, err := igService.InterpretUserQuery(ctx, req)
		duration := time.Since(start)
		
		// 记录到监控系统
		monitor.RecordHTTPRequest("POST", "/api/intelligence/query", 200, duration)
		
		if err != nil {
			t.Errorf("Intelligence Gateway processing failed: %v", err)
			return
		}
		
		if resp.Intent != "general_query" {
			t.Errorf("Expected intent 'general_query', got %s", resp.Intent)
		}
		
		// 验证上下文创建
		context, err := igService.GetConversationContext(ctx, userID, tenantID)
		if err != nil {
			t.Errorf("Failed to get conversation context: %v", err)
		}
		
		if len(context.History) != 2 {
			t.Errorf("Expected 2 messages in history, got %d", len(context.History))
		}
	})
	
	// 测试工作流处理
	t.Run("Workflow Processing", func(t *testing.T) {
		input := map[string]interface{}{
			"query":     "Workflow test query",
			"user_id":   userID.String(),
			"tenant_id": tenantID.String(),
		}
		
		start := time.Now()
		execution, err := workflowEngine.StartWorkflow(ctx, "intelligence-processing", input)
		
		if err != nil {
			t.Errorf("Failed to start workflow: %v", err)
			return
		}
		
		// 等待工作流完成
		maxWait := time.Second * 5
		startWait := time.Now()
		
		for {
			if time.Since(startWait) > maxWait {
				t.Fatal("Workflow did not complete in time")
			}
			
			updatedExecution, err := workflowEngine.GetExecution(execution.ID)
			if err != nil {
				t.Fatalf("Failed to get execution: %v", err)
			}
			
			if updatedExecution.Status == workflow.StatusCompleted {
				duration := time.Since(start)
				monitor.RecordHTTPRequest("POST", "/api/workflow/execute", 200, duration)
				
				// 验证工作流完成
				if len(updatedExecution.Steps) != 4 {
					t.Errorf("Expected 4 steps, got %d", len(updatedExecution.Steps))
				}
				
				for i, step := range updatedExecution.Steps {
					if step.Status != workflow.StatusCompleted {
						t.Errorf("Step %d (%s): Expected completed status, got %s", i, step.Name, step.Status)
					}
				}
				
				if updatedExecution.Output == nil {
					t.Error("Expected workflow output")
				}
				
				break
			} else if updatedExecution.Status == workflow.StatusFailed {
				t.Errorf("Workflow failed: %s", updatedExecution.Error)
				return
			}
			
			time.Sleep(time.Millisecond * 100)
		}
	})
	
	// 测试批处理集成
	t.Run("Batch Processing Integration", func(t *testing.T) {
		batchReq := &intelligencegateway.BatchRequest{
			BatchID: "integration-test-batch",
			Requests: []intelligencegateway.InterpretUserQueryRequest{
				{Query: "Batch query 1", UserID: userID, TenantID: tenantID},
				{Query: "Batch query 2", UserID: userID, TenantID: tenantID},
				{Query: "Batch query 3", UserID: userID, TenantID: tenantID},
			},
		}
		
		start := time.Now()
		batchResp, err := igService.ProcessBatchRequests(ctx, batchReq)
		duration := time.Since(start)
		
		monitor.RecordHTTPRequest("POST", "/api/intelligence/batch", 200, duration)
		
		if err != nil {
			t.Errorf("Batch processing failed: %v", err)
			return
		}
		
		if len(batchResp.Responses) != 3 {
			t.Errorf("Expected 3 batch responses, got %d", len(batchResp.Responses))
		}
		
		if batchResp.Status != "completed" {
			t.Errorf("Expected batch status 'completed', got %s", batchResp.Status)
		}
	})
	
	// 测试监控系统集成
	t.Run("Monitoring System Integration", func(t *testing.T) {
		// 获取健康状态
		healthStatus := monitor.GetHealthStatus(ctx)
		
		if healthStatus.Service != "integration-test" {
			t.Errorf("Expected service name 'integration-test', got %s", healthStatus.Service)
		}
		
		if healthStatus.Status != "healthy" {
			t.Errorf("Expected healthy status, got %s", healthStatus.Status)
		}
		
		// 获取详细健康状态
		detailedHealth := monitor.GetDetailedHealthStatus(ctx)
		
		if len(detailedHealth.Checks) == 0 {
			t.Error("Expected health checks to be performed")
		}
		
		// 验证HTTP指标
		httpMetrics := monitor.GetHTTPMetrics()
		
		if httpMetrics.RequestCount == 0 {
			t.Error("Expected HTTP requests to be recorded")
		}
		
		if len(httpMetrics.StatusCodes) == 0 {
			t.Error("Expected status codes to be recorded")
		}
		
		if len(httpMetrics.EndpointMetrics) == 0 {
			t.Error("Expected endpoint metrics to be recorded")
		}
	})
	
	// 测试系统统计集成
	t.Run("System Statistics Integration", func(t *testing.T) {
		// Intelligence Gateway 统计
		igStats := igService.GetContextStats()
		
		if igStats["total_contexts"].(int) == 0 {
			t.Error("Expected some conversation contexts to exist")
		}
		
		// 工作流统计
		workflowStats := workflowEngine.GetWorkflowStats()
		
		if workflowStats["total_workflows"].(int) == 0 {
			t.Error("Expected registered workflows")
		}
		
		if workflowStats["total_executions"].(int) == 0 {
			t.Error("Expected workflow executions")
		}
		
		// 系统指标
		systemMetrics := monitor.GetSystemMetrics()
		
		if systemMetrics.CPU.Cores <= 0 {
			t.Error("Expected positive CPU core count")
		}
		
		if systemMetrics.Memory.GoroutineCount <= 0 {
			t.Error("Expected positive goroutine count")
		}
		
		if systemMetrics.HTTP.RequestCount == 0 {
			t.Error("Expected recorded HTTP requests")
		}
	})
}

// TestErrorHandlingIntegration 测试错误处理集成
func TestErrorHandlingIntegration(t *testing.T) {
	monitor := monitoring.NewMonitor(nil)
	igService := intelligencegateway.NewService()
	
	ctx := context.Background()
	
	// 测试无效请求处理
	t.Run("Invalid Request Handling", func(t *testing.T) {
		invalidReq := &intelligencegateway.InterpretUserQueryRequest{
			Query:    "", // 空查询
			UserID:   uuid.New(),
			TenantID: uuid.New(),
		}
		
		start := time.Now()
		resp, err := igService.InterpretUserQuery(ctx, invalidReq)
		duration := time.Since(start)
		
		// 记录错误请求
		monitor.RecordHTTPRequest("POST", "/api/intelligence/query", 400, duration)
		
		if err == nil {
			t.Error("Expected error for invalid request")
		}
		
		if resp != nil {
			t.Error("Expected no response for invalid request")
		}
		
		// 验证错误被正确记录
		httpMetrics := monitor.GetHTTPMetrics()
		if httpMetrics.ErrorRate == 0 {
			t.Error("Expected error rate to be recorded")
		}
	})
	
	// 测试批处理错误处理
	t.Run("Batch Error Handling", func(t *testing.T) {
		batchReq := &intelligencegateway.BatchRequest{
			BatchID: "error-test-batch",
			Requests: []intelligencegateway.InterpretUserQueryRequest{
				{Query: "Valid query", UserID: uuid.New(), TenantID: uuid.New()},
				{Query: "", UserID: uuid.New(), TenantID: uuid.New()}, // 无效
				{Query: "Another valid", UserID: uuid.New(), TenantID: uuid.New()},
			},
		}
		
		batchResp, err := igService.ProcessBatchRequests(ctx, batchReq)
		
		if err != nil {
			t.Errorf("Batch processing should not fail completely, got %v", err)
		}
		
		if len(batchResp.Responses) != 3 {
			t.Errorf("Expected 3 responses, got %d", len(batchResp.Responses))
		}
		
		// 检查错误响应
		if batchResp.Responses[1].Intent != "error" {
			t.Errorf("Expected error intent for invalid request, got %s", batchResp.Responses[1].Intent)
		}
		
		// 检查有效响应
		if batchResp.Responses[0].Intent != "general_query" {
			t.Errorf("Expected general_query intent for valid request, got %s", batchResp.Responses[0].Intent)
		}
	})
}

// TestPerformanceIntegration 测试性能集成
func TestPerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}
	
	monitor := monitoring.NewMonitor(nil)
	igService := intelligencegateway.NewService()
	
	ctx := context.Background()
	userID := uuid.New()
	tenantID := uuid.New()
	
	// 测试并发处理
	t.Run("Concurrent Processing", func(t *testing.T) {
		const numRequests = 100
		results := make(chan error, numRequests)
		
		start := time.Now()
		
		for i := 0; i < numRequests; i++ {
			go func(i int) {
				req := &intelligencegateway.InterpretUserQueryRequest{
					Query:    fmt.Sprintf("Concurrent query %d", i),
					UserID:   userID,
					TenantID: tenantID,
				}
				
				reqStart := time.Now()
				_, err := igService.InterpretUserQuery(ctx, req)
				duration := time.Since(reqStart)
				
				monitor.RecordHTTPRequest("POST", "/api/intelligence/query", 200, duration)
				results <- err
			}(i)
		}
		
		// 收集结果
		errorCount := 0
		for i := 0; i < numRequests; i++ {
			if err := <-results; err != nil {
				errorCount++
				t.Logf("Request %d failed: %v", i, err)
			}
		}
		
		totalDuration := time.Since(start)
		
		if errorCount > 0 {
			t.Errorf("Expected no errors, got %d errors out of %d requests", errorCount, numRequests)
		}
		
		// 性能验证
		avgDuration := totalDuration / numRequests
		t.Logf("Processed %d requests in %v (avg: %v per request)", numRequests, totalDuration, avgDuration)
		
		if avgDuration > time.Millisecond*100 {
			t.Errorf("Average request duration too high: %v", avgDuration)
		}
		
		// 验证监控数据
		httpMetrics := monitor.GetHTTPMetrics()
		if httpMetrics.RequestCount < int64(numRequests) {
			t.Errorf("Expected at least %d recorded requests, got %d", numRequests, httpMetrics.RequestCount)
		}
	})
}

// TestResourceManagementIntegration 测试资源管理集成
func TestResourceManagementIntegration(t *testing.T) {
	monitor := monitoring.NewMonitor(nil)
	igService := intelligencegateway.NewService()
	
	ctx := context.Background()
	
	// 测试上下文清理
	t.Run("Context Cleanup", func(t *testing.T) {
		userID := uuid.New()
		tenantID := uuid.New()
		
		// 创建多个对话
		for i := 0; i < 10; i++ {
			req := &intelligencegateway.InterpretUserQueryRequest{
				Query:    fmt.Sprintf("Test query %d", i),
				UserID:   userID,
				TenantID: tenantID,
			}
			
			_, err := igService.InterpretUserQuery(ctx, req)
			if err != nil {
				t.Errorf("Failed to create conversation: %v", err)
				continue
			}
		}
		
		// 验证上下文存在
		context, err := igService.GetConversationContext(ctx, userID, tenantID)
		if err != nil {
			t.Errorf("Expected context to exist: %v", err)
		}
		
		if len(context.History) == 0 {
			t.Error("Expected conversation history")
		}
		
		// 清理上下文
		err = igService.ClearContext(userID, tenantID)
		if err != nil {
			t.Errorf("Failed to clear context: %v", err)
		}
		
		// 验证上下文已清理
		context, err = igService.GetConversationContext(ctx, userID, tenantID)
		if err == nil {
			t.Error("Expected error for cleared context")
		}
		if context != nil {
			t.Error("Expected nil context after clearing")
		}
	})
	
	// 测试内存使用
	t.Run("Memory Usage", func(t *testing.T) {
		initialMetrics := monitor.GetSystemMetrics()
		initialMemory := initialMetrics.Memory.UsedBytes
		
		// 创建大量数据
		for i := 0; i < 1000; i++ {
			userID := uuid.New()
			tenantID := uuid.New()
			
			req := &intelligencegateway.InterpretUserQueryRequest{
				Query:    fmt.Sprintf("Memory test query %d with some additional data", i),
				UserID:   userID,
				TenantID: tenantID,
			}
			
			igService.InterpretUserQuery(ctx, req)
		}
		
		finalMetrics := monitor.GetSystemMetrics()
		finalMemory := finalMetrics.Memory.UsedBytes
		
		memoryIncrease := finalMemory - initialMemory
		t.Logf("Memory increased by %d bytes", memoryIncrease)
		
		// 验证内存使用合理
		if memoryIncrease > 50*1024*1024 { // 50MB
			t.Errorf("Memory increase too high: %d bytes", memoryIncrease)
		}
	})
}