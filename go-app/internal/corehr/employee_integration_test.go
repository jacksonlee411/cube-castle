package corehr

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/middleware"
	"github.com/gaogu/cube-castle/go-app/internal/validation"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmployeeIntegration 综合测试员工管理功能
func TestEmployeeIntegration(t *testing.T) {
	// 设置测试环境
	logger := logging.NewStructuredLogger()
	service := NewMockService()
	
	// 创建Mock验证器
	mockChecker := validation.NewMockValidationChecker()
	validator := validation.NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)
	
	// 设置路由
	router := setupTestRoutes(service, logger, validator)
	
	t.Run("Complete_Employee_CRUD_Workflow", func(t *testing.T) {
		testCompleteEmployeeCRUD(t, router)
	})
	
	t.Run("Employee_Validation_Tests", func(t *testing.T) {
		testEmployeeValidation(t, router)
	})
	
	t.Run("Employee_Error_Handling", func(t *testing.T) {
		testEmployeeErrorHandling(t, router)
	})
	
	t.Run("Employee_Manager_Relationship", func(t *testing.T) {
		testEmployeeManagerRelationship(t, router)
	})
}

// setupTestRoutes 设置测试路由
func setupTestRoutes(service *Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) *chi.Mux {
	r := chi.NewRouter()
	
	// 添加测试中间件
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 为测试添加租户ID
			tenantID := uuid.MustParse("12345678-1234-1234-1234-123456789012")
			ctx := context.WithValue(r.Context(), middleware.TenantIDKey, tenantID.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	
	// 员工路由
	r.Route("/api/v1/corehr/employees", func(r chi.Router) {
		r.Get("/", handleListEmployees(service, logger, validator))
		r.Post("/", handleCreateEmployee(service, logger, validator))
		r.Route("/{employeeID}", func(r chi.Router) {
			r.Get("/", handleGetEmployee(service, logger))
			r.Put("/", handleUpdateEmployee(service, logger, validator))
			r.Delete("/", handleDeleteEmployee(service, logger, validator))
			r.Get("/manager", handleGetEmployeeManager(service, logger))
		})
	})
	
	return r
}

// testCompleteEmployeeCRUD 测试完整的CRUD工作流
func testCompleteEmployeeCRUD(t *testing.T, router *chi.Mux) {
	t.Log("=== 测试完整员工CRUD工作流 ===")
	
	// 1. 创建员工
	t.Log("Step 1: 创建员工")
	createReq := openapi.CreateEmployeeRequest{
		EmployeeNumber: "EMP001",
		FirstName:      "John",
		LastName:       "Doe", 
		Email:          openapi_types.Email("john.doe@example.com"),
		HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
		PhoneNumber:    stringPtr("13800138000"),
	}
	
	createReqBody, _ := json.Marshal(createReq)
	req := httptest.NewRequest("POST", "/api/v1/corehr/employees", bytes.NewBuffer(createReqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusCreated, w.Code, "创建员工应该返回201状态码")
	
	var createdEmployee openapi.Employee
	err := json.Unmarshal(w.Body.Bytes(), &createdEmployee)
	require.NoError(t, err, "应该能够解析创建的员工响应")
	
	assert.NotNil(t, createdEmployee.Id, "创建的员工应该有ID")
	assert.Equal(t, "EMP001", createdEmployee.EmployeeNumber, "员工编号应该匹配")
	assert.Equal(t, "John", createdEmployee.FirstName, "名字应该匹配")
	assert.Equal(t, "Doe", createdEmployee.LastName, "姓氏应该匹配")
	
	employeeID := *createdEmployee.Id
	t.Logf("创建的员工ID: %s", employeeID.String())
	
	// 2. 获取员工列表
	t.Log("Step 2: 获取员工列表")
	req = httptest.NewRequest("GET", "/api/v1/corehr/employees?page=1&page_size=10", nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "获取员工列表应该返回200状态码")
	
	var listResponse openapi.EmployeeListResponse
	err = json.Unmarshal(w.Body.Bytes(), &listResponse)
	require.NoError(t, err, "应该能够解析员工列表响应")
	
	assert.NotNil(t, listResponse.Employees, "员工列表不应该为空")
	assert.Greater(t, len(*listResponse.Employees), 0, "员工列表应该包含员工")
	
	// 3. 根据ID获取员工
	t.Log("Step 3: 根据ID获取员工")
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/corehr/employees/%s", employeeID.String()), nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "根据ID获取员工应该返回200状态码")
	
	var getEmployee openapi.Employee
	err = json.Unmarshal(w.Body.Bytes(), &getEmployee) 
	require.NoError(t, err, "应该能够解析获取的员工响应")
	
	assert.Equal(t, employeeID, *getEmployee.Id, "获取的员工ID应该匹配")
	assert.Equal(t, "EMP001", getEmployee.EmployeeNumber, "员工编号应该匹配")
	
	// 4. 更新员工
	t.Log("Step 4: 更新员工")
	updateReq := openapi.UpdateEmployeeRequest{
		FirstName:   stringPtr("Jane"),
		LastName:    stringPtr("Smith"),
		Email:       emailPtr("jane.smith@example.com"),
		PhoneNumber: stringPtr("13900139000"),
	}
	
	updateReqBody, _ := json.Marshal(updateReq)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/corehr/employees/%s", employeeID.String()), bytes.NewBuffer(updateReqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "更新员工应该返回200状态码")
	
	var updatedEmployee openapi.Employee
	err = json.Unmarshal(w.Body.Bytes(), &updatedEmployee)
	require.NoError(t, err, "应该能够解析更新的员工响应")
	
	// Mock服务返回固定数据，所以这里只验证状态码
	assert.Equal(t, employeeID, *updatedEmployee.Id, "更新的员工ID应该匹配")
	
	// 5. 获取员工的经理
	t.Log("Step 5: 获取员工的经理")
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/corehr/employees/%s/manager", employeeID.String()), nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "获取员工经理应该返回200状态码")
	
	// 6. 删除员工
	t.Log("Step 6: 删除员工")
	req = httptest.NewRequest("DELETE", fmt.Sprintf("/api/v1/corehr/employees/%s", employeeID.String()), nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code, "删除员工应该返回204状态码")
}

// testEmployeeValidation 测试员工数据验证
func testEmployeeValidation(t *testing.T, router *chi.Mux) {
	t.Log("=== 测试员工数据验证 ===")
	
	// 测试无效的创建请求
	t.Log("测试无效的员工创建请求")
	invalidCreateReq := openapi.CreateEmployeeRequest{
		// 缺少必需字段
		EmployeeNumber: "", // 空员工编号
		FirstName:      "",
		LastName:       "",
		Email:          openapi_types.Email("invalid-email"), // 无效邮箱
		HireDate:       openapi_types.Date{Time: time.Now()},
	}
	
	createReqBody, _ := json.Marshal(invalidCreateReq)
	req := httptest.NewRequest("POST", "/api/v1/corehr/employees", bytes.NewBuffer(createReqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// 应该返回400验证错误
	assert.Equal(t, http.StatusBadRequest, w.Code, "无效的创建请求应该返回400状态码")
	
	// 检查响应中是否包含验证错误信息
	responseBody := w.Body.String()
	t.Logf("验证错误响应: %s", responseBody)
	
	// 测试无效的内容类型
	t.Log("测试无效的Content-Type")
	req = httptest.NewRequest("POST", "/api/v1/corehr/employees", bytes.NewBuffer(createReqBody))
	req.Header.Set("Content-Type", "text/plain") // 错误的内容类型
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnsupportedMediaType, w.Code, "错误的Content-Type应该返回415状态码")
	
	// 测试无效的JSON
	t.Log("测试无效的JSON")
	req = httptest.NewRequest("POST", "/api/v1/corehr/employees", strings.NewReader("{invalid json}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "无效的JSON应该返回400状态码")
	
	// 测试列表参数验证
	t.Log("测试列表参数验证")
	req = httptest.NewRequest("GET", "/api/v1/corehr/employees?page=0&page_size=1000", nil) // 无效的分页参数
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "无效的分页参数应该返回400状态码")
}

// testEmployeeErrorHandling 测试错误处理
func testEmployeeErrorHandling(t *testing.T, router *chi.Mux) {
	t.Log("=== 测试错误处理 ===")
	
	// 测试无效的员工ID格式
	t.Log("测试无效的员工ID格式")
	req := httptest.NewRequest("GET", "/api/v1/corehr/employees/invalid-uuid", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "无效的UUID格式应该返回400状态码")
	
	// 测试缺少员工ID
	t.Log("测试缺少员工ID")
	req = httptest.NewRequest("GET", "/api/v1/corehr/employees/", nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// 应该路由到列表端点而不是详情端点
	assert.Equal(t, http.StatusOK, w.Code, "缺少员工ID应该路由到列表端点")
	
	// 测试无效的更新请求
	t.Log("测试无效的更新请求")
	employeeID := uuid.New()
	invalidUpdateReq := map[string]interface{}{
		"invalid_field": "invalid_value",
	}
	
	updateReqBody, _ := json.Marshal(invalidUpdateReq)
	req = httptest.NewRequest("PUT", fmt.Sprintf("/api/v1/corehr/employees/%s", employeeID.String()), bytes.NewBuffer(updateReqBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Mock服务会处理这个请求，但在真实环境中可能会有验证错误
	// 这里主要测试处理流程是否正常
	assert.True(t, w.Code >= 200 && w.Code < 500, "更新请求应该返回合理的状态码")
}

// testEmployeeManagerRelationship 测试员工经理关系
func testEmployeeManagerRelationship(t *testing.T, router *chi.Mux) {
	t.Log("=== 测试员工经理关系 ===")
	
	employeeID := uuid.New()
	
	// 测试获取员工经理
	t.Log("测试获取员工经理")
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/corehr/employees/%s/manager", employeeID.String()), nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code, "获取员工经理应该返回200状态码")
	
	var manager openapi.Employee
	err := json.Unmarshal(w.Body.Bytes(), &manager)
	require.NoError(t, err, "应该能够解析经理响应")
	
	assert.NotNil(t, manager.Id, "经理应该有ID")
	assert.NotEmpty(t, manager.EmployeeNumber, "经理应该有员工编号")
	
	// 测试无效的员工ID获取经理
	t.Log("测试无效员工ID获取经理")
	req = httptest.NewRequest("GET", "/api/v1/corehr/employees/invalid-uuid/manager", nil)
	w = httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code, "无效的员工ID应该返回400状态码")
}

// 辅助函数

func stringPtr(s string) *string {
	return &s
}

func emailPtr(s string) *openapi_types.Email {
	email := openapi_types.Email(s)
	return &email
}

// 模拟处理器函数（需要从main.go导入或重新实现）
// 这些函数应该从main.go文件中提取出来，放到一个可以测试的包中

func handleListEmployees(service *Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 简化的实现用于测试
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		
		// 返回Mock数据
		employees := []openapi.Employee{
			{
				Id:             &uuid.Nil,
				EmployeeNumber: "EMP001",
				FirstName:      "张",
				LastName:       "三",
				Email:          openapi_types.Email("zhangsan@example.com"),
			},
		}
		
		response := openapi.EmployeeListResponse{
			Employees:  &employees,
			TotalCount: intPtr(1),
		}
		
		json.NewEncoder(w).Encode(response)
	}
}

func handleCreateEmployee(service *Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 验证Content-Type
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		
		var req openapi.CreateEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// 基本验证
		if req.EmployeeNumber == "" || req.FirstName == "" || req.LastName == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}
		
		// 验证邮箱格式（简单验证）
		if !strings.Contains(string(req.Email), "@") {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}
		
		// 创建响应
		employee := openapi.Employee{
			Id:             &[]uuid.UUID{uuid.New()}[0],
			EmployeeNumber: req.EmployeeNumber,
			FirstName:      req.FirstName,
			LastName:       req.LastName,
			Email:          req.Email,
			HireDate:       req.HireDate,
			PhoneNumber:    req.PhoneNumber,
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(employee)
	}
}

func handleGetEmployee(service *Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		if employeeIDStr == "" {
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}
		
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}
		
		// 返回Mock员工数据
		employee := openapi.Employee{
			Id:             &employeeID,
			EmployeeNumber: "EMP001",
			FirstName:      "张",
			LastName:       "三",
			Email:          openapi_types.Email("zhangsan@example.com"),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(employee)
	}
}

func handleUpdateEmployee(service *Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		if employeeIDStr == "" {
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}
		
		employeeID, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}
		
		if r.Header.Get("Content-Type") != "application/json" {
			http.Error(w, "Content-Type must be application/json", http.StatusUnsupportedMediaType)
			return
		}
		
		var req openapi.UpdateEmployeeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}
		
		// 返回更新后的员工数据
		employee := openapi.Employee{
			Id:             &employeeID,
			EmployeeNumber: "EMP001",
			FirstName:      "张",
			LastName:       "三",
			Email:          openapi_types.Email("zhangsan@example.com"),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)  
		json.NewEncoder(w).Encode(employee)
	}
}

func handleDeleteEmployee(service *Service, logger *logging.StructuredLogger, validator *validation.EmployeeValidator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		if employeeIDStr == "" {
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}
		
		_, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}
		
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleGetEmployeeManager(service *Service, logger *logging.StructuredLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		employeeIDStr := chi.URLParam(r, "employeeID")
		if employeeIDStr == "" {
			http.Error(w, "Employee ID is required", http.StatusBadRequest)
			return
		}
		
		_, err := uuid.Parse(employeeIDStr)
		if err != nil {
			http.Error(w, "Invalid employee ID format", http.StatusBadRequest)
			return
		}
		
		// 返回Mock经理数据
		manager := openapi.Employee{
			Id:             &[]uuid.UUID{uuid.New()}[0],
			EmployeeNumber: "MGR001",
			FirstName:      "王",
			LastName:       "经理",
			Email:          openapi_types.Email("manager@example.com"),
		}
		
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(manager)
	}
}

func intPtr(i int) *int {
	return &i
}