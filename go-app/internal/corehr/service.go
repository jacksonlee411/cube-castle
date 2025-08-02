package corehr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/events"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Service CoreHR 服务层
type Service struct {
	repo     *Repository
	eventBus events.EventBus
}

// NewService 创建新的 CoreHR 服务
func NewService(repo *Repository) *Service {
	return &Service{
		repo:     repo,
		eventBus: nil, // EventBus 将通过 SetEventBus 方法设置
	}
}

// NewServiceWithEventBus 创建带有EventBus的CoreHR服务
func NewServiceWithEventBus(repo *Repository, eventBus events.EventBus) *Service {
	return &Service{
		repo:     repo,
		eventBus: eventBus,
	}
}

// SetEventBus 设置EventBus（用于依赖注入）
func (s *Service) SetEventBus(eventBus events.EventBus) {
	s.eventBus = eventBus
}

// NewMockService 创建 Mock 服务（用于测试）
func NewMockService() *Service {
	return &Service{
		repo:     nil,
		eventBus: nil,
	}
}

// ListEmployees 获取员工列表
func (s *Service) ListEmployees(ctx context.Context, tenantID uuid.UUID, page, pageSize int, search string) (*openapi.EmployeeListResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	employees, totalCount, err := s.repo.ListEmployees(ctx, tenantID, page, pageSize, search)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	// 转换为OpenAPI响应格式
	openapiEmployees := make([]openapi.Employee, len(employees))
	for i, emp := range employees {
		openapiEmployees[i] = s.convertToOpenAPIEmployee(emp)
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := openapi.PaginationInfo{
		Page:       &page,
		PageSize:   &pageSize,
		TotalPages: &totalPages,
		HasNext:    &hasNext,
		HasPrev:    &hasPrev,
	}

	return &openapi.EmployeeListResponse{
		Employees:  &openapiEmployees,
		Pagination: &pagination,
		TotalCount: &totalCount,
	}, nil
}

// GetEmployee 根据 ID 获取员工
func (s *Service) GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*openapi.Employee, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	employee, err := s.repo.GetEmployeeByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}

	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}

	openapiEmployee := s.convertToOpenAPIEmployee(*employee)
	return &openapiEmployee, nil
}

// CreateEmployee 创建员工
func (s *Service) CreateEmployee(ctx context.Context, tenantID uuid.UUID, req *openapi.CreateEmployeeRequest) (*openapi.Employee, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	// 检查员工编号是否已存在
	existingEmployee, err := s.repo.GetEmployeeByNumber(ctx, tenantID, req.EmployeeNumber)
	if err != nil {
		// 如果是"no rows"错误，说明员工编号不存在，这是正常的
		if strings.Contains(err.Error(), "no rows") {
			// 员工编号不存在，可以继续创建
		} else {
			return nil, fmt.Errorf("failed to check employee number: %w", err)
		}
	} else if existingEmployee != nil {
		return nil, fmt.Errorf("employee number already exists")
	}

	// 创建员工实体
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EmployeeNumber: req.EmployeeNumber,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          string(req.Email),
		PhoneNumber:    req.PhoneNumber,
		Position:       nil, // 使用PositionId替代
		Department:     nil, // 使用OrganizationId替代
		HireDate:       req.HireDate.Time,
		ManagerID:      nil, // 暂时设为nil
		Status:         "active",
	}

	// 创建员工
	err = s.repo.CreateEmployee(ctx, employee)
	if err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	// 发布员工创建事件
	if s.eventBus != nil {
		event := events.NewEmployeeCreated(tenantID, employee.ID, employee.EmployeeNumber, 
			employee.FirstName, employee.LastName, employee.Email, employee.HireDate)
		
		if err := s.eventBus.Publish(ctx, event); err != nil {
			// 记录事件发布失败，但不阻止主流程
			// 在生产环境中应该有重试机制或者死信队列
			fmt.Printf("Failed to publish EmployeeCreated event: %v\n", err)
		}
	}

	openapiEmployee := s.convertToOpenAPIEmployee(*employee)
	return &openapiEmployee, nil
}

