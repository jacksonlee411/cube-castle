# 组织单元API彻底激进优化实施指南

**版本**: v1.0  
**创建日期**: 2025-08-05  
**适用场景**: 组织单元管理API完全重构  
**实施周期**: 3天激进实施计划  
**状态**: 待实施

## 📋 方案概述

**核心理念**: 无历史包袱，追求极致简洁和性能，完全重构组织单元API架构

**关键特点**:
- 7位code直接作为数据库主键
- 零ID转换开销
- 极简API设计
- 最优性能架构

## 🎯 设计原则

### 1. 彻底简化原则
- **单一标识符**: 只使用7位code，完全隐藏UUID
- **直接映射**: 数据库到API零转换层
- **极简架构**: 移除所有不必要的抽象层

### 2. 性能优先原则
- **主键查询**: 直接使用code作为数据库主键
- **索引优化**: 针对7位code的专门索引策略
- **缓存简化**: 消除ID映射缓存的开销

### 3. 用户体验原则
- **认知统一**: 前后端使用相同的7位编码
- **业务语义**: 编码对业务人员有直接意义
- **集成简化**: 第三方系统集成复杂度最小化

## 🏗️ 激进架构设计

### 0. 编码位数策略说明 ⭐

**重要说明**: 本指南专注于组织单元的7位编码优化。其他实体的编码位数分配如下：

| 实体类型 | 位数 | 范围                 | 容量      | 设计理由                    |
|----------|------|---------------------|-----------|----------------------------|
| **组织单元** | 7位  | 1000000-9999999     | 900万     | 层级复杂，需要大量编码空间  |
| **员工**     | 8位  | 10000000-99999999   | 9000万    | 企业人员规模可能很大，需要充足空间 |
| **职位**     | 7位  | 1000000-9999999     | 900万     | 职位种类和实例较多，需要较大空间   |
| **作业档案** | 5位  | 10000-99999         | 9万       | 标准化程度高，数量相对可控  |

各实体采用独立的编码位数设计，避免耦合，便于独立扩展和维护。详见[标识符命名标准](../standards/identifier-naming-standards.md)。

### 1. 数据库重构 (彻底简化)

#### 核心表结构
```sql
-- 新的极简数据库设计 (组织单元专用7位编码)
CREATE TABLE organization_units (
    code VARCHAR(10) PRIMARY KEY,              -- 7位编码直接作为主键
    parent_code VARCHAR(10) REFERENCES organization_units(code),
    tenant_id UUID NOT NULL,                  -- 租户隔离
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM')),
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'PLANNED')),
    level INTEGER NOT NULL DEFAULT 1,
    path VARCHAR(1000),                       -- 层级路径: /1000000/1000001/1000002
    sort_order INTEGER DEFAULT 0,            -- 同级排序
    description TEXT,
    profile JSONB DEFAULT '{}',               -- 多态配置
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    
    -- 约束定义
    CONSTRAINT pk_organization_units PRIMARY KEY (code),
    CONSTRAINT fk_parent_code FOREIGN KEY (parent_code) REFERENCES organization_units(code),
    CONSTRAINT uk_tenant_code UNIQUE (tenant_id, code)
);
```

#### 7位编码生成机制
```sql
-- 7位编码生成序列 (组织单元专用)
CREATE SEQUENCE org_unit_code_seq 
    START WITH 1000000 
    INCREMENT BY 1 
    MAXVALUE 9999999;

-- 自动生成7位编码的触发器
CREATE OR REPLACE FUNCTION generate_org_unit_code()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_unit_code_seq')::text, 7, '0');
    END IF;
    -- 自动计算层级和路径
    IF NEW.parent_code IS NOT NULL THEN
        SELECT level + 1, path || '/' || NEW.code 
        INTO NEW.level, NEW.path
        FROM organization_units 
        WHERE code = NEW.parent_code;
    ELSE
        NEW.level := 1;
        NEW.path := '/' || NEW.code;
    END IF;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER set_org_unit_code 
    BEFORE INSERT OR UPDATE ON organization_units 
    FOR EACH ROW EXECUTE FUNCTION generate_org_unit_code();
```

#### 高性能索引策略
```sql
-- 高性能索引策略
CREATE INDEX idx_org_units_parent_code ON organization_units(parent_code);
CREATE INDEX idx_org_units_tenant_status ON organization_units(tenant_id, status);
CREATE INDEX idx_org_units_type_level ON organization_units(unit_type, level);
CREATE INDEX idx_org_units_path_gin ON organization_units USING gin(path gin_trgm_ops);
CREATE INDEX idx_org_units_name_gin ON organization_units USING gin(name gin_trgm_ops);
```

### 2. Go后端架构 (零转换)

