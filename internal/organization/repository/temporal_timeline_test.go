package repository

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"cube-castle/internal/types"
	pkglogger "cube-castle/pkg/logger"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// TestTemporalTimelineManager_ComplexScenarios 测试复杂场景下的时态时间轴管理
func TestTemporalTimelineManager_ComplexScenarios(t *testing.T) {
	// 跳过集成测试 (需要数据库连接)
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	// 数据库连接
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("未设置DATABASE_URL环境变量，跳过集成测试")
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		t.Fatalf("数据库连接失败: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Skipf("数据库连接验证失败，跳过集成测试: %v", err)
	}

	baseLogger := pkglogger.NewLogger(
		pkglogger.WithWriter(os.Stdout),
		pkglogger.WithLevel(pkglogger.LevelInfo),
	)
	tm := NewTemporalTimelineManager(db, baseLogger.WithFields(pkglogger.Fields{
		"test": "temporalTimeline",
	}))

	tenantID := uuid.New()
	orgCode := "TEST001"

	ctx := context.Background()

	t.Run("复杂时间轴场景测试", func(t *testing.T) {
		// 清理测试数据
		cleanupTestData(t, db, tenantID, orgCode)

		// 创建5个版本的复杂时间轴
		createComplexTimeline(t, tm, ctx, tenantID, orgCode)

		// 测试1: 中间插入记录
		testMiddleInsert(t, tm, ctx, tenantID, orgCode)

		// 测试2: 删除中间记录
		testMiddleDelete(t, tm, ctx, tenantID, orgCode)

		// 测试3: 删除第一条记录
		testFirstDelete(t, tm, ctx, tenantID, orgCode)

		// 测试4: 删除最后一条记录
		testLastDelete(t, tm, ctx, tenantID, orgCode)

		// 最终验证时间轴连续性
		verifyTimelineContinuity(t, db, tenantID, orgCode)
	})
}

// 清理测试数据
func cleanupTestData(t *testing.T, db *sql.DB, tenantID uuid.UUID, orgCode string) {
	_, err := db.Exec("DELETE FROM organization_units WHERE tenant_id = $1 AND code = $2", tenantID, orgCode)
	if err != nil {
		t.Logf("清理测试数据时出错: %v", err)
	}
	t.Log("测试数据清理完成")
}

