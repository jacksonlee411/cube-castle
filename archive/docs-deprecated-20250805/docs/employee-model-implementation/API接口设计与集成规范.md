# API接口设计与集成规范

## 概述

本文档定义了员工模型系统的RESTful API接口设计、GraphQL Schema定义和系统集成规范，确保与现有Cube Castle架构的无缝集成。

---

## 第一部分：RESTful API设计

### 1.1 API版本控制与路由

#### 基础路由结构

```go
// internal/api/routes/employee_routes.go
package routes

import (
    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
    
    "github.com/gaogu/cube-castle/internal/api/handlers"
    "github.com/gaogu/cube-castle/internal/api/middlewares"
)

// EmployeeAPIRouter 员工API路由配置
func EmployeeAPIRouter(h *handlers.EmployeeHandler) chi.Router {
    r := chi.NewRouter()
    
    // 通用中间件
    r.Use(middleware.RequestID)
    r.Use(middleware.RealIP)
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)
    
    // 认证和授权中间件
    r.Use(middlewares.JWTAuthMiddleware)
    r.Use(middlewares.TenantContextMiddleware)
    r.Use(middlewares.RBACMiddleware)
    
    // API版本控制
    r.Route("/api/v1", func(r chi.Router) {
        r.Route("/employees", func(r chi.Router) {
            // 员工基础操作
            r.Get("/", h.ListEmployees)           // GET /api/v1/employees
            r.Post("/", h.CreateEmployee)         // POST /api/v1/employees
            
            r.Route("/{employeeID}", func(r chi.Router) {
                r.Use(middlewares.EmployeeContextMiddleware) // 注入员工上下文
                
                r.Get("/", h.GetEmployee)         // GET /api/v1/employees/{id}
                r.Put("/", h.UpdateEmployee)      // PUT /api/v1/employees/{id}
                r.Delete("/", h.DeleteEmployee)   // DELETE /api/v1/employees/{id}
                
                // 职位管理
                r.Route("/positions", func(r chi.Router) {
                    r.Get("/", h.GetPositionHistory)         // GET /api/v1/employees/{id}/positions
                    r.Post("/", h.CreatePositionChange)      // POST /api/v1/employees/{id}/positions
                    r.Get("/current", h.GetCurrentPosition)  // GET /api/v1/employees/{id}/positions/current
                })
                
                // 组织关系
                r.Route("/organization", func(r chi.Router) {
                    r.Get("/reports", h.GetDirectReports)    // GET /api/v1/employees/{id}/organization/reports
                    r.Get("/manager", h.GetManager)          // GET /api/v1/employees/{id}/organization/manager
                    r.Get("/hierarchy", h.GetHierarchy)      // GET /api/v1/employees/{id}/organization/hierarchy
                })
                
                // 工作流操作
                r.Route("/workflows", func(r chi.Router) {
                    r.Post("/lifecycle", h.StartLifecycleWorkflow)    // POST /api/v1/employees/{id}/workflows/lifecycle
                    r.Post("/position-change", h.StartPositionChangeWorkflow) // POST /api/v1/employees/{id}/workflows/position-change
                    r.Get("/status", h.GetWorkflowStatus)             // GET /api/v1/employees/{id}/workflows/status
                })
            })
        })
        
        // 组织架构API
        r.Route("/organization", func(r chi.Router) {
            r.Get("/chart", h.GetOrganizationChart)       // GET /api/v1/organization/chart
            r.Get("/departments", h.GetDepartments)       // GET /api/v1/organization/departments
            r.Get("/positions", h.GetPositions)           // GET /api/v1/organization/positions
        })
        
        // 智能查询API
        r.Route("/intelligence", func(r chi.Router) {
            r.Post("/query", h.ProcessNaturalLanguageQuery)  // POST /api/v1/intelligence/query
            r.Get("/intents", h.GetSupportedIntents)          // GET /api/v1/intelligence/intents
            r.Post("/feedback", h.SubmitQueryFeedback)        // POST /api/v1/intelligence/feedback
        })
        
        // 高级查询和报表API
        r.Route("/analytics", func(r chi.Router) {
            r.Get("/headcount", h.GetHeadcountAnalytics)      // GET /api/v1/analytics/headcount
            r.Get("/turnover", h.GetTurnoverAnalytics)        // GET /api/v1/analytics/turnover
            r.Get("/demographics", h.GetDemographics)         // GET /api/v1/analytics/demographics
        })
    })
    
    return r
}
```

### 1.2 核心API处理器实现

#### 员工管理处理器

