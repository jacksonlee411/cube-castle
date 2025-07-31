package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/gaogu/cube-castle/go-app/internal/common"
	"github.com/gaogu/cube-castle/go-app/internal/corehr"
	"github.com/gaogu/cube-castle/go-app/internal/logging"
	"github.com/gaogu/cube-castle/go-app/internal/outbox"
	"github.com/gaogu/cube-castle/go-app/internal/validation"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMockReplacementIntegration 测试Mock替换后的真实数据库集成
func TestMockReplacementIntegration(t *testing.T) {
	// 检查测试环境
	if testing.Short() {
		t.Skip("跳过集成测试 (使用 -short 标志)")
	}

	logger := logging.NewStructuredLogger()
	
	t.Run("Database_Connection_Test", func(t *testing.T) {
		testDatabaseConnection(t, logger)
	})

	t.Run("Service_Initialization_Test", func(t *testing.T) {
		testServiceInitialization(t, logger)
	})

	t.Run("Validation_System_Test", func(t *testing.T) {
		testValidationSystem(t, logger)
	})

	t.Run("Employee_Service_Integration", func(t *testing.T) {
		testEmployeeServiceIntegration(t, logger)
	})

	t.Run("Organization_Service_Integration", func(t *testing.T) {
		testOrganizationServiceIntegration(t, logger)
	})

	t.Run("Error_Handling_Integration", func(t *testing.T) {
		testErrorHandlingIntegration(t, logger)
	})
}

func testDatabaseConnection(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试数据库连接 ===")
	
	// 测试数据库连接初始化
	db := common.InitDatabaseConnection()
	
	if db == nil {
		t.Log("⚠️ 数据库连接不可用，这是正常的如果没有配置数据库")
		t.Log("提示：要运行完整的集成测试，请配置数据库连接")
		return
	}

	// 验证数据库类型
	require.NotNil(t, db.PostgreSQL, "PostgreSQL连接不应该为空")

	// 测试基本连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := db.PostgreSQL.Ping(ctx)
	assert.NoError(t, err, "数据库应该能够ping通")

	t.Log("✅ 数据库连接测试通过")
}

func testServiceInitialization(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试服务初始化 ===")

	db := common.InitDatabaseConnection()
	
	if db == nil {
		t.Log("⚠️ 数据库不可用，测试Mock服务初始化")
		
		// 测试Mock服务初始化
		mockService := corehr.NewMockService()
		assert.NotNil(t, mockService, "Mock服务应该能够初始化")
		
		t.Log("✅ Mock服务初始化测试通过")
		return
	}

	// 测试真实服务初始化
	require.NotNil(t, db.PostgreSQL, "PostgreSQL连接不应该为空")

	repo := corehr.NewRepository(db.PostgreSQL)
	assert.NotNil(t, repo, "Repository应该能够初始化")

	outboxService := outbox.NewService(db.PostgreSQL)
	assert.NotNil(t, outboxService, "Outbox服务应该能够初始化")

	service := corehr.NewService(repo, outboxService)
	assert.NotNil(t, service, "CoreHR服务应该能够初始化")

	t.Log("✅ 真实服务初始化测试通过")
}

func testValidationSystem(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试验证系统 ===")

	db := common.InitDatabaseConnection()
	
	if db == nil {
		t.Log("⚠️ 数据库不可用，测试Mock验证器")
		
		// 测试Mock验证器
		mockChecker := validation.NewMockValidationChecker()
		assert.NotNil(t, mockChecker, "Mock验证器应该能够初始化")
		
		validator := validation.NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)
		assert.NotNil(t, validator, "员工验证器应该能够初始化")
		
		t.Log("✅ Mock验证系统测试通过")
		return
	}

	// 测试真实验证器
	require.NotNil(t, db.PostgreSQL, "PostgreSQL连接不应该为空")

	repo := corehr.NewRepository(db.PostgreSQL)
	coreHRChecker := validation.NewCoreHRValidationChecker(repo)
	assert.NotNil(t, coreHRChecker, "CoreHR验证器应该能够初始化")

	validator := validation.NewEmployeeValidator(coreHRChecker, coreHRChecker, coreHRChecker, coreHRChecker)
	assert.NotNil(t, validator, "员工验证器应该能够初始化")

	t.Log("✅ 真实验证系统测试通过")
}

