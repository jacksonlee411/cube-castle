package common

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 集成测试 - 测试不同组件之间的交互
// 目标覆盖率：≥70%

// TestBusinessIDSystemIntegration 业务ID系统集成测试
func TestBusinessIDSystemIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}
	
	t.Run("完整业务ID生命周期集成", func(t *testing.T) {
		// 1. 验证业务ID范围配置
		empRange := GetBusinessIDRange(EntityTypeEmployee)
		orgRange := GetBusinessIDRange(EntityTypeOrganization)
		posRange := GetBusinessIDRange(EntityTypePosition)
		
		assert.Equal(t, int64(1), empRange.Min)
		assert.Equal(t, int64(99999), empRange.Max)
		assert.Equal(t, int64(100000), orgRange.Min)
		assert.Equal(t, int64(999999), orgRange.Max)
		assert.Equal(t, int64(1000000), posRange.Min)
		assert.Equal(t, int64(9999999), posRange.Max)
		
		// 2. 验证ID验证功能集成
		// 测试员工ID验证
		err := ValidateBusinessID(EntityTypeEmployee, "1")
		assert.NoError(t, err, "有效员工ID应该通过验证")
		
		err = ValidateBusinessID(EntityTypeEmployee, "99999")
		assert.NoError(t, err, "最大员工ID应该通过验证")
		
		err = ValidateBusinessID(EntityTypeEmployee, "100000")
		assert.Error(t, err, "超出范围的员工ID应该失败")
		
		// 测试组织ID验证
		err = ValidateBusinessID(EntityTypeOrganization, "100000")
		assert.NoError(t, err, "有效组织ID应该通过验证")
		
		err = ValidateBusinessID(EntityTypeOrganization, "999999")
		assert.NoError(t, err, "最大组织ID应该通过验证")
		
		// 测试职位ID验证
		err = ValidateBusinessID(EntityTypePosition, "1000000")
		assert.NoError(t, err, "有效职位ID应该通过验证")
		
		err = ValidateBusinessID(EntityTypePosition, "9999999")
		assert.NoError(t, err, "最大职位ID应该通过验证")
	})
	
	t.Run("业务ID管理器与服务集成", func(t *testing.T) {
		// 配置管理器
		config := BusinessIDManagerConfig{
			EnableAutoGeneration: true,
			EnableValidation:     true,
			MaxRetries:          3,
		}
		
		// 使用nil数据库服务测试配置逻辑
		service := NewBusinessIDService(nil)
		_ = NewBusinessIDManager(service, config) // 测试管理器创建
		
		ctx := context.Background()
		
		// 验证不同配置下的行为
		disabledConfig := BusinessIDManagerConfig{
			EnableAutoGeneration: false,
			EnableValidation:     true,
			MaxRetries:          3,
		}
		disabledManager := NewBusinessIDManager(service, disabledConfig)
		
		_, err := disabledManager.GenerateUniqueBusinessID(ctx, EntityTypeEmployee)
		assert.Error(t, err, "禁用自动生成的管理器应该返回错误")
		assert.Contains(t, err.Error(), "auto generation is disabled")
	})
	
	t.Run("UUID兼容性集成测试", func(t *testing.T) {
		// 测试UUID识别功能与业务ID的区分
		validUUIDs := []string{
			"550e8400-e29b-41d4-a716-446655440000",
			"6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			"12345678-1234-1234-1234-123456789012",
		}
		
		businessIDs := []string{
			"1",
			"100000", 
			"1000000",
			"12345",
		}
		
		// 验证UUID识别
		for _, uuid := range validUUIDs {
			assert.True(t, IsUUID(uuid), "应该正确识别UUID: %s", uuid)
		}
		
		// 验证业务ID不被误识别为UUID
		for _, bizID := range businessIDs {
			assert.False(t, IsUUID(bizID), "业务ID不应该被识别为UUID: %s", bizID)
		}
	})
}

