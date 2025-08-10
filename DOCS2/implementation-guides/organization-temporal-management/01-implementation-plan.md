# 时态管理API升级分步骤实施方案

**基于**: [ADR-007 时态管理API升级方案](../architecture-decisions/ADR-007-temporal-management-api-upgrade-plan.md)  
**版本**: v1.0  
**制定日期**: 2025-08-10  
**预计总工期**: 13周  
**实施团队**: 后端开发、前端开发、数据库管理、测试工程师

## 📋 实施概述

### 核心目标
- 实现元合约v6.0合规的时态管理能力
- 支持时间点查询和历史版本管理
- 建立智能结束日期管理策略
- 保证向后兼容性和零业务中断

### 三大阶段分解
1. **阶段1** (Week 1-4): 数据模型扩展与基础设施
2. **阶段2** (Week 5-7): API扩展与时态查询能力
3. **阶段3** (Week 8-13): 事件驱动重构与完整合规

---

## 🚀 阶段1：数据模型扩展 (Week 1-4)

### Week 1: 数据库架构设计与准备

#### 1.1 数据库设计确认 (Day 1-2)
```sql
-- 1. 备份现有数据
pg_dump -h localhost -U user -d cubecastle > backup_pre_temporal_$(date +%Y%m%d).sql

-- 2. 确认表结构扩展设计
\d organization_units;  -- 查看现有结构
```

**任务清单**:
- [x] 分析现有organization_units表结构
- [ ] 设计时态字段映射策略
- [ ] 制定数据迁移计划
- [ ] 准备回滚方案

#### 1.2 开发环境搭建 (Day 3-5)
```bash
# 创建专用开发分支
git checkout -b feature/temporal-db-migration

# 准备测试数据
cp backup_pre_temporal_*.sql test_data/
```

**任务清单**:
- [ ] 搭建独立开发环境
- [ ] 准备完整测试数据集
- [ ] 配置数据库连接池
- [ ] 建立监控指标收集

### Week 2: 核心表结构扩展

#### 2.1 organization_units表扩展 (Day 1-3)
```sql
-- Step 1: 添加时态字段
BEGIN TRANSACTION;

-- 添加新字段
ALTER TABLE organization_units 
ADD COLUMN effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
ADD COLUMN end_date DATE,
ADD COLUMN version INTEGER NOT NULL DEFAULT 1,
ADD COLUMN supersedes_version INTEGER,
ADD COLUMN change_reason VARCHAR(500),
ADD COLUMN is_current BOOLEAN NOT NULL DEFAULT true;

-- Step 2: 迁移现有数据
UPDATE organization_units 
SET effective_date = created_at::DATE,
    version = 1,
    is_current = true
WHERE effective_date IS NULL;

-- Step 3: 修改主键约束
ALTER TABLE organization_units DROP CONSTRAINT organization_units_pkey;
ALTER TABLE organization_units 
ADD CONSTRAINT organization_units_pkey PRIMARY KEY (code, version);

-- Step 4: 添加索引优化查询性能
CREATE INDEX idx_org_effective_date ON organization_units(effective_date);
CREATE INDEX idx_org_current_version ON organization_units(code, is_current) WHERE is_current = true;
CREATE INDEX idx_org_version_chain ON organization_units(code, version);

COMMIT;
```

**验证脚本**:
```sql
-- 验证数据完整性
SELECT 
    COUNT(*) as total_records,
    COUNT(DISTINCT code) as unique_orgs,
    COUNT(*) FILTER (WHERE is_current = true) as current_versions,
    COUNT(*) FILTER (WHERE effective_date IS NULL) as missing_dates
FROM organization_units;
```

**任务清单**:
- [ ] 执行表结构扩展脚本
- [ ] 验证数据迁移完整性  
- [ ] 测试索引性能影响
- [ ] 更新应用程序连接配置

#### 2.2 结束日期自动管理触发器 (Day 4-5)
```sql
-- 创建结束日期自动管理函数
CREATE OR REPLACE FUNCTION auto_manage_end_date()
RETURNS TRIGGER AS $$
DECLARE
    affected_rows INTEGER;
BEGIN
    -- 记录操作开始日志
    RAISE NOTICE '开始处理组织 % 的版本 % 结束日期管理', NEW.code, NEW.version;
    
    -- 自动设置前版本的end_date
    UPDATE organization_units 
    SET end_date = NEW.effective_date - INTERVAL '1 day',
        is_current = false
    WHERE code = NEW.code 
      AND is_current = true 
      AND version != NEW.version;
    
    GET DIAGNOSTICS affected_rows = ROW_COUNT;
    RAISE NOTICE '更新了 % 条前版本记录的结束日期', affected_rows;
    
    -- 验证时间线一致性
    IF EXISTS (
        SELECT 1 FROM organization_units 
        WHERE code = NEW.code 
        AND version != NEW.version
        AND effective_date >= NEW.effective_date
    ) THEN
        RAISE EXCEPTION '时间线冲突：不能在现有版本之前插入新版本';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 创建触发器
DROP TRIGGER IF EXISTS trigger_auto_end_date ON organization_units;
CREATE TRIGGER trigger_auto_end_date
    BEFORE INSERT ON organization_units
    FOR EACH ROW 
    EXECUTE FUNCTION auto_manage_end_date();
```

**测试脚本**:
```sql
-- 测试触发器功能
BEGIN;
-- 创建测试数据
INSERT INTO organization_units (code, name, unit_type, tenant_id, effective_date, version)
VALUES ('TEST001', '测试部门V1', 'DEPARTMENT', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '2025-01-01', 1);

-- 添加新版本，验证触发器
INSERT INTO organization_units (code, name, unit_type, tenant_id, effective_date, version)
VALUES ('TEST001', '测试部门V2', 'DEPARTMENT', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '2025-06-01', 2);

-- 验证结果
SELECT code, version, effective_date, end_date, is_current 
FROM organization_units 
WHERE code = 'TEST001'
ORDER BY version;

ROLLBACK;
```

### Week 3: 事件表和版本表创建

#### 3.1 organization_events表创建 (Day 1-2)
```sql
-- 创建组织事件表
CREATE TABLE organization_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 约束
    CONSTRAINT chk_event_type CHECK (
        event_type IN ('CREATE', 'UPDATE', 'RESTRUCTURE', 'DISSOLVE', 'ACTIVATE', 'DEACTIVATE')
    ),
    CONSTRAINT chk_end_date_after_effective CHECK (
        end_date IS NULL OR end_date > effective_date
    ),
    
    -- 外键约束
    CONSTRAINT fk_org_events_org FOREIGN KEY (organization_code) 
        REFERENCES organization_units(code) ON DELETE RESTRICT
);

-- 创建索引
CREATE INDEX idx_org_events_code ON organization_events(organization_code);
CREATE INDEX idx_org_events_type ON organization_events(event_type);
CREATE INDEX idx_org_events_date ON organization_events(effective_date);
CREATE INDEX idx_org_events_tenant ON organization_events(tenant_id);

-- 为event_data创建GIN索引支持JSON查询
CREATE INDEX idx_org_events_data_gin ON organization_events USING GIN (event_data);
```

#### 3.2 organization_versions表创建 (Day 3-4)
```sql
-- 创建版本历史表
CREATE TABLE organization_versions (
    version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    version INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    snapshot_data JSONB NOT NULL,
    change_reason VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    -- 唯一约束
    CONSTRAINT uk_org_version UNIQUE (organization_code, version),
    
    -- 检查约束
    CONSTRAINT chk_version_positive CHECK (version > 0),
    CONSTRAINT chk_snapshot_not_empty CHECK (snapshot_data != '{}'::jsonb)
);

-- 创建索引
CREATE INDEX idx_org_versions_code_version ON organization_versions(organization_code, version);
CREATE INDEX idx_org_versions_effective ON organization_versions(effective_date);
CREATE INDEX idx_org_versions_tenant ON organization_versions(tenant_id);
```

