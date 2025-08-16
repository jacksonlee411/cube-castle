package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ===== 测试数据和工具函数 =====

var testDB *sql.DB
var testHandler *TemporalOrganizationHandler

func setupTestDB(t *testing.T) {
	var err error
	testDB, err = sql.Open("postgres", "postgres://user:password@localhost:5432/cubecastle?sslmode=disable")
	require.NoError(t, err)

	err = testDB.Ping()
	require.NoError(t, err)

	// 创建测试处理器
	testHandler = NewTemporalOrganizationHandler(testDB)
}

func teardownTestDB() {
	if testDB != nil {
		testDB.Close()
	}
}

func createTestOrganization(t *testing.T, code string) *Organization {
	org := &Organization{
		TenantID:      DefaultTenantID.String(),
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

	// 插入测试数据
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

// ===== 单元测试 =====

func TestMain(m *testing.M) {
	setupTestDB(&testing.T{})
	defer teardownTestDB()
	m.Run()
}

func TestParseTemporalQuery(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected *TemporalQueryOptions
		hasError bool
	}{
		{
			name:  "空查询参数",
			query: "",
			expected: &TemporalQueryOptions{
				IncludeHistory:   false,
				IncludeFuture:    false,
				IncludeDissolved: false,
			},
		},
		{
			name:  "as_of_date查询",
			query: "as_of_date=2025-08-10",
			expected: &TemporalQueryOptions{
				AsOfDate: func() *time.Time {
					t, _ := time.Parse("2006-01-02", "2025-08-10")
					return &t
				}(),
			},
		},
		{
			name:  "范围查询",
			query: "effective_from=2025-01-01&effective_to=2025-12-31",
			expected: &TemporalQueryOptions{
				EffectiveFrom: func() *time.Time {
					t, _ := time.Parse("2006-01-02", "2025-01-01")
					return &t
				}(),
				EffectiveTo: func() *time.Time {
					t, _ := time.Parse("2006-01-02", "2025-12-31")
					return &t
				}(),
			},
		},
		{
			name:  "布尔参数",
			query: "include_history=true&include_future=true&include_dissolved=true",
			expected: &TemporalQueryOptions{
				IncludeHistory:   true,
				IncludeFuture:    true,
				IncludeDissolved: true,
			},
		},
		{
			name:     "无效日期格式",
			query:    "as_of_date=invalid-date",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/?"+tt.query, nil)

			opts, err := ParseTemporalQuery(req)

			if tt.hasError {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.expected.IncludeHistory, opts.IncludeHistory)
			assert.Equal(t, tt.expected.IncludeFuture, opts.IncludeFuture)
			assert.Equal(t, tt.expected.IncludeDissolved, opts.IncludeDissolved)

			if tt.expected.AsOfDate != nil {
				require.NotNil(t, opts.AsOfDate)
				assert.Equal(t, tt.expected.AsOfDate.Format("2006-01-02"), opts.AsOfDate.Format("2006-01-02"))
			}
		})
	}
}

func TestTemporalOrganizationRepository_GetByCodeTemporal(t *testing.T) {
	repo := NewTemporalOrganizationRepository(testDB)
	testOrg := createTestOrganization(t, "TEST001")

	tests := []struct {
		name     string
		opts     *TemporalQueryOptions
		expected int // 期望的记录数量
	}{
		{
			name: "当前记录查询",
			opts: &TemporalQueryOptions{
				IncludeHistory: false,
				IncludeFuture:  false,
			},
			expected: 1,
		},
		{
			name: "历史记录查询",
			opts: &TemporalQueryOptions{
				IncludeHistory: true,
				MaxRecords:     10,
			},
			expected: 1,
		},
		{
			name: "时间点查询",
			opts: &TemporalQueryOptions{
				AsOfDate: func() *time.Time {
					t := time.Now().Add(-24 * time.Hour) // 昨天
					return &t
				}(),
			},
			expected: 0, // 组织是今天创建的，昨天不存在
		},
		{
			name: "未来记录查询",
			opts: &TemporalQueryOptions{
				IncludeFuture: true,
				MaxRecords:    5,
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orgs, err := repo.GetByCodeTemporal(context.Background(), DefaultTenantID, testOrg.Code, tt.opts)

			assert.NoError(t, err)
			assert.Len(t, orgs, tt.expected)

			if tt.expected > 0 {
				assert.Equal(t, testOrg.Code, orgs[0].Code)
				assert.Equal(t, testOrg.Name, orgs[0].Name)
			}
		})
	}
}

func TestOrganizationChangeEvent_CREATE(t *testing.T) {
	// 创建事件请求
	req := OrganizationChangeEvent{
		EventType:     "UPDATE",
		EffectiveDate: time.Now().Add(24 * time.Hour), // 明天生效
		ChangeData: map[string]interface{}{
			"name":        "更新后的组织名称",
			"description": "更新后的描述",
		},
		ChangeReason: "单元测试更新",
	}

	// 先创建测试组织
	testOrg := createTestOrganization(t, "TEST002")

	// 创建HTTP请求
	reqBody, _ := json.Marshal(req)
	httpReq := httptest.NewRequest("POST", "/api/v1/organization-units/"+testOrg.Code+"/events", bytes.NewBuffer(reqBody))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Tenant-ID", DefaultTenantID.String())

	// 设置URL参数
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("code", testOrg.Code)
	httpReq = httpReq.WithContext(context.WithValue(httpReq.Context(), chi.RouteCtxKey, rctx))

	// 执行请求
	rr := httptest.NewRecorder()
	testHandler.CreateOrganizationEvent(rr, httpReq)

	// 验证响应
	assert.Equal(t, http.StatusCreated, rr.Code)

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "UPDATE", response["event_type"])
	assert.Equal(t, testOrg.Code, response["organization"])
	assert.Equal(t, "processed", response["status"])
	assert.NotEmpty(t, response["event_id"])
}

