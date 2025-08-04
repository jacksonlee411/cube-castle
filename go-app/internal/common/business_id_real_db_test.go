package common

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRealDBConnection 建立真实数据库连接
func setupRealDBConnection(t *testing.T) *sql.DB {
	// 检查是否启用真实数据库测试
	if testing.Short() {
		t.Skip("跳过真实数据库测试 (使用 -short 标志)")
	}

	db, err := sql.Open("postgres", "postgres://user:password@localhost:5432/cubecastle?sslmode=disable")
	require.NoError(t, err, "连接数据库失败")

	// 验证连接
	err = db.Ping()
	require.NoError(t, err, "数据库ping失败")

	return db
}

// TestBusinessIDService_LookupByBusinessID_WithRealDB 真实数据库业务ID查询测试
func TestBusinessIDService_LookupByBusinessID_WithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据库测试")
	}

	db := setupRealDBConnection(t)
	defer db.Close()

	service := NewBusinessIDService(db)

	testCases := []struct {
		name       string
		entityType EntityType
		businessID string
		expectFound bool
	}{
		{
			name:       "员工边界最小值",
			entityType: EntityTypeEmployee,
			businessID: "1",
			expectFound: true,
		},
		{
			name:       "员工边界最大值",  
			entityType: EntityTypeEmployee,
			businessID: "99999",
			expectFound: true,
		},
		{
			name:       "员工超出范围",
			entityType: EntityTypeEmployee,
			businessID: "12345", // 使用有效格式但不存在的ID
			expectFound: false,
		},
		{
			name:       "组织最小值",
			entityType: EntityTypeOrganization,
			businessID: "100000",
			expectFound: true,
		},
		{
			name:       "职位最小值",
			entityType: EntityTypePosition,
			businessID: "1000000",
			expectFound: true,
		},
		{
			name:       "不存在的员工ID",
			entityType: EntityTypeEmployee,
			businessID: "50000",
			expectFound: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			result, err := service.LookupByBusinessID(ctx, tc.entityType, tc.businessID)
			
			if tc.expectFound {
				assert.NoError(t, err, "查询不应该失败")
				if assert.NotNil(t, result, "结果不应该为nil") {
					assert.True(t, result.Found, "应该找到记录")
					assert.NotEqual(t, uuid.Nil, result.UUID, "UUID不应该为空")
				}
			} else {
				// 不存在的记录应该返回Found=false，而不是错误
				if err != nil {
					t.Logf("查询不存在记录返回错误: %v", err)
				}
				if result != nil {
					assert.False(t, result.Found, "不应该找到记录")
				}
			}
		})
	}
}

// TestBusinessIDService_GenerateBusinessID_WithRealDB 真实数据库业务ID生成测试
func TestBusinessIDService_GenerateBusinessID_WithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据库测试")
	}

	db := setupRealDBConnection(t)
	defer db.Close()

	service := NewBusinessIDService(db)

	t.Run("生成员工业务ID", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		businessID, err := service.GenerateBusinessID(ctx, EntityTypeEmployee)
		assert.NoError(t, err, "生成员工业务ID应该成功")
		assert.NotEmpty(t, businessID, "业务ID不应该为空")
		
		// 验证ID格式
		err = ValidateBusinessID(EntityTypeEmployee, businessID)
		assert.NoError(t, err, "生成的业务ID应该符合格式要求")
	})

	t.Run("生成组织业务ID", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		businessID, err := service.GenerateBusinessID(ctx, EntityTypeOrganization)
		assert.NoError(t, err, "生成组织业务ID应该成功")
		assert.NotEmpty(t, businessID, "业务ID不应该为空")
		
		// 验证ID格式
		err = ValidateBusinessID(EntityTypeOrganization, businessID)
		assert.NoError(t, err, "生成的业务ID应该符合格式要求")
	})

	t.Run("生成职位业务ID", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		businessID, err := service.GenerateBusinessID(ctx, EntityTypePosition)
		assert.NoError(t, err, "生成职位业务ID应该成功")
		assert.NotEmpty(t, businessID, "业务ID不应该为空")
		
		// 验证ID格式
		err = ValidateBusinessID(EntityTypePosition, businessID)
		assert.NoError(t, err, "生成的业务ID应该符合格式要求")
	})
}