```go
// internal/api/handlers/employee_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "strconv"
    "time"
    
    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "go.uber.org/zap"
    
    "github.com/gaogu/cube-castle/internal/service"
    "github.com/gaogu/cube-castle/internal/api/dto"
    "github.com/gaogu/cube-castle/internal/api/response"
)

type EmployeeHandler struct {
    employeeService    *service.EmployeeService
    queryService       *service.EmployeeQueryService
    workflowService    *service.WorkflowService
    intelligenceService *service.IntelligenceService
    logger             *zap.Logger
}

// CreateEmployee 创建员工
func (h *EmployeeHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := middlewares.GetTenantID(ctx)
    userID := middlewares.GetUserID(ctx)
    
    var req dto.CreateEmployeeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body", err)
        return
    }
    
    // 验证请求
    if err := req.Validate(); err != nil {
        response.BadRequest(w, "Validation failed", err)
        return
    }
    
    // 调用业务服务
    result, err := h.employeeService.CreateEmployee(ctx, service.CreateEmployeeRequest{
        TenantID:        tenantID,
        PersonData:      req.PersonData,
        InitialPosition: req.InitialPosition,
        StartDate:       req.StartDate,
        RequestedBy:     userID,
    })
    
    if err != nil {
        h.logger.Error("Failed to create employee", 
            zap.Error(err),
            zap.String("tenant_id", tenantID.String()))
        response.InternalServerError(w, "Failed to create employee", err)
        return
    }
    
    // 转换响应
    resp := dto.CreateEmployeeResponse{
        EmployeeID:       result.EmployeeID,
        PersonID:         result.PersonID,
        EmployeeNumber:   result.EmployeeNumber,
        WorkflowID:       result.WorkflowID,
        Status:           result.Status,
        CreatedAt:        result.CreatedAt,
    }
    
    response.Created(w, resp)
}

// GetEmployee 获取员工详情
func (h *EmployeeHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := middlewares.GetTenantID(ctx)
    
    employeeIDStr := chi.URLParam(r, "employeeID")
    employeeID, err := uuid.Parse(employeeIDStr)
    if err != nil {
        response.BadRequest(w, "Invalid employee ID", err)
        return
    }
    
    // 解析查询参数
    asOfDateStr := r.URL.Query().Get("as_of_date")
    var asOfDate *time.Time
    if asOfDateStr != "" {
        parsed, err := time.Parse("2006-01-02", asOfDateStr)
        if err != nil {
            response.BadRequest(w, "Invalid as_of_date format, use YYYY-MM-DD", err)
            return
        }
        asOfDate = &parsed
    }
    
    includeHistory := r.URL.Query().Get("include_history") == "true"
    
    // 获取员工基础信息
    employee, err := h.queryService.GetEmployeeWithCurrentPosition(
        ctx, tenantID, employeeID, asOfDate)
    if err != nil {
        if service.IsNotFoundError(err) {
            response.NotFound(w, "Employee not found")
            return
        }
        h.logger.Error("Failed to get employee",
            zap.Error(err),
            zap.String("employee_id", employeeID.String()))
        response.InternalServerError(w, "Failed to get employee", err)
        return
    }
    
    resp := dto.EmployeeDetailResponse{
        PersonID:         employee.PersonID,
        EmployeeID:       employee.EmployeeID,
        EmployeeNumber:   employee.EmployeeNumber,
        FirstName:        employee.FirstName,
        LastName:         employee.LastName,
        Email:            employee.Email,
        HireDate:         employee.HireDate,
        EmploymentType:   employee.EmploymentType,
        EmploymentStatus: employee.EmploymentStatus,
        TerminationDate:  employee.TerminationDate,
        CurrentPosition:  convertPositionInfo(employee.CurrentPosition),
        CreatedAt:        employee.CreatedAt,
        UpdatedAt:        employee.UpdatedAt,
    }
    
    // 如果需要包含历史记录
    if includeHistory {
        history, err := h.queryService.GetEmployeePositionHistory(
            ctx, tenantID, employeeID)
        if err != nil {
            h.logger.Error("Failed to get position history",
                zap.Error(err),
                zap.String("employee_id", employeeID.String()))
            // 不因为历史记录失败而中断主要响应
        } else {
            resp.PositionHistory = convertPositionHistory(history)
        }
    }
    
    response.OK(w, resp)
}

// ListEmployees 列出员工
func (h *EmployeeHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := middlewares.GetTenantID(ctx)
    
    // 解析查询参数
    options := service.EmployeeQueryOptions{
        TenantID: tenantID,
    }
    
    if dept := r.URL.Query().Get("department"); dept != "" {
        options.Department = dept
    }
    
    if empType := r.URL.Query().Get("employment_type"); empType != "" {
        options.EmploymentType = empType
    }
    
    options.IncludeTerminated = r.URL.Query().Get("include_terminated") == "true"
    
    if asOfDateStr := r.URL.Query().Get("as_of_date"); asOfDateStr != "" {
        parsed, err := time.Parse("2006-01-02", asOfDateStr)
        if err != nil {
            response.BadRequest(w, "Invalid as_of_date format", err)
            return
        }
        options.AsOfDate = &parsed
    }
    
    // 分页参数
    if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
        if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
            options.Limit = limit
        }
    }
    if options.Limit == 0 {
        options.Limit = 50 // 默认分页大小
    }
    
    if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
        if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
            options.Offset = offset
        }
    }
    
    // 执行查询
    employees, total, err := h.queryService.ListEmployeesWithPositions(ctx, options)
    if err != nil {
        h.logger.Error("Failed to list employees",
            zap.Error(err),
            zap.String("tenant_id", tenantID.String()))
        response.InternalServerError(w, "Failed to list employees", err)
        return
    }
    
    // 转换响应
    employeeList := make([]dto.EmployeeSummary, len(employees))
    for i, emp := range employees {
        employeeList[i] = dto.EmployeeSummary{
            EmployeeID:       emp.EmployeeID,
            EmployeeNumber:   emp.EmployeeNumber,
            FirstName:        emp.FirstName,
            LastName:         emp.LastName,
            Email:            emp.Email,
            EmploymentStatus: emp.EmploymentStatus,
            CurrentPosition:  convertPositionInfo(emp.CurrentPosition),
        }
    }
    
    resp := dto.EmployeeListResponse{
        Employees:  employeeList,
        Pagination: dto.PaginationInfo{
            Total:  total,
            Limit:  options.Limit,
            Offset: options.Offset,
            HasMore: options.Offset+len(employees) < total,
        },
    }
    
    response.OK(w, resp)
}

// StartPositionChangeWorkflow 启动职位变更工作流
func (h *EmployeeHandler) StartPositionChangeWorkflow(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := middlewares.GetTenantID(ctx)
    userID := middlewares.GetUserID(ctx)
    
    employeeIDStr := chi.URLParam(r, "employeeID")
    employeeID, err := uuid.Parse(employeeIDStr)
    if err != nil {
        response.BadRequest(w, "Invalid employee ID", err)
        return
    }
    
    var req dto.StartPositionChangeRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body", err)
        return
    }
    
    if err := req.Validate(); err != nil {
        response.BadRequest(w, "Validation failed", err)
        return
    }
    
    // 启动工作流
    result, err := h.workflowService.StartPositionChangeWorkflow(ctx,
        service.PositionChangeWorkflowRequest{
            TenantID:      tenantID,
            EmployeeID:    employeeID,
            NewPosition:   req.NewPosition,
            EffectiveDate: req.EffectiveDate,
            ChangeReason:  req.ChangeReason,
            RequestedBy:   userID,
        })
    
    if err != nil {
        h.logger.Error("Failed to start position change workflow",
            zap.Error(err),
            zap.String("employee_id", employeeID.String()))
        response.InternalServerError(w, "Failed to start workflow", err)
        return
    }
    
    resp := dto.WorkflowResponse{
        WorkflowID:   result.WorkflowID,
        RunID:        result.RunID,
        Status:       result.Status,
        StartedAt:    result.StartedAt,
    }
    
    response.Accepted(w, resp)
}

// ProcessNaturalLanguageQuery 处理自然语言查询
func (h *EmployeeHandler) ProcessNaturalLanguageQuery(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    tenantID := middlewares.GetTenantID(ctx)
    userID := middlewares.GetUserID(ctx)
    
    var req dto.NaturalLanguageQueryRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        response.BadRequest(w, "Invalid request body", err)
        return
    }
    
    if err := req.Validate(); err != nil {
        response.BadRequest(w, "Validation failed", err)
        return
    }
    
    // 构建用户上下文
    userContext := service.UserContext{
        UserID:      userID,
        TenantID:    tenantID,
        Roles:       middlewares.GetUserRoles(ctx),
        Permissions: middlewares.GetUserPermissions(ctx),
        Language:    req.Language,
    }
    
    // 处理查询
    result, err := h.intelligenceService.ProcessQuery(ctx,
        service.IntelligenceQueryRequest{
            Query:       req.Query,
            UIContext:   req.UIContext,
            UserContext: userContext,
            SessionID:   req.SessionID,
        })
    
    if err != nil {
        h.logger.Error("Failed to process intelligence query",
            zap.Error(err),
            zap.String("query", req.Query))
        response.InternalServerError(w, "Failed to process query", err)
        return
    }
    
    resp := dto.IntelligenceQueryResponse{
        Intent:     convertIntent(result.Intent),
        Entities:   convertEntities(result.Entities),
        Confidence: result.Confidence,
        Status:     result.Status,
        Message:    result.Message,
        NextAction: convertNextAction(result.NextAction),
        SessionID:  req.SessionID,
    }
    
    response.OK(w, resp)
}
```