#### 数据模型定义
```go
// models/organization_unit.go - 极简模型
package models

import (
    "time"
    "encoding/json"
)

// OrganizationUnit 组织单元模型 - 直接使用7位code作为主键
type OrganizationUnit struct {
    Code        string          `json:"code" db:"code" validate:"required,len=7,numeric"`
    ParentCode  *string         `json:"parent_code,omitempty" db:"parent_code" validate:"omitempty,len=7,numeric"`
    TenantID    string          `json:"-" db:"tenant_id" validate:"required,uuid"`
    Name        string          `json:"name" db:"name" validate:"required,max=255"`
    UnitType    string          `json:"unit_type" db:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
    Status      string          `json:"status" db:"status" validate:"required,oneof=ACTIVE INACTIVE PLANNED"`
    Level       int             `json:"level" db:"level"`
    Path        string          `json:"path" db:"path"`
    SortOrder   int             `json:"sort_order" db:"sort_order"`
    Description *string         `json:"description,omitempty" db:"description"`
    Profile     json.RawMessage `json:"profile" db:"profile"`
    CreatedAt   time.Time       `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

// CreateOrganizationUnitRequest 创建请求
type CreateOrganizationUnitRequest struct {
    Name        string          `json:"name" validate:"required,max=255"`
    ParentCode  *string         `json:"parent_code,omitempty" validate:"omitempty,len=7,numeric"`
    UnitType    string          `json:"unit_type" validate:"required,oneof=DEPARTMENT COST_CENTER COMPANY PROJECT_TEAM"`
    Description *string         `json:"description,omitempty" validate:"omitempty,max=1000"`
    Profile     json.RawMessage `json:"profile,omitempty"`
    SortOrder   *int            `json:"sort_order,omitempty"`
}

// UpdateOrganizationUnitRequest 更新请求
type UpdateOrganizationUnitRequest struct {
    Name        *string         `json:"name,omitempty" validate:"omitempty,max=255"`
    ParentCode  *string         `json:"parent_code,omitempty" validate:"omitempty,len=7,numeric"`
    Status      *string         `json:"status,omitempty" validate:"omitempty,oneof=ACTIVE INACTIVE PLANNED"`
    Description *string         `json:"description,omitempty" validate:"omitempty,max=1000"`
    Profile     json.RawMessage `json:"profile,omitempty"`
    SortOrder   *int            `json:"sort_order,omitempty"`
}

// ListOrganizationUnitsResponse 列表响应
type ListOrganizationUnitsResponse struct {
    Organizations []OrganizationUnit `json:"organizations"`
    TotalCount    int64              `json:"total_count"`
    Page          int                `json:"page"`
    PageSize      int                `json:"page_size"`
}
```

#### 仓储层实现
```go
// repository/organization_unit_repository.go - 极简仓储层
package repository

import (
    "context"
    "database/sql"
    "fmt"
    "strings"
    
    "github.com/jmoiron/sqlx"
    "your-project/models"
)

type OrganizationUnitRepository struct {
    db *sqlx.DB
}

func NewOrganizationUnitRepository(db *sqlx.DB) *OrganizationUnitRepository {
    return &OrganizationUnitRepository{db: db}
}

// FindByCode 通过7位编码查询 - 直接主键查询，最高性能
func (r *OrganizationUnitRepository) FindByCode(ctx context.Context, tenantID, code string) (*models.OrganizationUnit, error) {
    var unit models.OrganizationUnit
    query := `
        SELECT code, parent_code, name, unit_type, status, level, path, 
               sort_order, description, profile, created_at, updated_at
        FROM organization_units 
        WHERE tenant_id = $1 AND code = $2
    `
    err := r.db.GetContext(ctx, &unit, query, tenantID, code)
    if err == sql.ErrNoRows {
        return nil, nil
    }
    return &unit, err
}

// List 列表查询 - 优化的分页查询
func (r *OrganizationUnitRepository) List(ctx context.Context, tenantID string, opts *ListOptions) ([]models.OrganizationUnit, int64, error) {
    var units []models.OrganizationUnit
    var totalCount int64
    
    // 构建查询条件
    conditions := []string{"tenant_id = $1"}
    args := []interface{}{tenantID}
    argIndex := 2
    
    if opts.ParentCode != nil {
        conditions = append(conditions, fmt.Sprintf("parent_code = $%d", argIndex))
        args = append(args, *opts.ParentCode)
        argIndex++
    }
    
    if opts.Status != nil {
        conditions = append(conditions, fmt.Sprintf("status = $%d", argIndex))
        args = append(args, *opts.Status)
        argIndex++
    }
    
    if opts.UnitType != nil {
        conditions = append(conditions, fmt.Sprintf("unit_type = $%d", argIndex))
        args = append(args, *opts.UnitType)
        argIndex++
    }
    
    whereClause := "WHERE " + strings.Join(conditions, " AND ")
    
    // 获取总数
    countQuery := fmt.Sprintf("SELECT COUNT(*) FROM organization_units %s", whereClause)
    err := r.db.GetContext(ctx, &totalCount, countQuery, args...)
    if err != nil {
        return nil, 0, err
    }
    
    // 获取数据 - 按path排序确保层级顺序
    dataQuery := fmt.Sprintf(`
        SELECT code, parent_code, name, unit_type, status, level, path,
               sort_order, description, profile, created_at, updated_at
        FROM organization_units %s
        ORDER BY path, sort_order, code
        LIMIT $%d OFFSET $%d
    `, whereClause, argIndex, argIndex+1)
    
    args = append(args, opts.Limit, opts.Offset)
    err = r.db.SelectContext(ctx, &units, dataQuery, args...)
    
    return units, totalCount, err
}

// Create 创建 - 自动生成7位编码
func (r *OrganizationUnitRepository) Create(ctx context.Context, tenantID string, req *models.CreateOrganizationUnitRequest) (*models.OrganizationUnit, error) {
    query := `
        INSERT INTO organization_units (tenant_id, name, parent_code, unit_type, description, profile, sort_order)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING code, parent_code, name, unit_type, status, level, path,
                  sort_order, description, profile, created_at, updated_at
    `
    
    var unit models.OrganizationUnit
    sortOrder := 0
    if req.SortOrder != nil {
        sortOrder = *req.SortOrder
    }
    
    err := r.db.GetContext(ctx, &unit, query,
        tenantID, req.Name, req.ParentCode, req.UnitType,
        req.Description, req.Profile, sortOrder)
    
    return &unit, err
}

// GetTree 获取组织树 - 高性能树形查询
func (r *OrganizationUnitRepository) GetTree(ctx context.Context, tenantID string, rootCode *string) ([]models.OrganizationUnit, error) {
    var query string
    var args []interface{}
    
    if rootCode != nil {
        // 获取指定节点及其所有子树
        query = `
            WITH RECURSIVE org_tree AS (
                SELECT code, parent_code, name, unit_type, status, level, path,
                       sort_order, description, profile, created_at, updated_at
                FROM organization_units 
                WHERE tenant_id = $1 AND code = $2
                
                UNION ALL
                
                SELECT o.code, o.parent_code, o.name, o.unit_type, o.status, o.level, o.path,
                       o.sort_order, o.description, o.profile, o.created_at, o.updated_at
                FROM organization_units o
                INNER JOIN org_tree t ON o.parent_code = t.code
                WHERE o.tenant_id = $1
            )
            SELECT * FROM org_tree ORDER BY path, sort_order, code
        `
        args = []interface{}{tenantID, *rootCode}
    } else {
        // 获取整个组织树
        query = `
            SELECT code, parent_code, name, unit_type, status, level, path,
                   sort_order, description, profile, created_at, updated_at
            FROM organization_units 
            WHERE tenant_id = $1 
            ORDER BY path, sort_order, code
        `
        args = []interface{}{tenantID}
    }
    
    var units []models.OrganizationUnit
    err := r.db.SelectContext(ctx, &units, query, args...)
    return units, err
}

type ListOptions struct {
    ParentCode *string
    Status     *string
    UnitType   *string
    Limit      int
    Offset     int
}
```

