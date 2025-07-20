package corehr

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/outbox"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Service CoreHR 服务层
type Service struct {
	repo    *Repository
	outbox  *outbox.Service
}

// NewService 创建新的 CoreHR 服务
func NewService(repo *Repository, outbox *outbox.Service) *Service {
	return &Service{repo: repo, outbox: outbox}
}

// NewMockService 创建 Mock 服务（用于测试）
func NewMockService() *Service {
	return &Service{repo: nil, outbox: nil}
}

// ListEmployees 获取员工列表
func (s *Service) ListEmployees(ctx context.Context, tenantID uuid.UUID, page, pageSize int, search string) (*openapi.EmployeeListResponse, error) {
	if s.repo == nil {
		return s.listEmployeesMock(ctx, page, pageSize, search)
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
		Employees:   &openapiEmployees,
		Pagination:  &pagination,
		TotalCount:  &totalCount,
	}, nil
}

// GetEmployee 根据 ID 获取员工
func (s *Service) GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*openapi.Employee, error) {
	if s.repo == nil {
		return s.getEmployeeMock(ctx, employeeID)
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
		return s.createEmployeeMock(ctx, req)
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

	// 创建员工创建事件
	if s.outbox != nil {
		employeeData := map[string]interface{}{
			"employee_number": employee.EmployeeNumber,
			"first_name":      employee.FirstName,
			"last_name":       employee.LastName,
			"email":           employee.Email,
			"position":        employee.Position,
			"department":      employee.Department,
			"hire_date":       employee.HireDate.Format(time.RFC3339),
		}
		
		err = s.outbox.CreateEmployeeCreatedEvent(ctx, employee.ID, employeeData)
		if err != nil {
			// 记录错误但不影响员工创建
			fmt.Printf("Warning: failed to create employee event: %v\n", err)
		}
	}

	openapiEmployee := s.convertToOpenAPIEmployee(*employee)
	return &openapiEmployee, nil
}

// UpdateEmployee 更新员工信息
func (s *Service) UpdateEmployee(ctx context.Context, tenantID, employeeID uuid.UUID, req *openapi.UpdateEmployeeRequest) (*openapi.Employee, error) {
	if s.repo == nil {
		return s.updateEmployeeMock(ctx, employeeID, req)
	}

	// 获取现有员工
	employee, err := s.repo.GetEmployeeByID(ctx, tenantID, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}

	// 记录更新的字段
	updatedFields := make(map[string]interface{})
	var oldPhoneNumber string
	if employee.PhoneNumber != nil {
		oldPhoneNumber = *employee.PhoneNumber
	}

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

	// 创建事件
	if s.outbox != nil {
		// 创建员工更新事件
		if len(updatedFields) > 0 {
			err = s.outbox.CreateEmployeeUpdatedEvent(ctx, employee.ID, updatedFields)
			if err != nil {
				// 记录错误但不影响员工更新
				fmt.Printf("Warning: failed to create employee updated event: %v\n", err)
			}
		}

		// 如果电话号码更新了，创建专门的电话更新事件
		if req.PhoneNumber != nil && oldPhoneNumber != *req.PhoneNumber {
			err = s.outbox.CreateEmployeePhoneUpdatedEvent(ctx, employee.ID, oldPhoneNumber, *req.PhoneNumber)
			if err != nil {
				// 记录错误但不影响员工更新
				fmt.Printf("Warning: failed to create phone updated event: %v\n", err)
			}
		}
	}

	openapiEmployee := s.convertToOpenAPIEmployee(*employee)
	return &openapiEmployee, nil
}

// DeleteEmployee 删除员工
func (s *Service) DeleteEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	if s.repo == nil {
		return nil // Mock模式直接返回成功
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

	// 创建员工删除事件
	if s.outbox != nil {
		payload, err := s.outbox.CreateEvent(ctx, &outbox.CreateEventRequest{
			AggregateID:   employeeID,
			AggregateType: outbox.AggregateTypeEmployee,
			EventType:     outbox.EventTypeEmployeeDeleted,
			EventVersion:  1,
			Payload:       []byte(fmt.Sprintf(`{"employee_id":"%s","deleted_at":"%s"}`, employeeID, time.Now().Format(time.RFC3339))),
		})
		if err != nil {
			// 记录错误但不影响员工删除
			fmt.Printf("Warning: failed to create employee deleted event: %v\n", err)
		}
		_ = payload // 避免未使用变量警告
	}

	return nil
}

// ListOrganizations 获取组织列表
func (s *Service) ListOrganizations(ctx context.Context, tenantID uuid.UUID) (*openapi.OrganizationListResponse, error) {
	if s.repo == nil {
		return s.listOrganizationsMock(ctx)
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
		return s.getOrganizationTreeMock(ctx)
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
		return s.createOrganizationMock(ctx, name, code, parentID)
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

	// 创建组织创建事件
	if s.outbox != nil {
		eventErr := s.outbox.CreateOrganizationCreatedEvent(ctx, org.ID, org.Name, org.Code, org.ParentID)
		if eventErr != nil {
			// 记录错误但不影响组织创建
			fmt.Printf("Warning: failed to create organization event: %v\n", eventErr)
		}
	}

	openapiOrg := s.convertToOpenAPIOrganization(*org)
	return &openapiOrg, nil
}

// GetManagerByEmployeeId 根据员工ID获取经理
func (s *Service) GetManagerByEmployeeId(ctx context.Context, tenantID, employeeID uuid.UUID) (*openapi.Employee, error) {
	if s.repo == nil {
		return s.getEmployeeMock(ctx, employeeID) // Mock模式返回员工本身
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
		Employees:   &employees,
		Pagination:  &pagination,
		TotalCount:  &totalCount,
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