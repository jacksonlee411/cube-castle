package validation

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gaogu/cube-castle/go-app/generated/openapi"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
)

// TestEmployeeValidatorIntegration 测试员工验证器的集成功能
func TestEmployeeValidatorIntegration(t *testing.T) {
	// 创建Mock验证器
	mockChecker := NewMockValidationChecker()
	validator := NewEmployeeValidator(mockChecker, mockChecker, mockChecker, mockChecker)
	
	ctx := context.Background()
	tenantID := uuid.New()
	
	t.Run("ValidateCreateEmployee_Success", func(t *testing.T) {
		testValidateCreateEmployeeSuccess(t, validator, ctx, tenantID)
	})
	
	t.Run("ValidateCreateEmployee_ValidationErrors", func(t *testing.T) {
		testValidateCreateEmployeeErrors(t, validator, ctx, tenantID)
	})
	
	t.Run("ValidateUpdateEmployee_Success", func(t *testing.T) {
		testValidateUpdateEmployeeSuccess(t, validator, ctx, tenantID)
	})
	
	t.Run("ValidateUpdateEmployee_ValidationErrors", func(t *testing.T) {
		testValidateUpdateEmployeeErrors(t, validator, ctx, tenantID)
	})
	
	t.Run("ValidateListEmployeesParams", func(t *testing.T) {
		testValidateListEmployeesParams(t, validator)
	})
	
	t.Run("ValidateEmployeeTermination", func(t *testing.T) {
		testValidateEmployeeTermination(t, validator, ctx, tenantID)
	})
	
	t.Run("ValidateEmployeeStatusTransition", func(t *testing.T) {
		testValidateEmployeeStatusTransition(t, validator)
	})
}

func testValidateCreateEmployeeSuccess(t *testing.T, validator *EmployeeValidator, ctx context.Context, tenantID uuid.UUID) {
	t.Log("=== 测试有效的员工创建验证 ===")
	
	// 有效的创建请求
	validReq := &openapi.CreateEmployeeRequest{
		EmployeeNumber: "EMP001",
		FirstName:      "John",
		LastName:       "Doe",
		Email:          openapi_types.Email("john.doe@example.com"),
		HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
		PhoneNumber:    stringPtr("13800138000"),
	}
	
	err := validator.ValidateCreateEmployee(ctx, tenantID, validReq)
	assert.NoError(t, err, "有效的创建请求应该通过验证")
}

