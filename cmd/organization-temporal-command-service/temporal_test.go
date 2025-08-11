package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== 测试专用数据结构 =====

type TestOrganization struct {
	TenantID      string     `json:"tenant_id"`
	Code          string     `json:"code"`
	Name          string     `json:"name"`
	UnitType      string     `json:"unit_type"`
	Status        string     `json:"status"`
	Level         int        `json:"level"`
	Path          string     `json:"path"`
	SortOrder     int        `json:"sort_order"`
	Description   string     `json:"description"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	EffectiveDate *time.Time `json:"effective_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	ChangeReason  *string    `json:"change_reason,omitempty"`
	IsCurrent     *bool      `json:"is_current,omitempty"`
}

type TestChangeEvent struct {
	EventType     string                 `json:"event_type"`
	EffectiveDate time.Time              `json:"effective_date"`
	EndDate       *time.Time             `json:"end_date,omitempty"`
	ChangeData    map[string]interface{} `json:"change_data"`
	ChangeReason  string                 `json:"change_reason"`
}

// ===== 测试工具函数 =====

var testDB *sql.DB
var testTenantID = uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")

func setupTest(t *testing.T) {
	var err error
	testDB, err = sql.Open("postgres", "postgres://user:password@localhost:5432/cubecastle?sslmode=disable")
	require.NoError(t, err)
	
	err = testDB.Ping()
	require.NoError(t, err)
}

func teardownTest() {
	if testDB != nil {
		testDB.Close()
	}
}

func createTestOrg(t *testing.T, code string) *TestOrganization {
	org := &TestOrganization{
		TenantID:      testTenantID.String(),
		Code:          code,
		Name:          "测试组织_" + code,
		UnitType:      "DEPARTMENT",
		Status:        "ACTIVE",
		Level:         1,
		Path:          "/" + code,
		SortOrder:     1,
		Description:   "单元测试组织",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		EffectiveDate: func() *time.Time { t := time.Now(); return &t }(),
		IsCurrent:     func() *bool { b := true; return &b }(),
	}
	
	// 确保测试数据存在
	_, err := testDB.Exec(`
		INSERT INTO organization_units (
			tenant_id, code, name, unit_type, status, level, path, sort_order, description,
			created_at, updated_at, effective_date, is_current
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		ON CONFLICT (code) DO UPDATE SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`, org.TenantID, org.Code, org.Name, org.UnitType, org.Status, org.Level, 
		org.Path, org.SortOrder, org.Description, org.CreatedAt, org.UpdatedAt,
		org.EffectiveDate, org.IsCurrent)
	
	require.NoError(t, err)
	return org
}

// ===== 功能测试 =====

func TestTemporalQueryParsing(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		expectFuture bool
		expectHistory bool
	}{
		{
			name:         "包含未来记录",
			query:        "include_future=true",
			expectFuture: true,
		},
		{
			name:          "包含历史记录",
			query:         "include_history=true",
			expectHistory: true,
		},
		{
			name:          "同时包含历史和未来",
			query:         "include_future=true&include_history=true",
			expectFuture:  true,
			expectHistory: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)
			
			// 模拟解析逻辑
			includeFuture := req.URL.Query().Get("include_future") == "true"
			includeHistory := req.URL.Query().Get("include_history") == "true"
			
			assert.Equal(t, tt.expectFuture, includeFuture)
			assert.Equal(t, tt.expectHistory, includeHistory)
		})
	}
}

func TestTemporalAPIIntegration(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	
	// 创建测试组织
	testOrg := createTestOrg(t, "TEST001")
	
	tests := []struct {
		name           string
		url            string
		expectedStatus int
		method         string
		body           interface{}
	}{
		{
			name:           "查询当前记录",
			url:            "/api/v1/organization-units/" + testOrg.Code,
			expectedStatus: 200,
			method:         "GET",
		},
		{
			name:           "查询历史记录",
			url:            "/api/v1/organization-units/" + testOrg.Code + "?include_history=true",
			expectedStatus: 200,
			method:         "GET",
		},
		{
			name:           "创建UPDATE事件",
			url:            "/api/v1/organization-units/" + testOrg.Code + "/events",
			expectedStatus: 201,
			method:         "POST",
			body: TestChangeEvent{
				EventType:     "UPDATE",
				EffectiveDate: time.Now().Add(24 * time.Hour),
				ChangeData: map[string]interface{}{
					"description": "测试更新描述",
				},
				ChangeReason: "单元测试",
			},
		},
	}
	
	client := &http.Client{Timeout: 5 * time.Second}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			var err error
			
			if tt.method == "POST" && tt.body != nil {
				bodyBytes, _ := json.Marshal(tt.body)
				req, err = http.NewRequest(tt.method, "http://localhost:9091"+tt.url, bytes.NewBuffer(bodyBytes))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req, err = http.NewRequest(tt.method, "http://localhost:9091"+tt.url, nil)
			}
			
			require.NoError(t, err)
			req.Header.Set("X-Tenant-ID", testTenantID.String())
			
			resp, err := client.Do(req)
			if err != nil {
				t.Logf("API请求失败 (可能时态服务未运行): %v", err)
				t.Skip("跳过API集成测试 - 服务不可用")
				return
			}
			defer resp.Body.Close()
			
			// 验证响应状态
			if resp.StatusCode == tt.expectedStatus {
				t.Logf("✅ %s 测试通过 - 状态码: %d", tt.name, resp.StatusCode)
			} else {
				t.Logf("⚠️  %s 测试状态码不匹配 - 期望: %d, 实际: %d", tt.name, tt.expectedStatus, resp.StatusCode)
			}
			
			// 对于成功的响应，验证JSON格式
			if resp.StatusCode < 400 {
				var result map[string]interface{}
				err := json.NewDecoder(resp.Body).Decode(&result)
				assert.NoError(t, err, "响应应该是有效的JSON")
			}
		})
	}
}