#### 3.3 数据一致性验证 (Day 5)
```sql
-- 创建数据一致性检查函数
CREATE OR REPLACE FUNCTION validate_temporal_consistency()
RETURNS TABLE (
    organization_code VARCHAR(10),
    issue_type VARCHAR(50), 
    description TEXT
) AS $$
BEGIN
    -- 检查时间线间隙
    RETURN QUERY
    SELECT 
        o1.code,
        'TIMELINE_GAP'::VARCHAR(50),
        format('版本%s结束日期%s与版本%s生效日期%s之间存在间隙', 
               o1.version, o1.end_date, o2.version, o2.effective_date)
    FROM organization_units o1
    JOIN organization_units o2 ON o1.code = o2.code
    WHERE o1.version < o2.version
      AND o1.end_date IS NOT NULL
      AND o1.end_date + INTERVAL '1 day' != o2.effective_date;
    
    -- 检查重叠版本
    RETURN QUERY  
    SELECT 
        o1.code,
        'VERSION_OVERLAP'::VARCHAR(50),
        format('版本%s与版本%s存在时间重叠', o1.version, o2.version)
    FROM organization_units o1
    JOIN organization_units o2 ON o1.code = o2.code
    WHERE o1.version != o2.version
      AND o1.effective_date < COALESCE(o2.end_date, CURRENT_DATE + INTERVAL '100 years')
      AND COALESCE(o1.end_date, CURRENT_DATE + INTERVAL '100 years') > o2.effective_date;
      
    -- 检查当前版本标记
    RETURN QUERY
    SELECT 
        code,
        'MULTIPLE_CURRENT'::VARCHAR(50),
        format('存在多个当前版本：%s', string_agg(version::text, ','))
    FROM organization_units 
    WHERE is_current = true
    GROUP BY code
    HAVING COUNT(*) > 1;
END;
$$ LANGUAGE plpgsql;

-- 执行一致性检查
SELECT * FROM validate_temporal_consistency();
```

### Week 4: 应用程序适配与兼容性

#### 4.1 现有API兼容性保护 (Day 1-3)
```go
// 扩展现有Organization结构体，保持向后兼容
type Organization struct {
    // 现有字段保持不变
    TenantID    string    `json:"tenant_id" db:"tenant_id"`
    Code        string    `json:"code" db:"code"`
    ParentCode  *string   `json:"parent_code,omitempty" db:"parent_code"`
    Name        string    `json:"name" db:"name"`
    UnitType    string    `json:"unit_type" db:"unit_type"`
    Status      string    `json:"status" db:"status"`
    Level       int       `json:"level" db:"level"`
    Path        string    `json:"path" db:"path"`
    SortOrder   int       `json:"sort_order" db:"sort_order"`
    Description string    `json:"description" db:"description"`
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
    
    // 新增时态字段（可选返回，保证兼容性）
    EffectiveDate     *time.Time `json:"effective_date,omitempty" db:"effective_date"`
    EndDate           *time.Time `json:"end_date,omitempty" db:"end_date"`
    Version           *int       `json:"version,omitempty" db:"version"`
    SupersedesVersion *int       `json:"supersedes_version,omitempty" db:"supersedes_version"`
    ChangeReason      *string    `json:"change_reason,omitempty" db:"change_reason"`
    IsCurrent         *bool      `json:"is_current,omitempty" db:"is_current"`
}

// 兼容性查询函数 - 默认只返回当前版本
func (r *OrganizationRepository) GetByCodeCompatible(ctx context.Context, tenantID uuid.UUID, code string) (*Organization, error) {
    query := `
        SELECT tenant_id, code, parent_code, name, unit_type, status,
               level, path, sort_order, description, created_at, updated_at
               -- 时态字段默认不返回，保证现有API兼容性
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2 AND is_current = true
    `
    
    var org Organization
    err := r.db.QueryRowContext(ctx, query, tenantID.String(), code).Scan(
        &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
        &org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
        &org.Description, &org.CreatedAt, &org.UpdatedAt,
    )
    
    return &org, err
}

// 新的时态查询函数
func (r *OrganizationRepository) GetByCodeTemporal(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) (*Organization, error) {
    // 实现时态查询逻辑
    // 支持as_of_date, version等参数
}
```

#### 4.2 配置管理和环境变量 (Day 4)
```go
// 添加时态管理配置
type TemporalConfig struct {
    Enabled                    bool   `env:"TEMPORAL_MANAGEMENT_ENABLED" envDefault:"true"`
    AutoEndDateManagement      bool   `env:"AUTO_END_DATE_MANAGEMENT" envDefault:"true"`
    TimelineConsistencyPolicy  string `env:"TIMELINE_CONSISTENCY_POLICY" envDefault:"NO_GAPS_ALLOWED"`
    SupportsRetroactivity      bool   `env:"SUPPORTS_RETROACTIVITY" envDefault:"true"`
    MaxRetroactiveDays         int    `env:"MAX_RETROACTIVE_DAYS" envDefault:"365"`
    DefaultQueryMode           string `env:"DEFAULT_QUERY_MODE" envDefault:"CURRENT_ONLY"`
}

// 环境配置文件更新
// .env 文件
TEMPORAL_MANAGEMENT_ENABLED=true
AUTO_END_DATE_MANAGEMENT=true
TIMELINE_CONSISTENCY_POLICY=NO_GAPS_ALLOWED
DEFAULT_QUERY_MODE=CURRENT_ONLY
```

#### 4.3 集成测试与验证 (Day 5)
```bash
#!/bin/bash
# 阶段1集成测试脚本

echo "=== 阶段1：数据模型扩展集成测试 ==="

# 1. 数据库连接测试
echo "1. 测试数据库连接..."
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT version();"

# 2. 表结构验证
echo "2. 验证表结构扩展..."
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
\d organization_units;
\d organization_events; 
\d organization_versions;
"

# 3. 触发器功能测试
echo "3. 测试结束日期自动管理触发器..."
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "
BEGIN;
INSERT INTO organization_units (code, name, unit_type, tenant_id, effective_date, version)
VALUES ('TEST999', '集成测试部门V1', 'DEPARTMENT', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '2025-01-01', 1);

INSERT INTO organization_units (code, name, unit_type, tenant_id, effective_date, version)  
VALUES ('TEST999', '集成测试部门V2', 'DEPARTMENT', '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9', '2025-06-01', 2);

SELECT '触发器测试结果：' as test_type, code, version, effective_date, end_date, is_current 
FROM organization_units WHERE code = 'TEST999' ORDER BY version;
ROLLBACK;
"

# 4. API兼容性测试
echo "4. 测试API兼容性..."
curl -X GET "http://localhost:9090/api/v1/organization-units/1000001" \
  -H "Content-Type: application/json" \
  -w "HTTP Status: %{http_code}\n"

# 5. 数据一致性检查
echo "5. 执行数据一致性检查..."
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT * FROM validate_temporal_consistency();"

echo "=== 阶段1测试完成 ==="
```

---

## 🔧 阶段2：API扩展与时态查询 (Week 5-7)

### Week 5: 时态查询API开发

#### 5.1 时态查询参数设计 (Day 1-2)
```go
// 时态查询选项
type TemporalQueryOptions struct {
    AsOfDate        *time.Time `json:"as_of_date,omitempty"`        // 时间点查询
    EffectiveFrom   *time.Time `json:"effective_from,omitempty"`    // 生效起始时间
    EffectiveTo     *time.Time `json:"effective_to,omitempty"`      // 生效结束时间  
    IncludeHistory  bool       `json:"include_history,omitempty"`   // 包含历史版本
    IncludeFuture   bool       `json:"include_future,omitempty"`    // 包含未来版本
    IncludeDissolved bool      `json:"include_dissolved,omitempty"` // 包含已解散组织
    Version         *int       `json:"version,omitempty"`           // 特定版本
    MaxVersions     int        `json:"max_versions,omitempty"`      // 最大版本数量
}

// HTTP查询参数解析
func ParseTemporalQuery(r *http.Request) (*TemporalQueryOptions, error) {
    opts := &TemporalQueryOptions{}
    
    // 解析as_of_date参数
    if asOfStr := r.URL.Query().Get("as_of_date"); asOfStr != "" {
        if asOfDate, err := time.Parse("2006-01-02", asOfStr); err == nil {
            opts.AsOfDate = &asOfDate
        } else {
            return nil, fmt.Errorf("无效的as_of_date格式，期望：YYYY-MM-DD")
        }
    }
    
    // 解析日期范围
    if fromStr := r.URL.Query().Get("effective_from"); fromStr != "" {
        if from, err := time.Parse("2006-01-02", fromStr); err == nil {
            opts.EffectiveFrom = &from
        }
    }
    
    if toStr := r.URL.Query().Get("effective_to"); toStr != "" {
        if to, err := time.Parse("2006-01-02", toStr); err == nil {
            opts.EffectiveTo = &to
        }
    }
    
    // 解析布尔参数
    opts.IncludeHistory = r.URL.Query().Get("include_history") == "true"
    opts.IncludeFuture = r.URL.Query().Get("include_future") == "true" 
    opts.IncludeDissolved = r.URL.Query().Get("include_dissolved") == "true"
    
    // 解析版本参数
    if versionStr := r.URL.Query().Get("version"); versionStr != "" {
        if version, err := strconv.Atoi(versionStr); err == nil {
            opts.Version = &version
        }
    }
    
    return opts, nil
}
```

