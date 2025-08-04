package common

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateBusinessID 测试业务ID验证功能
func TestValidateBusinessID(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		businessID string
		expectErr  bool
		errorMsg   string
	}{
		// 员工业务ID测试
		{
			name:       "Valid employee ID - single digit",
			entityType: EntityTypeEmployee,
			businessID: "1",
			expectErr:  false,
		},
		{
			name:       "Valid employee ID - multiple digits",
			entityType: EntityTypeEmployee,
			businessID: "12345",
			expectErr:  false,
		},
		{
			name:       "Valid employee ID - max length",
			entityType: EntityTypeEmployee,
			businessID: "99999999",
			expectErr:  false,
		},
		{
			name:       "Invalid employee ID - starts with zero",
			entityType: EntityTypeEmployee,
			businessID: "01234",
			expectErr:  true,
		},
		{
			name:       "Invalid employee ID - too long",
			entityType: EntityTypeEmployee,
			businessID: "123456789",
			expectErr:  true,
		},
		{
			name:       "Invalid employee ID - contains letters",
			entityType: EntityTypeEmployee,
			businessID: "123a4",
			expectErr:  true,
		},
		{
			name:       "Invalid employee ID - empty",
			entityType: EntityTypeEmployee,
			businessID: "",
			expectErr:  true,
		},

		// 组织业务ID测试
		{
			name:       "Valid organization ID - min value",
			entityType: EntityTypeOrganization,
			businessID: "100000",
			expectErr:  false,
		},
		{
			name:       "Valid organization ID - max value",
			entityType: EntityTypeOrganization,
			businessID: "999999",
			expectErr:  false,
		},
		{
			name:       "Valid organization ID - middle value",
			entityType: EntityTypeOrganization,
			businessID: "123456",
			expectErr:  false,
		},
		{
			name:       "Invalid organization ID - too short",
			entityType: EntityTypeOrganization,
			businessID: "12345",
			expectErr:  true,
		},
		{
			name:       "Invalid organization ID - too long",
			entityType: EntityTypeOrganization,
			businessID: "1234567",
			expectErr:  true,
		},
		{
			name:       "Invalid organization ID - starts with zero",
			entityType: EntityTypeOrganization,
			businessID: "012345",
			expectErr:  true,
		},
		{
			name:       "Invalid organization ID - below min range",
			entityType: EntityTypeOrganization,
			businessID: "099999",
			expectErr:  true,
		},

		// 职位业务ID测试
		{
			name:       "Valid position ID - min value",
			entityType: EntityTypePosition,
			businessID: "1000000",
			expectErr:  false,
		},
		{
			name:       "Valid position ID - max value",
			entityType: EntityTypePosition,
			businessID: "9999999",
			expectErr:  false,
		},
		{
			name:       "Invalid position ID - too short",
			entityType: EntityTypePosition,
			businessID: "123456",
			expectErr:  true,
		},
		{
			name:       "Invalid position ID - too long",
			entityType: EntityTypePosition,
			businessID: "12345678",
			expectErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateBusinessID(tt.entityType, tt.businessID)
			
			if tt.expectErr {
				assert.Error(t, err, "Expected validation error for %s", tt.businessID)
			} else {
				assert.NoError(t, err, "Expected no validation error for %s", tt.businessID)
			}
		})
	}
}

// TestGetBusinessIDRange 测试业务ID范围获取
func TestGetBusinessIDRange(t *testing.T) {
	tests := []struct {
		name       string
		entityType EntityType
		expectedMin int64
		expectedMax int64
		expectedLen int
	}{
		{
			name:        "Employee range",
			entityType:  EntityTypeEmployee,
			expectedMin: 1,
			expectedMax: 99999999,
			expectedLen: 8,
		},
		{
			name:        "Organization range",
			entityType:  EntityTypeOrganization,
			expectedMin: 100000,
			expectedMax: 999999,
			expectedLen: 6,
		},
		{
			name:        "Position range",
			entityType:  EntityTypePosition,
			expectedMin: 1000000,
			expectedMax: 9999999,
			expectedLen: 7,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rangeDef := GetBusinessIDRange(tt.entityType)
			
			assert.Equal(t, tt.expectedMin, rangeDef.Min)
			assert.Equal(t, tt.expectedMax, rangeDef.Max)
			assert.Equal(t, tt.expectedLen, rangeDef.Length)
		})
	}
}