#### 业务服务层
```go
// service/organization_unit_service.go - 无转换业务层
package service

import (
    "context"
    "fmt"
    
    "your-project/models"
    "your-project/repository"
)

type OrganizationUnitService struct {
    repo *repository.OrganizationUnitRepository
}

func NewOrganizationUnitService(repo *repository.OrganizationUnitRepository) *OrganizationUnitService {
    return &OrganizationUnitService{repo: repo}
}

// GetByCode 通过编码获取 - 直接调用，无转换
func (s *OrganizationUnitService) GetByCode(ctx context.Context, tenantID, code string) (*models.OrganizationUnit, error) {
    if len(code) != 7 {
        return nil, fmt.Errorf("invalid organization code: must be 7 digits")
    }
    return s.repo.FindByCode(ctx, tenantID, code)
}

// List 列表查询 - 直接调用，无转换
func (s *OrganizationUnitService) List(ctx context.Context, tenantID string, opts *repository.ListOptions) (*models.ListOrganizationUnitsResponse, error) {
    // 参数验证
    if opts.Limit <= 0 || opts.Limit > 100 {
        opts.Limit = 50
    }
    if opts.Offset < 0 {
        opts.Offset = 0
    }
    
    // 验证parent_code格式
    if opts.ParentCode != nil && len(*opts.ParentCode) != 7 {
        return nil, fmt.Errorf("invalid parent_code: must be 7 digits")
    }
    
    units, total, err := s.repo.List(ctx, tenantID, opts)
    if err != nil {
        return nil, err
    }
    
    page := (opts.Offset / opts.Limit) + 1
    
    return &models.ListOrganizationUnitsResponse{
        Organizations: units,
        TotalCount:    total,
        Page:          page,
        PageSize:      opts.Limit,
    }, nil
}

// Create 创建 - 直接调用，无转换
func (s *OrganizationUnitService) Create(ctx context.Context, tenantID string, req *models.CreateOrganizationUnitRequest) (*models.OrganizationUnit, error) {
    // 验证parent_code存在性
    if req.ParentCode != nil {
        parent, err := s.repo.FindByCode(ctx, tenantID, *req.ParentCode)
        if err != nil {
            return nil, err
        }
        if parent == nil {
            return nil, fmt.Errorf("parent organization unit not found: %s", *req.ParentCode)
        }
    }
    
    return s.repo.Create(ctx, tenantID, req)
}

// GetTree 获取组织树 - 直接调用，无转换
func (s *OrganizationUnitService) GetTree(ctx context.Context, tenantID string, rootCode *string) ([]models.OrganizationUnit, error) {
    if rootCode != nil && len(*rootCode) != 7 {
        return nil, fmt.Errorf("invalid root_code: must be 7 digits")
    }
    return s.repo.GetTree(ctx, tenantID, rootCode)
}
```

