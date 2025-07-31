package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/metrics"
	"github.com/gaogu/cube-castle/go-app/internal/middleware"
	"github.com/gaogu/cube-castle/go-app/internal/workflow"
)

// TestStructuredLogging 测试结构化日志
func TestStructuredLogging(t *testing.T) {
	logger := logging.NewStructuredLogger()
	require.NotNil(t, logger)

	// 测试基础日志方法
	employeeID := uuid.New()
	tenantID := uuid.New()

	// 这些调用不应该panic
	logger.LogEmployeeCreated(employeeID, tenantID, "EMP001")
	logger.LogEmployeeUpdated(employeeID, tenantID, map[string]interface{}{
		"phone_number": "123-456-7890",
	})
	logger.LogEmployeeDeleted(employeeID, tenantID, "EMP001")

	// 测试AI请求日志
	logger.LogAIRequest("session123", "list_employees", time.Millisecond*100, true)

	// 测试数据库操作日志
	logger.LogDatabaseOperation("SELECT", "employees", 5, time.Millisecond*50, true)

	// 测试错误日志
	testErr := assert.AnError
	logger.LogError("test_error", "Test error message", testErr, map[string]interface{}{
		"test_context": "unit_test",
	})
}

// TestPrometheusMetrics 测试Prometheus指标
func TestPrometheusMetrics(t *testing.T) {
	// 重置指标用于测试
	metrics.ResetMetricsForTesting()

	// 测试业务指标记录
	tenantID := uuid.New().String()

	metrics.RecordEmployeeCreated(tenantID)
	metrics.RecordEmployeeUpdated(tenantID)
	metrics.RecordEmployeeDeleted(tenantID)

	// 测试AI指标
	metrics.RecordAIRequest("list_employees", "success", time.Millisecond*200)
	metrics.UpdateAISessionsActive(5)

	// 测试数据库指标
	metrics.RecordDatabaseOperation("SELECT", "employees", "success", time.Millisecond*10)
	metrics.UpdateDatabaseConnections(10, 5)

	// 测试错误指标
	metrics.RecordError("test_component", "test_error")

	// 验证指标端点
	handler := metrics.MetricsHandler()
	req := httptest.NewRequest("GET", "/metrics", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "cube_castle_employees_created_total")
}

// TestMiddleware 测试中间件
func TestMiddleware(t *testing.T) {
	logger := logging.NewStructuredLogger()

	// 创建测试路由
	r := chi.NewRouter()
	r.Use(middleware.LoggingMiddleware(logger))
	r.Use(middleware.RecoveryMiddleware(logger))
	r.Use(metrics.PrometheusMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(middleware.TenantMiddleware)

	r.Get("/test", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// 测试正常请求
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Tenant-ID", uuid.New().String())
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	// 验证CORS头
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))

	var response map[string]string
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "ok", response["status"])
}

// TestHealthCheck 测试健康检查
func TestHealthCheck(t *testing.T) {
	logger := logging.NewStructuredLogger()

	handler := middleware.HealthCheckMiddleware(logger)
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}

// TestCoreHRServiceIntegration 测试CoreHR服务集成
func TestCoreHRServiceIntegration(t *testing.T) {
	// 使用Mock服务进行测试
	service := corehr.NewMockService()
	require.NotNil(t, service)

	ctx := context.Background()
	tenantID := uuid.New()

	// 测试员工列表
	employees, err := service.ListEmployees(ctx, tenantID, 1, 10, "")
	require.NoError(t, err)
	require.NotNil(t, employees)
	assert.NotNil(t, employees.Employees)
	assert.Greater(t, len(*employees.Employees), 0)

	// 测试创建员工
	createReq := &corehr.CreateEmployeeRequest{
		EmployeeNumber: "TEST001",
		FirstName:      "Test",
		LastName:       "User",
		Email:          "test@example.com",
		HireDate:       time.Now(),
	}

	createdEmployee, err := service.CreateEmployee(ctx, tenantID, createReq)
	require.NoError(t, err)
	require.NotNil(t, createdEmployee)
	assert.Equal(t, "TEST001", createdEmployee.EmployeeNumber)
	assert.Equal(t, "Test", createdEmployee.FirstName)
	assert.Equal(t, "User", createdEmployee.LastName)

	// 测试获取员工
	employee, err := service.GetEmployee(ctx, tenantID, *createdEmployee.Id)
	require.NoError(t, err)
	require.NotNil(t, employee)
	assert.Equal(t, createdEmployee.Id, employee.Id)
}

