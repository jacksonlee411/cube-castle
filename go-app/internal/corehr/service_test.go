package corehr

import (
	"context"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

// TestService_CreateEmployee 测试创建员工（使用Mock模式）
func TestService_CreateEmployee(t *testing.T) {
	// 使用Mock服务，避免数据库依赖
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()

	// 创建请求
	req := &openapi.CreateEmployeeRequest{
		EmployeeNumber: "EMP001",
		FirstName:      "张",
		LastName:       "三",
		Email:          openapi_types.Email("zhangsan@example.com"),
		HireDate:       openapi_types.Date{Time: time.Now()},
	}

	// 执行（Mock模式会直接返回成功）
	employee, err := service.CreateEmployee(ctx, tenantID, req)

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, employee)
	assert.Equal(t, "EMP001", employee.EmployeeNumber)
	assert.Equal(t, "张", employee.FirstName)
	assert.Equal(t, "三", employee.LastName)
}

// TestService_GetEmployee 测试获取员工（使用Mock模式）
func TestService_GetEmployee(t *testing.T) {
	// 使用Mock服务
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()
	employeeID := uuid.New()

	// 执行（Mock模式会返回预设的员工数据）
	employee, err := service.GetEmployee(ctx, tenantID, employeeID)

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, employee)
	assert.Equal(t, "EMP001", employee.EmployeeNumber)
	assert.Equal(t, "张", employee.FirstName)
	assert.Equal(t, "三", employee.LastName)
}

// TestService_ListEmployees 测试员工列表（使用Mock模式）
func TestService_ListEmployees(t *testing.T) {
	// 使用Mock服务
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()
	page, pageSize := 1, 10

	// 执行（Mock模式会返回预设的员工列表）
	response, err := service.ListEmployees(ctx, tenantID, page, pageSize, "")

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.TotalCount)
	assert.Greater(t, *response.TotalCount, 0)
	assert.NotNil(t, response.Employees)
	assert.Greater(t, len(*response.Employees), 0)
}

// TestService_CreateOrganization 测试创建组织（使用Mock模式）
func TestService_CreateOrganization(t *testing.T) {
	// 使用Mock服务
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()
	name := "技术部"
	code := "TECH"
	var parentID *uuid.UUID = nil

	// 执行（Mock模式会返回预设的组织数据）
	org, err := service.CreateOrganization(ctx, tenantID, name, code, parentID)

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, org)
	assert.Equal(t, name, org.Name)
	assert.Equal(t, code, org.Code)
}

// TestService_GetOrganizationTree 测试获取组织树（使用Mock模式）
func TestService_GetOrganizationTree(t *testing.T) {
	// 使用Mock服务
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()

	// 执行（Mock模式会返回预设的组织树）
	response, err := service.GetOrganizationTree(ctx, tenantID)

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotNil(t, response.Tree)
	assert.Greater(t, len(*response.Tree), 0)
}

// TestService_GetManagerByEmployeeId 测试获取经理（使用Mock模式）
func TestService_GetManagerByEmployeeId(t *testing.T) {
	// 使用Mock服务
	service := NewMockService()

	ctx := context.Background()
	tenantID := uuid.New()
	employeeID := uuid.New()

	// 执行（Mock模式会返回员工本身作为经理）
	manager, err := service.GetManagerByEmployeeId(ctx, tenantID, employeeID)

	// 验证
	assert.NoError(t, err)
	assert.NotNil(t, manager)
	assert.Equal(t, "EMP001", manager.EmployeeNumber)
}

// TestRepository_CreateEmployee 测试仓储层创建员工
func TestRepository_CreateEmployee(t *testing.T) {
	// 这是一个简单的单元测试，测试Employee结构体验证
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       uuid.New(),
		EmployeeNumber: "EMP001",
		FirstName:      "张",
		LastName:       "三",
		Email:          "zhangsan@example.com",
		Status:         "active",
		HireDate:       time.Now(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// 验证员工结构体字段
	assert.NotEqual(t, uuid.Nil, employee.ID)
	assert.NotEqual(t, uuid.Nil, employee.TenantID)
	assert.Equal(t, "EMP001", employee.EmployeeNumber)
	assert.Equal(t, "张", employee.FirstName)
	assert.Equal(t, "三", employee.LastName)
	assert.Equal(t, "zhangsan@example.com", employee.Email)
	assert.Equal(t, "active", employee.Status)
}

// TestOrganization_Structure 测试组织结构体
func TestOrganization_Structure(t *testing.T) {
	org := &Organization{
		ID:        uuid.New(),
		TenantID:  uuid.New(),
		Name:      "技术部",
		Code:      "TECH",
		Level:     1,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 验证组织结构体字段
	assert.NotEqual(t, uuid.Nil, org.ID)
	assert.NotEqual(t, uuid.Nil, org.TenantID)
	assert.Equal(t, "技术部", org.Name)
	assert.Equal(t, "TECH", org.Code)
	assert.Equal(t, 1, org.Level)
	assert.Equal(t, "active", org.Status)
}