### 1.3 数据传输对象(DTO)定义

#### 员工相关DTO

```go
// internal/api/dto/employee_dto.go
package dto

import (
    "time"
    "github.com/google/uuid"
)

// CreateEmployeeRequest 创建员工请求
type CreateEmployeeRequest struct {
    PersonData      PersonCreationData `json:"person_data" validate:"required"`
    InitialPosition PositionData       `json:"initial_position" validate:"required"`
    StartDate       time.Time          `json:"start_date" validate:"required"`
}

type PersonCreationData struct {
    FirstName   string `json:"first_name" validate:"required,min=1,max=50"`
    LastName    string `json:"last_name" validate:"required,min=1,max=50"`
    Email       string `json:"email" validate:"required,email,max=255"`
    PhoneNumber string `json:"phone_number,omitempty" validate:"omitempty,max=20"`
    DateOfBirth string `json:"date_of_birth,omitempty" validate:"omitempty,datetime=2006-01-02"`
}

type PositionData struct {
    PositionTitle   string     `json:"position_title" validate:"required,min=1,max=100"`
    Department      string     `json:"department" validate:"required,min=1,max=100"`
    ReportsTo       *uuid.UUID `json:"reports_to,omitempty"`
    JobLevel        string     `json:"job_level,omitempty" validate:"omitempty,max=50"`
    EmploymentType  string     `json:"employment_type" validate:"required,oneof=FULL_TIME PART_TIME CONTRACT INTERN"`
    Location        string     `json:"location,omitempty" validate:"omitempty,max=100"`
}

func (r *CreateEmployeeRequest) Validate() error {
    // 自定义验证逻辑
    return nil
}

// CreateEmployeeResponse 创建员工响应
type CreateEmployeeResponse struct {
    EmployeeID     uuid.UUID `json:"employee_id"`
    PersonID       uuid.UUID `json:"person_id"`
    EmployeeNumber string    `json:"employee_number"`
    WorkflowID     string    `json:"workflow_id"`
    Status         string    `json:"status"`
    CreatedAt      time.Time `json:"created_at"`
}

// EmployeeDetailResponse 员工详情响应
type EmployeeDetailResponse struct {
    PersonID         uuid.UUID       `json:"person_id"`
    EmployeeID       uuid.UUID       `json:"employee_id"`
    EmployeeNumber   string          `json:"employee_number"`
    FirstName        string          `json:"first_name"`
    LastName         string          `json:"last_name"`
    Email            string          `json:"email"`
    HireDate         time.Time       `json:"hire_date"`
    EmploymentType   string          `json:"employment_type"`
    EmploymentStatus string          `json:"employment_status"`
    TerminationDate  *time.Time      `json:"termination_date,omitempty"`
    CurrentPosition  *PositionInfo   `json:"current_position,omitempty"`
    PositionHistory  []PositionInfo  `json:"position_history,omitempty"`
    CreatedAt        time.Time       `json:"created_at"`
    UpdatedAt        time.Time       `json:"updated_at"`
}

type PositionInfo struct {
    PositionHistoryID uuid.UUID  `json:"position_history_id"`
    PositionTitle     string     `json:"position_title"`
    Department        string     `json:"department"`
    JobLevel          string     `json:"job_level"`
    Location          string     `json:"location"`
    ReportsToID       *uuid.UUID `json:"reports_to_id,omitempty"`
    ReportsToName     string     `json:"reports_to_name,omitempty"`
    EffectiveDate     time.Time  `json:"effective_date"`
    EndDate           *time.Time `json:"end_date,omitempty"`
}

// EmployeeSummary 员工摘要信息
type EmployeeSummary struct {
    EmployeeID       uuid.UUID     `json:"employee_id"`
    EmployeeNumber   string        `json:"employee_number"`
    FirstName        string        `json:"first_name"`
    LastName         string        `json:"last_name"`
    Email            string        `json:"email"`
    EmploymentStatus string        `json:"employment_status"`
    CurrentPosition  *PositionInfo `json:"current_position,omitempty"`
}

// EmployeeListResponse 员工列表响应
type EmployeeListResponse struct {
    Employees  []EmployeeSummary `json:"employees"`
    Pagination PaginationInfo    `json:"pagination"`
}

type PaginationInfo struct {
    Total   int  `json:"total"`
    Limit   int  `json:"limit"`
    Offset  int  `json:"offset"`
    HasMore bool `json:"has_more"`
}

// StartPositionChangeRequest 启动职位变更请求
type StartPositionChangeRequest struct {
    NewPosition   PositionData `json:"new_position" validate:"required"`
    EffectiveDate time.Time    `json:"effective_date" validate:"required"`
    ChangeReason  string       `json:"change_reason,omitempty" validate:"omitempty,max=500"`
}

func (r *StartPositionChangeRequest) Validate() error {
    // 验证生效日期不能太久远
    if r.EffectiveDate.Before(time.Now().AddDate(-1, 0, 0)) {
        return fmt.Errorf("effective_date cannot be more than 1 year in the past")
    }
    if r.EffectiveDate.After(time.Now().AddDate(1, 0, 0)) {
        return fmt.Errorf("effective_date cannot be more than 1 year in the future")
    }
    return nil
}

// WorkflowResponse 工作流响应
type WorkflowResponse struct {
    WorkflowID string    `json:"workflow_id"`
    RunID      string    `json:"run_id"`
    Status     string    `json:"status"`
    StartedAt  time.Time `json:"started_at"`
}

// 智能查询相关DTO
type NaturalLanguageQueryRequest struct {
    Query     string    `json:"query" validate:"required,min=1,max=1000"`
    UIContext UIContext `json:"ui_context"`
    Language  string    `json:"language,omitempty" validate:"omitempty,oneof=zh-CN en-US"`
    SessionID string    `json:"session_id,omitempty"`
}

type UIContext struct {
    PageID          string                 `json:"page_id"`
    PageType        string                 `json:"page_type"`
    DataContext     map[string]interface{} `json:"data_context,omitempty"`
    AffordedIntents []string               `json:"afforded_intents,omitempty"`
}

func (r *NaturalLanguageQueryRequest) Validate() error {
    if r.Language == "" {
        r.Language = "zh-CN"
    }
    return nil
}

type IntelligenceQueryResponse struct {
    Intent     IntentInfo        `json:"intent"`
    Entities   []EntityInfo      `json:"entities"`
    Confidence float64           `json:"confidence"`
    Status     string            `json:"status"`
    Message    string            `json:"message,omitempty"`
    NextAction *NextActionInfo   `json:"next_action,omitempty"`
    SessionID  string            `json:"session_id"`
}

type IntentInfo struct {
    Name        string  `json:"name"`
    Category    string  `json:"category"`
    Description string  `json:"description"`
    Confidence  float64 `json:"confidence"`
}

type EntityInfo struct {
    Name       string      `json:"name"`
    Type       string      `json:"type"`
    Value      interface{} `json:"value"`
    Confidence float64     `json:"confidence"`
    Source     string      `json:"source"`
}

type NextActionInfo struct {
    ActionType   string                 `json:"action_type"`
    Target       string                 `json:"target,omitempty"`
    Parameters   map[string]interface{} `json:"parameters,omitempty"`
    Confirmation *ConfirmationInfo      `json:"confirmation,omitempty"`
}

type ConfirmationInfo struct {
    Message    string                 `json:"message"`
    Parameters map[string]interface{} `json:"parameters"`
    TimeoutMs  int                    `json:"timeout_ms"`
}
```

