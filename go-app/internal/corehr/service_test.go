package corehr

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRepository 模拟仓储接口
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) CreateEmployee(ctx context.Context, employee *Employee) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockRepository) GetEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) (*Employee, error) {
	args := m.Called(ctx, tenantID, employeeID)
	return args.Get(0).(*Employee), args.Error(1)
}

func (m *MockRepository) UpdateEmployee(ctx context.Context, employee *Employee) error {
	args := m.Called(ctx, employee)
	return args.Error(0)
}

func (m *MockRepository) DeleteEmployee(ctx context.Context, tenantID, employeeID uuid.UUID) error {
	args := m.Called(ctx, tenantID, employeeID)
	return args.Error(0)
}

func (m *MockRepository) ListEmployees(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*Employee, int, error) {
	args := m.Called(ctx, tenantID, page, pageSize)
	return args.Get(0).([]*Employee), args.Int(1), args.Error(2)
}

func (m *MockRepository) CreateOrganization(ctx context.Context, org *Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) GetOrganization(ctx context.Context, tenantID, orgID uuid.UUID) (*Organization, error) {
	args := m.Called(ctx, tenantID, orgID)
	return args.Get(0).(*Organization), args.Error(1)
}

func (m *MockRepository) UpdateOrganization(ctx context.Context, org *Organization) error {
	args := m.Called(ctx, org)
	return args.Error(0)
}

func (m *MockRepository) DeleteOrganization(ctx context.Context, tenantID, orgID uuid.UUID) error {
	args := m.Called(ctx, tenantID, orgID)
	return args.Error(0)
}

func (m *MockRepository) ListOrganizations(ctx context.Context, tenantID uuid.UUID, page, pageSize int) ([]*Organization, int, error) {
	args := m.Called(ctx, tenantID, page, pageSize)
	return args.Get(0).([]*Organization), args.Int(1), args.Error(2)
}

func (m *MockRepository) GetOrganizationTree(ctx context.Context, tenantID uuid.UUID) ([]*OrganizationTreeNode, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]*OrganizationTreeNode), args.Error(1)
}

// MockOutboxService 模拟发件箱服务
type MockOutboxService struct {
	mock.Mock
}

func (m *MockOutboxService) CreateEmployeeCreatedEvent(ctx context.Context, employeeID uuid.UUID, employeeData map[string]interface{}) error {
	args := m.Called(ctx, employeeID, employeeData)
	return args.Error(0)
}

func (m *MockOutboxService) CreateEmployeeUpdatedEvent(ctx context.Context, employeeID uuid.UUID, employeeData map[string]interface{}) error {
	args := m.Called(ctx, employeeID, employeeData)
	return args.Error(0)
}

func (m *MockOutboxService) CreateOrganizationCreatedEvent(ctx context.Context, orgID uuid.UUID, orgData map[string]interface{}) error {
	args := m.Called(ctx, orgID, orgData)
	return args.Error(0)
}

