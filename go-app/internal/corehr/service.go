package corehr

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Service CoreHR 服务层
type Service struct {
	repo *Repository
}

// NewService 创建新的 CoreHR 服务
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// NewMockService 创建 Mock 服务（用于测试）
func NewMockService() *Service {
	return &Service{repo: nil}
}

// ListEmployees 获取员工列表
func (s *Service) ListEmployees(ctx context.Context, page, pageSize int, search string) (*openapi.EmployeeListResponse, error) {
	// 使用 mock 数据
	return s.listEmployeesMock(ctx, page, pageSize, search)
}

// GetEmployee 根据 ID 获取员工
func (s *Service) GetEmployee(ctx context.Context, employeeID uuid.UUID) (*openapi.Employee, error) {
	// 使用 mock 数据
	return s.getEmployeeMock(ctx, employeeID)
}

// CreateEmployee 创建员工
func (s *Service) CreateEmployee(ctx context.Context, req *openapi.CreateEmployeeRequest) (*openapi.Employee, error) {
	// 使用 mock 数据
	status := openapi.EmployeeStatus("active")
	now := time.Now()
	id := uuid.New()
	
	employee := openapi.Employee{
		Id:             &id,
		EmployeeNumber: req.EmployeeNumber,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          req.Email,
		HireDate:       req.HireDate,
		Status:         &status,
		CreatedAt:      &now,
		UpdatedAt:      &now,
	}

	return &employee, nil
}

// UpdateEmployee 更新员工信息
func (s *Service) UpdateEmployee(ctx context.Context, employeeID uuid.UUID, req *openapi.UpdateEmployeeRequest) (*openapi.Employee, error) {
	// 使用 mock 数据
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

// DeleteEmployee 删除员工
func (s *Service) DeleteEmployee(ctx context.Context, employeeID uuid.UUID) error {
	// Mock 实现
	return nil
}

// ListOrganizations 获取组织列表
func (s *Service) ListOrganizations(ctx context.Context) (*openapi.OrganizationListResponse, error) {
	// 使用 mock 数据
	return s.listOrganizationsMock(ctx)
}

// GetOrganizationTree 获取组织树
func (s *Service) GetOrganizationTree(ctx context.Context) (*openapi.OrganizationTreeResponse, error) {
	// 使用 mock 数据
	return s.getOrganizationTreeMock(ctx)
}

// Mock 方法实现

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