---

## 第二部分：GraphQL Schema设计

### 2.1 核心Schema定义

#### Employee相关Schema

```graphql
# schema/employee.graphql

"""
员工图查询Schema
支持复杂的关系查询和图遍历
"""

# 基础类型定义
scalar UUID
scalar DateTime
scalar Date

# 员工节点
type Employee {
  # 基础标识
  id: UUID!
  employeeNumber: String!
  
  # 个人信息
  person: Person!
  
  # 就业信息
  hireDate: Date!
  terminationDate: Date
  employmentType: EmploymentType!
  employmentStatus: EmploymentStatus!
  
  # 当前职位
  currentPosition(asOfDate: Date): Position
  
  # 职位历史
  positionHistory(
    fromDate: Date
    toDate: Date
    limit: Int = 50
    offset: Int = 0
  ): PositionHistoryConnection!
  
  # 组织关系
  directReports(
    includeTerminated: Boolean = false
    asOfDate: Date
  ): [Employee!]!
  
  manager(asOfDate: Date): Employee
  
  # 层级关系查询
  reportingChain(
    direction: HierarchyDirection = UP
    maxLevels: Int = 5
    asOfDate: Date
  ): [Employee!]!
  
  # 协作关系
  colleagues(
    department: String
    includeOtherDepartments: Boolean = false
  ): [Employee!]!
  
  # 审计信息
  createdAt: DateTime!
  updatedAt: DateTime!
}

# 人员基础信息
type Person {
  id: UUID!
  firstName: String!
  lastName: String!
  fullName: String! # 计算字段
  email: String!
  phoneNumber: String
  dateOfBirth: Date
  status: PersonStatus!
}

# 职位信息
type Position {
  id: UUID!
  title: String!
  department: String!
  jobLevel: String
  location: String
  employmentType: EmploymentType!
  
  # 汇报关系
  reportsTo: Employee
  
  # 时态信息
  effectiveDate: Date!
  endDate: Date
  changeReason: String
  
  # 薪酬信息（权限控制）
  salaryRange: SalaryRange @auth(requires: ["hr.compensation.read"])
}

# 薪酬范围
type SalaryRange {
  minSalary: Float
  maxSalary: Float
  currency: String!
}

# 职位历史连接
type PositionHistoryConnection {
  edges: [PositionHistoryEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type PositionHistoryEdge {
  node: Position!
  cursor: String!
}

type PageInfo {
  hasNextPage: Boolean!
  hasPreviousPage: Boolean!
  startCursor: String
  endCursor: String
}

# 枚举类型
enum EmploymentType {
  FULL_TIME
  PART_TIME
  CONTRACT
  INTERN
}

enum EmploymentStatus {
  ACTIVE
  ON_LEAVE
  TERMINATED
  SUSPENDED
}

enum PersonStatus {
  ACTIVE
  INACTIVE
  PENDING
  TERMINATED
}

enum HierarchyDirection {
  UP      # 向上查询管理层级
  DOWN    # 向下查询下属层级
  BOTH    # 双向查询
}

# 查询根类型
type Query {
  # 员工查询
  employee(id: UUID!): Employee
  
  employees(
    # 过滤条件
    department: String
    employmentType: EmploymentType
    employmentStatus: EmploymentStatus
    includeTerminated: Boolean = false
    
    # 时间条件
    asOfDate: Date
    hiredAfter: Date
    hiredBefore: Date
    
    # 搜索条件
    searchTerm: String
    
    # 分页
    first: Int = 50
    after: String
    last: Int
    before: String
    
    # 排序
    orderBy: EmployeeOrderBy = LAST_NAME_ASC
  ): EmployeeConnection!
  
  # 组织查询
  organizationChart(
    rootDepartment: String
    maxLevels: Int = 5
    includeTerminated: Boolean = false
    asOfDate: Date
  ): OrganizationChart!
  
  # 部门查询
  departments: [Department!]!
  
  # 高级图查询
  findReportingPath(
    fromEmployee: UUID!
    toEmployee: UUID!
    asOfDate: Date
  ): [Employee!]
  
  # 查找共同上级
  findCommonManager(
    employees: [UUID!]!
    asOfDate: Date
  ): Employee
}

# 变更操作
type Mutation {
  # 员工创建
  createEmployee(input: CreateEmployeeInput!): CreateEmployeePayload!
  
  # 职位变更
  changePosition(input: ChangePositionInput!): ChangePositionPayload!
  
  # 员工更新
  updateEmployee(input: UpdateEmployeeInput!): UpdateEmployeePayload!
  
  # 员工终止
  terminateEmployee(input: TerminateEmployeeInput!): TerminateEmployeePayload!
}

# 输入类型
input CreateEmployeeInput {
  personData: PersonInput!
  initialPosition: PositionInput!
  startDate: Date!
}

input PersonInput {
  firstName: String!
  lastName: String!
  email: String!
  phoneNumber: String
  dateOfBirth: Date
}

input PositionInput {
  title: String!
  department: String!
  jobLevel: String
  location: String
  employmentType: EmploymentType!
  reportsTo: UUID
}

input ChangePositionInput {
  employeeId: UUID!
  newPosition: PositionInput!
  effectiveDate: Date!
  changeReason: String
}

input UpdateEmployeeInput {
  employeeId: UUID!
  personData: PersonUpdateInput
}

input PersonUpdateInput {
  firstName: String
  lastName: String
  phoneNumber: String
}

input TerminateEmployeeInput {
  employeeId: UUID!
  terminationDate: Date!
  reason: String!
}

# 响应载荷类型
type CreateEmployeePayload {
  employee: Employee
  errors: [ValidationError!]
}

type ChangePositionPayload {
  employee: Employee
  position: Position
  workflowId: String
  errors: [ValidationError!]
}

type UpdateEmployeePayload {
  employee: Employee
  errors: [ValidationError!]
}

type TerminateEmployeePayload {
  employee: Employee
  workflowId: String
  errors: [ValidationError!]
}

type ValidationError {
  field: String!
  message: String!
}

# 连接类型
type EmployeeConnection {
  edges: [EmployeeEdge!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type EmployeeEdge {
  node: Employee!
  cursor: String!
}

# 组织架构
type OrganizationChart {
  rootNodes: [OrganizationNode!]!
  totalEmployees: Int!
  asOfDate: Date!
}

type OrganizationNode {
  employee: Employee!
  directReports: [OrganizationNode!]!
  level: Int!
}

type Department {
  name: String!
  description: String
  headCount: Int!
  manager: Employee
  subDepartments: [Department!]!
}

# 排序枚举
enum EmployeeOrderBy {
  LAST_NAME_ASC
  LAST_NAME_DESC
  FIRST_NAME_ASC
  FIRST_NAME_DESC
  HIRE_DATE_ASC
  HIRE_DATE_DESC
  EMPLOYEE_NUMBER_ASC
  EMPLOYEE_NUMBER_DESC
}

# 权限指令
directive @auth(
  requires: [String!]!
) on FIELD_DEFINITION

# 订阅类型（实时更新）
type Subscription {
  # 员工变更订阅
  employeeUpdated(employeeId: UUID): Employee!
  
  # 组织架构变更订阅
  organizationChanged(department: String): OrganizationChart!
  
  # 工作流状态变更
  workflowStatusChanged(workflowId: String!): WorkflowStatus!
}

type WorkflowStatus {
  workflowId: String!
  status: String!
  updatedAt: DateTime!
}
```