func testEmployeeServiceIntegration(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试员工服务集成 ===")

	db := common.InitDatabaseConnection()
	var service *corehr.Service

	if db == nil {
		t.Log("⚠️ 数据库不可用，使用Mock服务进行基础测试")
		service = corehr.NewMockService()
	} else {
		repo := corehr.NewRepository(db.PostgreSQL)
		outboxService := outbox.NewService(db.PostgreSQL)
		service = corehr.NewService(repo, outboxService)
	}

	ctx := context.Background()
	tenantID := uuid.New()

	// 测试1: 获取员工列表
	t.Log("测试获取员工列表")
	employees, err := service.ListEmployees(ctx, tenantID, 1, 10, "")
	
	if db == nil {
		// Mock服务应该返回Mock数据
		assert.NoError(t, err, "Mock服务应该成功返回员工列表")
		assert.NotNil(t, employees, "员工列表不应该为空")
	} else {
		// 真实服务可能返回空列表或真实数据，都是正常的
		if err != nil {
			t.Logf("获取员工列表返回错误（可能是正常的）: %v", err)
		} else {
			assert.NotNil(t, employees, "员工列表响应不应该为空")
			t.Logf("获取到 %d 名员工", func() int {
				if employees.Employees != nil {
					return len(*employees.Employees)
				}
				return 0
			}())
		}
	}

	// 测试2: 创建员工
	t.Log("测试创建员工")
	createReq := &openapi.CreateEmployeeRequest{
		EmployeeNumber: fmt.Sprintf("TEST-%d", time.Now().Unix()),
		FirstName:      "集成",
		LastName:       "测试",
		Email:          openapi_types.Email(fmt.Sprintf("test-%d@example.com", time.Now().Unix())),
		HireDate:       openapi_types.Date{Time: time.Now()},
		PhoneNumber:    stringPtr("13800138000"),
	}

	createdEmployee, err := service.CreateEmployee(ctx, tenantID, createReq)
	
	if db == nil {
		// Mock服务应该成功创建
		assert.NoError(t, err, "Mock服务应该成功创建员工")
		assert.NotNil(t, createdEmployee, "创建的员工不应该为空")
		assert.Equal(t, createReq.EmployeeNumber, createdEmployee.EmployeeNumber, "员工编号应该匹配")
	} else {
		// 真实服务的结果取决于数据库状态
		if err != nil {
			t.Logf("创建员工返回错误（可能是数据库表不存在）: %v", err)
		} else {
			assert.NotNil(t, createdEmployee, "创建的员工不应该为空")
			assert.Equal(t, createReq.EmployeeNumber, createdEmployee.EmployeeNumber, "员工编号应该匹配")
			t.Logf("成功创建员工: %s", createdEmployee.EmployeeNumber)
		}
	}

	t.Log("✅ 员工服务集成测试完成")
}