### 3. API Handler层 (极简响应)

```go
// handlers/organization_unit_handler.go - 零配置API处理器
package handlers

import (
    "net/http"
    "strconv"
    
    "github.com/gin-gonic/gin"
    "your-project/models"
    "your-project/service"
    "your-project/repository"
)

type OrganizationUnitHandler struct {
    service *service.OrganizationUnitService
}

func NewOrganizationUnitHandler(service *service.OrganizationUnitService) *OrganizationUnitHandler {
    return &OrganizationUnitHandler{service: service}
}

// GetOrganizationUnits 获取组织单元列表
// GET /api/v1/organization-units
func (h *OrganizationUnitHandler) GetOrganizationUnits(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    
    // 解析查询参数
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
    offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
    
    opts := &repository.ListOptions{
        Limit:  limit,
        Offset: offset,
    }
    
    if parentCode := c.Query("parent_code"); parentCode != "" {
        opts.ParentCode = &parentCode
    }
    
    if status := c.Query("status"); status != "" {
        opts.Status = &status
    }
    
    if unitType := c.Query("unit_type"); unitType != "" {
        opts.UnitType = &unitType
    }
    
    // 直接调用服务，无转换
    response, err := h.service.List(c.Request.Context(), tenantID, opts)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, response)
}

// GetOrganizationUnit 获取单个组织单元
// GET /api/v1/organization-units/{code}
func (h *OrganizationUnitHandler) GetOrganizationUnit(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    code := c.Param("code")
    
    // 直接使用7位编码查询
    unit, err := h.service.GetByCode(c.Request.Context(), tenantID, code)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    if unit == nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Organization unit not found"})
        return
    }
    
    c.JSON(http.StatusOK, unit)
}

// CreateOrganizationUnit 创建组织单元
// POST /api/v1/organization-units
func (h *OrganizationUnitHandler) CreateOrganizationUnit(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    
    var req models.CreateOrganizationUnitRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    // 直接创建，无转换
    unit, err := h.service.Create(c.Request.Context(), tenantID, &req)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusCreated, unit)
}

// GetOrganizationTree 获取组织树
// GET /api/v1/organization-units/tree
func (h *OrganizationUnitHandler) GetOrganizationTree(c *gin.Context) {
    tenantID := c.GetString("tenant_id")
    
    var rootCode *string
    if root := c.Query("root_code"); root != "" {
        rootCode = &root
    }
    
    // 直接获取树结构，无转换
    units, err := h.service.GetTree(c.Request.Context(), tenantID, rootCode)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "organizations": units,
        "total_count":   len(units),
    })
}
```

### 4. 路由配置 (统一端点)

```go
// routes/organization_routes.go - 简化路由配置
package routes

import (
    "github.com/gin-gonic/gin"
    "your-project/handlers"
    "your-project/middleware"
)

func SetupOrganizationRoutes(router *gin.Engine, handler *handlers.OrganizationUnitHandler) {
    v1 := router.Group("/api/v1")
    v1.Use(middleware.AuthMiddleware())
    v1.Use(middleware.TenantMiddleware())
    
    // 主要API端点 - 使用7位编码
    orgUnits := v1.Group("/organization-units")
    {
        orgUnits.GET("", handler.GetOrganizationUnits)
        orgUnits.POST("", handler.CreateOrganizationUnit)
        orgUnits.GET("/tree", handler.GetOrganizationTree)
        orgUnits.GET("/:code", handler.GetOrganizationUnit)     // 7位编码参数
        orgUnits.PUT("/:code", handler.UpdateOrganizationUnit)  // 7位编码参数
        orgUnits.DELETE("/:code", handler.DeleteOrganizationUnit) // 7位编码参数
    }
    
    // CoreHR兼容端点
    coreHR := v1.Group("/corehr")
    {
        coreHR.GET("/organizations", handler.GetCoreHROrganizations)
        coreHR.POST("/organizations", handler.CreateCoreHROrganization)
        coreHR.GET("/organizations/tree", handler.GetOrganizationTree)
    }
}
```

## 📅 3天激进实施计划

### Day 1: 数据库和后端重构 ⚡