### 2.2 GraphQL解析器实现

#### 员工解析器

```go
// internal/graphql/resolvers/employee_resolver.go
package resolvers

import (
    "context"
    "fmt"
    
    "github.com/google/uuid"
    "github.com/gaogu/cube-castle/internal/service"
    "github.com/gaogu/cube-castle/internal/graphql/generated"
    "github.com/gaogu/cube-castle/internal/graphql/model"
)

type EmployeeResolver struct {
    employeeService *service.EmployeeService
    queryService    *service.EmployeeQueryService
    authService     *service.AuthorizationService
}

// Employee 解析员工字段
func (r *EmployeeResolver) Employee(ctx context.Context, obj *model.Employee) (*model.Employee, error) {
    return obj, nil
}

// CurrentPosition 解析当前职位
func (r *EmployeeResolver) CurrentPosition(ctx context.Context, obj *model.Employee, asOfDate *string) (*model.Position, error) {
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    employee, err := r.queryService.GetEmployeeWithCurrentPosition(
        ctx, tenantID, obj.ID, asOf)
    if err != nil {
        return nil, err
    }
    
    if employee.CurrentPosition == nil {
        return nil, nil
    }
    
    return convertToGraphQLPosition(employee.CurrentPosition), nil
}

// PositionHistory 解析职位历史
func (r *EmployeeResolver) PositionHistory(
    ctx context.Context, 
    obj *model.Employee, 
    fromDate, toDate *string,
    limit, offset *int,
) (*model.PositionHistoryConnection, error) {
    
    tenantID := GetTenantID(ctx)
    
    // 解析日期参数
    var from, to *time.Time
    if fromDate != nil {
        parsed, err := time.Parse("2006-01-02", *fromDate)
        if err != nil {
            return nil, fmt.Errorf("invalid fromDate format: %w", err)
        }
        from = &parsed
    }
    if toDate != nil {
        parsed, err := time.Parse("2006-01-02", *toDate)
        if err != nil {
            return nil, fmt.Errorf("invalid toDate format: %w", err)
        }
        to = &parsed
    }
    
    // 设置默认值
    if limit == nil {
        defaultLimit := 50
        limit = &defaultLimit
    }
    if offset == nil {
        defaultOffset := 0
        offset = &defaultOffset
    }
    
    // 查询历史记录
    positions, err := r.queryService.GetEmployeePositionHistoryWithFilter(
        ctx, tenantID, obj.ID, service.PositionHistoryFilter{
            FromDate: from,
            ToDate:   to,
            Limit:    *limit,
            Offset:   *offset,
        })
    if err != nil {
        return nil, err
    }
    
    // 转换为GraphQL连接格式
    edges := make([]*model.PositionHistoryEdge, len(positions.Items))
    for i, pos := range positions.Items {
        edges[i] = &model.PositionHistoryEdge{
            Node:   convertToGraphQLPosition(pos),
            Cursor: encodeCursor(pos.EffectiveDate),
        }
    }
    
    return &model.PositionHistoryConnection{
        Edges: edges,
        PageInfo: &model.PageInfo{
            HasNextPage:     positions.HasMore,
            HasPreviousPage: *offset > 0,
            StartCursor:     getStartCursor(edges),
            EndCursor:       getEndCursor(edges),
        },
        TotalCount: positions.Total,
    }, nil
}

// DirectReports 解析直接下属
func (r *EmployeeResolver) DirectReports(
    ctx context.Context,
    obj *model.Employee,
    includeTerminated *bool,
    asOfDate *string,
) ([]*model.Employee, error) {
    
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    includeTerminatedFlag := false
    if includeTerminated != nil {
        includeTerminatedFlag = *includeTerminated
    }
    
    reports, err := r.queryService.GetDirectReports(ctx, 
        service.DirectReportsQuery{
            TenantID:          tenantID,
            ManagerID:         obj.ID,
            IncludeTerminated: includeTerminatedFlag,
            AsOfDate:          asOf,
        })
    if err != nil {
        return nil, err
    }
    
    // 转换为GraphQL模型
    result := make([]*model.Employee, len(reports))
    for i, report := range reports {
        result[i] = convertToGraphQLEmployee(report)
    }
    
    return result, nil
}

// Manager 解析上级经理
func (r *EmployeeResolver) Manager(
    ctx context.Context,
    obj *model.Employee,
    asOfDate *string,
) (*model.Employee, error) {
    
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    manager, err := r.queryService.GetManager(ctx, tenantID, obj.ID, asOf)
    if err != nil {
        if service.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, err
    }
    
    return convertToGraphQLEmployee(manager), nil
}

// ReportingChain 解析汇报链
func (r *EmployeeResolver) ReportingChain(
    ctx context.Context,
    obj *model.Employee,
    direction *model.HierarchyDirection,
    maxLevels *int,
    asOfDate *string,
) ([]*model.Employee, error) {
    
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    dir := model.HierarchyDirectionUp
    if direction != nil {
        dir = *direction
    }
    
    maxLvl := 5
    if maxLevels != nil {
        maxLvl = *maxLevels
    }
    
    chain, err := r.queryService.GetReportingChain(ctx,
        service.ReportingChainQuery{
            TenantID:   tenantID,
            EmployeeID: obj.ID,
            Direction:  convertHierarchyDirection(dir),
            MaxLevels:  maxLvl,
            AsOfDate:   asOf,
        })
    if err != nil {
        return nil, err
    }
    
    result := make([]*model.Employee, len(chain))
    for i, emp := range chain {
        result[i] = convertToGraphQLEmployee(emp)
    }
    
    return result, nil
}

// Query解析器
func (r *QueryResolver) Employee(ctx context.Context, id uuid.UUID) (*model.Employee, error) {
    tenantID := GetTenantID(ctx)
    
    employee, err := r.queryService.GetEmployeeWithCurrentPosition(
        ctx, tenantID, id, nil)
    if err != nil {
        if service.IsNotFoundError(err) {
            return nil, nil
        }
        return nil, err
    }
    
    return convertToGraphQLEmployee(employee), nil
}

func (r *QueryResolver) Employees(
    ctx context.Context,
    department, searchTerm *string,
    employmentType *model.EmploymentType,
    employmentStatus *model.EmploymentStatus,
    includeTerminated *bool,
    asOfDate, hiredAfter, hiredBefore *string,
    first *int,
    after *string,
    last *int,
    before *string,
    orderBy *model.EmployeeOrderBy,
) (*model.EmployeeConnection, error) {
    
    tenantID := GetTenantID(ctx)
    
    // 构建查询选项
    options := service.EmployeeQueryOptions{
        TenantID: tenantID,
    }
    
    if department != nil {
        options.Department = *department
    }
    
    if employmentType != nil {
        options.EmploymentType = string(*employmentType)
    }
    
    if includeTerminated != nil {
        options.IncludeTerminated = *includeTerminated
    }
    
    // 处理日期参数
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid asOfDate format: %w", err)
        }
        options.AsOfDate = &parsed
    }
    
    // 处理分页
    if first != nil {
        options.Limit = *first
    } else if last != nil {
        options.Limit = *last
    } else {
        options.Limit = 50
    }
    
    if after != nil {
        offset, err := decodeCursor(*after)
        if err != nil {
            return nil, fmt.Errorf("invalid after cursor: %w", err)
        }
        options.Offset = offset
    }
    
    // 执行查询
    employees, total, err := r.queryService.ListEmployeesWithPositions(ctx, options)
    if err != nil {
        return nil, err
    }
    
    // 转换为连接格式
    edges := make([]*model.EmployeeEdge, len(employees))
    for i, emp := range employees {
        edges[i] = &model.EmployeeEdge{
            Node:   convertToGraphQLEmployee(emp),
            Cursor: encodeCursorFromEmployee(emp),
        }
    }
    
    return &model.EmployeeConnection{
        Edges: edges,
        PageInfo: &model.PageInfo{
            HasNextPage:     options.Offset+len(employees) < total,
            HasPreviousPage: options.Offset > 0,
            StartCursor:     getStartCursor(edges),
            EndCursor:       getEndCursor(edges),
        },
        TotalCount: total,
    }, nil
}

// 组织架构查询
func (r *QueryResolver) OrganizationChart(
    ctx context.Context,
    rootDepartment *string,
    maxLevels *int,
    includeTerminated *bool,
    asOfDate *string,
) (*model.OrganizationChart, error) {
    
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    maxLvl := 5
    if maxLevels != nil {
        maxLvl = *maxLevels
    }
    
    includeTerminatedFlag := false
    if includeTerminated != nil {
        includeTerminatedFlag = *includeTerminated
    }
    
    chart, err := r.queryService.GetOrganizationChart(ctx,
        service.OrganizationChartQuery{
            TenantID:          tenantID,
            RootDepartment:    rootDepartment,
            MaxLevels:         maxLvl,
            IncludeTerminated: includeTerminatedFlag,
            AsOfDate:          asOf,
        })
    if err != nil {
        return nil, err
    }
    
    return convertToGraphQLOrganizationChart(chart), nil
}

// 高级图查询 - 查找汇报路径
func (r *QueryResolver) FindReportingPath(
    ctx context.Context,
    fromEmployee, toEmployee uuid.UUID,
    asOfDate *string,
) ([]*model.Employee, error) {
    
    tenantID := GetTenantID(ctx)
    
    var asOf *time.Time
    if asOfDate != nil {
        parsed, err := time.Parse("2006-01-02", *asOfDate)
        if err != nil {
            return nil, fmt.Errorf("invalid date format: %w", err)
        }
        asOf = &parsed
    }
    
    path, err := r.queryService.FindReportingPath(ctx,
        service.ReportingPathQuery{
            TenantID:     tenantID,
            FromEmployee: fromEmployee,
            ToEmployee:   toEmployee,
            AsOfDate:     asOf,
        })
    if err != nil {
        return nil, err
    }
    
    result := make([]*model.Employee, len(path))
    for i, emp := range path {
        result[i] = convertToGraphQLEmployee(emp)
    }
    
    return result, nil
}

// 辅助函数
func convertToGraphQLEmployee(emp *service.EmployeeWithPosition) *model.Employee {
    return &model.Employee{
        ID:             emp.EmployeeID,
        EmployeeNumber: emp.EmployeeNumber,
        Person: &model.Person{
            ID:        emp.PersonID,
            FirstName: emp.FirstName,
            LastName:  emp.LastName,
            FullName:  emp.FirstName + " " + emp.LastName,
            Email:     emp.Email,
        },
        HireDate:         emp.HireDate,
        TerminationDate:  emp.TerminationDate,
        EmploymentType:   model.EmploymentType(emp.EmploymentType),
        EmploymentStatus: model.EmploymentStatus(emp.EmploymentStatus),
        CreatedAt:        emp.CreatedAt,
        UpdatedAt:        emp.UpdatedAt,
    }
}

func convertToGraphQLPosition(pos *service.PositionInfo) *model.Position {
    if pos == nil {
        return nil
    }
    
    return &model.Position{
        ID:            pos.PositionHistoryID,
        Title:         pos.PositionTitle,
        Department:    pos.Department,
        JobLevel:      pos.JobLevel,
        Location:      pos.Location,
        EffectiveDate: pos.EffectiveDate,
        EndDate:       pos.EndDate,
    }
}
```