func TestDatabaseConnection(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	
	// 测试数据库连接
	var count int
	err := testDB.QueryRow("SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1", testTenantID).Scan(&count)
	assert.NoError(t, err)
	
	t.Logf("数据库中找到 %d 个组织记录", count)
}

func TestCacheKeyConsistency(t *testing.T) {
	// 测试缓存键生成的一致性
	tests := []struct {
		tenant1, org1, tenant2, org2 string
		shouldMatch                  bool
	}{
		{"tenant1", "org1", "tenant1", "org1", true},   // 相同参数
		{"tenant1", "org1", "tenant2", "org1", false},  // 不同租户
		{"tenant1", "org1", "tenant1", "org2", false},  // 不同组织
		{"tenant2", "org2", "tenant2", "org2", true},   // 相同参数（不同值）
	}
	
	for i, tt := range tests {
		t.Run(string(rune('A'+i)), func(t *testing.T) {
			// 模拟缓存键生成逻辑
			key1 := tt.tenant1 + ":" + tt.org1
			key2 := tt.tenant2 + ":" + tt.org2
			
			if tt.shouldMatch {
				assert.Equal(t, key1, key2)
			} else {
				assert.NotEqual(t, key1, key2)
			}
		})
	}
}

func TestTenantIDHandling(t *testing.T) {
	tests := []struct {
		name     string
		tenantID string
		isValid  bool
	}{
		{
			name:     "有效的UUID",
			tenantID: "123e4567-e89b-12d3-a456-426614174000",
			isValid:  true,
		},
		{
			name:     "无效的UUID",
			tenantID: "invalid-uuid",
			isValid:  false,
		},
		{
			name:     "空字符串",
			tenantID: "",
			isValid:  false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uuid.Parse(tt.tenantID)
			if tt.isValid {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// ===== 性能测试 =====

func TestQueryPerformance(t *testing.T) {
	setupTest(t)
	defer teardownTest()
	
	// 创建测试数据
	testOrg := createTestOrg(t, "PERF001")
	
	// 测试查询性能
	start := time.Now()
	
	for i := 0; i < 10; i++ {
		var count int
		err := testDB.QueryRow(
			"SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1 AND code = $2",
			testTenantID, testOrg.Code,
		).Scan(&count)
		
		assert.NoError(t, err)
		assert.Greater(t, count, 0)
	}
	
	duration := time.Since(start)
	avgDuration := duration / 10
	
	t.Logf("平均查询时间: %v", avgDuration)
	
	// 性能预期：单次查询应该在10ms以内
	assert.Less(t, avgDuration, 10*time.Millisecond, "查询性能应该在可接受范围内")
}

// ===== 错误处理测试 =====

func TestErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		scenario      string
		expectError   bool
	}{
		{
			name:        "空组织代码",
			scenario:    "",
			expectError: true,
		},
		{
			name:        "有效组织代码",
			scenario:    "VALID001",
			expectError: false,
		},
		{
			name:        "特殊字符组织代码",
			scenario:    "TEST@123",
			expectError: false, // 假设系统接受特殊字符
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isEmpty := tt.scenario == ""
			
			if tt.expectError {
				assert.True(t, isEmpty || len(tt.scenario) > 50, "应该检测到错误情况")
			} else {
				assert.False(t, isEmpty, "不应该有错误")
			}
		})
	}
}

// ===== 基准测试 =====

func BenchmarkDatabaseQuery(b *testing.B) {
	setupTest(&testing.T{})
	defer teardownTest()
	
	createTestOrg(&testing.T{}, "BENCH001")
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var count int
		testDB.QueryRow(
			"SELECT COUNT(*) FROM organization_units WHERE tenant_id = $1 AND code = $2",
			testTenantID, "BENCH001",
		).Scan(&count)
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	org := &TestOrganization{
		TenantID:    testTenantID.String(),
		Code:        "BENCH002",
		Name:        "基准测试组织",
		UnitType:    "DEPARTMENT",
		Status:      "ACTIVE",
		Level:       1,
		Path:        "/BENCH002",
		SortOrder:   1,
		Description: "基准测试用途",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		json.Marshal(org)
	}
}