#### 上午 (9:00-12:00): 数据库重构
```bash
# 1. 备份现有数据
pg_dump -h localhost -U user -d cubecastle > backup_org_units_$(date +%Y%m%d).sql

# 2. 创建迁移脚本
cat > migration_to_7digit_codes.sql << 'EOF'
BEGIN;

-- 1. 创建新表结构
CREATE TABLE organization_units_new (
    code VARCHAR(10) PRIMARY KEY,
    parent_code VARCHAR(10),
    tenant_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    unit_type VARCHAR(50) NOT NULL CHECK (unit_type IN ('DEPARTMENT', 'COST_CENTER', 'COMPANY', 'PROJECT_TEAM')),
    status VARCHAR(20) DEFAULT 'ACTIVE' CHECK (status IN ('ACTIVE', 'INACTIVE', 'PLANNED')),
    level INTEGER NOT NULL DEFAULT 1,
    path VARCHAR(1000),
    sort_order INTEGER DEFAULT 0,
    description TEXT,
    profile JSONB DEFAULT '{}',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- 2. 创建7位编码序列
CREATE SEQUENCE org_unit_code_seq START WITH 1000000 INCREMENT BY 1 MAXVALUE 9999999;

-- 3. 迁移数据 (生成7位编码)
INSERT INTO organization_units_new 
SELECT 
    LPAD((ROW_NUMBER() OVER (ORDER BY created_at) + 999999)::text, 7, '0') as code,
    (SELECT LPAD((ROW_NUMBER() OVER (ORDER BY p.created_at) + 999999)::text, 7, '0') 
     FROM organization_units p WHERE p.id = o.parent_unit_id) as parent_code,
    tenant_id, name, unit_type, status, 1 as level, 
    ('/' || LPAD((ROW_NUMBER() OVER (ORDER BY created_at) + 999999)::text, 7, '0')) as path,
    0 as sort_order, description, '{}' as profile,
    created_at, updated_at
FROM organization_units o;

-- 4. 添加约束和索引
ALTER TABLE organization_units_new 
    ADD CONSTRAINT fk_parent_code 
    FOREIGN KEY (parent_code) 
    REFERENCES organization_units_new(code);

CREATE INDEX idx_org_units_parent_code ON organization_units_new(parent_code);
CREATE INDEX idx_org_units_tenant_status ON organization_units_new(tenant_id, status);
CREATE INDEX idx_org_units_type_level ON organization_units_new(unit_type, level);

-- 5. 原子替换
DROP TABLE organization_units;
ALTER TABLE organization_units_new RENAME TO organization_units;

-- 6. 创建触发器
CREATE OR REPLACE FUNCTION generate_org_unit_code()
RETURNS TRIGGER AS $body$
BEGIN
    IF NEW.code IS NULL THEN
        NEW.code := LPAD(nextval('org_unit_code_seq')::text, 7, '0');
    END IF;
    IF NEW.parent_code IS NOT NULL THEN
        SELECT level + 1, path || '/' || NEW.code 
        INTO NEW.level, NEW.path
        FROM organization_units 
        WHERE code = NEW.parent_code;
    ELSE
        NEW.level := 1;
        NEW.path := '/' || NEW.code;
    END IF;
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$body$ LANGUAGE plpgsql;

CREATE TRIGGER set_org_unit_code 
    BEFORE INSERT OR UPDATE ON organization_units 
    FOR EACH ROW EXECUTE FUNCTION generate_org_unit_code();

COMMIT;
EOF

# 3. 执行迁移
psql -h localhost -U user -d cubecastle -f migration_to_7digit_codes.sql

# 4. 验证数据完整性
psql -h localhost -U user -d cubecastle -c "
SELECT COUNT(*) as total_units,
       COUNT(DISTINCT code) as unique_codes,
       MIN(LENGTH(code)) as min_code_len,
       MAX(LENGTH(code)) as max_code_len
FROM organization_units;"
```

#### 下午 (13:00-18:00): Go后端重构
```bash
# 1. 重构模型层
# 创建新的模型文件
mkdir -p models
cat > models/organization_unit.go << 'EOF'
// [上面提供的完整Go模型代码]
EOF

# 2. 重构仓储层  
mkdir -p repository
cat > repository/organization_unit_repository.go << 'EOF'
// [上面提供的完整仓储层代码]
EOF

# 3. 重构服务层
mkdir -p service
cat > service/organization_unit_service.go << 'EOF'
// [上面提供的完整服务层代码]
EOF

# 4. 重构API处理器
mkdir -p handlers
cat > handlers/organization_unit_handler.go << 'EOF'
// [上面提供的完整处理器代码]
EOF

# 5. 更新路由
cat > routes/organization_routes.go << 'EOF'
// [上面提供的完整路由代码]
EOF

# 6. 编译和测试
go mod tidy
go build -o bin/server cmd/server/main.go
go test ./... -v
```

### Day 2: 前端和测试 ⚡