---

## 第三部分：服务集成规范

### 3.1 中间件设计

#### 认证和授权中间件

```go
// internal/api/middlewares/auth_middleware.go
package middlewares

import (
    "context"
    "net/http"
    "strings"
    
    "github.com/google/uuid"
    "go.uber.org/zap"
    
    "github.com/gaogu/cube-castle/internal/auth"
    "github.com/gaogu/cube-castle/internal/api/response"
    "github.com/gaogu/cube-castle/internal/opa"
)

type contextKey string

const (
    UserIDKey      contextKey = "user_id"
    TenantIDKey    contextKey = "tenant_id"
    UserRolesKey   contextKey = "user_roles"
    PermissionsKey contextKey = "permissions"
)

// JWTAuthMiddleware JWT认证中间件
func JWTAuthMiddleware(jwtService *auth.JWTService, logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // 提取Authorization头
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                response.Unauthorized(w, "Missing authorization header")
                return
            }
            
            // 验证Bearer token格式
            parts := strings.SplitN(authHeader, " ", 2)
            if len(parts) != 2 || parts[0] != "Bearer" {
                response.Unauthorized(w, "Invalid authorization header format")
                return
            }
            
            token := parts[1]
            
            // 验证JWT token
            claims, err := jwtService.ValidateToken(token)
            if err != nil {
                logger.Warn("Invalid JWT token", zap.Error(err))
                response.Unauthorized(w, "Invalid token")
                return
            }
            
            // 提取用户信息
            userID, err := uuid.Parse(claims.Subject)
            if err != nil {
                response.Unauthorized(w, "Invalid user ID in token")
                return
            }
            
            tenantID, err := uuid.Parse(claims.TenantID)
            if err != nil {
                response.Unauthorized(w, "Invalid tenant ID in token")
                return
            }
            
            // 将用户信息注入到context
            ctx := context.WithValue(r.Context(), UserIDKey, userID)
            ctx = context.WithValue(ctx, TenantIDKey, tenantID)
            ctx = context.WithValue(ctx, UserRolesKey, claims.Roles)
            ctx = context.WithValue(ctx, PermissionsKey, claims.Permissions)
            
            // 继续处理
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// TenantContextMiddleware 租户上下文中间件
func TenantContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()
        tenantID := GetTenantID(ctx)
        
        // 为数据库查询设置租户上下文
        if tenantID != uuid.Nil {
            // 这里可以设置数据库会话变量或其他租户上下文
            ctx = context.WithValue(ctx, "db_tenant_id", tenantID.String())
        }
        
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}

// RBACMiddleware 基于角色的访问控制中间件
func RBACMiddleware(opaService *opa.Service, logger *zap.Logger) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            
            // 构建授权输入
            authInput := opa.HTTPAuthorizationInput{
                UserID:      GetUserID(ctx),
                TenantID:    GetTenantID(ctx),
                Method:      r.Method,
                Path:        r.URL.Path,
                Roles:       GetUserRoles(ctx),
                Permissions: GetUserPermissions(ctx),
                Headers:     r.Header,
                QueryParams: r.URL.Query(),
            }
            
            // 检查授权
            result, err := opaService.AuthorizeHTTPRequest(ctx, authInput)
            if err != nil {
                logger.Error("Authorization check failed", 
                    zap.Error(err),
                    zap.String("path", r.URL.Path),
                    zap.String("method", r.Method))
                response.InternalServerError(w, "Authorization check failed", err)
                return
            }
            
            if !result.Allowed {
                logger.Warn("Access denied",
                    zap.String("user_id", authInput.UserID.String()),
                    zap.String("path", r.URL.Path),
                    zap.String("method", r.Method),
                    zap.String("reason", result.Reason))
                response.Forbidden(w, "Access denied")
                return
            }
            
            next.ServeHTTP(w, r)
        })
    }
}

// EmployeeContextMiddleware 员工上下文中间件
func EmployeeContextMiddleware(employeeService *service.EmployeeService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := r.Context()
            tenantID := GetTenantID(ctx)
            
            employeeIDStr := chi.URLParam(r, "employeeID")
            if employeeIDStr == "" {
                response.BadRequest(w, "Missing employee ID", nil)
                return
            }
            
            employeeID, err := uuid.Parse(employeeIDStr)
            if err != nil {
                response.BadRequest(w, "Invalid employee ID", err)
                return
            }
            
            // 验证员工是否存在且属于当前租户
            exists, err := employeeService.EmployeeExists(ctx, tenantID, employeeID)
            if err != nil {
                response.InternalServerError(w, "Failed to verify employee", err)
                return
            }
            
            if !exists {
                response.NotFound(w, "Employee not found")
                return
            }
            
            // 将员工ID注入到context
            ctx = context.WithValue(ctx, "employee_id", employeeID)
            
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// 上下文访问器函数
func GetUserID(ctx context.Context) uuid.UUID {
    if userID, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
        return userID
    }
    return uuid.Nil
}

func GetTenantID(ctx context.Context) uuid.UUID {
    if tenantID, ok := ctx.Value(TenantIDKey).(uuid.UUID); ok {
        return tenantID
    }
    return uuid.Nil
}

func GetUserRoles(ctx context.Context) []string {
    if roles, ok := ctx.Value(UserRolesKey).([]string); ok {
        return roles
    }
    return []string{}
}

func GetUserPermissions(ctx context.Context) []string {
    if permissions, ok := ctx.Value(PermissionsKey).([]string); ok {
        return permissions
    }
    return []string{}
}

func GetEmployeeID(ctx context.Context) uuid.UUID {
    if employeeID, ok := ctx.Value("employee_id").(uuid.UUID); ok {
        return employeeID
    }
    return uuid.Nil
}
```