// UpdateEmployee 更新员工信息
func (s *Service) UpdateEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, req *openapi.UpdateEmployeeRequest) (*openapi.Employee, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	// 获取现有员工
	employee, err := s.repo.GetEmployeeByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}

	// Record the old phone number for potential event publishing
	var oldPhoneNumber string
	if employee.PhoneNumber != nil {
		oldPhoneNumber = *employee.PhoneNumber
	}

	// Record updated fields for event publishing
	updatedFields := make(map[string]interface{})

	// 更新字段
	if req.FirstName != nil {
		employee.FirstName = *req.FirstName
		updatedFields["first_name"] = *req.FirstName
	}
	if req.LastName != nil {
		employee.LastName = *req.LastName
		updatedFields["last_name"] = *req.LastName
	}
	if req.Email != nil {
		employee.Email = string(*req.Email)
		updatedFields["email"] = string(*req.Email)
	}
	if req.PhoneNumber != nil {
		employee.PhoneNumber = req.PhoneNumber
		updatedFields["phone_number"] = *req.PhoneNumber
	}
	if req.PositionId != nil {
		// 暂时设为nil，因为Position字段不存在
		updatedFields["position_id"] = req.PositionId.String()
	}
	if req.OrganizationId != nil {
		// 暂时设为nil，因为Department字段不存在
		updatedFields["organization_id"] = req.OrganizationId.String()
	}
	// ManagerId字段不存在，暂时跳过
	if req.Status != nil {
		employee.Status = string(*req.Status)
		updatedFields["status"] = string(*req.Status)
	}

	// 更新员工
	err = s.repo.UpdateEmployee(ctx, employee)
	if err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	// 发布员工更新事件
	if s.eventBus != nil && len(updatedFields) > 0 {
		event := events.NewEmployeeUpdated(tenantID, employee.ID, employee.EmployeeNumber, updatedFields)
		
		if err := s.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish EmployeeUpdated event: %v\n", err)
		}
		
		// 如果电话号码发生变化，发布专门的电话更新事件
		if newPhone, exists := updatedFields["phone_number"]; exists {
			phoneEvent := events.NewEmployeePhoneUpdated(tenantID, employee.ID, 
				employee.EmployeeNumber, oldPhoneNumber, newPhone.(string))
			
			if err := s.eventBus.Publish(ctx, phoneEvent); err != nil {
				fmt.Printf("Failed to publish EmployeePhoneUpdated event: %v\n", err)
			}
		}
	}

	openapiEmployee := s.convertToOpenAPIEmployee(*employee)
	return &openapiEmployee, nil
}

// DeleteEmployee 删除员工
func (s *Service) DeleteEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	if s.repo == nil {
		return fmt.Errorf("service not properly initialized: repository is nil")
	}

	// 获取员工信息用于事件创建
	employee, err := s.repo.GetEmployeeByID(ctx, tenantID, employeeID)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return fmt.Errorf("employee not found")
	}

	// 删除员工
	err = s.repo.DeleteEmployee(ctx, tenantID, employeeID)
	if err != nil {
		return fmt.Errorf("failed to delete employee: %w", err)
	}

	// 发布员工删除事件
	if s.eventBus != nil {
		event := events.NewEmployeeDeleted(tenantID, employee.ID, employee.EmployeeNumber)
		
		if err := s.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish EmployeeDeleted event: %v\n", err)
		}
	}

	return nil
}

// ListOrganizations 获取组织列表
func (s *Service) ListOrganizations(ctx context.Context, tenantID uuid.UUID) (*openapi.OrganizationListResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	organizations, err := s.repo.ListOrganizations(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}

	// 转换为OpenAPI响应格式
	openapiOrganizations := make([]openapi.Organization, len(organizations))
	for i, org := range organizations {
		openapiOrganizations[i] = s.convertToOpenAPIOrganization(org)
	}

	totalCount := len(openapiOrganizations)
	return &openapi.OrganizationListResponse{
		Organizations: &openapiOrganizations,
		TotalCount:    &totalCount,
	}, nil
}

// GetOrganizationTree 获取组织树
func (s *Service) GetOrganizationTree(ctx context.Context, tenantID uuid.UUID) (*openapi.OrganizationTreeResponse, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	orgTrees, err := s.repo.GetOrganizationTree(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization tree: %w", err)
	}

	// 转换为OpenAPI响应格式
	openapiOrgTrees := make([]openapi.OrganizationTreeNode, len(orgTrees))
	for i, tree := range orgTrees {
		openapiOrgTrees[i] = s.convertToOpenAPITreeNode(tree)
	}

	return &openapi.OrganizationTreeResponse{
		Tree: &openapiOrgTrees,
	}, nil
}

// CreateOrganization 创建组织
func (s *Service) CreateOrganization(ctx context.Context, tenantID uuid.UUID, name, code string, parentID *uuid.UUID) (*openapi.Organization, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	// 创建组织实体
	org := &Organization{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     name,
		Code:     code,
		ParentID: parentID,
		Level:    1, // 默认层级
		Status:   "active",
	}

	// 创建组织
	err := s.repo.CreateOrganization(ctx, org)
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// 发布组织创建事件
	if s.eventBus != nil {
		event := events.NewOrganizationCreated(tenantID, org.ID, org.Name, org.Code, org.ParentID, org.Level)
		
		if err := s.eventBus.Publish(ctx, event); err != nil {
			fmt.Printf("Failed to publish OrganizationCreated event: %v\n", err)
		}
	}

	openapiOrg := s.convertToOpenAPIOrganization(*org)
	return &openapiOrg, nil
}