#### 上午 (9:00-12:00): 前端更新
```typescript
// 1. 更新TypeScript类型定义
// types/organization.ts
interface Organization {
  code: string;                    // 7位编码
  parent_code?: string;           // 父级7位编码
  name: string;
  unit_type: 'DEPARTMENT' | 'COST_CENTER' | 'COMPANY' | 'PROJECT_TEAM';
  status: 'ACTIVE' | 'INACTIVE' | 'PLANNED';
  level: number;
  path: string;
  sort_order: number;
  description?: string;
  profile: any;
  created_at: string;
  updated_at: string;
}

interface CreateOrganizationRequest {
  name: string;
  parent_code?: string;
  unit_type: string;
  description?: string;
  profile?: any;
  sort_order?: number;
}

interface ListOrganizationResponse {
  organizations: Organization[];
  total_count: number;
  page: number;
  page_size: number;
}

// 2. 更新API客户端
// api/organizations.ts
export class OrganizationAPI {
  private baseURL = '/api/v1/organization-units';
  
  async getAll(params?: {
    parent_code?: string;
    status?: string;
    unit_type?: string;
    limit?: number;
    offset?: number;
  }): Promise<ListOrganizationResponse> {
    const queryParams = new URLSearchParams();
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, String(value));
        }
      });
    }
    
    const response = await fetch(`${this.baseURL}?${queryParams}`);
    if (!response.ok) throw new Error('Failed to fetch organizations');
    return response.json();
  }
  
  async getByCode(code: string): Promise<Organization> {
    const response = await fetch(`${this.baseURL}/${code}`);
    if (!response.ok) throw new Error('Organization not found');
    return response.json();
  }
  
  async create(data: CreateOrganizationRequest): Promise<Organization> {
    const response = await fetch(this.baseURL, {
      method: 'POST',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(data)
    });
    if (!response.ok) throw new Error('Failed to create organization');
    return response.json();
  }
  
  async update(code: string, data: Partial<Organization>): Promise<Organization> {
    const response = await fetch(`${this.baseURL}/${code}`, {
      method: 'PUT',
      headers: {'Content-Type': 'application/json'},
      body: JSON.stringify(data)
    });
    if (!response.ok) throw new Error('Failed to update organization');
    return response.json();
  }
  
  async delete(code: string): Promise<void> {
    const response = await fetch(`${this.baseURL}/${code}`, {
      method: 'DELETE'
    });
    if (!response.ok) throw new Error('Failed to delete organization');
  }
  
  async getTree(rootCode?: string): Promise<{organizations: Organization[], total_count: number}> {
    const params = rootCode ? `?root_code=${rootCode}` : '';
    const response = await fetch(`${this.baseURL}/tree${params}`);
    if (!response.ok) throw new Error('Failed to fetch organization tree');
    return response.json();
  }
}

// 3. 更新React组件
// components/OrganizationSelector.tsx
import React, { useState, useEffect } from 'react';
import { Organization, OrganizationAPI } from '../api/organizations';

interface Props {
  value?: string;
  onChange: (code: string) => void;
  placeholder?: string;
}

export function OrganizationSelector({ value, onChange, placeholder }: Props) {
  const [organizations, setOrganizations] = useState<Organization[]>([]);
  const [loading, setLoading] = useState(false);
  
  useEffect(() => {
    const loadOrganizations = async () => {
      setLoading(true);
      try {
        const api = new OrganizationAPI();
        const response = await api.getAll({ limit: 100 });
        setOrganizations(response.organizations);
      } catch (error) {
        console.error('Failed to load organizations:', error);
      } finally {
        setLoading(false);
      }
    };
    
    loadOrganizations();
  }, []);
  
  return (
    <select 
      value={value || ''} 
      onChange={e => onChange(e.target.value)}
      disabled={loading}
    >
      <option value="">{placeholder || 'Select Organization'}</option>
      {organizations.map(org => (
        <option key={org.code} value={org.code}>
          {'  '.repeat(org.level - 1)}{org.name} ({org.code})
        </option>
      ))}
    </select>
  );
}
```

#### 下午 (13:00-18:00): 全面测试
```bash
# 1. 单元测试
echo "Running Go unit tests..."
go test ./... -v -cover -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html

# 2. 集成测试
echo "Running integration tests..."
cat > tests/integration_test.go << 'EOF'
package tests

import (
    "testing"
    "net/http"
    "net/http/httptest"
    "bytes"
    "encoding/json"
    "your-project/handlers"
    "your-project/service"
    "your-project/repository"
)

func TestOrganizationUnitAPI(t *testing.T) {
    // 设置测试数据库和服务
    db := setupTestDB()
    repo := repository.NewOrganizationUnitRepository(db)
    svc := service.NewOrganizationUnitService(repo)
    handler := handlers.NewOrganizationUnitHandler(svc)
    
    // 测试创建组织单元
    createReq := `{"name":"测试部门","unit_type":"DEPARTMENT"}`
    req := httptest.NewRequest("POST", "/api/v1/organization-units", bytes.NewString(createReq))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    
    handler.CreateOrganizationUnit(w, req)
    
    if w.Code != http.StatusCreated {
        t.Errorf("Expected status 201, got %d", w.Code)
    }
    
    var org models.OrganizationUnit
    json.Unmarshal(w.Body.Bytes(), &org)
    
    // 验证7位编码格式
    if len(org.Code) != 7 {
        t.Errorf("Expected 7-digit code, got %s", org.Code)
    }
    
    // 测试查询
    req = httptest.NewRequest("GET", "/api/v1/organization-units/"+org.Code, nil)
    w = httptest.NewRecorder()
    
    handler.GetOrganizationUnit(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("Expected status 200, got %d", w.Code)
    }
}
EOF

go test ./tests -v

# 3. API测试 (使用newman/postman)
echo "Running API tests..."
cat > api_test.json << 'EOF'
{
  "info": {
    "name": "Organization Units API Test",
    "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
  },
  "item": [
    {
      "name": "Create Organization Unit",
      "request": {
        "method": "POST",
        "header": [{"key": "Content-Type", "value": "application/json"}],
        "body": {
          "mode": "raw",
          "raw": "{\"name\":\"测试部门\",\"unit_type\":\"DEPARTMENT\"}"
        },
        "url": "{{baseUrl}}/api/v1/organization-units"
      },
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "pm.test('Status is 201', () => pm.response.to.have.status(201));",
              "pm.test('Response has 7-digit code', () => {",
              "  const org = pm.response.json();",
              "  pm.expect(org.code).to.match(/^[0-9]{7}$/);",
              "  pm.globals.set('org_code', org.code);",
              "});"
            ]
          }
        }
      ]
    },
    {
      "name": "Get Organization Unit",
      "request": {
        "method": "GET",
        "url": "{{baseUrl}}/api/v1/organization-units/{{org_code}}"
      },
      "event": [
        {
          "listen": "test",
          "script": {
            "exec": [
              "pm.test('Status is 200', () => pm.response.to.have.status(200));",
              "pm.test('Response has correct structure', () => {",
              "  const org = pm.response.json();",
              "  pm.expect(org).to.have.property('code');",
              "  pm.expect(org).to.have.property('name');",
              "  pm.expect(org).to.have.property('unit_type');",
              "});"
            ]
          }
        }
      ]
    }
  ]
}
EOF

# 如果有newman，运行API测试
if command -v newman &> /dev/null; then
    newman run api_test.json --env-var "baseUrl=http://localhost:8080"
fi

# 4. 性能测试
echo "Running performance tests..."
# 启动服务器
./bin/server &
SERVER_PID=$!
sleep 2

# 使用ab进行性能测试
echo "Testing GET /api/v1/organization-units performance..."
ab -n 1000 -c 10 http://localhost:8080/api/v1/organization-units

echo "Testing POST /api/v1/organization-units performance..."
ab -n 100 -c 5 -p post_data.json -T application/json http://localhost:8080/api/v1/organization-units

# 停止服务器
kill $SERVER_PID

# 5. 前端测试
echo "Running frontend tests..."
cd frontend
npm test
npm run e2e
cd ..
```