### 3.2 响应格式标准化

#### 统一响应结构

```go
// internal/api/response/response.go
package response

import (
    "encoding/json"
    "net/http"
    "time"
)

// APIResponse 统一API响应结构
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     *APIError   `json:"error,omitempty"`
    Meta      *Meta       `json:"meta,omitempty"`
    Timestamp time.Time   `json:"timestamp"`
    RequestID string      `json:"request_id,omitempty"`
}

// APIError 错误信息结构
type APIError struct {
    Code    string      `json:"code"`
    Message string      `json:"message"`
    Details interface{} `json:"details,omitempty"`
    Field   string      `json:"field,omitempty"`
}

// Meta 元数据结构
type Meta struct {
    Pagination *PaginationMeta `json:"pagination,omitempty"`
    Timing     *TimingMeta     `json:"timing,omitempty"`
    Version    string          `json:"version,omitempty"`
}

type PaginationMeta struct {
    Total      int  `json:"total"`
    Page       int  `json:"page"`
    PerPage    int  `json:"per_page"`
    TotalPages int  `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}

type TimingMeta struct {
    ProcessingTime string `json:"processing_time"`
    DatabaseTime   string `json:"database_time,omitempty"`
}

// 成功响应函数
func OK(w http.ResponseWriter, data interface{}) {
    writeResponse(w, http.StatusOK, &APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now().UTC(),
    })
}