func testValidateCreateEmployeeErrors(t *testing.T, validator *EmployeeValidator, ctx context.Context, tenantID uuid.UUID) {
	t.Log("=== 测试无效的员工创建验证 ===")
	
	testCases := []struct {
		name        string
		req         *openapi.CreateEmployeeRequest
		expectError bool
		description string
	}{
		{
			name: "空员工编号",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "", // 空值
				FirstName:      "John",
				LastName:       "Doe",
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
			},
			expectError: true,
			description: "员工编号不能为空",
		},
		{
			name: "员工编号过长",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP123456789012345678901", // 超过20字符
				FirstName:      "John",
				LastName:       "Doe",
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
			},
			expectError: true,
			description: "员工编号不能超过20字符",
		},
		{
			name: "空名字",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP001",
				FirstName:      "", // 空值
				LastName:       "Doe",
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
			},
			expectError: true,
			description: "名字不能为空",
		},
		{
			name: "空姓氏",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP001",
				FirstName:      "John",
				LastName:       "", // 空值
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
			},
			expectError: true,
			description: "姓氏不能为空",
		},
		{
			name: "无效邮箱格式",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP001",
				FirstName:      "John",
				LastName:       "Doe",
				Email:          openapi_types.Email("invalid-email"), // 无效格式
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
			},
			expectError: true,
			description: "邮箱格式无效",
		},
		{
			name: "未来的入职日期",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP001",
				FirstName:      "John",
				LastName:       "Doe",
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(1, 0, 0)}, // 未来日期
			},
			expectError: true,
			description: "入职日期不能是未来日期",
		},
		{
			name: "无效电话号码格式",
			req: &openapi.CreateEmployeeRequest{
				EmployeeNumber: "EMP001",
				FirstName:      "John",
				LastName:       "Doe",
				Email:          openapi_types.Email("john.doe@example.com"),
				HireDate:       openapi_types.Date{Time: time.Now().AddDate(0, 0, -1)}, // 昨天
				PhoneNumber:    stringPtr("123"), // 太短
			},
			expectError: true,
			description: "电话号码格式无效",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateCreateEmployee(ctx, tenantID, tc.req)
			if tc.expectError {
				assert.Error(t, err, tc.description)
				
				// 检查是否是ValidationErrors类型
				if validationErrors, ok := err.(ValidationErrors); ok {
					assert.Greater(t, len(validationErrors.Errors), 0, "应该有具体的验证错误信息")
					t.Logf("验证错误: %+v", validationErrors.Errors)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func testValidateUpdateEmployeeSuccess(t *testing.T, validator *EmployeeValidator, ctx context.Context, tenantID uuid.UUID) {
	t.Log("=== 测试有效的员工更新验证 ===")
	
	employeeID := uuid.New()
	
	// 有效的更新请求
	validReq := &openapi.UpdateEmployeeRequest{
		FirstName:   stringPtr("Jane"),
		LastName:    stringPtr("Smith"),
		Email:       emailPtr("jane.smith@example.com"),
		PhoneNumber: stringPtr("13900139000"),
	}
	
	err := validator.ValidateUpdateEmployee(ctx, tenantID, employeeID, validReq)
	assert.NoError(t, err, "有效的更新请求应该通过验证")
}

func testValidateUpdateEmployeeErrors(t *testing.T, validator *EmployeeValidator, ctx context.Context, tenantID uuid.UUID) {
	t.Log("=== 测试无效的员工更新验证 ===")
	
	employeeID := uuid.New()
	
	testCases := []struct {
		name        string
		req         *openapi.UpdateEmployeeRequest
		expectError bool
		description string
	}{
		{
			name: "无效邮箱格式",
			req: &openapi.UpdateEmployeeRequest{
				Email: emailPtr("invalid-email"), // 无效格式
			},
			expectError: true,
			description: "更新的邮箱格式无效",
		},
		{
			name: "无效电话号码",
			req: &openapi.UpdateEmployeeRequest{
				PhoneNumber: stringPtr("123"), // 太短
			},
			expectError: true,
			description: "更新的电话号码格式无效",
		},
		{
			name: "空名字",
			req: &openapi.UpdateEmployeeRequest{
				FirstName: stringPtr(""), // 空值
			},
			expectError: true,
			description: "更新的名字不能为空",
		},
		{
			name: "空姓氏",
			req: &openapi.UpdateEmployeeRequest{
				LastName: stringPtr(""), // 空值
			},
			expectError: true,
			description: "更新的姓氏不能为空",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateUpdateEmployee(ctx, tenantID, employeeID, tc.req)
			if tc.expectError {
				assert.Error(t, err, tc.description)
				
				// 检查是否是ValidationErrors类型
				if validationErrors, ok := err.(ValidationErrors); ok {
					assert.Greater(t, len(validationErrors.Errors), 0, "应该有具体的验证错误信息")
					t.Logf("验证错误: %+v", validationErrors.Errors)
				}
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func testValidateListEmployeesParams(t *testing.T, validator *EmployeeValidator) {
	t.Log("=== 测试员工列表参数验证 ===")
	
	testCases := []struct {
		name        string
		page        int
		pageSize    int
		search      string
		expectError bool
		description string
	}{
		{
			name:        "有效参数",
			page:        1,
			pageSize:    20,
			search:      "张三",
			expectError: false,
			description: "有效的列表参数",
		},
		{
			name:        "页码为0",
			page:        0,
			pageSize:    20,
			search:      "",
			expectError: true,
			description: "页码不能为0",
		},
		{
			name:        "负页码",
			page:        -1,
			pageSize:    20,
			search:      "",
			expectError: true,
			description: "页码不能为负数",
		},
		{
			name:        "页面大小为0",
			page:        1,
			pageSize:    0,
			search:      "",
			expectError: true,
			description: "页面大小不能为0",
		},
		{
			name:        "页面大小过大",
			page:        1,
			pageSize:    101,
			search:      "",
			expectError: true,
			description: "页面大小不能超过100",
		},
		{
			name:        "搜索关键词过长",
			page:        1,
			pageSize:    20,
			search:      strings.Repeat("a", 101), // 超过100字符
			expectError: true,  
			description: "搜索关键词不能超过100字符",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateListEmployeesParams(tc.page, tc.pageSize, tc.search)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

func testValidateEmployeeTermination(t *testing.T, validator *EmployeeValidator, ctx context.Context, tenantID uuid.UUID) {
	t.Log("=== 测试员工离职验证 ===")
	
	employeeID := uuid.New()
	
	// Mock验证器总是允许删除，所以这里主要测试方法调用是否正常
	err := validator.ValidateEmployeeTermination(ctx, employeeID, tenantID)
	assert.NoError(t, err, "Mock验证器应该允许员工离职")
}

func testValidateEmployeeStatusTransition(t *testing.T, validator *EmployeeValidator) {
	t.Log("=== 测试员工状态转换验证 ===")
	
	testCases := []struct {
		name           string
		currentStatus  string
		newStatus      string
		expectError    bool
		description    string
	}{
		{
			name:          "活跃到非活跃",
			currentStatus: "active",
			newStatus:     "inactive",
			expectError:   false,
			description:   "从活跃状态转换到非活跃状态应该被允许",
		},
		{
			name:          "活跃到离职",
			currentStatus: "active",
			newStatus:     "terminated",
			expectError:   false,
			description:   "从活跃状态转换到离职状态应该被允许",
		},
		{
			name:          "离职到活跃",
			currentStatus: "terminated",
			newStatus:     "active",
			expectError:   true,
			description:   "从离职状态转换到活跃状态应该被拒绝",
		},
		{
			name:          "无效状态",
			currentStatus: "active",
			newStatus:     "invalid_status",
			expectError:   true,
			description:   "转换到无效状态应该被拒绝",
		},
		{
			name:          "相同状态",
			currentStatus: "active",
			newStatus:     "active",
			expectError:   true,
			description:   "转换到相同状态应该被拒绝",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validator.ValidateEmployeeStatusTransition(tc.currentStatus, tc.newStatus)
			if tc.expectError {
				assert.Error(t, err, tc.description)
			} else {
				assert.NoError(t, err, tc.description)
			}
		})
	}
}

// TestValidationErrorsStructure 测试验证错误结构
func TestValidationErrorsStructure(t *testing.T) {
	t.Log("=== 测试验证错误结构 ===")
	
	// 创建验证错误
	validationError := ValidationError{
		Field:   "employee_number",
		Message: "Employee number is required",
		Code:    "REQUIRED",
		Value:   "",
	}
	
	assert.Equal(t, "employee_number", validationError.Field)
	assert.Equal(t, "Employee number is required", validationError.Message)
	assert.Equal(t, "REQUIRED", validationError.Code)
	assert.Equal(t, "", validationError.Value)
	
	// 测试Error()方法
	expectedErrorMsg := "Validation failed for field 'employee_number': Employee number is required (code: REQUIRED)"
	assert.Equal(t, expectedErrorMsg, validationError.Error())
	
	// 创建多个验证错误
	validationErrors := ValidationErrors{
		Errors: []ValidationError{
			{
				Field:   "first_name",
				Message: "First name is required",
				Code:    "REQUIRED",
			},
			{
				Field:   "email",
				Message: "Invalid email format",
				Code:    "INVALID_FORMAT",
				Value:   "invalid-email",
			},
		},
	}
	
	assert.Equal(t, 2, len(validationErrors.Errors))
	assert.Contains(t, validationErrors.Error(), "first_name")
	assert.Contains(t, validationErrors.Error(), "email")
}

// TestMockValidationChecker 测试Mock验证检查器
func TestMockValidationChecker(t *testing.T) {
	t.Log("=== 测试Mock验证检查器 ===")
	
	checker := NewMockValidationChecker()
	ctx := context.Background()
	tenantID := uuid.New()
	
	// 测试员工编号存在检查
	exists, err := checker.IsEmployeeNumberExists(ctx, tenantID, "EMP001", nil)
	assert.NoError(t, err)
	assert.False(t, exists, "Mock检查器应该返回false（不存在）")
	
	// 测试邮箱存在检查
	exists, err = checker.IsEmailExists(ctx, tenantID, "test@example.com", nil)
	assert.NoError(t, err)
	assert.False(t, exists, "Mock检查器应该返回false（不存在）")
	
	// 测试组织存在检查
	orgID := uuid.New()
	exists, err = checker.IsOrganizationExists(ctx, tenantID, orgID)
	assert.NoError(t, err)
	assert.True(t, exists, "Mock检查器应该返回true（存在）")
	
	// 测试职位存在检查
	positionID := uuid.New()
	exists, err = checker.IsPositionExists(ctx, tenantID, positionID)
	assert.NoError(t, err)
	assert.True(t, exists, "Mock检查器应该返回true（存在）")
}

// 辅助函数
func stringPtr(s string) *string {
	return &s
}

func emailPtr(s string) *openapi_types.Email {
	email := openapi_types.Email(s)
	return &email
}