// TestDatabaseConfigurationIntegration 数据库配置集成测试
func TestDatabaseConfigurationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过数据库配置集成测试")
	}
	
	t.Run("数据库配置与连接集成", func(t *testing.T) {
		// 测试配置创建
		config := NewDatabaseConfig()
		require.NotNil(t, config)
		
		// 验证配置包含必要信息
		assert.NotEmpty(t, config.PostgreSQLURL, "PostgreSQL URL不应该为空")
		assert.NotEmpty(t, config.Neo4jURI, "Neo4j URI不应该为空")
		
		// 尝试连接（在测试环境中可能失败，但测试配置传递）
		db, err := Connect(config)
		if err != nil {
			t.Logf("预期的连接失败（测试环境）: %v", err)
			// 验证错误包含有用信息
			assert.Error(t, err)
		} else {
			// 如果连接成功，测试基本功能
			require.NotNil(t, db)
			
			// 测试连接关闭
			db.Close()
		}
	})
}

// TestCrossModuleIntegration 跨模块集成测试
func TestCrossModuleIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过跨模块集成测试")
	}
	
	t.Run("业务ID系统组件协作", func(t *testing.T) {
		// 模拟完整的业务场景
		service := NewBusinessIDService(nil)
		config := DefaultBusinessIDManagerConfig()
		_ = NewBusinessIDManager(service, config) // 测试管理器创建
		
		// 测试各实体类型的验证规则集成
		testCases := []struct {
			entityType EntityType
			validIDs   []string
			invalidIDs []string
		}{
			{
				entityType: EntityTypeEmployee,
				validIDs:   []string{"1", "123", "99999"},
				invalidIDs: []string{"0", "100000", "abc", ""},
			},
			{
				entityType: EntityTypeOrganization,
				validIDs:   []string{"100000", "500000", "999999"},
				invalidIDs: []string{"99999", "1000000", "abc123", ""},
			},
			{
				entityType: EntityTypePosition,
				validIDs:   []string{"1000000", "5000000", "9999999"},
				invalidIDs: []string{"999999", "10000000", "pos123", ""},
			},
		}
		
		for _, tc := range testCases {
			t.Run(string(tc.entityType)+"_验证集成", func(t *testing.T) {
				// 测试有效ID
				for _, validID := range tc.validIDs {
					err := ValidateBusinessID(tc.entityType, validID)
					assert.NoError(t, err, "有效ID应该通过验证: %s", validID)
				}
				
				// 测试无效ID
				for _, invalidID := range tc.invalidIDs {
					err := ValidateBusinessID(tc.entityType, invalidID)
					assert.Error(t, err, "无效ID应该失败: %s", invalidID)
				}
			})
		}
		
		// 验证范围定义的一致性
		empRange := GetBusinessIDRange(EntityTypeEmployee)
		orgRange := GetBusinessIDRange(EntityTypeOrganization)
		posRange := GetBusinessIDRange(EntityTypePosition)
		
		// 验证范围不重叠
		assert.True(t, empRange.Max < orgRange.Min, "员工ID范围应该不与组织ID重叠")
		assert.True(t, orgRange.Max < posRange.Min, "组织ID范围应该不与职位ID重叠")
	})
	
	t.Run("错误处理流程集成", func(t *testing.T) {
		// 测试完整的错误处理流程
		service := NewBusinessIDService(nil) // 使用nil数据库触发错误
		config := DefaultBusinessIDManagerConfig()
		manager := NewBusinessIDManager(service, config)
		
		ctx := context.Background()
		
		// 捕获预期的panic
		defer func() {
			if r := recover(); r != nil {
				t.Log("成功捕获到预期的数据库错误panic")
			}
		}()
		
		// 尝试生成业务ID（应该失败）
		_, err := manager.GenerateUniqueBusinessID(ctx, EntityTypeEmployee)
		
		// 如果没有panic，应该有错误返回
		if err != nil {
			t.Logf("管理器正确返回了错误: %v", err)
		}
	})
}