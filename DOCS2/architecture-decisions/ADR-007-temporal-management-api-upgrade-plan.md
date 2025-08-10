# ADR-007: 组织架构时态管理API升级方案

**状态**: 提案  
**决策日期**: 2025-08-10  
**决策者**: 系统架构师  
**影响范围**: 组织架构API、数据模型、时态查询能力  

## 背景与问题陈述

### 问题描述
当前组织架构API规范与元合约v6.0规范中关于时态管理的要求存在重大差距，项目实际实现也缺乏完整的时态管理能力。主要问题包括：

1. **API规范差距**: 组织架构API规范v2.0缺乏时态查询和事件驱动能力
2. **元合约合规性**: 不符合元合约v6.0对EVENT_DRIVEN核心业务实体的强制要求  
3. **实现能力缺失**: 无法支持"某时间点组织架构状态查询"等关键业务需求

### 深度差距分析

#### 1. API文档 vs 元合约v6.0 时态要求差距

**组织架构API规范 v2.0 现状**：
- ✅ 基础数据模型：支持created_at、updated_at时间戳
- ✅ 状态管理：ACTIVE、INACTIVE、PLANNED状态枚举  
- ❌ **缺失生效日期**：没有effective_date字段支持
- ❌ **缺失时态查询**：无"某时间点组织架构状态"查询能力
- ❌ **缺失事件驱动**：直接CRUD模式，非EVENT_DRIVEN范式

**元合约v6.0 时态要求**：
- 🔴 **强制要求**：核心业务实体(OrganizationUnit)必须采用EVENT_DRIVEN模式
- 🔴 **强制要求**：timeline_query_parameters对EVENT_DRIVEN资源是强制性的
- 🔴 **强制要求**：supports_future_dating和supports_retroactivity必需配置
- 🔴 **强制要求**：timeline_management_actions替代传统DELETE操作

#### 2. 项目实际实现 vs 元合约要求差距

**当前数据库表结构分析**：
```sql
-- 现有organization_units表字段
✅ code, tenant_id, name, unit_type, status
✅ created_at, updated_at (基础时间戳)
✅ parent_code, level, path (层级关系)

-- 缺失的关键时态字段
❌ effective_date     -- 生效日期(EVENT_DRIVEN必需)
❌ end_date           -- 失效日期(版本管理必需) 
❌ version            -- 版本号(历史追踪必需)
❌ supersedes_version -- 版本链引用
❌ change_reason      -- 变更原因(审计必需)
```

**当前API实现分析**：
- ✅ 基础CRUD操作：POST、PUT、DELETE、GET
- ❌ **缺失时间点查询**：无`as_of_date`参数支持
- ❌ **缺失历史版本API**：无法查询组织架构变更历史
- ❌ **缺失事件驱动API**：UPDATE直接修改记录，非事件创建

## 决策方案

### 选择方案：渐进式时态管理升级

采用三阶段渐进式升级方案，确保兼容性同时逐步达成元合约v6.0合规要求。

## 详细实施方案

### 阶段1：时态数据模型扩展 (4周实施)

**1.1 扩展核心表结构**
```sql
-- 扩展organization_units表增加时态字段
ALTER TABLE organization_units ADD COLUMN effective_date DATE NOT NULL DEFAULT CURRENT_DATE;
ALTER TABLE organization_units ADD COLUMN end_date DATE;
ALTER TABLE organization_units ADD COLUMN version INTEGER NOT NULL DEFAULT 1;
ALTER TABLE organization_units ADD COLUMN supersedes_version INTEGER;
ALTER TABLE organization_units ADD COLUMN change_reason VARCHAR(500);
ALTER TABLE organization_units ADD COLUMN is_current BOOLEAN NOT NULL DEFAULT true;

-- 修改主键支持版本管理
ALTER TABLE organization_units DROP CONSTRAINT organization_units_pkey;
ALTER TABLE organization_units ADD CONSTRAINT organization_units_pkey 
    PRIMARY KEY (code, version);
```

**1.2 新增事件表**
```sql
-- 创建组织事件表
CREATE TABLE organization_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_code VARCHAR(10) NOT NULL,
    event_type VARCHAR(50) NOT NULL, -- CREATE, UPDATE, ACTIVATE, DEACTIVATE, RESTRUCTURE
    event_data JSONB NOT NULL,
    effective_date DATE NOT NULL,
    created_by VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    tenant_id UUID NOT NULL,
    
    CONSTRAINT fk_org_events_org FOREIGN KEY (organization_code) 
        REFERENCES organization_units(code)
);
```

**1.3 版本管理表**
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
    tenant_id UUID NOT NULL
);
```

### 阶段2：API扩展支持时态查询 (3周实施)

**2.1 扩展现有API支持时态参数**
```go
// 扩展查询参数结构
type OrganizationQuery struct {
    AsOfDate      *time.Time `json:"as_of_date,omitempty"`      // 时间点查询
    EffectiveFrom *time.Time `json:"effective_from,omitempty"`  // 生效起始时间  
    EffectiveTo   *time.Time `json:"effective_to,omitempty"`    // 生效结束时间
    IncludeHistory bool      `json:"include_history,omitempty"` // 是否包含历史版本
    Version       *int       `json:"version,omitempty"`         // 特定版本查询
}