#### 5.2 时态查询数据库层实现 (Day 3-4)
```go
// 时态查询仓储实现
func (r *OrganizationRepository) GetByCodeTemporal(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) ([]*Organization, error) {
    var conditions []string
    var args []interface{}
    argIndex := 1
    
    // 基础条件
    conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argIndex))
    args = append(args, tenantID.String())
    argIndex++
    
    conditions = append(conditions, fmt.Sprintf("code = $%d", argIndex))
    args = append(args, code)
    argIndex++
    
    // 时间点查询：查询在指定日期有效的版本
    if opts.AsOfDate != nil {
        conditions = append(conditions, fmt.Sprintf(
            "effective_date <= $%d AND (end_date IS NULL OR end_date >= $%d)", 
            argIndex, argIndex))
        args = append(args, *opts.AsOfDate)
        argIndex++
    }
    
    // 日期范围查询
    if opts.EffectiveFrom != nil {
        conditions = append(conditions, fmt.Sprintf("effective_date >= $%d", argIndex))
        args = append(args, *opts.EffectiveFrom)
        argIndex++
    }
    
    if opts.EffectiveTo != nil {
        conditions = append(conditions, fmt.Sprintf("effective_date <= $%d", argIndex))
        args = append(args, *opts.EffectiveTo)
        argIndex++
    }
    
    // 特定版本查询
    if opts.Version != nil {
        conditions = append(conditions, fmt.Sprintf("version = $%d", argIndex))
        args = append(args, *opts.Version)
        argIndex++
    }
    
    // 当前版本过滤
    if !opts.IncludeHistory && opts.AsOfDate == nil && opts.Version == nil {
        conditions = append(conditions, "is_current = true")
    }
    
    // 未来版本过滤
    if !opts.IncludeFuture {
        conditions = append(conditions, "effective_date <= CURRENT_DATE")
    }
    
    // 已解散组织过滤
    if !opts.IncludeDissolved {
        conditions = append(conditions, "(end_date IS NULL OR end_date > CURRENT_DATE)")
    }
    
    // 构建查询
    query := fmt.Sprintf(`
        SELECT tenant_id, code, parent_code, name, unit_type, status,
               level, path, sort_order, description, created_at, updated_at,
               effective_date, end_date, version, supersedes_version, change_reason, is_current
        FROM organization_units 
        WHERE %s
        ORDER BY version DESC
        %s
    `, strings.Join(conditions, " AND "), 
       func() string {
           if opts.MaxVersions > 0 {
               return fmt.Sprintf("LIMIT %d", opts.MaxVersions)
           }
           return ""
       }())
    
    rows, err := r.db.QueryContext(ctx, query, args...)
    if err != nil {
        return nil, fmt.Errorf("时态查询失败: %w", err)
    }
    defer rows.Close()
    
    var organizations []*Organization
    for rows.Next() {
        org := &Organization{}
        err := rows.Scan(
            &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
            &org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
            &org.Description, &org.CreatedAt, &org.UpdatedAt,
            &org.EffectiveDate, &org.EndDate, &org.Version, &org.SupersedesVersion,
            &org.ChangeReason, &org.IsCurrent,
        )
        if err != nil {
            return nil, fmt.Errorf("扫描时态查询结果失败: %w", err)
        }
        organizations = append(organizations, org)
    }
    
    return organizations, nil
}
```