// GetManagerByEmployeeId 根据员工ID获取经理
func (s *Service) GetManagerByEmployeeId(ctx context.Context, tenantID, employeeID uuid.UUID) (*openapi.Employee, error) {
	if s.repo == nil {
		return nil, fmt.Errorf("service not properly initialized: repository is nil")
	}

	manager, err := s.repo.GetManagerByEmployeeID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get manager: %w", err)
	}

	if manager == nil {
		return nil, fmt.Errorf("manager not found")
	}

	openapiManager := s.convertToOpenAPIEmployee(*manager)
	return &openapiManager, nil
}

// 辅助方法：转换为OpenAPI格式

func (s *Service) convertToOpenAPIEmployee(emp Employee) openapi.Employee {
	status := openapi.EmployeeStatus(emp.Status)
	email := openapi_types.Email(emp.Email)
	hireDate := openapi_types.Date{Time: emp.HireDate}

	return openapi.Employee{
		Id:             &emp.ID,
		EmployeeNumber: emp.EmployeeNumber,
		FirstName:      emp.FirstName,
		LastName:       emp.LastName,
		Email:          email,
		HireDate:       hireDate,
		Status:         &status,
		CreatedAt:      &emp.CreatedAt,
		UpdatedAt:      &emp.UpdatedAt,
	}
}

func (s *Service) convertToOpenAPIOrganization(org Organization) openapi.Organization {
	status := openapi.OrganizationStatus(org.Status)
	return openapi.Organization{
		Id:        &org.ID,
		Name:      org.Name,
		Code:      org.Code,
		Level:     org.Level,
		Status:    &status,
		CreatedAt: &org.CreatedAt,
		UpdatedAt: &org.UpdatedAt,
	}
}

func (s *Service) convertToOpenAPITreeNode(tree OrganizationTree) openapi.OrganizationTreeNode {
	name := tree.Name
	code := tree.Code
	level := tree.Level

	node := openapi.OrganizationTreeNode{
		Id:    &tree.ID,
		Name:  &name,
		Code:  &code,
		Level: &level,
	}

	if len(tree.Children) > 0 {
		children := make([]openapi.OrganizationTreeNode, len(tree.Children))
		for i, child := range tree.Children {
			children[i] = s.convertToOpenAPITreeNode(child)
		}
		node.Children = &children
	}

	return node
}

// Mock 方法实现（保留用于测试和降级）

