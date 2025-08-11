package main

import (
	"context"
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

// 诚实测试原则：测试删除organization_versions表后的系统完整性

func TestDatabaseIntegrityAfterVersionsTableRemoval(t *testing.T) {
	// 连接测试数据库
	db := getTestDB(t)
	defer db.Close()

	t.Run("验证organization_versions表已完全删除", func(t *testing.T) {
		var tableExists bool
		err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM pg_tables 
				WHERE tablename = 'organization_versions'
			)
		`).Scan(&tableExists)
		
		require.NoError(t, err, "查询表存在性失败")
		assert.False(t, tableExists, "❌ CRITICAL: organization_versions表仍然存在，删除操作失败")
	})

	t.Run("验证备份表数据完整性", func(t *testing.T) {
		var backupCount int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM organization_versions_backup_before_deletion
		`).Scan(&backupCount)
		
		require.NoError(t, err, "查询备份表失败")
		assert.Equal(t, 2, backupCount, "❌ 备份表数据不完整，期望2条记录")
	})

	t.Run("验证organization_units表时态字段完整性", func(t *testing.T) {
		var fieldCount int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM information_schema.columns 
			WHERE table_name = 'organization_units' 
			AND column_name IN ('effective_date', 'end_date', 'change_reason', 'is_current')
		`).Scan(&fieldCount)
		
		require.NoError(t, err, "查询时态字段失败")
		assert.Equal(t, 4, fieldCount, "❌ CRITICAL: 时态字段不完整，纯日期生效模型受损")
	})

	t.Run("验证相关触发器已彻底删除", func(t *testing.T) {
		var triggerCount int
		err := db.QueryRow(`
			SELECT COUNT(*) FROM pg_proc WHERE proname = 'auto_manage_end_date_v2'
		`).Scan(&triggerCount)
		
		require.NoError(t, err, "查询触发器状态失败")
		assert.Equal(t, 0, triggerCount, "❌ 相关触发器未完全清理，存在技术债务")
	})

	t.Run("验证时态数据查询功能", func(t *testing.T) {
		var activeCount, temporalCount int
		
		// 检查活跃组织数量
		err := db.QueryRow(`
			SELECT COUNT(*) FROM organization_units WHERE is_current = true
		`).Scan(&activeCount)
		require.NoError(t, err, "查询活跃组织失败")
		
		// 检查时态数据完整性
		err = db.QueryRow(`
			SELECT COUNT(*) FROM organization_units WHERE effective_date IS NOT NULL
		`).Scan(&temporalCount)
		require.NoError(t, err, "查询时态数据失败")
		
		assert.Greater(t, activeCount, 0, "❌ CRITICAL: 无活跃组织数据，系统功能受损")
		assert.Greater(t, temporalCount, 0, "❌ CRITICAL: 时态数据缺失，功能不可用")
		assert.GreaterOrEqual(t, temporalCount, activeCount, "❌ 时态数据完整性异常")
	})
}

func TestTemporalAPIEndpointsComprehensive(t *testing.T) {
	// 创建测试服务器
	db := getTestDB(t)
	defer db.Close()
	
	handler := NewTemporalOrganizationHandler(db)
	server := httptest.NewServer(setupRoutes(handler))
	defer server.Close()

	// 测试组织代码
	testOrgCode := "1000056"

	t.Run("基础时态查询端点功能", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/organization-units/" + testOrgCode + "/temporal")
		require.NoError(t, err, "HTTP请求失败")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "❌ API端点返回错误状态码")
		
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "响应JSON解析失败")
		
		// 诚实测试：验证响应结构完整性
		assert.Contains(t, response, "organizations", "❌ 响应缺少organizations字段")
		assert.Contains(t, response, "result_count", "❌ 响应缺少result_count字段")
		assert.Contains(t, response, "queried_at", "❌ 响应缺少queried_at字段")
		
		organizations := response["organizations"].([]interface{})
		assert.Greater(t, len(organizations), 0, "❌ CRITICAL: 时态查询返回空结果")
		
		// 验证第一条记录的字段完整性
		firstOrg := organizations[0].(map[string]interface{})
		requiredFields := []string{"code", "name", "effective_date", "is_current", "tenant_id"}
		for _, field := range requiredFields {
			assert.Contains(t, firstOrg, field, "❌ 组织记录缺少必需字段: %s", field)
		}
	})

	t.Run("时间点查询功能验证", func(t *testing.T) {
		testDate := "2025-08-01"
		resp, err := http.Get(server.URL + "/api/v1/organization-units/" + testOrgCode + "/temporal?as_of_date=" + testDate)
		require.NoError(t, err, "时间点查询HTTP请求失败")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "❌ 时间点查询失败")
		
		var response map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err, "时间点查询响应解析失败")
		
		// 验证查询选项被正确处理
		assert.Contains(t, response, "query_options", "❌ 时间点查询缺少query_options")
	})

	t.Run("错误处理机制验证", func(t *testing.T) {
		// 测试不存在的组织
		resp, err := http.Get(server.URL + "/api/v1/organization-units/9999999/temporal")
		require.NoError(t, err, "错误测试HTTP请求失败")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusNotFound, resp.StatusCode, "❌ 错误处理不正确，应返回404")
		
		var errorResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&errorResponse)
		require.NoError(t, err, "错误响应解析失败")
		
		assert.Contains(t, errorResponse, "error_code", "❌ 错误响应缺少error_code")
		assert.Equal(t, "NOT_FOUND", errorResponse["error_code"], "❌ 错误码不正确")
	})

	t.Run("健康检查端点验证", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		require.NoError(t, err, "健康检查HTTP请求失败")
		defer resp.Body.Close()
		
		assert.Equal(t, http.StatusOK, resp.StatusCode, "❌ 健康检查失败")
		
		var healthResponse map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&healthResponse)
		require.NoError(t, err, "健康检查响应解析失败")
		
		assert.Equal(t, "healthy", healthResponse["status"], "❌ 服务状态异常")
		assert.Contains(t, healthResponse, "service", "❌ 服务信息缺失")
	})
}

func TestTemporalQueryLogicWithoutVersions(t *testing.T) {
	db := getTestDB(t)
	defer db.Close()
	
	repo := NewTemporalOrganizationRepository(db)
	tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
	
	t.Run("纯日期生效模型查询验证", func(t *testing.T) {
		opts := &TemporalQueryOptions{
			IncludeHistory: false,
			IncludeFuture:  false,
		}
		
		orgs, err := repo.GetByCodeTemporal(context.Background(), tenantID, "1000056", opts)
		require.NoError(t, err, "时态查询执行失败")
		assert.Greater(t, len(orgs), 0, "❌ CRITICAL: 纯日期模型查询无结果")
		
		// 验证返回的组织记录包含完整时态信息
		org := orgs[0]
		assert.NotNil(t, org.EffectiveDate, "❌ 生效日期字段缺失")
		assert.NotNil(t, org.IsCurrent, "❌ 当前标识字段缺失")
		assert.True(t, *org.IsCurrent, "❌ 查询结果不是当前有效记录")
	})

	t.Run("时间范围查询功能验证", func(t *testing.T) {
		fromDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		toDate := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
		
		opts := &TemporalQueryOptions{
			EffectiveFrom:  &fromDate,
			EffectiveTo:    &toDate,
			IncludeHistory: true,
		}
		
		orgs, err := repo.GetByCodeTemporal(context.Background(), tenantID, "1000056", opts)
		require.NoError(t, err, "时间范围查询失败")
		assert.Greater(t, len(orgs), 0, "❌ 时间范围查询无结果")
		
		// 验证查询结果在指定时间范围内
		for _, org := range orgs {
			if org.EffectiveDate != nil {
				assert.True(t, org.EffectiveDate.After(fromDate) || org.EffectiveDate.Equal(fromDate),
					"❌ 查询结果不在指定时间范围内")
			}
		}
	})
}

func TestPerformanceBenchmarksHonest(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过性能测试")
	}
	
	db := getTestDB(t)
	defer db.Close()
	
	handler := NewTemporalOrganizationHandler(db)
	server := httptest.NewServer(setupRoutes(handler))
	defer server.Close()

	t.Run("API响应时间基准测试", func(t *testing.T) {
		const iterations = 10
		var totalDuration time.Duration
		
		for i := 0; i < iterations; i++ {
			start := time.Now()
			
			resp, err := http.Get(server.URL + "/api/v1/organization-units/1000056/temporal")
			require.NoError(t, err, "性能测试HTTP请求失败")
			resp.Body.Close()
			
			duration := time.Since(start)
			totalDuration += duration
		}
		
		avgDuration := totalDuration / iterations
		
		// 诚实测试：严格的性能要求
		assert.Less(t, avgDuration, 500*time.Millisecond, 
			"❌ PERFORMANCE FAIL: 平均响应时间 %v 超过500ms基准", avgDuration)
		
		t.Logf("✅ 性能测试通过: 平均响应时间 %v", avgDuration)
	})
}

// 辅助函数
func getTestDB(t *testing.T) *sql.DB {
	dbURL := "postgres://user:password@localhost:5432/cubecastle?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	require.NoError(t, err, "数据库连接失败")
	
	err = db.Ping()
	require.NoError(t, err, "数据库连接测试失败")
	
	return db
}

func setupRoutes(handler *TemporalOrganizationHandler) http.Handler {
	// 这里需要根据实际的路由设置来实现
	// 简化版本，实际应该使用项目中的路由配置
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/organization-units/", handler.GetOrganizationTemporal)
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "healthy",
			"service": "organization-temporal-command-service-test",
		})
	})
	return mux
}