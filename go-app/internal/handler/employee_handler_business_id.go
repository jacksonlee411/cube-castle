package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/middleware"
)

// EmployeeHandlerBusinessID 增强的员工处理器，支持业务ID
type EmployeeHandlerBusinessID struct {
	service *corehr.EmployeeService
}

// NewEmployeeHandlerBusinessID 创建员工处理器实例
func NewEmployeeHandlerBusinessID(service *corehr.EmployeeService) *EmployeeHandlerBusinessID {
	return &EmployeeHandlerBusinessID{service: service}
}

// GetEmployee 获取单个员工信息
// 支持业务ID和UUID查询模式
func (h *EmployeeHandlerBusinessID) GetEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")
	includeUUID := r.URL.Query().Get("include_uuid") == "true"
	uuidLookup := r.URL.Query().Get("uuid_lookup") == "true"

	// 构建查询选项
	opts := corehr.DefaultEmployeeQueryOptions()
	opts.IncludeUUID = includeUUID
	
	// 解析其他查询参数
	if r.URL.Query().Get("with_position") == "true" {
		opts.WithPosition = true
	}
	if r.URL.Query().Get("with_organization") == "true" {
		opts.WithOrgUnit = true
	}
	if r.URL.Query().Get("with_manager") == "true" {
		opts.WithManager = true
	}

	var employee *corehr.EmployeeResponse
	var err error

	if uuidLookup {
		// UUID查询模式
		id, parseErr := uuid.Parse(employeeID)
		if parseErr != nil {
			h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_UUID", "Invalid UUID format", nil)
			return
		}
		employee, err = h.service.GetEmployeeByUUID(r.Context(), id, opts)
	} else {
		// 业务ID查询模式 (默认)
		employee, err = h.service.GetEmployeeByBusinessID(r.Context(), employeeID, opts)
	}

	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "EMPLOYEE_NOT_FOUND", "Employee not found", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, employee)
}

// CreateEmployee 创建新员工
func (h *EmployeeHandlerBusinessID) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req corehr.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// 基本验证
	if err := h.validateCreateEmployeeRequest(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "VALIDATION_ERROR", err.Error(), nil)
		return
	}

	employee, err := h.service.CreateEmployee(r.Context(), req)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "already exists") {
			h.writeErrorResponse(w, http.StatusConflict, "EMPLOYEE_ALREADY_EXISTS", "Employee already exists", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusCreated, employee)
}

// UpdateEmployee 更新员工信息
func (h *EmployeeHandlerBusinessID) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")
	includeUUID := r.URL.Query().Get("include_uuid") == "true"

	var req corehr.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body", nil)
		return
	}

	// 构建查询选项
	opts := corehr.DefaultEmployeeQueryOptions()
	opts.IncludeUUID = includeUUID

	employee, err := h.service.UpdateEmployee(r.Context(), employeeID, req, opts)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "EMPLOYEE_NOT_FOUND", "Employee not found", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, employee)
}