// TestIsUUID 测试UUID检测功能
func TestIsUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Valid UUID v4",
			input:    "550e8400-e29b-41d4-a716-446655440000",
			expected: true,
		},
		{
			name:     "Valid UUID with different format",
			input:    "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
			expected: true,
		},
		{
			name:     "Invalid UUID - too short",
			input:    "550e8400-e29b-41d4-a716",
			expected: false,
		},
		{
			name:     "Invalid UUID - no hyphens",
			input:    "550e8400e29b41d4a716446655440000",
			expected: true, // Go's uuid.Parse实际上会接受这种格式
		},
		{
			name:     "Invalid UUID - contains non-hex chars",
			input:    "550e8400-e29b-41d4-a716-44665544000g",
			expected: false,
		},
		{
			name:     "Business ID (not UUID)",
			input:    "123456",
			expected: false,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsUUID(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// BusinessIDMockDB 业务ID测试专用的模拟数据库接口
type BusinessIDMockDB struct {
	sequences map[string]int64
	data      map[string]map[string]interface{}
}

// NewBusinessIDMockDB 创建模拟数据库
func NewBusinessIDMockDB() *BusinessIDMockDB {
	return &BusinessIDMockDB{
		sequences: map[string]int64{
			"employee_business_id_seq": 1,
			"org_business_id_seq":      0,
			"position_business_id_seq": 0,
		},
		data: make(map[string]map[string]interface{}),
	}
}

// DatabaseInterface 数据库接口，用于测试时的依赖注入
type DatabaseInterface interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner
}

// RowScanner 行扫描器接口
type RowScanner interface {
	Scan(dest ...interface{}) error
}

// TestableBusinessIDService 可测试的业务ID服务
type TestableBusinessIDService struct {
	db DatabaseInterface
}

// NewTestableBusinessIDService 创建可测试的业务ID服务
func NewTestableBusinessIDService(db DatabaseInterface) *TestableBusinessIDService {
	return &TestableBusinessIDService{db: db}
}

// GenerateBusinessID 生成新的业务ID
func (s *TestableBusinessIDService) GenerateBusinessID(ctx context.Context, entityType EntityType) (string, error) {
	var sequenceName string
	var offset int64

	switch entityType {
	case EntityTypeEmployee:
		sequenceName = "employee_business_id_seq"
		offset = 0
	case EntityTypeOrganization:
		sequenceName = "org_business_id_seq"
		offset = 100000
	case EntityTypePosition:
		sequenceName = "position_business_id_seq"
		offset = 1000000
	default:
		return "", fmt.Errorf("unknown entity type: %s", entityType)
	}

	var nextVal int64
	query := fmt.Sprintf("SELECT nextval('%s')", sequenceName)
	err := s.db.QueryRowContext(ctx, query).Scan(&nextVal)
	if err != nil {
		return "", fmt.Errorf("failed to generate business ID for %s: %w", entityType, err)
	}

	businessID := strconv.FormatInt(nextVal+offset, 10)

	// 验证生成的ID是否在有效范围内
	if err := ValidateBusinessID(entityType, businessID); err != nil {
		return "", fmt.Errorf("generated invalid business ID: %w", err)
	}

	return businessID, nil
}

// 让BusinessIDMockDB实现DatabaseInterface接口
func (m *BusinessIDMockDB) QueryRowContext(ctx context.Context, query string, args ...interface{}) RowScanner {
	return m.queryRowContext(ctx, query, args...)
}

// queryRowContext 内部方法，返回具体的mock row类型
func (m *BusinessIDMockDB) queryRowContext(ctx context.Context, query string, args ...interface{}) *BusinessIDMockRow {
	// 模拟序列查询
	if query == "SELECT nextval('employee_business_id_seq')" {
		val := m.sequences["employee_business_id_seq"]
		m.sequences["employee_business_id_seq"]++
		return &BusinessIDMockRow{value: val}
	}
	if query == "SELECT nextval('org_business_id_seq')" {
		val := m.sequences["org_business_id_seq"]
		m.sequences["org_business_id_seq"]++
		return &BusinessIDMockRow{value: val}
	}
	if query == "SELECT nextval('position_business_id_seq')" {
		val := m.sequences["position_business_id_seq"]
		m.sequences["position_business_id_seq"]++
		return &BusinessIDMockRow{value: val}
	}

	// 模拟查找查询
	return &BusinessIDMockRow{err: sql.ErrNoRows}
}

// BusinessIDMockRow 模拟数据库行
type BusinessIDMockRow struct {
	value interface{}
	err   error
}

// Scan 模拟扫描结果
func (r *BusinessIDMockRow) Scan(dest ...interface{}) error {
	if r.err != nil {
		return r.err
	}
	
	if len(dest) > 0 {
		switch v := dest[0].(type) {
		case *int64:
			if val, ok := r.value.(int64); ok {
				*v = val
			}
		case *string:
			if val, ok := r.value.(string); ok {
				*v = val
			}
		}
	}
	
	return nil
}

// TestableBusinessIDManager 可测试的业务ID管理器
type TestableBusinessIDManager struct {
	service   *TestableBusinessIDService
	config    BusinessIDManagerConfig
}

// NewTestableBusinessIDManager 创建可测试的业务ID管理器
func NewTestableBusinessIDManager(service *TestableBusinessIDService, config BusinessIDManagerConfig) *TestableBusinessIDManager {
	return &TestableBusinessIDManager{
		service: service,
		config:  config,
	}
}

// GenerateUniqueBusinessID 生成唯一的业务ID（带重试机制）
func (m *TestableBusinessIDManager) GenerateUniqueBusinessID(ctx context.Context, entityType EntityType) (string, error) {
	if !m.config.EnableAutoGeneration {
		return "", fmt.Errorf("auto generation is disabled")
	}

	var lastErr error
	for i := 0; i < m.config.MaxRetries; i++ {
		businessID, err := m.service.GenerateBusinessID(ctx, entityType)
		if err != nil {
			lastErr = err
			continue
		}

		if m.config.EnableValidation {
			if err := ValidateBusinessID(entityType, businessID); err != nil {
				lastErr = err
				continue
			}
		}

		return businessID, nil
	}

	return "", fmt.Errorf("failed to generate unique business ID after %d retries: %w", 
		m.config.MaxRetries, lastErr)
}
func TestBusinessIDService(t *testing.T) {
	mockDB := NewBusinessIDMockDB()
	service := NewTestableBusinessIDService(mockDB)

	t.Run("Generate Employee Business ID", func(t *testing.T) {
		ctx := context.Background()
		
		businessID, err := service.GenerateBusinessID(ctx, EntityTypeEmployee)
		require.NoError(t, err)
		
		assert.Regexp(t, `^[1-9][0-9]{0,7}$`, businessID)
		assert.Equal(t, "1", businessID) // 第一个生成的ID应该是1
	})

	t.Run("Generate Organization Business ID", func(t *testing.T) {
		ctx := context.Background()
		
		businessID, err := service.GenerateBusinessID(ctx, EntityTypeOrganization)
		require.NoError(t, err)
		
		assert.Regexp(t, `^[1-9][0-9]{5}$`, businessID)
		assert.Equal(t, "100000", businessID) // 100000 + 0
	})

	t.Run("Generate Position Business ID", func(t *testing.T) {
		ctx := context.Background()
		
		businessID, err := service.GenerateBusinessID(ctx, EntityTypePosition)
		require.NoError(t, err)
		
		assert.Regexp(t, `^[1-9][0-9]{6}$`, businessID)
		assert.Equal(t, "1000000", businessID) // 1000000 + 0
	})

	t.Run("Invalid Entity Type", func(t *testing.T) {
		ctx := context.Background()
		
		_, err := service.GenerateBusinessID(ctx, EntityType("invalid"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown entity type")
	})
}

// TestBusinessIDManager 测试业务ID管理器
func TestBusinessIDManager(t *testing.T) {
	mockDB := NewBusinessIDMockDB()
	service := NewTestableBusinessIDService(mockDB)
	config := DefaultBusinessIDManagerConfig()
	manager := NewTestableBusinessIDManager(service, config)

	t.Run("Generate Unique Business ID with Validation", func(t *testing.T) {
		ctx := context.Background()
		
		businessID, err := manager.GenerateUniqueBusinessID(ctx, EntityTypeEmployee)
		require.NoError(t, err)
		
		// 验证生成的ID格式正确
		assert.NoError(t, ValidateBusinessID(EntityTypeEmployee, businessID))
		assert.Regexp(t, `^[1-9][0-9]{0,7}$`, businessID)
	})

	t.Run("Auto Generation Disabled", func(t *testing.T) {
		ctx := context.Background()
		disabledConfig := config
		disabledConfig.EnableAutoGeneration = false
		disabledManager := NewTestableBusinessIDManager(service, disabledConfig)
		
		_, err := disabledManager.GenerateUniqueBusinessID(ctx, EntityTypeEmployee)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "auto generation is disabled")
	})
}

// TestBusinessIDService 测试业务ID服务

// BenchmarkValidateBusinessID 业务ID验证性能基准测试
func BenchmarkValidateBusinessID(b *testing.B) {
	testCases := []struct {
		name       string
		entityType EntityType
		businessID string
	}{
		{"Employee", EntityTypeEmployee, "12345"},
		{"Organization", EntityTypeOrganization, "123456"},
		{"Position", EntityTypePosition, "1234567"},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ValidateBusinessID(tc.entityType, tc.businessID)
			}
		})
	}
}

// BenchmarkBusinessIDGeneration 业务ID生成性能基准测试
func BenchmarkBusinessIDGeneration(b *testing.B) {
	mockDB := NewBusinessIDMockDB()
	service := NewTestableBusinessIDService(mockDB)
	ctx := context.Background()

	testCases := []struct {
		name       string
		entityType EntityType
	}{
		{"Employee", EntityTypeEmployee},
		{"Organization", EntityTypeOrganization},
		{"Position", EntityTypePosition},
	}

	for _, tc := range testCases {
		b.Run(tc.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				service.GenerateBusinessID(ctx, tc.entityType)
			}
		})
	}
}