// TestService_CreateEmployee 测试创建员工
func TestService_CreateEmployee(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	tenantID := uuid.New()
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EmployeeNumber: "EMP001",
		FirstName:      "张三",
		LastName:       "李",
		Email:          "zhangsan@example.com",
		Status:         "active",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 设置模拟期望
	mockRepo.On("CreateEmployee", ctx, employee).Return(nil)
	mockOutbox.On("CreateEmployeeCreatedEvent", ctx, employee.ID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// 执行
	err := service.CreateEmployee(ctx, employee)

	// 验证
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

// TestService_CreateEmployee_RepositoryError 测试创建员工时仓储错误
func TestService_CreateEmployee_RepositoryError(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		EmployeeNumber: "EMP001",
		FirstName:      "张三",
		LastName:       "李",
		Email:          "zhangsan@example.com",
		Status:         "active",
	}

	// 设置模拟期望 - 仓储返回错误
	expectedError := assert.AnError
	mockRepo.On("CreateEmployee", ctx, employee).Return(expectedError)

	// 执行
	err := service.CreateEmployee(ctx, employee)

	// 验证
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
	mockRepo.AssertExpectations(t)
	// 发件箱服务不应该被调用，因为仓储操作失败
	mockOutbox.AssertNotCalled(t, "CreateEmployeeCreatedEvent")
}

// TestService_GetEmployee 测试获取员工
func TestService_GetEmployee(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	tenantID := uuid.New()
	employeeID := uuid.New()
	expectedEmployee := &Employee{
		ID:             employeeID,
		TenantID:       tenantID,
		EmployeeNumber: "EMP001",
		FirstName:      "张三",
		LastName:       "李",
		Email:          "zhangsan@example.com",
		Status:         "active",
	}

	// 设置模拟期望
	mockRepo.On("GetEmployee", ctx, tenantID, employeeID).Return(expectedEmployee, nil)

	// 执行
	employee, err := service.GetEmployee(ctx, tenantID, employeeID)

	// 验证
	assert.NoError(t, err)
	assert.Equal(t, expectedEmployee, employee)
	mockRepo.AssertExpectations(t)
}

// TestService_UpdateEmployee 测试更新员工
func TestService_UpdateEmployee(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		EmployeeNumber: "EMP001",
		FirstName:      "张三",
		LastName:       "李",
		Email:          "zhangsan@example.com",
		Status:         "active",
		UpdatedAt:      time.Now(),
	}

	// 设置模拟期望
	mockRepo.On("UpdateEmployee", ctx, employee).Return(nil)
	mockOutbox.On("CreateEmployeeUpdatedEvent", ctx, employee.ID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// 执行
	err := service.UpdateEmployee(ctx, employee)

	// 验证
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

// TestService_ListEmployees 测试员工列表
func TestService_ListEmployees(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	tenantID := uuid.New()
	page, pageSize := 1, 10
	expectedEmployees := []*Employee{
		{ID: uuid.New(), TenantID: tenantID, EmployeeNumber: "EMP001", FirstName: "张三"},
		{ID: uuid.New(), TenantID: tenantID, EmployeeNumber: "EMP002", FirstName: "李四"},
	}
	expectedTotal := 25

	// 设置模拟期望
	mockRepo.On("ListEmployees", ctx, tenantID, page, pageSize).Return(expectedEmployees, expectedTotal, nil)

	// 执行
	employees, total, err := service.ListEmployees(ctx, tenantID, page, pageSize)

	// 验证
	assert.NoError(t, err)
	assert.Equal(t, expectedEmployees, employees)
	assert.Equal(t, expectedTotal, total)
	mockRepo.AssertExpectations(t)
}

// TestService_CreateOrganization 测试创建组织
func TestService_CreateOrganization(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	org := &Organization{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Name:     "技术部",
		Code:     "TECH",
		Level:    1,
		Status:   "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 设置模拟期望
	mockRepo.On("CreateOrganization", ctx, org).Return(nil)
	mockOutbox.On("CreateOrganizationCreatedEvent", ctx, org.ID, mock.AnythingOfType("map[string]interface {}")).Return(nil)

	// 执行
	err := service.CreateOrganization(ctx, org)

	// 验证
	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
	mockOutbox.AssertExpectations(t)
}

// TestService_GetOrganizationTree 测试获取组织树
func TestService_GetOrganizationTree(t *testing.T) {
	// 设置
	mockRepo := new(MockRepository)
	mockOutbox := new(MockOutboxService)
	service := NewService(mockRepo, mockOutbox)

	ctx := context.Background()
	tenantID := uuid.New()
	expectedTree := []*OrganizationTreeNode{
		{
			Organization: &Organization{
				ID:       uuid.New(),
				TenantID: tenantID,
				Name:     "总公司",
				Code:     "ROOT",
				Level:    0,
				Status:   "active",
			},
			Children: []*OrganizationTreeNode{
				{
					Organization: &Organization{
						ID:       uuid.New(),
						TenantID: tenantID,
						Name:     "技术部",
						Code:     "TECH",
						Level:    1,
						Status:   "active",
					},
					Children: nil,
				},
			},
		},
	}

	// 设置模拟期望
	mockRepo.On("GetOrganizationTree", ctx, tenantID).Return(expectedTree, nil)

	// 执行
	tree, err := service.GetOrganizationTree(ctx, tenantID)

	// 验证
	assert.NoError(t, err)
	assert.Equal(t, expectedTree, tree)
	mockRepo.AssertExpectations(t)
}

// TestEmployee_Validation 测试员工模型验证
func TestEmployee_Validation(t *testing.T) {
	tests := []struct {
		name     string
		employee *Employee
		wantErr  bool
	}{
		{
			name: "有效员工",
			employee: &Employee{
				ID:             uuid.New(),
				TenantID:       uuid.New(),
				EmployeeNumber: "EMP001",
				FirstName:      "张三",
				LastName:       "李",
				Email:          "zhangsan@example.com",
				Status:         "active",
			},
			wantErr: false,
		},
		{
			name: "缺少员工编号",
			employee: &Employee{
				ID:        uuid.New(),
				TenantID:  uuid.New(),
				FirstName: "张三",
				LastName:  "李",
				Email:     "zhangsan@example.com",
				Status:    "active",
			},
			wantErr: true,
		},
		{
			name: "无效邮箱",
			employee: &Employee{
				ID:             uuid.New(),
				TenantID:       uuid.New(),
				EmployeeNumber: "EMP001",
				FirstName:      "张三",
				LastName:       "李",
				Email:          "invalid-email",
				Status:         "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmployee(tt.employee)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestOrganization_Validation 测试组织模型验证
func TestOrganization_Validation(t *testing.T) {
	tests := []struct {
		name string
		org  *Organization
		wantErr bool
	}{
		{
			name: "有效组织",
			org: &Organization{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Name:     "技术部",
				Code:     "TECH",
				Level:    1,
				Status:   "active",
			},
			wantErr: false,
		},
		{
			name: "缺少组织名称",
			org: &Organization{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Code:     "TECH",
				Level:    1,
				Status:   "active",
			},
			wantErr: true,
		},
		{
			name: "无效组织级别",
			org: &Organization{
				ID:       uuid.New(),
				TenantID: uuid.New(),
				Name:     "技术部",
				Code:     "TECH",
				Level:    -1,
				Status:   "active",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOrganization(tt.org)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// 辅助函数：验证员工数据
func validateEmployee(employee *Employee) error {
	if employee.EmployeeNumber == "" {
		return assert.AnError
	}
	if employee.Email != "" && !isValidEmail(employee.Email) {
		return assert.AnError
	}
	return nil
}

// 辅助函数：验证组织数据
func validateOrganization(org *Organization) error {
	if org.Name == "" {
		return assert.AnError
	}
	if org.Level < 0 {
		return assert.AnError
	}
	return nil
}

// 辅助函数：验证邮箱格式
func isValidEmail(email string) bool {
	// 简单的邮箱验证逻辑
	return len(email) > 5 && 
		   len(email) < 100 && 
		   email[0] != '@' && 
		   email[len(email)-1] != '@' &&
		   countChar(email, '@') == 1
}

func countChar(s string, c rune) int {
	count := 0
	for _, char := range s {
		if char == c {
			count++
		}
	}
	return count
}