### Day 3: 部署和验证 ⚡

#### 上午 (9:00-12:00): 部署准备
```bash
# 1. 更新OpenAPI规范
echo "Updating OpenAPI specification..."
cat > docs/openapi.yaml << 'EOF'
openapi: 3.0.0
info:
  title: Organization Units API
  version: 2.0.0
  description: 彻底激进优化后的组织单元管理API

paths:
  /api/v1/organization-units:
    get:
      summary: 获取组织单元列表
      parameters:
        - name: parent_code
          in: query
          schema:
            type: string
            pattern: '^[0-9]{7}$'
        - name: status
          in: query
          schema:
            type: string
            enum: [ACTIVE, INACTIVE, PLANNED]
        - name: limit
          in: query
          schema:
            type: integer
            default: 50
            maximum: 100
        - name: offset
          in: query
          schema:
            type: integer
            default: 0
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ListOrganizationUnitsResponse'
    
    post:
      summary: 创建组织单元
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateOrganizationUnitRequest'
      responses:
        '201':
          description: 创建成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizationUnit'

  /api/v1/organization-units/{code}:
    get:
      summary: 获取单个组织单元
      parameters:
        - name: code
          in: path
          required: true
          schema:
            type: string
            pattern: '^[0-9]{7}$'
      responses:
        '200':
          description: 成功
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/OrganizationUnit'

components:
  schemas:
    OrganizationUnit:
      type: object
      properties:
        code:
          type: string
          pattern: '^[0-9]{7}$'
          description: 7位组织编码
        parent_code:
          type: string
          pattern: '^[0-9]{7}$'
          description: 父级组织编码
        name:
          type: string
          maxLength: 255
        unit_type:
          type: string
          enum: [DEPARTMENT, COST_CENTER, COMPANY, PROJECT_TEAM]
        status:
          type: string
          enum: [ACTIVE, INACTIVE, PLANNED]
        level:
          type: integer
        path:
          type: string
        sort_order:
          type: integer
        description:
          type: string
        profile:
          type: object
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
      required: [code, name, unit_type, status, level, path]
    
    CreateOrganizationUnitRequest:
      type: object
      properties:
        name:
          type: string
          maxLength: 255
        parent_code:
          type: string
          pattern: '^[0-9]{7}$'
        unit_type:
          type: string
          enum: [DEPARTMENT, COST_CENTER, COMPANY, PROJECT_TEAM]
        description:
          type: string
        profile:
          type: object
        sort_order:
          type: integer
      required: [name, unit_type]
    
    ListOrganizationUnitsResponse:
      type: object
      properties:
        organizations:
          type: array
          items:
            $ref: '#/components/schemas/OrganizationUnit'
        total_count:
          type: integer
        page:
          type: integer
        page_size:
          type: integer
      required: [organizations, total_count, page, page_size]
EOF

# 2. 更新文档
echo "Updating documentation..."
# 生成API文档
swagger-codegen generate -i docs/openapi.yaml -l html2 -o docs/api-docs

# 3. 准备部署脚本
cat > scripts/deploy.sh << 'EOF'
#!/bin/bash
set -e

echo "Starting deployment..."

# 1. 构建应用
echo "Building application..."
go build -o bin/server cmd/server/main.go

# 2. 运行数据库迁移
echo "Running database migration..."
psql $DATABASE_URL -f migration_to_7digit_codes.sql

# 3. 重启服务
echo "Restarting service..."
systemctl restart organization-api

# 4. 验证部署
echo "Verifying deployment..."
sleep 5
curl -f http://localhost:8080/health || exit 1

echo "Deployment completed successfully!"
EOF

chmod +x scripts/deploy.sh
```