// listEmployeesMock 获取员工列表 (Mock 实现)
func (s *Service) listEmployeesMock(ctx context.Context, page, pageSize int, search string) (*openapi.EmployeeListResponse, error) {
	// Mock 数据
	status1 := openapi.EmployeeStatus("active")
	status2 := openapi.EmployeeStatus("active")
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	employees := []openapi.Employee{
		{
			Id:             &id1,
			EmployeeNumber: "EMP001",
			FirstName:      "张",
			LastName:       "三",
			Email:          openapi_types.Email("zhangsan@example.com"),
			HireDate:       openapi_types.Date{Time: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
			Status:         &status1,
			CreatedAt:      &now,
			UpdatedAt:      &now,
		},
		{
			Id:             &id2,
			EmployeeNumber: "EMP002",
			FirstName:      "李",
			LastName:       "四",
			Email:          openapi_types.Email("lisi@example.com"),
			HireDate:       openapi_types.Date{Time: time.Date(2023, 2, 20, 0, 0, 0, 0, time.UTC)},
			Status:         &status2,
			CreatedAt:      &now,
			UpdatedAt:      &now,
		},
	}

	// 如果有搜索关键词，过滤数据
	if search != "" {
		filteredEmployees := []openapi.Employee{}
		for _, emp := range employees {
			if emp.FirstName == search || emp.LastName == search || emp.EmployeeNumber == search {
				filteredEmployees = append(filteredEmployees, emp)
			}
		}
		employees = filteredEmployees
	}

	totalCount := len(employees)
	totalPages := (totalCount + pageSize - 1) / pageSize
	hasNext := page < totalPages
	hasPrev := page > 1

	pagination := openapi.PaginationInfo{
		Page:       &page,
		PageSize:   &pageSize,
		TotalPages: &totalPages,
		HasNext:    &hasNext,
		HasPrev:    &hasPrev,
	}

	return &openapi.EmployeeListResponse{
		Employees:  &employees,
		Pagination: &pagination,
		TotalCount: &totalCount,
	}, nil
}

// getEmployeeMock 根据 ID 获取员工 (Mock 实现)
func (s *Service) getEmployeeMock(ctx context.Context, employeeID uuid.UUID) (*openapi.Employee, error) {
	// Mock 数据
	status := openapi.EmployeeStatus("active")
	now := time.Now()

	employee := openapi.Employee{
		Id:             &employeeID,
		EmployeeNumber: "EMP001",
		FirstName:      "张",
		LastName:       "三",
		Email:          openapi_types.Email("zhangsan@example.com"),
		HireDate:       openapi_types.Date{Time: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
		Status:         &status,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	return &employee, nil
}

// createEmployeeMock 创建员工 (Mock 实现)
func (s *Service) createEmployeeMock(ctx context.Context, req *openapi.CreateEmployeeRequest) (*openapi.Employee, error) {
	status := openapi.EmployeeStatus("active")
	now := time.Now()
	id := uuid.New()

	employee := openapi.Employee{
		Id:             &id,
		EmployeeNumber: req.EmployeeNumber,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		HireDate:       openapi_types.Date{Time: req.HireDate.Time},
		Status:         &status,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	return &employee, nil
}

// updateEmployeeMock 更新员工 (Mock 实现)
func (s *Service) updateEmployeeMock(ctx context.Context, employeeID uuid.UUID, req *openapi.UpdateEmployeeRequest) (*openapi.Employee, error) {
	status := openapi.EmployeeStatus("active")
	now := time.Now()

	employee := openapi.Employee{
		Id:             &employeeID,
		EmployeeNumber: "EMP001",
		FirstName:      "张",
		LastName:       "三",
		Email:          openapi_types.Email("zhangsan@example.com"),
		HireDate:       openapi_types.Date{Time: time.Date(2023, 1, 15, 0, 0, 0, 0, time.UTC)},
		Status:         &status,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	return &employee, nil
}

// listOrganizationsMock 获取组织列表 (Mock 实现)
func (s *Service) listOrganizationsMock(ctx context.Context) (*openapi.OrganizationListResponse, error) {
	// Mock 数据
	status := openapi.OrganizationStatus("active")
	now := time.Now()
	id1 := uuid.New()
	id2 := uuid.New()

	organizations := []openapi.Organization{
		{
			Id:        &id1,
			Name:      "技术部",
			Code:      "TECH",
			Level:     1,
			Status:    &status,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
		{
			Id:        &id2,
			Name:      "人事部",
			Code:      "HR",
			Level:     1,
			Status:    &status,
			CreatedAt: &now,
			UpdatedAt: &now,
		},
	}

	totalCount := len(organizations)
	return &openapi.OrganizationListResponse{
		Organizations: &organizations,
		TotalCount:    &totalCount,
	}, nil
}

// getOrganizationTreeMock 获取组织树 (Mock 实现)
func (s *Service) getOrganizationTreeMock(ctx context.Context) (*openapi.OrganizationTreeResponse, error) {
	// Mock 数据
	id1 := uuid.New()
	id2 := uuid.New()
	id3 := uuid.New()
	id4 := uuid.New()

	name1 := "技术部"
	name2 := "人事部"
	name3 := "前端组"
	name4 := "后端组"
	code1 := "TECH"
	code2 := "HR"
	code3 := "FRONTEND"
	code4 := "BACKEND"
	level1 := 1
	level2 := 2

	tree := []openapi.OrganizationTreeNode{
		{
			Id:    &id1,
			Name:  &name1,
			Code:  &code1,
			Level: &level1,
			Children: &[]openapi.OrganizationTreeNode{
				{
					Id:    &id3,
					Name:  &name3,
					Code:  &code3,
					Level: &level2,
				},
				{
					Id:    &id4,
					Name:  &name4,
					Code:  &code4,
					Level: &level2,
				},
			},
		},
		{
			Id:    &id2,
			Name:  &name2,
			Code:  &code2,
			Level: &level1,
		},
	}

	return &openapi.OrganizationTreeResponse{
		Tree: &tree,
	}, nil
}

// createOrganizationMock 创建组织 (Mock 实现)
func (s *Service) createOrganizationMock(ctx context.Context, name, code string, parentID *uuid.UUID) (*openapi.Organization, error) {
	status := openapi.OrganizationStatus("active")
	now := time.Now()
	id := uuid.New()

	level := 1
	org := openapi.Organization{
		Id:        &id,
		Name:      name,
		Code:      code,
		Level:     level,
		Status:    &status,
		CreatedAt: &now,
		UpdatedAt: &now,
	}

	return &org, nil
}