// 扩展API端点
GET /api/v1/organization-units?as_of_date=2025-01-01          // 时间点查询
GET /api/v1/organization-units/{code}/history                // 历史版本查询
GET /api/v1/organization-units/{code}/versions/{version}     // 特定版本查询
```

**2.2 新增事件驱动状态变更API**
```go
// 事件驱动变更请求
type OrganizationChangeEvent struct {
    EventType     string     `json:"event_type"`      // CREATE, UPDATE, RESTRUCTURE
    EffectiveDate time.Time  `json:"effective_date"`  // 生效日期
    ChangeData    ChangeData `json:"change_data"`     // 变更内容
    ChangeReason  string     `json:"change_reason"`   // 变更原因
}

// 新增事件API端点
POST /api/v1/organization-units/{code}/events    // 创建变更事件
GET  /api/v1/organization-units/{code}/events    // 查询变更事件历史
```

### 阶段3：完整事件驱动重构 (6周实施)

**3.1 时间线一致性检查**
```go
// 实现timeline_consistency_policy
type TimelineConsistencyPolicy string
const (
    NO_GAPS_ALLOWED    TimelineConsistencyPolicy = "NO_GAPS"      // 不允许时间线间隙
    NO_OVERLAPS        TimelineConsistencyPolicy = "NO_OVERLAPS"  // 不允许重叠
    CONTINUOUS_HISTORY TimelineConsistencyPolicy = "CONTINUOUS"   // 连续历史记录
)
```

**3.2 追溯处理支持**
```go
// 追溯处理配置
type RetroactivityConfig struct {
    SupportsRetroactivity            bool     `json:"supports_retroactivity"`
    RetroactivityTriggersRecalculation []string `json:"retroactivity_triggers"` // ["PAYROLL", "ACCRUALS"]
    MaxRetroactiveDays              int      `json:"max_retroactive_days"`
}
```

**3.3 时间线管理操作**
```go
// 替代传统DELETE的时间线操作
POST /api/v1/organization-units/{code}/timeline/correct   // 校正历史记录
POST /api/v1/organization-units/{code}/timeline/cancel    // 取消未来变更  
POST /api/v1/organization-units/{code}/timeline/void      // 撤销已生效变更
```

## API规范文档更新

### 扩展数据模型
```json
{
  "code": "1000001",
  "name": "技术部", 
  "unit_type": "DEPARTMENT",
  "status": "ACTIVE",
  "effective_date": "2025-08-01",        // 新增：生效日期
  "end_date": null,                      // 新增：失效日期  
  "version": 1,                          // 新增：版本号
  "supersedes_version": null,            // 新增：替代版本
  "change_reason": "组织架构调整",         // 新增：变更原因
  "is_current": true,                    // 新增：当前版本标识
  "created_at": "2025-08-04T00:00:00Z",
  "updated_at": "2025-08-04T00:00:00Z"
}
```

### 元合约符合性配置
```yaml
# 组织架构API元合约配置
temporality_paradigm: EVENT_DRIVEN
timeline_consistency_policy: NO_GAPS_ALLOWED  
supports_future_dating: true
supports_retroactivity: true
retroactivity_triggers_recalculation: ["PAYROLL", "POSITION_ASSIGNMENTS"]

timeline_query_parameters:
  as_of_date: 
    type: "date"
    description: "查询指定时间点的组织架构状态"
  effective_range:
    from_date: "date" 
    to_date: "date"
    description: "查询指定时间范围内的变更历史"
```

## 实施优先级与风险控制

### Phase 1 优先级 (高)
1. 时间点查询能力（业务需求最迫切）
2. 基础版本管理（数据完整性保障）
3. 兼容性API封装（现有功能无影响）

### Phase 2 优先级 (中)
1. 事件驱动变更API
2. 历史版本查询
3. 追溯处理支持

### Phase 3 优先级 (低)
1. 完整时间线一致性检查
2. 复杂业务规则验证
3. 下游系统集成

### 风险控制措施
- 🛡️ **双轨运行**：新旧API同时支持6个月过渡期
- 🛡️ **渐进迁移**：现有数据自动生成version=1, effective_date=created_at
- 🛡️ **兼容性保证**：现有前端代码无需修改
- 🛡️ **回滚机制**：每个阶段都支持快速回滚到前一版本

## 业务价值评估

### 立即价值
- ✅ 支持"查看2024年12月31日的组织架构"等业务查询
- ✅ 完整的组织变更审计跟踪
- ✅ 支持HR系统的追溯薪酬计算

### 中期价值  
- ✅ 符合企业级HR系统合规要求
- ✅ 支持复杂的组织重组场景
- ✅ 为AI分析提供完整的时序数据

### 长期价值
- ✅ 完全符合元合约v6.0企业级标准
- ✅ 支持多租户时态数据隔离
- ✅ 可扩展到员工、职位等其他核心实体

## 决策结果

**采纳该渐进式升级方案**，建议优先启动Phase 1实施，预计4周完成基础时态能力，为业务提供立即价值，同时为后续完整事件驱动架构奠定基础。

## 后续行动

1. **立即执行**: Phase 1数据模型扩展设计与实施
2. **4周后**: Phase 2 API扩展开发
3. **7周后**: Phase 3事件驱动重构
4. **13周后**: 完整合规性验证与性能优化

---

**文档版本**: v1.0  
**最后更新**: 2025-08-10  
**相关文档**: 
- [元合约v6.0规范](../architecture-foundations/metacontract-v6.0-specification.md)
- [组织架构API规范](../api-specifications/organization-units-api-specification.md)