#### 下午 (13:00-18:00): 生产部署和验证
```bash
# 1. 生产环境数据库迁移
echo "Migrating production database..."
# 备份生产数据库
pg_dump $PROD_DATABASE_URL > prod_backup_$(date +%Y%m%d_%H%M%S).sql

# 执行迁移（确保在维护窗口进行）
psql $PROD_DATABASE_URL -f migration_to_7digit_codes.sql

# 2. 应用部署
echo "Deploying to production..."
./scripts/deploy.sh

# 3. 全面验证
echo "Running production verification..."

# API健康检查
curl -f https://api.yourdomain.com/health

# 功能验证
curl -H "Authorization: Bearer $TOKEN" \
     https://api.yourdomain.com/api/v1/organization-units \
     | jq '.organizations[0].code' | grep -E '^"[0-9]{7}"$'

# 性能基准测试
echo "Running performance benchmarks..."
ab -n 1000 -c 10 -H "Authorization: Bearer $TOKEN" \
   https://api.yourdomain.com/api/v1/organization-units

# 4. 监控和告警
echo "Setting up monitoring..."
# 设置新的监控指标
# - 7位编码生成成功率
# - API响应时间（应该有显著改善）
# - 错误率（应该保持低水平）

# 5. 文档发布
echo "Publishing updated documentation..."
# 发布新的API文档
# 通知相关团队关于API变更

echo "🎉 激进优化实施完成！"
echo "预期性能提升："
echo "- 查询性能: +40-60%"
echo "- 内存使用: -20-30%"
echo "- 代码复杂度: -35%"
echo "- 维护成本: -50%"
```

## 🚀 预期性能提升

### 查询性能优化
- **单条查询**: 从UUID索引查询 → 主键直接查询，提升 **50%**
- **列表查询**: 消除ID转换JOIN操作，提升 **40%**  
- **树形查询**: 路径索引+递归CTE优化，提升 **60%**
- **批量操作**: 统一编码体系，提升 **45%**

### 内存使用优化
- **模型简化**: 移除UUID字段，减少 **30%**
- **缓存消除**: 无需ID映射缓存，减少 **20%**
- **序列化优化**: JSON响应体积减少 **25%**
- **GC压力**: 对象分配减少，GC压力降低 **35%**

### 代码复杂度降低
- **转换逻辑**: 移除所有ID转换代码，减少 **35%**
- **错误处理**: 消除转换错误场景，简化 **40%**
- **测试用例**: 测试场景简化，减少 **30%**
- **维护成本**: 统一架构，维护成本降低 **50%**

## 📊 成功验证指标

### 技术KPI
```yaml
性能指标:
  - 单条查询响应时间: < 20ms (原 < 50ms)
  - 列表查询响应时间: < 50ms (原 < 100ms)  
  - 树形查询响应时间: < 100ms (原 < 200ms)
  - 内存使用减少: > 25%
  - CPU使用减少: > 20%

质量指标:
  - 代码覆盖率: > 90%
  - API可用性: > 99.9%
  - 错误率: < 0.1%
  - 响应时间P95: < 100ms
```

### 业务KPI
```yaml
用户体验:
  - API集成时间缩短: > 60%
  - 用户认知复杂度降低: > 80%
  - 文档理解度提升: > 70%
  - 开发者满意度: > 90%

开发效率:
  - 功能开发速度提升: > 50%
  - 代码审查时间减少: > 40%
  - 缺陷修复时间减少: > 45%
  - 系统稳定性提升: > 99.9%
```

## 🔍 风险控制措施

### 实施前风险评估
```yaml
高风险项:
  - 数据库结构变更: 完整备份 + 回滚计划
  - API接口变更: 版本兼容 + 渐进迁移
  - 性能影响: 基准测试 + 监控告警

中风险项:
  - 前端集成: 类型检查 + 集成测试
  - 第三方依赖: 接口文档 + 沟通协调
  - 团队协作: 培训计划 + 文档支持
```

### 应急预案
```yaml
回滚策略:
  - 数据库: 使用备份快速恢复
  - 应用: 部署前一版本
  - 时间窗口: 30分钟内完成回滚

应急联系:
  - 技术负责人: 24小时待命
  - DBA团队: 数据库紧急支持
  - 运维团队: 基础设施支持
```

## 📞 支持和反馈

### 实施支持
- **技术咨询**: 架构团队 (实施期间全程支持)
- **数据库支持**: DBA团队 (迁移期间在线支持)  
- **应用支持**: 开发团队 (代码实施指导)
- **测试支持**: QA团队 (质量保证验证)

### 后续优化
- **性能调优**: 根据实际运行数据进一步优化
- **功能扩展**: 基于新架构添加高级功能
- **标准推广**: 将成功经验推广到其他模块

---

**文档维护**: 架构团队  
**实施负责**: 全栈开发团队  
**质量保证**: QA测试团队  
**运营支持**: DevOps团队

**创建时间**: 2025-08-05  
**预计实施**: 2025-08-06 ~ 2025-08-08  
**文档版本**: v1.0