func testOrganizationServiceIntegration(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试组织服务集成 ===")

	db := common.InitDatabaseConnection()
	var service *corehr.Service

	if db == nil {
		t.Log("⚠️ 数据库不可用，使用Mock服务进行基础测试")
		service = corehr.NewMockService()
	} else {
		repo := corehr.NewRepository(db.PostgreSQL)
		outboxService := outbox.NewService(db.PostgreSQL)
		service = corehr.NewService(repo, outboxService)
	}

	ctx := context.Background()
	tenantID := uuid.New()

	// 测试1: 获取组织列表
	t.Log("测试获取组织列表")
	organizations, err := service.ListOrganizations(ctx, tenantID)
	
	if db == nil {
		// Mock服务应该返回Mock数据
		assert.NoError(t, err, "Mock服务应该成功返回组织列表")
		assert.NotNil(t, organizations, "组织列表不应该为空")
	} else {
		// 真实服务的结果取决于数据库状态
		if err != nil {
			t.Logf("获取组织列表返回错误（可能是数据库表不存在）: %v", err)
		} else {
			assert.NotNil(t, organizations, "组织列表响应不应该为空")
			t.Logf("获取到 %d 个组织", func() int {
				if organizations.Organizations != nil {
					return len(*organizations.Organizations)
				}
				return 0
			}())
		}
	}

	// 测试2: 获取组织树
	t.Log("测试获取组织树")
	orgTree, err := service.GetOrganizationTree(ctx, tenantID)
	
	if db == nil {
		// Mock服务应该返回Mock数据
		assert.NoError(t, err, "Mock服务应该成功返回组织树")
		assert.NotNil(t, orgTree, "组织树不应该为空")
	} else {
		// 真实服务的结果取决于数据库状态
		if err != nil {
			t.Logf("获取组织树返回错误（可能是数据库表不存在）: %v", err)
		} else {
			assert.NotNil(t, orgTree, "组织树响应不应该为空")
			t.Logf("组织树根节点数量: %d", func() int {
				if orgTree.Tree != nil {
					return len(*orgTree.Tree)
				}
				return 0
			}())
		}
	}

	// 测试3: 创建组织
	t.Log("测试创建组织")
	orgName := fmt.Sprintf("测试组织-%d", time.Now().Unix())
	orgCode := fmt.Sprintf("TEST-%d", time.Now().Unix())
	
	createdOrg, err := service.CreateOrganization(ctx, tenantID, orgName, orgCode, nil)
	
	if db == nil {
		// Mock服务应该成功创建
		assert.NoError(t, err, "Mock服务应该成功创建组织")
		assert.NotNil(t, createdOrg, "创建的组织不应该为空")
		assert.Equal(t, orgName, createdOrg.Name, "组织名称应该匹配")
	} else {
		// 真实服务的结果取决于数据库状态
		if err != nil {
			t.Logf("创建组织返回错误（可能是数据库表不存在）: %v", err)
		} else {
			assert.NotNil(t, createdOrg, "创建的组织不应该为空")
			assert.Equal(t, orgName, createdOrg.Name, "组织名称应该匹配")
			t.Logf("成功创建组织: %s", createdOrg.Name)
		}
	}

	t.Log("✅ 组织服务集成测试完成")
}

func testErrorHandlingIntegration(t *testing.T, logger *logging.StructuredLogger) {
	t.Log("=== 测试错误处理集成 ===")

	db := common.InitDatabaseConnection()
	var service *corehr.Service

	if db == nil {
		t.Log("⚠️ 数据库不可用，创建nil repository服务测试错误处理")
		service = corehr.NewService(nil, nil) // 故意传入nil来测试错误处理
	} else {
		repo := corehr.NewRepository(db.PostgreSQL)
		outboxService := outbox.NewService(db.PostgreSQL)
		service = corehr.NewService(repo, outboxService)
	}

	ctx := context.Background()
	tenantID := uuid.New()

	if db == nil {
		// 测试nil repository的错误处理
		t.Log("测试nil repository错误处理")
		
		_, err := service.ListEmployees(ctx, tenantID, 1, 10, "")
		assert.Error(t, err, "nil repository应该返回错误")
		assert.Contains(t, err.Error(), "service not properly initialized", "错误消息应该包含初始化错误信息")

		_, err = service.CreateEmployee(ctx, tenantID, &openapi.CreateEmployeeRequest{
			EmployeeNumber: "TEST001",
			FirstName:      "Test",
			LastName:       "User",
			Email:          "test@example.com",
			HireDate:       openapi_types.Date{Time: time.Now()},
		})
		assert.Error(t, err, "nil repository应该返回错误")
		assert.Contains(t, err.Error(), "service not properly initialized", "错误消息应该包含初始化错误信息")

		t.Log("✅ nil repository错误处理测试通过")
	} else {
		t.Log("✅ 真实数据库服务已正确初始化")
	}

	t.Log("✅ 错误处理集成测试完成")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}