func Created(w http.ResponseWriter, data interface{}) {
    writeResponse(w, http.StatusCreated, &APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now().UTC(),
    })
}

func Accepted(w http.ResponseWriter, data interface{}) {
    writeResponse(w, http.StatusAccepted, &APIResponse{
        Success:   true,
        Data:      data,
        Timestamp: time.Now().UTC(),
    })
}

func NoContent(w http.ResponseWriter) {
    w.WriteHeader(http.StatusNoContent)
}

// 错误响应函数
func BadRequest(w http.ResponseWriter, message string, details interface{}) {
    writeErrorResponse(w, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func Unauthorized(w http.ResponseWriter, message string) {
    writeErrorResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(w http.ResponseWriter, message string) {
    writeErrorResponse(w, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(w http.ResponseWriter, message string) {
    writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func Conflict(w http.ResponseWriter, message string, details interface{}) {
    writeErrorResponse(w, http.StatusConflict, "CONFLICT", message, details)
}

func UnprocessableEntity(w http.ResponseWriter, message string, details interface{}) {
    writeErrorResponse(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", message, details)
}

func InternalServerError(w http.ResponseWriter, message string, err error) {
    var details interface{}
    if err != nil {
        details = err.Error()
    }
    writeErrorResponse(w, http.StatusInternalServerError, "INTERNAL_ERROR", message, details)
}

// 分页响应
func OKWithPagination(w http.ResponseWriter, data interface{}, pagination PaginationMeta) {
    writeResponse(w, http.StatusOK, &APIResponse{
        Success: true,
        Data:    data,
        Meta: &Meta{
            Pagination: &pagination,
        },
        Timestamp: time.Now().UTC(),
    })
}

// 内部辅助函数
func writeResponse(w http.ResponseWriter, statusCode int, response *APIResponse) {
    // 设置请求ID（如果存在）
    if requestID := w.Header().Get("X-Request-ID"); requestID != "" {
        response.RequestID = requestID
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
        http.Error(w, "Failed to encode response", http.StatusInternalServerError)
    }
}

func writeErrorResponse(w http.ResponseWriter, statusCode int, code, message string, details interface{}) {
    response := &APIResponse{
        Success: false,
        Error: &APIError{
            Code:    code,
            Message: message,
            Details: details,
        },
        Timestamp: time.Now().UTC(),
    }
    
    writeResponse(w, statusCode, response)
}

// 验证错误特殊处理
func ValidationError(w http.ResponseWriter, errors []ValidationFieldError) {
    writeResponse(w, http.StatusUnprocessableEntity, &APIResponse{
        Success: false,
        Error: &APIError{
            Code:    "VALIDATION_ERROR",
            Message: "Validation failed",
            Details: errors,
        },
        Timestamp: time.Now().UTC(),
    })
}

type ValidationFieldError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Value   interface{} `json:"value,omitempty"`
}
```

### 3.3 配置管理

#### 环境配置

```yaml
# config/employee-model.yaml
employee_model:
  api:
    version: "v1"
    base_path: "/api/v1"
    request_timeout: "30s"
    max_request_size: "10MB"
    
  database:
    # PostgreSQL配置
    postgres:
      host: "${DB_HOST:localhost}"
      port: "${DB_PORT:5432}"
      database: "${DB_NAME:cube_castle}"
      username: "${DB_USER:postgres}"
      password: "${DB_PASSWORD:}"
      ssl_mode: "${DB_SSL_MODE:disable}"
      max_open_conns: 25
      max_idle_conns: 5
      conn_max_lifetime: "5m"
      
    # Neo4j配置
    neo4j:
      uri: "${NEO4J_URI:bolt://localhost:7687}"
      username: "${NEO4J_USER:neo4j}"
      password: "${NEO4J_PASSWORD:}"
      max_connection_pool_size: 50
      connection_timeout: "30s"
      
  graphql:
    playground_enabled: "${GRAPHQL_PLAYGROUND:true}"
    introspection_enabled: "${GRAPHQL_INTROSPECTION:true}"
    query_cache_size: 1000
    query_complexity_limit: 200
    query_depth_limit: 10
    
  temporal:
    host_port: "${TEMPORAL_HOST_PORT:localhost:7233}"
    namespace: "${TEMPORAL_NAMESPACE:default}"
    task_queue: "employee-workflows"
    workflow_timeout: "1h"
    activity_timeout: "10m"
    
  intelligence:
    sam:
      confidence_threshold: 0.7
      max_entity_count: 10
      context_timeout_ms: 5000
      enable_debug_logging: "${SAM_DEBUG:false}"
      
    llm:
      provider: "${LLM_PROVIDER:openai}"
      model: "${LLM_MODEL:gpt-4}"
      max_tokens: 2000
      temperature: 0.3
      
  opa:
    embedded: true
    policy_dir: "./policies"
    decision_cache_size: 1000
    
  monitoring:
    metrics_enabled: true
    tracing_enabled: true
    jaeger_endpoint: "${JAEGER_ENDPOINT:http://localhost:14268/api/traces}"
    
  security:
    jwt_secret: "${JWT_SECRET:}"
    jwt_expiration: "24h"
    bcrypt_cost: 12
    
  cache:
    redis:
      host: "${REDIS_HOST:localhost}"
      port: "${REDIS_PORT:6379}"
      password: "${REDIS_PASSWORD:}"
      database: 0
      max_retries: 3
      
  logging:
    level: "${LOG_LEVEL:info}"
    format: "json"
    output: "stdout"
```

---

本API接口设计与集成规范为员工模型系统提供了完整的接口定义、GraphQL Schema和集成指南，确保与现有Cube Castle架构的无缝集成和高质量的开发者体验。