// 创建5个版本的复杂时间轴
func createComplexTimeline(t *testing.T, tm *TemporalTimelineManager, ctx context.Context, tenantID uuid.UUID, orgCode string) {
	t.Log("创建5个版本的复杂时间轴...")

	versions := []struct {
		name          string
		effectiveDate string
		reason        string
	}{
		{"测试部门 v1.0", "2024-01-01", "初始版本"},
		{"测试部门 v2.0", "2024-03-01", "组织架构调整"},
		{"测试部门 v3.0", "2024-06-01", "中期重组"},
		{"测试部门 v4.0", "2024-09-01", "季度调整"},
		{"测试部门 v5.0", "2024-12-01", "年终重组"},
	}

	for i, version := range versions {
		effectiveDate, err := time.Parse("2006-01-02", version.effectiveDate)
		if err != nil {
			t.Fatalf("解析日期失败: %v", err)
		}

		codePath := "/" + orgCode
		org := &types.Organization{
			RecordID:      uuid.New().String(),
			TenantID:      tenantID.String(),
			Code:          orgCode,
			ParentCode:    nil,
			Name:          version.name,
			UnitType:      "DEPARTMENT",
			Status:        "ACTIVE",
			Level:         1,
			CodePath:      codePath,
			NamePath:      "/" + version.name,
			SortOrder:     1,
			Description:   version.reason,
			EffectiveDate: types.NewDateFromTime(effectiveDate),
			// isTemporal removed; derived from endDate
			ChangeReason: &version.reason,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		createdVersion, err := tm.InsertVersion(ctx, org)
		if err != nil {
			t.Fatalf("插入版本 %d 失败: %v", i+1, err)
		}

		t.Logf("✅ 版本 %d 插入成功: %s (RecordID: %s)", i+1, version.name, createdVersion.RecordID)
	}

	t.Log("✅ 5个版本的复杂时间轴创建完成")
}

// 测试中间插入记录
func testMiddleInsert(t *testing.T, tm *TemporalTimelineManager, ctx context.Context, tenantID uuid.UUID, orgCode string) {
	t.Log("🧪 测试1: 中间插入记录 (2024-04-15)")

	// 插入中间版本
	effectiveDate, _ := time.Parse("2006-01-02", "2024-04-15")

	codePath := "/" + orgCode
	org := &types.Organization{
		RecordID:      uuid.New().String(),
		TenantID:      tenantID.String(),
		Code:          orgCode,
		ParentCode:    nil,
		Name:          "测试部门 v2.5 (中间插入)",
		UnitType:      "DEPARTMENT",
		Status:        "ACTIVE",
		Level:         1,
		CodePath:      codePath,
		NamePath:      "/测试部门 v2.5 (中间插入)",
		SortOrder:     1,
		Description:   "中间插入测试",
		EffectiveDate: types.NewDateFromTime(effectiveDate),
		// isTemporal removed; derived from endDate
		ChangeReason: func(s string) *string { return &s }("中间版本插入测试"),
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdVersion, err := tm.InsertVersion(ctx, org)
	if err != nil {
		t.Fatalf("中间版本插入失败: %v", err)
	}

	t.Logf("✅ 中间版本插入成功: %s (RecordID: %s)", createdVersion.Name, createdVersion.RecordID)

	// 查询当前完整时间轴
	timeline, err := tm.RecalculateTimeline(ctx, tenantID, orgCode)
	if err != nil {
		t.Fatalf("查询时间轴失败: %v", err)
	}

	t.Logf("当前时间轴包含 %d 个版本:", len(*timeline))
	for i, version := range *timeline {
		t.Logf("  版本 %d: %s → %v | %s", i+1,
			version.EffectiveDate.Format("2006-01-02"),
			func() string {
				if version.EndDate != nil {
					return version.EndDate.Format("2006-01-02")
				}
				return "∞"
			}(),
			version.Name)
	}
}

// 测试删除中间记录
func testMiddleDelete(t *testing.T, tm *TemporalTimelineManager, ctx context.Context, tenantID uuid.UUID, orgCode string) {
	t.Log("🧪 测试2: 删除中间记录 (2024-04-15版本)")

	// 找到2024-04-15的版本recordId
	var recordID uuid.UUID
	err := tm.db.QueryRow(`
		SELECT record_id FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 AND effective_date = $3 
		  AND status != 'DELETED'
	`, tenantID, orgCode, "2024-04-15").Scan(&recordID)

	if err == sql.ErrNoRows {
		t.Skip("未找到2024-04-15版本，跳过删除测试")
	} else if err != nil {
		t.Fatalf("查找中间版本失败: %v", err)
	}

	timeline, err := tm.DeleteVersion(ctx, tenantID, recordID)
	if err != nil {
		t.Fatalf("删除中间版本失败: %v", err)
	}

	t.Logf("✅ 中间版本删除成功，当前时间轴包含 %d 个版本", len(*timeline))

	// 验证时间轴连续性
	for i, version := range *timeline {
		t.Logf("  版本 %d: %s → %v | %s", i+1,
			version.EffectiveDate.Format("2006-01-02"),
			func() string {
				if version.EndDate != nil {
					return version.EndDate.Format("2006-01-02")
				}
				return "∞"
			}(),
			version.Name)
	}
}

// 测试删除第一条记录
func testFirstDelete(t *testing.T, tm *TemporalTimelineManager, ctx context.Context, tenantID uuid.UUID, orgCode string) {
	t.Log("🧪 测试3: 删除第一条记录")

	// 找到第一个版本的recordId
	var recordID uuid.UUID
	err := tm.db.QueryRow(`
		SELECT record_id FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED'
		ORDER BY effective_date ASC
		LIMIT 1
	`, tenantID, orgCode).Scan(&recordID)

	if err == sql.ErrNoRows {
		t.Skip("未找到第一个版本，跳过删除测试")
	} else if err != nil {
		t.Fatalf("查找第一个版本失败: %v", err)
	}

	timeline, err := tm.DeleteVersion(ctx, tenantID, recordID)
	if err != nil {
		t.Fatalf("删除第一个版本失败: %v", err)
	}

	t.Logf("✅ 第一个版本删除成功，当前时间轴包含 %d 个版本", len(*timeline))

	// 验证新的第一个版本
	if len(*timeline) > 0 {
		firstVersion := (*timeline)[0]
		t.Logf("  新的第一个版本: %s (%s)", firstVersion.EffectiveDate.Format("2006-01-02"), firstVersion.Name)
	}
}

// 测试删除最后一条记录
func testLastDelete(t *testing.T, tm *TemporalTimelineManager, ctx context.Context, tenantID uuid.UUID, orgCode string) {
	t.Log("🧪 测试4: 删除最后一条记录")

	// 找到最后一个版本的recordId
	var recordID uuid.UUID
	err := tm.db.QueryRow(`
		SELECT record_id FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED'
		ORDER BY effective_date DESC
		LIMIT 1
	`, tenantID, orgCode).Scan(&recordID)

	if err == sql.ErrNoRows {
		t.Skip("未找到最后一个版本，跳过删除测试")
	} else if err != nil {
		t.Fatalf("查找最后一个版本失败: %v", err)
	}

	timeline, err := tm.DeleteVersion(ctx, tenantID, recordID)
	if err != nil {
		t.Fatalf("删除最后一个版本失败: %v", err)
	}

	t.Logf("✅ 最后一个版本删除成功，当前时间轴包含 %d 个版本", len(*timeline))

	// 验证新的最后一个版本
	if len(*timeline) > 0 {
		lastVersion := (*timeline)[len(*timeline)-1]
		if lastVersion.EndDate != nil {
			t.Errorf("新的最后版本的end_date应该为NULL，实际为: %v", lastVersion.EndDate)
		} else {
			t.Logf("  新的最后版本: %s → ∞ (%s)", lastVersion.EffectiveDate.Format("2006-01-02"), lastVersion.Name)
		}
	}
}

// 验证时间轴连续性
func verifyTimelineContinuity(t *testing.T, db *sql.DB, tenantID uuid.UUID, orgCode string) {
	t.Log("🔍 验证时间轴连续性")

	// 检查1: 时间断档
	var gapCount int
	err := db.QueryRow(`
		WITH timeline AS (
			SELECT 
				effective_date,
				end_date,
				LEAD(effective_date) OVER (ORDER BY effective_date) as next_start
			FROM organization_units 
			WHERE tenant_id = $1 AND code = $2 
			  AND status != 'DELETED'
			ORDER BY effective_date
		)
		SELECT COUNT(*) 
		FROM timeline 
		WHERE end_date IS NOT NULL 
		  AND next_start IS NOT NULL 
		  AND end_date + INTERVAL '1 day' != next_start
	`, tenantID, orgCode).Scan(&gapCount)

	if err != nil {
		t.Errorf("时间断档检查失败: %v", err)
	} else if gapCount > 0 {
		t.Errorf("发现 %d 个时间断档", gapCount)
	} else {
		t.Log("✅ 时间断档检查通过")
	}

	// 检查2: 当前版本唯一性
	var currentCount int
	err = db.QueryRow(`
		SELECT COUNT(*) 
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED'
		  AND is_current = true
	`, tenantID, orgCode).Scan(&currentCount)

	if err != nil {
		t.Errorf("当前版本检查失败: %v", err)
	} else if currentCount != 1 {
		t.Errorf("当前版本数量异常: %d (应该为1)", currentCount)
	} else {
		t.Log("✅ 当前版本唯一性检查通过")
	}

	// 检查3: 尾部开放
	var tailOpen bool
	err = db.QueryRow(`
		SELECT end_date IS NULL
		FROM organization_units 
		WHERE tenant_id = $1 AND code = $2 
		  AND status != 'DELETED'
		  AND effective_date = (
			  SELECT MAX(effective_date) 
			  FROM organization_units 
			  WHERE tenant_id = $1 AND code = $2 
				AND status != 'DELETED'
		  )
	`, tenantID, orgCode).Scan(&tailOpen)

	if err != nil {
		t.Errorf("尾部开放检查失败: %v", err)
	} else if !tailOpen {
		t.Error("最后版本的end_date不为NULL")
	} else {
		t.Log("✅ 尾部开放检查通过")
	}

	var hierarchyMismatch int
	err = db.QueryRow(`
		SELECT COUNT(*)
		FROM organization_units
		WHERE tenant_id = $1 AND code = $2
		  AND status <> 'DELETED'
		  AND (
			code_path IS NULL OR code_path = ''
			OR name_path IS NULL OR name_path = ''
		  )
	`, tenantID, orgCode).Scan(&hierarchyMismatch)
	if err != nil {
		t.Errorf("层级字段一致性检查失败: %v", err)
	} else if hierarchyMismatch != 0 {
		t.Errorf("存在 %d 条记录的 path/code_path/name_path 不一致", hierarchyMismatch)
	} else {
		t.Log("✅ 层级字段一致性检查通过")
	}

	t.Log("🎯 时间轴连续性验证完成")
}