// TestWorkflowIntegration 测试工作流集成
func TestWorkflowIntegration(t *testing.T) {
	// 注意：这个测试需要Temporal服务运行
	// 在CI/CD环境中可能需要跳过

	logger := logging.NewStructuredLogger()

	// 尝试创建Temporal管理器
	tm, err := workflow.NewTemporalManager("localhost:7233", logger)
	if err != nil {
		t.Skipf("Temporal service not available: %v", err)
		return
	}
	defer tm.Stop()

	ctx := context.Background()

	// 测试健康检查
	err = tm.HealthCheck(ctx)
	if err != nil {
		t.Skipf("Temporal health check failed: %v", err)
		return
	}

	// 测试启动员工入职工作流
	onboardingReq := workflow.EmployeeOnboardingRequest{
		EmployeeID: uuid.New(),
		TenantID:   uuid.New(),
		FirstName:  "Test",
		LastName:   "Employee",
		Email:      "test.employee@example.com",
		Department: "Technology",
		Position:   "Software Engineer",
		StartDate:  time.Now().AddDate(0, 0, 7), // 一周后开始
	}

	workflowID, err := tm.StartEmployeeOnboarding(ctx, onboardingReq)
	require.NoError(t, err)
	assert.NotEmpty(t, workflowID)

	// 等待一小段时间让工作流开始
	time.Sleep(2 * time.Second)

	// 测试获取工作流状态
	status, err := tm.GetWorkflowStatus(ctx, workflowID)
	require.NoError(t, err)
	require.NotNil(t, status)
	assert.Equal(t, workflowID, status.WorkflowID)
	assert.NotEmpty(t, status.Status)
}

// TestHTTPAPIIntegration 测试HTTP API集成
func TestHTTPAPIIntegration(t *testing.T) {
	logger := logging.NewStructuredLogger()
	coreHRService := corehr.NewMockService()

	// 创建测试路由
	router := setupRoutes(logger, coreHRService)

	// 测试健康检查端点
	t.Run("Health Check", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&response)
		require.NoError(t, err)
		assert.Equal(t, "healthy", response["status"])
	})

	// 测试指标端点
	t.Run("Metrics Endpoint", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/metrics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "# HELP")
		assert.Contains(t, w.Body.String(), "cube_castle")
	})

	// 测试员工API端点
	t.Run("Employees API", func(t *testing.T) {
		// 测试获取员工列表
		req := httptest.NewRequest("GET", "/api/v1/corehr/employees?page=1&page_size=10", nil)
		req.Header.Set("X-Tenant-ID", uuid.New().String())
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		// 由于我们的处理器还是占位符实现，检查是否能正常路由
		// 实际实现完成后，这里应该验证返回的员工数据
		assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotImplemented)
	})
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	logger := logging.NewStructuredLogger()

	// 创建会触发panic的路由
	r := chi.NewRouter()
	r.Use(middleware.RecoveryMiddleware(logger))

	r.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// 这不应该导致测试程序崩溃
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "Internal Server Error")
}

// TestConcurrency 测试并发处理
func TestConcurrency(t *testing.T) {
	logger := logging.NewStructuredLogger()
	coreHRService := corehr.NewMockService()
	router := setupRoutes(logger, coreHRService)

	// 并发发送请求
	const concurrentRequests = 10
	results := make(chan int, concurrentRequests)

	for i := 0; i < concurrentRequests; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)
			results <- w.Code
		}()
	}

	// 收集结果
	for i := 0; i < concurrentRequests; i++ {
		statusCode := <-results
		assert.Equal(t, http.StatusOK, statusCode)
	}
}

// BenchmarkHealthCheck 健康检查性能基准测试
func BenchmarkHealthCheck(b *testing.B) {
	logger := logging.NewStructuredLogger()
	handler := middleware.HealthCheckMiddleware(logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// BenchmarkEmployeesList 员工列表性能基准测试
func BenchmarkEmployeesList(b *testing.B) {
	logger := logging.NewStructuredLogger()
	coreHRService := corehr.NewMockService()
	handler := handleListEmployees(coreHRService, logger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/corehr/employees?page=1&page_size=20", nil)
		req.Header.Set("X-Tenant-ID", uuid.New().String())
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
	}
}

// TestIntegrationScenarios 集成场景测试
func TestIntegrationScenarios(t *testing.T) {
	t.Run("Complete Employee Lifecycle", func(t *testing.T) {
		// 这个测试模拟完整的员工生命周期
		// 1. 创建员工
		// 2. 更新员工信息
		// 3. 查询员工
		// 4. 删除员工

		service := corehr.NewMockService()
		ctx := context.Background()
		tenantID := uuid.New()

		// 1. 创建员工
		createReq := &corehr.CreateEmployeeRequest{
			EmployeeNumber: "LIFECYCLE001",
			FirstName:      "Lifecycle",
			LastName:       "Test",
			Email:          "lifecycle@example.com",
			HireDate:       time.Now(),
		}

		employee, err := service.CreateEmployee(ctx, tenantID, createReq)
		require.NoError(t, err)
		require.NotNil(t, employee)

		// 2. 查询员工
		retrievedEmployee, err := service.GetEmployee(ctx, tenantID, *employee.Id)
		require.NoError(t, err)
		assert.Equal(t, employee.Id, retrievedEmployee.Id)

		// 3. 更新员工
		updateReq := &corehr.UpdateEmployeeRequest{
			FirstName: stringPtr("Updated"),
			LastName:  stringPtr("Name"),
		}

		updatedEmployee, err := service.UpdateEmployee(ctx, tenantID, *employee.Id, updateReq)
		require.NoError(t, err)
		assert.Equal(t, "Updated", updatedEmployee.FirstName)

		// 4. 删除员工
		err = service.DeleteEmployee(ctx, tenantID, *employee.Id)
		require.NoError(t, err)
	})
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}