func TestGetOrganizationTemporal_HTTP(t *testing.T) {
	// 创建测试组织
	testOrg := createTestOrganization(t, "TEST003")

	tests := []struct {
		name           string
		url            string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "基础查询",
			url:            "/api/v1/organization-units/" + testOrg.Code + "/temporal",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "历史查询",
			url:            "/api/v1/organization-units/" + testOrg.Code + "/temporal?include_history=true",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "范围查询",
			url:            "/api/v1/organization-units/" + testOrg.Code + "/temporal?effective_from=2025-01-01&effective_to=2025-12-31",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "不存在的组织",
			url:            "/api/v1/organization-units/NOTEXIST/temporal",
			expectedStatus: http.StatusNotFound,
			expectedCount:  0,
		},
		{
			name:           "无效查询参数",
			url:            "/api/v1/organization-units/" + testOrg.Code + "/temporal?as_of_date=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.url, nil)
			req.Header.Set("X-Tenant-ID", DefaultTenantID.String())

			// 设置URL参数
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("code", testOrg.Code)
			if tt.name == "不存在的组织" {
				rctx.URLParams.Keys[0] = "code"
				rctx.URLParams.Values[0] = "NOTEXIST"
			}
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			testHandler.GetOrganizationTemporal(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				orgs := response["organizations"].([]interface{})
				assert.Len(t, orgs, tt.expectedCount)

				if tt.expectedCount > 0 {
					org := orgs[0].(map[string]interface{})
					assert.Equal(t, testOrg.Code, org["code"])
				}
			}
		})
	}
}

func TestCacheKeyGeneration(t *testing.T) {
	handler := NewTemporalOrganizationHandler(testDB)

	// 测试缓存键生成的一致性
	opts1 := &TemporalQueryOptions{
		AsOfDate: func() *time.Time {
			t, _ := time.Parse("2006-01-02", "2025-08-10")
			return &t
		}(),
		IncludeHistory: true,
	}

	opts2 := &TemporalQueryOptions{
		AsOfDate: func() *time.Time {
			t, _ := time.Parse("2006-01-02", "2025-08-10")
			return &t
		}(),
		IncludeHistory: true,
	}

	key1 := handler.getCacheKey("tenant1", "org1", opts1)
	key2 := handler.getCacheKey("tenant1", "org1", opts2)
	key3 := handler.getCacheKey("tenant2", "org1", opts1)

	// 相同参数应该生成相同的缓存键
	assert.Equal(t, key1, key2)
	// 不同租户应该生成不同的缓存键
	assert.NotEqual(t, key1, key3)
	// 所有缓存键应该有正确的前缀
	assert.Contains(t, key1, "cache:")
	assert.Contains(t, key3, "cache:")
}

func TestTenantIDExtraction(t *testing.T) {
	handler := NewTemporalOrganizationHandler(testDB)

	tests := []struct {
		name       string
		header     string
		expectedID uuid.UUID
	}{
		{
			name:       "有效的租户ID",
			header:     "123e4567-e89b-12d3-a456-426614174000",
			expectedID: uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
		},
		{
			name:       "无效的租户ID",
			header:     "invalid-uuid",
			expectedID: DefaultTenantID,
		},
		{
			name:       "空租户ID",
			header:     "",
			expectedID: DefaultTenantID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.header != "" {
				req.Header.Set("X-Tenant-ID", tt.header)
			}

			tenantID := handler.getTenantID(req)
			assert.Equal(t, tt.expectedID, tenantID)
		})
	}
}

// ===== 基准测试 =====

func BenchmarkParseTemporalQuery(b *testing.B) {
	req := httptest.NewRequest("GET", "/?as_of_date=2025-08-10&include_history=true&include_future=true", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ParseTemporalQuery(req)
	}
}

func BenchmarkGetByCodeTemporal(b *testing.B) {
	repo := NewTemporalOrganizationRepository(testDB)
	testOrg := createTestOrganization(&testing.T{}, "BENCH001")

	opts := &TemporalQueryOptions{
		IncludeHistory: true,
		MaxRecords:     10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		repo.GetByCodeTemporal(context.Background(), DefaultTenantID, testOrg.Code, opts)
	}
}

func BenchmarkCacheKeyGeneration(b *testing.B) {
	handler := NewTemporalOrganizationHandler(testDB)
	opts := &TemporalQueryOptions{
		AsOfDate: func() *time.Time {
			t := time.Now()
			return &t
		}(),
		IncludeHistory: true,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handler.getCacheKey("tenant", "org", opts)
	}
}