// DeleteEmployee 删除员工
func (h *EmployeeHandlerBusinessID) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	employeeID := chi.URLParam(r, "employee_id")

	err := h.service.DeleteEmployee(r.Context(), employeeID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.writeErrorResponse(w, http.StatusNotFound, "EMPLOYEE_NOT_FOUND", "Employee not found", nil)
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListEmployees 获取员工列表
func (h *EmployeeHandlerBusinessID) ListEmployees(w http.ResponseWriter, r *http.Request) {
	// 解析查询参数
	req := corehr.ListEmployeesRequest{
		Page:     1,
		PageSize: 20,
		QueryOptions: corehr.DefaultEmployeeQueryOptions(),
	}

	// 解析分页参数
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil && page > 0 {
			req.Page = page
		}
	}

	if pageSizeStr := r.URL.Query().Get("page_size"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil && pageSize > 0 && pageSize <= 100 {
			req.PageSize = pageSize
		}
	}

	// 解析搜索参数
	if search := r.URL.Query().Get("search"); search != "" {
		req.Search = &search
	}

	// 解析过滤参数
	if status := r.URL.Query().Get("status"); status != "" {
		req.Status = &status
	}

	if employeeType := r.URL.Query().Get("employee_type"); employeeType != "" {
		req.EmployeeType = &employeeType
	}

	if orgID := r.URL.Query().Get("organization_id"); orgID != "" {
		req.OrganizationID = &orgID
	}

	// 解析查询选项
	if r.URL.Query().Get("include_uuid") == "true" {
		req.QueryOptions.IncludeUUID = true
	}
	if r.URL.Query().Get("with_position") == "true" {
		req.QueryOptions.WithPosition = true
	}
	if r.URL.Query().Get("with_organization") == "true" {
		req.QueryOptions.WithOrgUnit = true
	}
	if r.URL.Query().Get("with_manager") == "true" {
		req.QueryOptions.WithManager = true
	}

	// 设置租户ID (从上下文或其他地方获取)
	// 这里需要根据实际的认证/授权机制来设置
	req.TenantID = uuid.MustParse("00000000-0000-0000-0000-000000000000") // 默认租户

	response, err := h.service.ListEmployees(r.Context(), req)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error", nil)
		return
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// validateCreateEmployeeRequest 验证创建员工请求
func (h *EmployeeHandlerBusinessID) validateCreateEmployeeRequest(req *corehr.CreateEmployeeRequest) error {
	if req.FirstName == "" {
		return fmt.Errorf("first_name is required")
	}
	if req.LastName == "" {
		return fmt.Errorf("last_name is required")
	}
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}
	if req.HireDate == "" {
		return fmt.Errorf("hire_date is required")
	}
	if req.EmployeeType == "" {
		return fmt.Errorf("employee_type is required")
	}

	// 验证关联的业务ID格式
	if req.PositionID != nil {
		if err := common.ValidateBusinessID(common.EntityTypePosition, *req.PositionID); err != nil {
			return fmt.Errorf("invalid position_id: %w", err)
		}
	}

	if req.OrganizationID != nil {
		if err := common.ValidateBusinessID(common.EntityTypeOrganization, *req.OrganizationID); err != nil {
			return fmt.Errorf("invalid organization_id: %w", err)
		}
	}

	if req.ManagerID != nil {
		if err := common.ValidateBusinessID(common.EntityTypeEmployee, *req.ManagerID); err != nil {
			return fmt.Errorf("invalid manager_id: %w", err)
		}
	}

	return nil
}

// writeJSONResponse 写入JSON响应
func (h *EmployeeHandlerBusinessID) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if err := json.NewEncoder(w).Encode(data); err != nil {
		// 如果编码失败，记录错误但不能再修改响应
		// 这里可以添加日志记录
		return
	}
}

// writeErrorResponse 写入错误响应  
func (h *EmployeeHandlerBusinessID) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string, details map[string]string) {
	errorResp := middleware.ErrorResponse{
		Error:     errorCode,
		Message:   message,
		Details:   details,
		Timestamp: time.Now(),
		RequestID: generateRequestID(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(errorResp)
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("req_%s", uuid.New().String()[:8])
}

// RegisterEmployeeRoutes 注册员工路由
func (h *EmployeeHandlerBusinessID) RegisterEmployeeRoutes(r chi.Router) {
	// 业务ID验证中间件
	businessIDValidator := middleware.BusinessIDValidator(common.EntityTypeEmployee)

	r.Route("/employees", func(r chi.Router) {
		r.Get("/", h.ListEmployees)
		r.Post("/", h.CreateEmployee)
		
		r.Route("/{employee_id}", func(r chi.Router) {
			r.Use(businessIDValidator) // 应用业务ID验证
			r.Get("/", h.GetEmployee)
			r.Put("/", h.UpdateEmployee)
			r.Delete("/", h.DeleteEmployee)
		})
	})
}

// GetEmployeeStatistics 获取员工统计信息
func (h *EmployeeHandlerBusinessID) GetEmployeeStatistics(w http.ResponseWriter, r *http.Request) {
	// 这里可以添加员工统计信息的逻辑
	// 比如按状态、类型、部门等维度的统计
	stats := map[string]interface{}{
		"total_employees": 0,
		"active_employees": 0,
		"by_type": map[string]int{
			"FULL_TIME": 0,
			"PART_TIME": 0,
			"CONTRACTOR": 0,
			"INTERN": 0,
		},
		"by_status": map[string]int{
			"ACTIVE": 0,
			"ON_LEAVE": 0,
			"TERMINATED": 0,
			"SUSPENDED": 0,
			"PENDING_START": 0,
		},
	}

	h.writeJSONResponse(w, http.StatusOK, stats)
}

// ValidateBusinessID 手动验证业务ID的端点 (用于测试)
func (h *EmployeeHandlerBusinessID) ValidateBusinessID(w http.ResponseWriter, r *http.Request) {
	businessID := r.URL.Query().Get("business_id")
	if businessID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_PARAMETER", "business_id parameter is required", nil)
		return
	}

	err := common.ValidateBusinessID(common.EntityTypeEmployee, businessID)
	
	result := map[string]interface{}{
		"business_id": businessID,
		"valid": err == nil,
	}

	if err != nil {
		result["error"] = err.Error()
	}

	h.writeJSONResponse(w, http.StatusOK, result)
}