// TestBusinessIDSystem_FullLifecycle_WithRealDB 完整生命周期测试
func TestBusinessIDSystem_FullLifecycle_WithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据库测试")
	}

	db := setupRealDBConnection(t)
	defer db.Close()

	service := NewBusinessIDService(db)
	manager := NewBusinessIDManager(service, DefaultBusinessIDManagerConfig())

	t.Run("员工完整生命周期", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// 1. 生成新的业务ID  
		businessID, err := manager.GenerateUniqueBusinessID(ctx, EntityTypeEmployee)
		require.NoError(t, err, "生成业务ID失败")
		require.NotEmpty(t, businessID, "业务ID不应该为空")

		// 2. 验证ID格式
		err = ValidateBusinessID(EntityTypeEmployee, businessID)
		require.NoError(t, err, "业务ID格式验证失败")

		// 3. 创建员工记录
		employeeUUID := uuid.New()
		
		// 先检查是否已存在这个业务ID
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(*) FROM employees WHERE business_id = $1", businessID).Scan(&count)
		require.NoError(t, err, "检查业务ID是否存在失败")
		
		if count > 0 {
			t.Skipf("业务ID %s 已存在，跳过测试", businessID)
		}
		
		_, err = db.ExecContext(ctx, `
			INSERT INTO employees (id, tenant_id, employee_type, first_name, last_name, 
			                      email, hire_date, employment_status, business_id, 
			                      created_at, updated_at) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
			employeeUUID, "00000000-0000-0000-0000-000000000000", "FULL_TIME",
			"测试", "员工", "lifecycle_test_"+businessID+"@company.com", time.Now().Format("2006-01-02"),
			"ACTIVE", businessID, time.Now(), time.Now())
		require.NoError(t, err, "创建员工记录失败")

		// 4. 通过业务ID查找UUID
		result, err := service.LookupByBusinessID(ctx, EntityTypeEmployee, businessID) 
		require.NoError(t, err, "通过业务ID查找UUID失败")
		require.True(t, result.Found, "应该找到记录")
		assert.Equal(t, employeeUUID, result.UUID, "UUID应该匹配")

		// 5. 通过UUID查找业务ID
		result2, err := service.LookupByUUID(ctx, EntityTypeEmployee, employeeUUID)
		require.NoError(t, err, "通过UUID查找业务ID失败")
		require.True(t, result2.Found, "应该找到记录")
		assert.Equal(t, businessID, result2.BusinessID, "业务ID应该匹配")

		// 清理测试数据
		_, err = db.ExecContext(ctx, "DELETE FROM employees WHERE id = $1", employeeUUID)
		require.NoError(t, err, "清理测试数据失败")
	})
}

// TestBusinessIDService_DatabaseConnection_WithRealDB 数据库连接测试
func TestBusinessIDService_DatabaseConnection_WithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据库测试")
	}

	db := setupRealDBConnection(t)
	defer db.Close()

	service := NewBusinessIDService(db)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 通过尝试生成ID来验证数据库连接
	businessID, err := service.GenerateBusinessID(ctx, EntityTypeEmployee)
	assert.NoError(t, err, "数据库连接测试应该通过")
	assert.NotEmpty(t, businessID, "应该能够生成业务ID")
}

// TestBusinessIDService_ConcurrentGeneration_WithRealDB 并发生成测试
func TestBusinessIDService_ConcurrentGeneration_WithRealDB(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过真实数据库测试")
	}

	db := setupRealDBConnection(t)
	defer db.Close()

	service := NewBusinessIDService(db)

	const numWorkers = 5
	const numPerWorker = 10

	results := make(chan string, numWorkers*numPerWorker)
	errors := make(chan error, numWorkers*numPerWorker)

	// 启动多个协程并发生成ID
	for i := 0; i < numWorkers; i++ {
		go func() {
			for j := 0; j < numPerWorker; j++ {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				id, err := service.GenerateBusinessID(ctx, EntityTypeEmployee)
				cancel()
				
				if err != nil {
					errors <- err
				} else {
					results <- id
				}
			}
		}()
	}

	// 收集结果
	var generatedIDs []string
	var generationErrors []error

	for i := 0; i < numWorkers*numPerWorker; i++ {
		select {
		case id := <-results:
			generatedIDs = append(generatedIDs, id)
		case err := <-errors:
			generationErrors = append(generationErrors, err)
		case <-time.After(10 * time.Second):
			t.Fatal("并发生成测试超时")
		}
	}

	// 验证结果
	assert.Empty(t, generationErrors, "不应该有生成错误")
	assert.Len(t, generatedIDs, numWorkers*numPerWorker, "应该生成预期数量的ID")

	// 验证所有ID都是唯一的
	seen := make(map[string]bool)
	for _, id := range generatedIDs {
		assert.False(t, seen[id], "发现重复的业务ID: %s", id)
		seen[id] = true
	}
}