#### 5.3 时态查询API端点实现 (Day 5)
```go
// 时态查询API处理器
func (h *OrganizationHandler) GetOrganizationTemporal(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    if code == "" {
        h.writeErrorResponse(w, http.StatusBadRequest, "MISSING_CODE", "缺少组织代码", nil)
        return
    }
    
    // 解析时态查询参数
    opts, err := ParseTemporalQuery(r)
    if err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_TEMPORAL_PARAMS", "时态查询参数无效", err)
        return
    }
    
    tenantID := h.getTenantID(r)
    
    // 执行时态查询
    organizations, err := h.repo.GetByCodeTemporal(r.Context(), tenantID, code, opts)
    if err != nil {
        monitoring.RecordOrganizationOperation("temporal_get", "failed", "command-service")
        h.writeErrorResponse(w, http.StatusInternalServerError, "TEMPORAL_QUERY_ERROR", "时态查询失败", err)
        return
    }
    
    if len(organizations) == 0 {
        h.writeErrorResponse(w, http.StatusNotFound, "NOT_FOUND", "未找到匹配的组织版本", nil)
        return
    }
    
    // 构建响应
    response := map[string]interface{}{
        "organizations": organizations,
        "query_options": opts,
        "result_count": len(organizations),
    }
    
    monitoring.RecordOrganizationOperation("temporal_get", "success", "command-service")
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

### Week 6: 历史版本和事件API

#### 6.1 历史版本查询API (Day 1-2)
```go
// 历史版本查询处理器
func (h *OrganizationHandler) GetOrganizationHistory(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    tenantID := h.getTenantID(r)
    
    // 查询所有历史版本
    query := `
        SELECT o.tenant_id, o.code, o.parent_code, o.name, o.unit_type, o.status,
               o.level, o.path, o.sort_order, o.description, o.created_at, o.updated_at,
               o.effective_date, o.end_date, o.version, o.supersedes_version, 
               o.change_reason, o.is_current,
               e.event_type, e.event_data, e.created_by as changed_by
        FROM organization_units o
        LEFT JOIN organization_events e ON o.code = e.organization_code 
            AND o.effective_date = e.effective_date
        WHERE o.tenant_id = $1 AND o.code = $2
        ORDER BY o.version ASC
    `
    
    rows, err := h.repo.db.QueryContext(r.Context(), query, tenantID.String(), code)
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "HISTORY_QUERY_ERROR", "查询历史版本失败", err)
        return
    }
    defer rows.Close()
    
    var history []map[string]interface{}
    for rows.Next() {
        var org Organization
        var eventType, changedBy sql.NullString
        var eventData sql.NullString
        
        err := rows.Scan(
            &org.TenantID, &org.Code, &org.ParentCode, &org.Name,
            &org.UnitType, &org.Status, &org.Level, &org.Path, &org.SortOrder,
            &org.Description, &org.CreatedAt, &org.UpdatedAt,
            &org.EffectiveDate, &org.EndDate, &org.Version, &org.SupersedesVersion,
            &org.ChangeReason, &org.IsCurrent, &eventType, &eventData, &changedBy,
        )
        if err != nil {
            h.writeErrorResponse(w, http.StatusInternalServerError, "SCAN_ERROR", "扫描历史记录失败", err)
            return
        }
        
        historyItem := map[string]interface{}{
            "organization": org,
            "event_type":   eventType.String,
            "changed_by":   changedBy.String,
        }
        
        if eventData.Valid {
            var data map[string]interface{}
            json.Unmarshal([]byte(eventData.String), &data)
            historyItem["changes"] = data
        }
        
        history = append(history, historyItem)
    }
    
    response := map[string]interface{}{
        "code":     code,
        "history":  history,
        "versions": len(history),
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

#### 6.2 事件驱动变更API (Day 3-5)
```go
// 组织变更事件请求
type OrganizationChangeEvent struct {
    EventType     string                 `json:"event_type"`      // CREATE, UPDATE, RESTRUCTURE, DISSOLVE
    EffectiveDate time.Time              `json:"effective_date"`  // 生效日期
    EndDate       *time.Time             `json:"end_date,omitempty"` // 结束日期(特殊场景)
    ChangeData    map[string]interface{} `json:"change_data"`     // 变更内容
    ChangeReason  string                 `json:"change_reason"`   // 变更原因
}

// 事件驱动变更处理器
func (h *OrganizationHandler) CreateOrganizationEvent(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    
    var req OrganizationChangeEvent
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
        return
    }
    
    // 验证事件类型
    validEventTypes := map[string]bool{
        "CREATE": true, "UPDATE": true, "RESTRUCTURE": true, "DISSOLVE": true,
        "ACTIVATE": true, "DEACTIVATE": true,
    }
    if !validEventTypes[req.EventType] {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_EVENT_TYPE", "无效的事件类型", nil)
        return
    }
    
    tenantID := h.getTenantID(r)
    
    // 开始事务
    tx, err := h.repo.db.BeginTx(r.Context(), nil)
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
        return
    }
    defer tx.Rollback()
    
    // 1. 记录事件
    eventData, _ := json.Marshal(req.ChangeData)
    eventID, err := h.createOrganizationEvent(r.Context(), tx, &OrganizationEvent{
        OrganizationCode: code,
        EventType:        req.EventType,
        EventData:        eventData,
        EffectiveDate:    req.EffectiveDate,
        EndDate:          req.EndDate,
        CreatedBy:        "system", // 从认证上下文获取
        TenantID:         tenantID.String(),
    })
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_CREATE_ERROR", "创建事件失败", err)
        return
    }
    
    // 2. 处理不同类型的事件
    switch req.EventType {
    case "UPDATE", "RESTRUCTURE":
        err = h.handleUpdateEvent(r.Context(), tx, tenantID, code, &req)
    case "DISSOLVE":
        err = h.handleDissolveEvent(r.Context(), tx, tenantID, code, &req)
    case "ACTIVATE", "DEACTIVATE":
        err = h.handleStatusEvent(r.Context(), tx, tenantID, code, &req)
    default:
        err = fmt.Errorf("未支持的事件类型: %s", req.EventType)
    }
    
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_PROCESS_ERROR", "处理事件失败", err)
        return
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
        return
    }
    
    response := map[string]interface{}{
        "event_id":       eventID,
        "event_type":     req.EventType,
        "organization":   code,
        "effective_date": req.EffectiveDate,
        "status":         "processed",
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}

// 处理更新事件
func (h *OrganizationHandler) handleUpdateEvent(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *OrganizationChangeEvent) error {
    // 获取当前版本
    currentOrg, err := h.getCurrentVersion(ctx, tx, tenantID, code)
    if err != nil {
        return fmt.Errorf("获取当前版本失败: %w", err)
    }
    
    // 创建新版本
    newVersion := currentOrg.Version + 1
    
    // 应用变更数据
    updatedOrg := *currentOrg
    updatedOrg.Version = newVersion
    updatedOrg.EffectiveDate = &req.EffectiveDate
    updatedOrg.EndDate = req.EndDate
    updatedOrg.ChangeReason = &req.ChangeReason
    updatedOrg.SupersedesVersion = &currentOrg.Version
    
    // 应用具体的字段变更
    for field, value := range req.ChangeData {
        switch field {
        case "name":
            if name, ok := value.(string); ok {
                updatedOrg.Name = name
            }
        case "unit_type":
            if unitType, ok := value.(string); ok {
                updatedOrg.UnitType = unitType
            }
        case "status":
            if status, ok := value.(string); ok {
                updatedOrg.Status = status
            }
        case "description":
            if desc, ok := value.(string); ok {
                updatedOrg.Description = desc
            }
        }
    }
    
    // 插入新版本（触发器会自动处理end_date）
    _, err = h.insertNewVersion(ctx, tx, &updatedOrg)
    return err
}
```

### Week 7: 时态查询优化与测试

#### 7.1 查询性能优化 (Day 1-3)
```sql
-- 创建时态查询专用索引
CREATE INDEX CONCURRENTLY idx_org_temporal_query 
ON organization_units(tenant_id, code, effective_date, end_date) 
WHERE is_current = true;

-- 时间点查询优化索引
CREATE INDEX CONCURRENTLY idx_org_as_of_date 
ON organization_units(code, effective_date DESC, end_date) 
WHERE tenant_id = '3b99930c-4dc6-4cc9-8e4d-7d960a931cb9';

-- 创建时态查询视图
CREATE VIEW organization_current_view AS
SELECT * FROM organization_units 
WHERE is_current = true;

CREATE VIEW organization_historical_view AS  
SELECT 
    o.*,
    e.event_type,
    e.created_by as changed_by,
    e.created_at as event_created_at
FROM organization_units o
LEFT JOIN organization_events e ON o.code = e.organization_code 
    AND o.effective_date = e.effective_date
ORDER BY o.code, o.version;
```

#### 7.2 时态查询缓存策略 (Day 4)
```go
// 时态查询缓存键生成
func generateTemporalCacheKey(tenantID string, code string, opts *TemporalQueryOptions) string {
    var keyParts []string
    keyParts = append(keyParts, "org_temporal", tenantID, code)
    
    if opts.AsOfDate != nil {
        keyParts = append(keyParts, "as_of", opts.AsOfDate.Format("2006-01-02"))
    }
    if opts.Version != nil {
        keyParts = append(keyParts, "version", strconv.Itoa(*opts.Version))
    }
    if opts.IncludeHistory {
        keyParts = append(keyParts, "with_history")
    }
    
    return strings.Join(keyParts, ":")
}

// 带缓存的时态查询
func (r *OrganizationRepository) GetByCodeTemporalCached(ctx context.Context, tenantID uuid.UUID, code string, opts *TemporalQueryOptions) ([]*Organization, error) {
    cacheKey := generateTemporalCacheKey(tenantID.String(), code, opts)
    
    // 尝试从缓存获取
    if cached := r.cache.Get(cacheKey); cached != nil {
        if orgs, ok := cached.([]*Organization); ok {
            return orgs, nil
        }
    }
    
    // 缓存未命中，执行数据库查询
    orgs, err := r.GetByCodeTemporal(ctx, tenantID, code, opts)
    if err != nil {
        return nil, err
    }
    
    // 缓存结果（时态查询结果相对稳定，可以较长时间缓存）
    cacheDuration := time.Hour * 1
    if opts.AsOfDate != nil && opts.AsOfDate.Before(time.Now().AddDate(0, 0, -7)) {
        // 历史查询缓存更长时间
        cacheDuration = time.Hour * 24
    }
    
    r.cache.Set(cacheKey, orgs, cacheDuration)
    return orgs, nil
}
```

#### 7.3 阶段2集成测试 (Day 5)
```bash
#!/bin/bash
# 阶段2集成测试脚本

echo "=== 阶段2：时态查询API集成测试 ==="

BASE_URL="http://localhost:9090/api/v1/organization-units"

# 1. 基础时态查询测试
echo "1. 测试当前版本查询..."
curl -X GET "${BASE_URL}/1000001" \
  -H "Content-Type: application/json" \
  -w "HTTP Status: %{http_code}\n"

# 2. 时间点查询测试
echo "2. 测试时间点查询..."
curl -X GET "${BASE_URL}/1000001?as_of_date=2025-01-01" \
  -H "Content-Type: application/json" \
  -w "HTTP Status: %{http_code}\n"

# 3. 历史版本查询测试
echo "3. 测试历史版本查询..."
curl -X GET "${BASE_URL}/1000001/history" \
  -H "Content-Type: application/json" \
  -w "HTTP Status: %{http_code}\n"

# 4. 事件驱动变更测试
echo "4. 测试事件驱动变更..."
curl -X POST "${BASE_URL}/1000001/events" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "UPDATE",
    "effective_date": "2025-09-01",
    "change_data": {
      "name": "技术研发部",
      "description": "负责产品研发和技术创新"
    },
    "change_reason": "部门职能调整"
  }' \
  -w "HTTP Status: %{http_code}\n"

# 5. 性能测试
echo "5. 时态查询性能测试..."
ab -n 100 -c 10 "${BASE_URL}/1000001?as_of_date=2025-01-01"

echo "=== 阶段2测试完成 ==="
```

---

## ⚙️ 阶段3：事件驱动重构与合规 (Week 8-13)

### Week 8-9: 时间线一致性验证系统

#### 8.1 时间线验证引擎 (Week 8 Day 1-3)
```go
// 时间线一致性验证引擎
type TimelineValidator struct {
    db     *sql.DB
    logger *log.Logger
    config *TimelineValidationConfig
}

type TimelineValidationConfig struct {
    Policy                    string   `json:"policy"` // NO_GAPS_ALLOWED, CONTINUOUS_HISTORY
    AllowManualEndDate       bool     `json:"allow_manual_end_date"`
    MaxRetroactiveDays       int      `json:"max_retroactive_days"`
    RequireChangeReason      bool     `json:"require_change_reason"`
    RestrictedEventTypes     []string `json:"restricted_event_types"`
}

// 验证时间线一致性
func (tv *TimelineValidator) ValidateTimeline(ctx context.Context, orgCode string, newVersion *Organization) error {
    // 1. 获取现有时间线
    timeline, err := tv.getTimeline(ctx, orgCode)
    if err != nil {
        return fmt.Errorf("获取时间线失败: %w", err)
    }
    
    // 2. 验证新版本插入位置
    if err := tv.validateInsertionPoint(timeline, newVersion); err != nil {
        return fmt.Errorf("插入点验证失败: %w", err)
    }
    
    // 3. 验证时间线策略
    switch tv.config.Policy {
    case "NO_GAPS_ALLOWED":
        if err := tv.validateNoGaps(timeline, newVersion); err != nil {
            return err
        }
    case "CONTINUOUS_HISTORY":
        if err := tv.validateContinuousHistory(timeline, newVersion); err != nil {
            return err
        }
    }
    
    // 4. 验证追溯处理限制
    if newVersion.EffectiveDate.Before(time.Now().AddDate(0, 0, -tv.config.MaxRetroactiveDays)) {
        return fmt.Errorf("超出最大追溯天数限制: %d天", tv.config.MaxRetroactiveDays)
    }
    
    return nil
}

// 获取组织时间线
func (tv *TimelineValidator) getTimeline(ctx context.Context, orgCode string) ([]*Organization, error) {
    query := `
        SELECT code, version, effective_date, end_date, is_current, change_reason
        FROM organization_units
        WHERE code = $1
        ORDER BY effective_date ASC, version ASC
    `
    
    rows, err := tv.db.QueryContext(ctx, query, orgCode)
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var timeline []*Organization
    for rows.Next() {
        org := &Organization{}
        err := rows.Scan(&org.Code, &org.Version, &org.EffectiveDate, 
                        &org.EndDate, &org.IsCurrent, &org.ChangeReason)
        if err != nil {
            return nil, err
        }
        timeline = append(timeline, org)
    }
    
    return timeline, nil
}

// 验证无间隙策略
func (tv *TimelineValidator) validateNoGaps(timeline []*Organization, newVersion *Organization) error {
    // 查找插入位置前后的版本
    var prevVersion, nextVersion *Organization
    
    for i, version := range timeline {
        if version.EffectiveDate.After(*newVersion.EffectiveDate) {
            nextVersion = version
            if i > 0 {
                prevVersion = timeline[i-1]
            }
            break
        }
        prevVersion = version
    }
    
    // 验证与前一版本的连续性
    if prevVersion != nil && prevVersion.EndDate != nil {
        expectedStart := prevVersion.EndDate.AddDate(0, 0, 1)
        if !newVersion.EffectiveDate.Equal(expectedStart) {
            return fmt.Errorf("时间线间隙：新版本生效日期应为 %s", expectedStart.Format("2006-01-02"))
        }
    }
    
    // 验证与后一版本的连续性
    if nextVersion != nil && newVersion.EndDate != nil {
        expectedNext := newVersion.EndDate.AddDate(0, 0, 1)
        if !nextVersion.EffectiveDate.Equal(expectedNext) {
            return fmt.Errorf("时间线间隙：后续版本生效日期应为 %s", expectedNext.Format("2006-01-02"))
        }
    }
    
    return nil
}
```

#### 8.2 结束日期管理规则引擎 (Week 8 Day 4-5)
```go
// 结束日期管理规则引擎
type EndDateRuleEngine struct {
    rules []EndDateRule
    db    *sql.DB
}

type EndDateRule struct {
    ID          string
    Name        string
    Condition   func(ctx context.Context, org *Organization, event *OrganizationEvent) bool
    Action      func(ctx context.Context, tx *sql.Tx, org *Organization, event *OrganizationEvent) error
    Priority    int
    Description string
}

// 初始化默认规则
func NewEndDateRuleEngine(db *sql.DB) *EndDateRuleEngine {
    engine := &EndDateRuleEngine{db: db}
    
    // 规则1: 正常版本更新自动设置end_date
    engine.AddRule(EndDateRule{
        ID:   "AUTO_SET_END_DATE_ON_UPDATE",
        Name: "自动设置结束日期",
        Condition: func(ctx context.Context, org *Organization, event *OrganizationEvent) bool {
            return event.EventType == "UPDATE" || event.EventType == "RESTRUCTURE"
        },
        Action: func(ctx context.Context, tx *sql.Tx, org *Organization, event *OrganizationEvent) error {
            // 自动设置前版本的end_date为新版本effective_date - 1天
            previousEndDate := event.EffectiveDate.AddDate(0, 0, -1)
            _, err := tx.ExecContext(ctx,
                "UPDATE organization_units SET end_date = $1, is_current = false WHERE code = $2 AND is_current = true",
                previousEndDate, org.Code)
            return err
        },
        Priority: 1,
        Description: "当创建新版本时，自动设置前一版本的结束日期",
    })
    
    // 规则2: 组织解散明确设置end_date
    engine.AddRule(EndDateRule{
        ID:   "EXPLICIT_END_DATE_ON_DISSOLVE",
        Name: "解散时明确设置结束日期",
        Condition: func(ctx context.Context, org *Organization, event *OrganizationEvent) bool {
            return event.EventType == "DISSOLVE"
        },
        Action: func(ctx context.Context, tx *sql.Tx, org *Organization, event *OrganizationEvent) error {
            endDate := event.EndDate
            if endDate == nil {
                // 默认使用生效日期作为结束日期
                endDate = &event.EffectiveDate
            }
            
            _, err := tx.ExecContext(ctx,
                "UPDATE organization_units SET end_date = $1, status = 'INACTIVE', is_current = false WHERE code = $2 AND is_current = true",
                *endDate, org.Code)
            return err
        },
        Priority: 2,
        Description: "组织解散时明确设置结束日期并更新状态",
    })
    
    // 规则3: 追溯修正重新计算后续版本
    engine.AddRule(EndDateRule{
        ID:   "RECALCULATE_ON_RETROACTIVE",
        Name: "追溯修正重新计算",
        Condition: func(ctx context.Context, org *Organization, event *OrganizationEvent) bool {
            return event.EffectiveDate.Before(time.Now()) && 
                   hasSubsequentVersions(ctx, tx, org.Code, event.EffectiveDate)
        },
        Action: func(ctx context.Context, tx *sql.Tx, org *Organization, event *OrganizationEvent) error {
            return recalculateSubsequentTimeline(ctx, tx, org.Code, event.EffectiveDate)
        },
        Priority: 3,
        Description: "追溯修正时重新计算所有后续版本的时间范围",
    })
    
    return engine
}

// 执行规则引擎
func (ere *EndDateRuleEngine) ProcessEndDate(ctx context.Context, tx *sql.Tx, org *Organization, event *OrganizationEvent) error {
    // 按优先级排序规则
    sort.Slice(ere.rules, func(i, j int) bool {
        return ere.rules[i].Priority < ere.rules[j].Priority
    })
    
    // 执行匹配的规则
    for _, rule := range ere.rules {
        if rule.Condition(ctx, org, event) {
            log.Printf("执行结束日期规则: %s (%s)", rule.Name, rule.ID)
            if err := rule.Action(ctx, tx, org, event); err != nil {
                return fmt.Errorf("执行规则 %s 失败: %w", rule.Name, err)
            }
            // 只执行第一个匹配的规则
            break
        }
    }
    
    return nil
}
```

### Week 9: 高级时态功能

#### 9.1 未来变更规划API (Day 1-3)
```go
// 未来变更规划请求
type FuturePlanRequest struct {
    PlannedChanges []PlannedChange `json:"planned_changes"`
    Reason         string          `json:"reason"`
    CreatedBy      string          `json:"created_by"`
}

type PlannedChange struct {
    EffectiveDate time.Time              `json:"effective_date"`
    EndDate       *time.Time             `json:"end_date,omitempty"`
    Changes       map[string]interface{} `json:"changes"`
    EventType     string                 `json:"event_type"`
}

// 未来变更规划处理器
func (h *OrganizationHandler) PlanFutureChanges(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    
    var req FuturePlanRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
        return
    }
    
    tenantID := h.getTenantID(r)
    
    // 验证规划的变更
    if err := h.validateFuturePlan(req.PlannedChanges); err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_PLAN", "变更规划无效", err)
        return
    }
    
    // 开始事务
    tx, err := h.repo.db.BeginTx(r.Context(), nil)
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
        return
    }
    defer tx.Rollback()
    
    var createdEvents []string
    
    // 创建未来变更版本
    for _, change := range req.PlannedChanges {
        // 获取基础版本（当前或上一个计划版本）
        baseVersion, err := h.getBaseVersionForPlan(r.Context(), tx, tenantID, code, change.EffectiveDate)
        if err != nil {
            h.writeErrorResponse(w, http.StatusInternalServerError, "BASE_VERSION_ERROR", "获取基础版本失败", err)
            return
        }
        
        // 创建未来版本
        futureVersion := *baseVersion
        futureVersion.Version = h.getNextVersion(r.Context(), tx, code)
        futureVersion.EffectiveDate = &change.EffectiveDate
        futureVersion.EndDate = change.EndDate
        futureVersion.ChangeReason = &req.Reason
        futureVersion.IsCurrent = &[]bool{false}[0] // 未来版本不是当前版本
        
        // 应用计划的变更
        h.applyPlannedChanges(&futureVersion, change.Changes)
        
        // 插入未来版本
        if err := h.insertFutureVersion(r.Context(), tx, &futureVersion); err != nil {
            h.writeErrorResponse(w, http.StatusInternalServerError, "INSERT_ERROR", "插入未来版本失败", err)
            return
        }
        
        // 记录规划事件
        eventID, err := h.createPlanningEvent(r.Context(), tx, code, &change, req.Reason, tenantID.String())
        if err != nil {
            h.writeErrorResponse(w, http.StatusInternalServerError, "EVENT_ERROR", "创建规划事件失败", err)
            return
        }
        
        createdEvents = append(createdEvents, eventID)
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
        return
    }
    
    response := map[string]interface{}{
        "organization":    code,
        "planned_changes": len(req.PlannedChanges),
        "events_created":  createdEvents,
        "status":          "planned",
    }
    
    w.Header().Set("Content-Type: application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(response)
}
```

#### 9.2 时间线操作API (Day 4-5)
```go
// 时间线操作请求
type TimelineOperationRequest struct {
    Operation     string     `json:"operation"`      // CORRECT, CANCEL, VOID
    TargetDate    time.Time  `json:"target_date"`    // 操作目标日期
    TargetVersion *int       `json:"target_version,omitempty"` // 目标版本
    NewData       map[string]interface{} `json:"new_data,omitempty"` // 校正数据
    Reason        string     `json:"reason"`         // 操作原因
}

// 时间线操作处理器
func (h *OrganizationHandler) ExecuteTimelineOperation(w http.ResponseWriter, r *http.Request) {
    code := chi.URLParam(r, "code")
    operation := chi.URLParam(r, "operation") // correct, cancel, void
    
    var req TimelineOperationRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        h.writeErrorResponse(w, http.StatusBadRequest, "INVALID_REQUEST", "请求格式无效", err)
        return
    }
    
    req.Operation = strings.ToUpper(operation)
    tenantID := h.getTenantID(r)
    
    // 验证操作权限
    if !h.hasTimelineOperationPermission(r, req.Operation) {
        h.writeErrorResponse(w, http.StatusForbidden, "PERMISSION_DENIED", "无权限执行时间线操作", nil)
        return
    }
    
    // 开始事务
    tx, err := h.repo.db.BeginTx(r.Context(), nil)
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "TRANSACTION_ERROR", "开始事务失败", err)
        return
    }
    defer tx.Rollback()
    
    var result map[string]interface{}
    
    switch req.Operation {
    case "CORRECT":
        result, err = h.executeCorrection(r.Context(), tx, tenantID, code, &req)
    case "CANCEL":
        result, err = h.executeCancellation(r.Context(), tx, tenantID, code, &req)
    case "VOID":
        result, err = h.executeVoid(r.Context(), tx, tenantID, code, &req)
    default:
        err = fmt.Errorf("不支持的时间线操作: %s", req.Operation)
    }
    
    if err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "OPERATION_ERROR", "时间线操作失败", err)
        return
    }
    
    // 提交事务
    if err := tx.Commit(); err != nil {
        h.writeErrorResponse(w, http.StatusInternalServerError, "COMMIT_ERROR", "提交事务失败", err)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(result)
}

// 执行历史校正
func (h *OrganizationHandler) executeCorrection(ctx context.Context, tx *sql.Tx, tenantID uuid.UUID, code string, req *TimelineOperationRequest) (map[string]interface{}, error) {
    // 1. 查找目标版本
    targetVersion, err := h.getVersionByDate(ctx, tx, tenantID, code, req.TargetDate)
    if err != nil {
        return nil, fmt.Errorf("查找目标版本失败: %w", err)
    }
    
    // 2. 创建校正版本
    correctedVersion := *targetVersion
    correctedVersion.Version = h.getNextVersion(ctx, tx, code)
    correctedVersion.ChangeReason = &req.Reason
    correctedVersion.SupersedesVersion = &targetVersion.Version
    
    // 应用校正数据
    h.applyCorrections(&correctedVersion, req.NewData)
    
    // 3. 插入校正版本
    if err := h.insertCorrectionVersion(ctx, tx, &correctedVersion); err != nil {
        return nil, fmt.Errorf("插入校正版本失败: %w", err)
    }
    
    // 4. 重新计算后续版本
    if err := h.recalculateSubsequentVersions(ctx, tx, code, req.TargetDate); err != nil {
        return nil, fmt.Errorf("重新计算后续版本失败: %w", err)
    }
    
    // 5. 记录校正事件
    eventID, err := h.createCorrectionEvent(ctx, tx, code, req, tenantID.String())
    if err != nil {
        return nil, fmt.Errorf("记录校正事件失败: %w", err)
    }
    
    return map[string]interface{}{
        "operation":         "CORRECT",
        "target_date":       req.TargetDate,
        "corrected_version": correctedVersion.Version,
        "event_id":         eventID,
        "affected_versions": "calculated", // 实际需要计算影响的版本数
    }, nil
}
```

### Week 10-11: 完整合规验证

#### 10.1 元合约合规检查器 (Week 10)
```go
// 元合约合规检查器
type MetaContractComplianceChecker struct {
    db     *sql.DB
    config *ComplianceConfig
}

type ComplianceConfig struct {
    TemporalityParadigm           string   `json:"temporality_paradigm"`           // EVENT_DRIVEN
    TimelineConsistencyPolicy     string   `json:"timeline_consistency_policy"`    // NO_GAPS_ALLOWED
    SupportsFutureDating          bool     `json:"supports_future_dating"`
    SupportsRetroactivity         bool     `json:"supports_retroactivity"`
    RetroactivityTriggersRecalculation []string `json:"retroactivity_triggers_recalculation"`
    RequiredTimelineQueryParams   []string `json:"required_timeline_query_params"`
}

// 合规检查报告
type ComplianceReport struct {
    OverallStatus    string                    `json:"overall_status"` // COMPLIANT, NON_COMPLIANT
    CheckedAt        time.Time                 `json:"checked_at"`
    Requirements     []RequirementCheck        `json:"requirements"`
    Recommendations  []string                  `json:"recommendations"`
    CriticalIssues   []string                  `json:"critical_issues"`
}

type RequirementCheck struct {
    ID          string `json:"id"`
    Name        string `json:"name"`
    Status      string `json:"status"`      // PASS, FAIL, WARNING
    Description string `json:"description"`
    Evidence    string `json:"evidence"`
}

// 执行完整合规检查
func (mcc *MetaContractComplianceChecker) CheckCompliance(ctx context.Context) (*ComplianceReport, error) {
    report := &ComplianceReport{
        CheckedAt: time.Now(),
    }
    
    // 检查1: EVENT_DRIVEN范式实现
    eventDrivenCheck := mcc.checkEventDrivenParadigm(ctx)
    report.Requirements = append(report.Requirements, eventDrivenCheck)
    
    // 检查2: 时间线查询参数支持
    timelineQueryCheck := mcc.checkTimelineQuerySupport(ctx)
    report.Requirements = append(report.Requirements, timelineQueryCheck)
    
    // 检查3: 时间线一致性策略
    consistencyCheck := mcc.checkTimelineConsistency(ctx)
    report.Requirements = append(report.Requirements, consistencyCheck)
    
    // 检查4: 未来日期支持
    futureDatingCheck := mcc.checkFutureDatingSupport(ctx)
    report.Requirements = append(report.Requirements, futureDatingCheck)
    
    // 检查5: 追溯处理支持
    retroactivityCheck := mcc.checkRetroactivitySupport(ctx)
    report.Requirements = append(report.Requirements, retroactivityCheck)
    
    // 检查6: 时间线管理操作
    timelineManagementCheck := mcc.checkTimelineManagementActions(ctx)
    report.Requirements = append(report.Requirements, timelineManagementCheck)
    
    // 计算总体状态
    report.OverallStatus = mcc.calculateOverallStatus(report.Requirements)
    
    // 生成建议和关键问题
    report.Recommendations = mcc.generateRecommendations(report.Requirements)
    report.CriticalIssues = mcc.extractCriticalIssues(report.Requirements)
    
    return report, nil
}

// 检查EVENT_DRIVEN范式实现
func (mcc *MetaContractComplianceChecker) checkEventDrivenParadigm(ctx context.Context) RequirementCheck {
    check := RequirementCheck{
        ID:   "REQ_001",
        Name: "EVENT_DRIVEN范式实现",
        Description: "核心业务实体必须采用EVENT_DRIVEN模式",
    }
    
    // 检查是否有事件表
    var eventTableExists bool
    err := mcc.db.QueryRowContext(ctx,
        "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_name = 'organization_events')").
        Scan(&eventTableExists)
    
    if err != nil || !eventTableExists {
        check.Status = "FAIL"
        check.Evidence = "组织事件表不存在，未实现EVENT_DRIVEN范式"
        return check
    }
    
    // 检查是否有版本管理
    var versionColumnExists bool
    err = mcc.db.QueryRowContext(ctx,
        "SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = 'organization_units' AND column_name = 'version')").
        Scan(&versionColumnExists)
    
    if err != nil || !versionColumnExists {
        check.Status = "FAIL"
        check.Evidence = "组织单元表缺少版本字段，未实现版本管理"
        return check
    }
    
    // 检查事件记录数量
    var eventCount int
    err = mcc.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM organization_events").Scan(&eventCount)
    
    if err != nil {
        check.Status = "WARNING"
        check.Evidence = "无法统计事件记录数量"
        return check
    }
    
    check.Status = "PASS"
    check.Evidence = fmt.Sprintf("已实现EVENT_DRIVEN范式，包含%d条事件记录", eventCount)
    return check
}
```

#### 10.2 性能基准测试 (Week 11)
```go
// 时态查询性能测试
func BenchmarkTemporalQueries(b *testing.B) {
    db := setupTestDB()
    repo := NewOrganizationRepository(db, nil)
    tenantID := uuid.MustParse("3b99930c-4dc6-4cc9-8e4d-7d960a931cb9")
    
    // 准备测试数据：创建100个组织，每个5个历史版本
    setupBenchmarkData(db, 100, 5)
    
    b.Run("CurrentVersionQuery", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            code := fmt.Sprintf("BENCH%03d", i%100)
            opts := &TemporalQueryOptions{} // 只查询当前版本
            _, err := repo.GetByCodeTemporal(context.Background(), tenantID, code, opts)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("AsOfDateQuery", func(b *testing.B) {
        asOfDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
        for i := 0; i < b.N; i++ {
            code := fmt.Sprintf("BENCH%03d", i%100)
            opts := &TemporalQueryOptions{AsOfDate: &asOfDate}
            _, err := repo.GetByCodeTemporal(context.Background(), tenantID, code, opts)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("HistoryQuery", func(b *testing.B) {
        for i := 0; i < b.N; i++ {
            code := fmt.Sprintf("BENCH%03d", i%100)
            opts := &TemporalQueryOptions{IncludeHistory: true, MaxVersions: 10}
            _, err := repo.GetByCodeTemporal(context.Background(), tenantID, code, opts)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
    
    b.Run("CachedQuery", func(b *testing.B) {
        asOfDate := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
        for i := 0; i < b.N; i++ {
            code := fmt.Sprintf("BENCH%03d", i%100)
            opts := &TemporalQueryOptions{AsOfDate: &asOfDate}
            _, err := repo.GetByCodeTemporalCached(context.Background(), tenantID, code, opts)
            if err != nil {
                b.Fatal(err)
            }
        }
    })
}

// 性能基准目标
var performanceTargets = map[string]time.Duration{
    "CurrentVersionQuery": 50 * time.Millisecond,   // 当前版本查询 < 50ms
    "AsOfDateQuery":      100 * time.Millisecond,   // 时间点查询 < 100ms  
    "HistoryQuery":       200 * time.Millisecond,   // 历史查询 < 200ms
    "CachedQuery":        10 * time.Millisecond,    // 缓存查询 < 10ms
}
```

### Week 12-13: 生产部署准备

#### 12.1 数据迁移脚本 (Week 12)
```sql
-- 生产环境数据迁移脚本
-- 文件: migrate_to_temporal_v1.sql

BEGIN;

-- 步骤1: 备份现有数据
CREATE TABLE organization_units_backup_pre_temporal AS
SELECT * FROM organization_units;

-- 步骤2: 添加时态字段（如果还未添加）
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                  WHERE table_name = 'organization_units' 
                  AND column_name = 'effective_date') THEN
        ALTER TABLE organization_units ADD COLUMN effective_date DATE NOT NULL DEFAULT CURRENT_DATE;
        ALTER TABLE organization_units ADD COLUMN end_date DATE;
        ALTER TABLE organization_units ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
        ALTER TABLE organization_units ADD COLUMN supersedes_version INTEGER;
        ALTER TABLE organization_units ADD COLUMN change_reason VARCHAR(500);
        ALTER TABLE organization_units ADD COLUMN is_current BOOLEAN NOT NULL DEFAULT true;
    END IF;
END
$$;

-- 步骤3: 迁移现有数据
UPDATE organization_units 
SET effective_date = created_at::DATE,
    version = 1,
    is_current = true,
    change_reason = '初始数据迁移'
WHERE effective_date IS NULL OR version IS NULL;

-- 步骤4: 修改主键约束
DO $$
BEGIN
    -- 检查是否需要修改主键
    IF EXISTS (SELECT 1 FROM information_schema.table_constraints 
              WHERE table_name = 'organization_units' 
              AND constraint_name = 'organization_units_pkey'
              AND constraint_type = 'PRIMARY KEY') THEN
        ALTER TABLE organization_units DROP CONSTRAINT organization_units_pkey;
        ALTER TABLE organization_units ADD CONSTRAINT organization_units_pkey 
            PRIMARY KEY (code, version);
    END IF;
END
$$;

-- 步骤5: 创建优化索引
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_effective_date 
ON organization_units(effective_date);

CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_org_current_version 
ON organization_units(code, is_current) WHERE is_current = true;

-- 步骤6: 创建事件表和版本表（如果不存在）
CREATE TABLE IF NOT EXISTS organization_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    event_type VARCHAR(50) NOT NULL,
    event_data JSONB NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL
);

CREATE TABLE IF NOT EXISTS organization_versions (
    version_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    version INTEGER NOT NULL,
    effective_date DATE NOT NULL,
    end_date DATE,
    snapshot_data JSONB NOT NULL,
    change_reason VARCHAR(500),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL
);

-- 步骤7: 创建触发器和函数
CREATE OR REPLACE FUNCTION auto_manage_end_date()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE organization_units 
    SET end_date = NEW.effective_date - INTERVAL '1 day',
        is_current = false
    WHERE code = NEW.code 
      AND is_current = true 
      AND version != NEW.version;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_auto_end_date ON organization_units;
CREATE TRIGGER trigger_auto_end_date
    BEFORE INSERT ON organization_units
    FOR EACH ROW 
    EXECUTE FUNCTION auto_manage_end_date();

-- 步骤8: 数据验证
DO $$
DECLARE
    issue_count INTEGER := 0;
    total_orgs INTEGER := 0;
    current_versions INTEGER := 0;
BEGIN
    -- 统计基本信息
    SELECT COUNT(*) INTO total_orgs FROM organization_units;
    SELECT COUNT(DISTINCT code) INTO current_versions FROM organization_units WHERE is_current = true;
    
    -- 检查数据一致性问题
    SELECT COUNT(*) INTO issue_count FROM (
        SELECT code FROM organization_units WHERE is_current = true GROUP BY code HAVING COUNT(*) > 1
        UNION ALL
        SELECT code FROM organization_units WHERE effective_date IS NULL
        UNION ALL  
        SELECT code FROM organization_units WHERE version IS NULL OR version < 1
    ) issues;
    
    -- 报告结果
    RAISE NOTICE '=== 数据迁移验证结果 ===';
    RAISE NOTICE '总组织记录数: %', total_orgs;
    RAISE NOTICE '当前版本组织数: %', current_versions;
    RAISE NOTICE '数据一致性问题: %', issue_count;
    
    IF issue_count > 0 THEN
        RAISE EXCEPTION '发现数据一致性问题，请检查后重新执行迁移';
    END IF;
    
    RAISE NOTICE '数据迁移验证通过！';
END
$$;

-- 提交事务
COMMIT;

-- 迁移后清理脚本（可选，在确认迁移成功后执行）
-- DROP TABLE IF EXISTS organization_units_backup_pre_temporal;
```

#### 12.2 监控和报警配置 (Week 12-13)
```yaml
# prometheus_alerts.yml
groups:
- name: temporal_management
  rules:
  - alert: TemporalQueryHighLatency
    expr: histogram_quantile(0.95, http_request_duration_seconds{endpoint=~".*temporal.*"}) > 0.5
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "时态查询响应时间过高"
      description: "时态查询95%分位数响应时间超过500ms，当前值: {{ $value }}s"
      
  - alert: TimelineConsistencyError
    expr: increase(timeline_consistency_errors_total[5m]) > 0
    for: 1m
    labels:
      severity: critical
    annotations:
      summary: "发现时间线一致性错误"
      description: "检测到时间线一致性违规，过去5分钟内发生{{ $value }}次错误"
      
  - alert: EndDateManagementFailure  
    expr: increase(end_date_management_errors_total[5m]) > 5
    for: 2m
    labels:
      severity: warning
    annotations:
      summary: "结束日期自动管理失败"
      description: "结束日期自动管理失败次数过多，过去5分钟内发生{{ $value }}次"
      
  - alert: TemporalDataInconsistency
    expr: temporal_data_consistency_score < 0.95
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "时态数据一致性分数过低"
      description: "时态数据一致性分数为{{ $value }}，低于95%阈值"
```

#### 12.3 最终验收测试 (Week 13)
```bash
#!/bin/bash
# 最终验收测试脚本
# 文件: final_acceptance_test.sh

echo "=== 时态管理API升级最终验收测试 ==="
echo "测试开始时间: $(date)"

BASE_URL="http://localhost:9090/api/v1/organization-units"
FAILED_TESTS=0
TOTAL_TESTS=0

# 测试函数
run_test() {
    local test_name="$1"
    local command="$2"
    local expected_status="$3"
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    echo -n "[$TOTAL_TESTS] $test_name ... "
    
    response=$(eval "$command" 2>/dev/null)
    status=$?
    
    if [ $status -eq $expected_status ]; then
        echo "✅ PASS"
    else
        echo "❌ FAIL (expected: $expected_status, got: $status)"
        FAILED_TESTS=$((FAILED_TESTS + 1))
    fi
}

# 1. 基础功能测试
echo "1. 基础功能验证"
run_test "健康检查" \
    "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/../health" \
    0

run_test "当前版本查询" \
    "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/1000001" \
    0

# 2. 时态查询测试  
echo -e "\n2. 时态查询功能验证"
run_test "时间点查询" \
    "curl -s -o /dev/null -w '%{http_code}' '$BASE_URL/1000001?as_of_date=2025-01-01'" \
    0

run_test "历史版本查询" \
    "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/1000001/history" \
    0

run_test "版本范围查询" \
    "curl -s -o /dev/null -w '%{http_code}' '$BASE_URL/1000001?include_history=true&max_versions=5'" \
    0

# 3. 事件驱动操作测试
echo -e "\n3. 事件驱动功能验证" 
run_test "创建变更事件" \
    "curl -s -o /dev/null -w '%{http_code}' -X POST $BASE_URL/TEST001/events -H 'Content-Type: application/json' -d '{\"event_type\":\"UPDATE\",\"effective_date\":\"2025-12-01\",\"change_data\":{\"name\":\"更新测试\"},\"change_reason\":\"验收测试\"}'" \
    0

run_test "未来变更规划" \
    "curl -s -o /dev/null -w '%{http_code}' -X POST $BASE_URL/TEST001/timeline/plan -H 'Content-Type: application/json' -d '{\"planned_changes\":[{\"effective_date\":\"2026-01-01\",\"changes\":{\"name\":\"未来版本\"},\"event_type\":\"UPDATE\"}],\"reason\":\"验收测试规划\"}'" \
    0

# 4. 时间线管理操作测试
echo -e "\n4. 时间线管理功能验证"
run_test "时间线校正" \
    "curl -s -o /dev/null -w '%{http_code}' -X POST $BASE_URL/TEST001/timeline/correct -H 'Content-Type: application/json' -d '{\"target_date\":\"2025-06-01\",\"new_data\":{\"description\":\"校正描述\"},\"reason\":\"验收测试校正\"}'" \
    0

# 5. 数据一致性验证
echo -e "\n5. 数据一致性验证"
run_test "时间线一致性检查" \
    "PGPASSWORD=password psql -h localhost -U user -d cubecastle -c 'SELECT COUNT(*) FROM validate_temporal_consistency();' | grep -q '0'" \
    0

# 6. 性能测试
echo -e "\n6. 性能基准验证"
run_test "并发查询性能" \
    "ab -n 100 -c 10 -s 30 '$BASE_URL/1000001?as_of_date=2025-01-01' 2>/dev/null | grep -q 'Complete'" \
    0

# 7. 元合约合规性检查
echo -e "\n7. 元合约合规性验证"
run_test "合规检查API" \
    "curl -s -o /dev/null -w '%{http_code}' $BASE_URL/../compliance/check" \
    0

# 测试结果汇总
echo -e "\n=== 验收测试结果汇总 ==="
echo "总测试数: $TOTAL_TESTS"
echo "失败数: $FAILED_TESTS" 
echo "通过率: $(echo "scale=2; ($TOTAL_TESTS - $FAILED_TESTS) * 100 / $TOTAL_TESTS" | bc)%"

if [ $FAILED_TESTS -eq 0 ]; then
    echo "🎉 所有测试通过！时态管理API升级验收成功！"
    echo "✅ 系统已准备好生产环境部署"
    exit 0
else
    echo "❌ 验收测试失败，请修复问题后重新测试"
    exit 1
fi
```

---

## 📅 实施时间表总览

| 阶段 | 周次 | 主要任务 | 关键交付物 | 验收标准 |
|------|------|----------|------------|----------|
| **阶段1** | Week 1 | 数据库设计与准备 | 表结构设计、迁移计划 | 设计评审通过 |
| | Week 2 | 核心表结构扩展 | 扩展表结构、触发器 | 数据迁移验证通过 |
| | Week 3 | 事件表创建 | 事件表、版本表 | 一致性检查通过 |
| | Week 4 | 应用程序适配 | 兼容性API | 现有功能不受影响 |
| **阶段2** | Week 5 | 时态查询开发 | 时态查询API | 查询功能验证通过 |
| | Week 6 | 历史版本API | 历史查询、事件API | API功能完整 |
| | Week 7 | 性能优化 | 缓存策略、索引优化 | 性能基准达标 |
| **阶段3** | Week 8-9 | 一致性验证系统 | 规则引擎、验证器 | 一致性保证通过 |
| | Week 10-11 | 合规验证 | 合规检查器、性能测试 | 元合约合规通过 |  
| | Week 12-13 | 生产部署准备 | 迁移脚本、监控配置 | 验收测试通过 |

## 🎯 成功标准

### 功能完整性
- ✅ 支持时间点查询（as_of_date）
- ✅ 历史版本管理和查询
- ✅ 事件驱动状态变更
- ✅ 智能结束日期管理
- ✅ 时间线一致性保证

### 性能指标
- ✅ 当前版本查询 < 50ms
- ✅ 时间点查询 < 100ms  
- ✅ 历史查询 < 200ms
- ✅ 缓存查询 < 10ms

### 合规性要求
- ✅ 完全符合元合约v6.0规范
- ✅ EVENT_DRIVEN范式实现
- ✅ 时间线查询参数支持
- ✅ 未来日期和追溯处理支持

### 稳定性保证
- ✅ 向后兼容性100%保持
- ✅ 零业务中断部署
- ✅ 完整的回滚机制
- ✅ 数据一致性验证通过

---

**文档版本**: v1.0  
**制定日期**: 2025-08-10  
**预计完成**: 2025-11-02 (13周后)  
**项目负责人**: 系统架构师  
**技术负责人**: 后端技术负责人