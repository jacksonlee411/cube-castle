package corehr

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRepository_EmployeeCRUD 测试员工CRUD操作
func TestRepository_EmployeeCRUD(t *testing.T) {
	// 跳过测试，如果没有数据库连接
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// 这里应该使用测试数据库连接
	// 为了简化，我们跳过实际的数据库测试
	t.Skip("Database tests require test database setup")

	ctx := context.Background()
	tenantID := uuid.New()

	// 创建测试员工
	employee := &Employee{
		ID:             uuid.New(),
		TenantID:       tenantID,
		EmployeeNumber: "TEST001",
		FirstName:      "测试",
		LastName:       "员工",
		Email:          "test@example.com",
		PhoneNumber:    stringPtr("13800138000"),
		Position:       stringPtr("软件工程师"),
		Department:     stringPtr("技术部"),
		HireDate:       time.Now(),
		Status:         "active",
	}

	// 测试创建员工
	t.Run("CreateEmployee", func(t *testing.T) {
		err := repo.CreateEmployee(ctx, employee)
		assert.NoError(t, err)
	})

	// 测试根据ID获取员工
	t.Run("GetEmployeeByID", func(t *testing.T) {
		found, err := repo.GetEmployeeByID(ctx, tenantID, employee.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, employee.EmployeeNumber, found.EmployeeNumber)
		assert.Equal(t, employee.FirstName, found.FirstName)
		assert.Equal(t, employee.LastName, found.LastName)
	})

	// 测试根据员工编号获取员工
	t.Run("GetEmployeeByNumber", func(t *testing.T) {
		found, err := repo.GetEmployeeByNumber(ctx, tenantID, employee.EmployeeNumber)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, employee.ID, found.ID)
	})

	// 测试更新员工
	t.Run("UpdateEmployee", func(t *testing.T) {
		employee.FirstName = "更新后的名字"
		err := repo.UpdateEmployee(ctx, employee)
		assert.NoError(t, err)

		// 验证更新
		found, err := repo.GetEmployeeByID(ctx, tenantID, employee.ID)
		assert.NoError(t, err)
		assert.Equal(t, "更新后的名字", found.FirstName)
	})

	// 测试删除员工
	t.Run("DeleteEmployee", func(t *testing.T) {
		err := repo.DeleteEmployee(ctx, tenantID, employee.ID)
		assert.NoError(t, err)

		// 验证删除
		found, err := repo.GetEmployeeByID(ctx, tenantID, employee.ID)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

// TestRepository_ListEmployees 测试员工列表查询
func TestRepository_ListEmployees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	t.Skip("Database tests require test database setup")

	ctx := context.Background()
	tenantID := uuid.New()

	// 创建测试员工
	employees := []*Employee{
		{
			ID:             uuid.New(),
			TenantID:       tenantID,
			EmployeeNumber: "TEST001",
			FirstName:      "张三",
			LastName:       "李",
			Email:          "zhangsan@example.com",
			HireDate:       time.Now(),
			Status:         "active",
		},
		{
			ID:             uuid.New(),
			TenantID:       tenantID,
			EmployeeNumber: "TEST002",
			FirstName:      "李四",
			LastName:       "王",
			Email:          "lisi@example.com",
			HireDate:       time.Now(),
			Status:         "active",
		},
	}

	// 创建员工
	for _, emp := range employees {
		err := repo.CreateEmployee(ctx, emp)
		require.NoError(t, err)
	}

	t.Run("ListEmployees_NoSearch", func(t *testing.T) {
		employees, count, err := repo.ListEmployees(ctx, tenantID, 1, 10, "")
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, count, 2)
		assert.Len(t, employees, 2)
	})

	t.Run("ListEmployees_WithSearch", func(t *testing.T) {
		employees, count, err := repo.ListEmployees(ctx, tenantID, 1, 10, "张三")
		assert.NoError(t, err)
		assert.Equal(t, 1, count)
		assert.Len(t, employees, 1)
		assert.Equal(t, "张三", employees[0].FirstName)
	})

	// 清理测试数据
	for _, emp := range employees {
		repo.DeleteEmployee(ctx, tenantID, emp.ID)
	}
}

// TestRepository_OrganizationCRUD 测试组织CRUD操作
func TestRepository_OrganizationCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	t.Skip("Database tests require test database setup")

	ctx := context.Background()
	tenantID := uuid.New()

	// 创建测试组织
	org := &Organization{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "测试部门",
		Code:     "TEST_DEPT",
		Level:    1,
		Status:   "active",
	}

	t.Run("CreateOrganization", func(t *testing.T) {
		err := repo.CreateOrganization(ctx, org)
		assert.NoError(t, err)
	})

	t.Run("GetOrganizationByID", func(t *testing.T) {
		found, err := repo.GetOrganizationByID(ctx, tenantID, org.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, org.Name, found.Name)
		assert.Equal(t, org.Code, found.Code)
	})

	t.Run("ListOrganizations", func(t *testing.T) {
		orgs, err := repo.ListOrganizations(ctx, tenantID)
		assert.NoError(t, err)
		assert.Len(t, orgs, 1)
		assert.Equal(t, org.Name, orgs[0].Name)
	})

	t.Run("UpdateOrganization", func(t *testing.T) {
		org.Name = "更新后的部门名"
		err := repo.UpdateOrganization(ctx, org)
		assert.NoError(t, err)

		found, err := repo.GetOrganizationByID(ctx, tenantID, org.ID)
		assert.NoError(t, err)
		assert.Equal(t, "更新后的部门名", found.Name)
	})

	t.Run("DeleteOrganization", func(t *testing.T) {
		err := repo.DeleteOrganization(ctx, tenantID, org.ID)
		assert.NoError(t, err)

		found, err := repo.GetOrganizationByID(ctx, tenantID, org.ID)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})
}

// TestRepository_GetOrganizationTree 测试组织树查询
func TestRepository_GetOrganizationTree(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	t.Skip("Database tests require test database setup")

	ctx := context.Background()
	tenantID := uuid.New()

	// 创建组织树结构
	parentOrg := &Organization{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "技术部",
		Code:     "TECH",
		Level:    1,
		Status:   "active",
	}

	childOrg := &Organization{
		ID:       uuid.New(),
		TenantID: tenantID,
		Name:     "前端组",
		Code:     "FRONTEND",
		ParentID: &parentOrg.ID,
		Level:    2,
		Status:   "active",
	}

	// 创建组织
	err := repo.CreateOrganization(ctx, parentOrg)
	require.NoError(t, err)
	err = repo.CreateOrganization(ctx, childOrg)
	require.NoError(t, err)

	t.Run("GetOrganizationTree", func(t *testing.T) {
		trees, err := repo.GetOrganizationTree(ctx, tenantID)
		assert.NoError(t, err)
		assert.Len(t, trees, 1) // 只有一个根节点

		root := trees[0]
		assert.Equal(t, parentOrg.Name, root.Name)
		assert.Len(t, root.Children, 1) // 有一个子节点

		child := root.Children[0]
		assert.Equal(t, childOrg.Name, child.Name)
		assert.Nil(t, child.Children) // 子节点没有子节点
	})

	// 清理测试数据
	repo.DeleteOrganization(ctx, tenantID, childOrg.ID)
	repo.DeleteOrganization(ctx, tenantID, parentOrg.ID)
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
} 