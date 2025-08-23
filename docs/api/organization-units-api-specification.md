# 组织单元管理API规范

**版本**: v4.2 - 数据模型命名规范统一版  
**创建日期**: 2025-08-04  
**最后更新**: 2025-08-23  
**基于实际实现**: ✅ 已验证  
**架构**: 严格CQRS + PostgreSQL原生单一数据源 + OAuth 2.0企业级安全  
**状态**: 开发中 + 数据模型命名一致性完成

## 📋 概述

组织单元管理API提供完整的企业组织架构管理功能，基于CQRS架构实现高性能读写分离，采用统一的企业级标准格式，支持时态数据管理、层级结构、多种单元类型和灵活的配置选项，实现多租户隔离和完整的CRUD操作。

### 🚀 CQRS架构特性 ⭐ **严格协议唯一性**

**统一企业级标准格式**:
- **查询操作**: GraphQL接口 - `/graphql` (时态查询优化)
- **命令操作**: REST API - `/api/v1/organization-units` (标准CRUD)
- **统一响应格式**: 企业级标准JSON结构，无向后兼容负担

**PostgreSQL原生架构**:
- **查询端**: PostgreSQL GraphQL服务，1.5-8ms极致响应性能
- **命令端**: PostgreSQL REST API服务，保证数据一致性
- **单一数据源**: 消除同步延迟，保证强一致性
- **26个专用索引**: 时态查询性能达到极致

**协议唯一性强制执行** 🚨:
```yaml
# 严格CQRS协议分离
查询操作 (只能用GraphQL):
  - 获取单个组织: organization(code: String!)
  - 获取组织列表: organizations(filter: OrganizationFilter)
  - 获取统计信息: organizationStats
  - 获取层级结构: organizationHierarchy, organizationSubtree
  
命令操作 (只能用REST API):
  - 创建: POST /api/v1/organization-units
  - 更新: PUT/PATCH /api/v1/organization-units/{code}
  - 状态操作: POST /api/v1/organization-units/{code}/suspend
  - 删除: DELETE /api/v1/organization-units/{code}

# 绝对禁止的协议违反
❌ REST GET端点: 不存在任何REST查询端点 (如GET /api/v1/organization-units/{code})
❌ GraphQL Mutation命令: 所有变更操作必须通过REST API完成
❌ 协议混用: 同一功能不能同时存在于两种协议中

# 唯一性原则保证
每种功能只有一种协议实现，消除API消费者的协议选择困惑。
```

### 🌳 高级层级管理系统 ⭐ **新增 v7.0**

**17级深度限制**:
- 支持1-17级的深度组织架构
- 自动约束检查防止超深层级
- 智能层级深度缓存优化查询性能

**双路径系统**:
- **编码路径** (`codePath`): `/1000000/1000001/1000002` - 主要层级路径字段
- **名称路径** (`namePath`): `/高谷集团/爱治理办公室/技术部` - 人类可读的路径表示

**智能级联更新** ⭐ **业务操作自动级联**:
- 父组织编码变更时自动更新所有子组织
- 组织名称变更时自动更新名称路径
- 异步通知机制确保性能不受影响
- 递归CTE优化大规模层级更新
- **触发时机**: 业务操作 (CREATE/UPDATE/SUSPEND/REACTIVATE) 自动触发

**手动层级维护** ⭐ **运维工具专用**:
```yaml
使用场景:
  数据修复: 处理由于意外情况导致的层级不一致
  系统迁移: 历史数据迁移后的层级重建
  数据修复: 人工修正数据库异常后的一致性恢复
  性能优化: 定期重建缓存索引和路径优化
  
机制区别:
  智能级联: 业务操作触发→自动级联更新→数据一致性保证
  手动刷新: 运维工具→手动触发→修复异常或优化性能

设计原则:
  主要机制: 智能级联更新承担日常业务操作的数据一致性维护
  备用机制: 手动刷新仅用于异常情况下的数据修复
  权限隔离: 运维工具需要特殊权限，防止误操作
```

**循环引用防护**:
- 智能检测循环引用并阻止操作
- 深度限制防止无限递归
- 完整性约束保证数据一致性

### 🏷️ 标识符设计说明 ⭐ **统一性原则强制执行**

**重要变更**: 本API采用全新的标识符命名策略，详见[ADR-006标识符命名策略](../architecture-decisions/ADR-006-identifier-naming-strategy.md)

**唯一性原则执行**: 为消除标识符使用的不一致性，本文档已清理所有旧的UUID引用示例

```yaml
对外标识符: 
  - 主要字段: "code" (7位数字编码，如 "1000001")
  - 关系引用: "parent_code" (引用父级组织编码)
  - 业务含义: 组织编码，业务人员直观理解

内部标识符:
  - UUID仅在系统内部使用，完全对外隐藏
  - 数据库主键继续使用UUID确保性能
  - API响应中不包含任何UUID字段

设计优势:
  - 降低用户认知负担 (只需理解一种ID)
  - 符合企业级HR系统行业标准
  - 提供更直观的业务语义

文档统一性清理 (v3.1):
  - 移除所有parent_unit_id引用，统一使用parentCode
  - 移除所有id字段引用，统一使用code
  - 移除所有unit_type引用，统一使用unitType
  - 标注违反统一性的旧示例为"已废弃"
  - 确保所有API示例符合新标识符规范
```

### 核心特性
- **高级层级结构**: 支持1-17级深度的组织架构，自动级联更新
- **双路径系统**: 编码路径+名称路径，支持多维度组织导航
- **智能级联更新**: 父组织变更时自动更新所有子组织层级路径
- **多种类型**: 部门、成本中心、公司、项目团队等
- **多态配置**: 基于单元类型的动态配置
- **多租户隔离**: 严格的租户数据边界
- **循环引用防护**: 智能检测和阻止循环引用
- **关联管理**: 与职位和员工的关联关系
- **REST标准合规**: 严格遵循PUT/PATCH语义，符合HTTP标准
- **极致架构简化**: 移除冗余CDC发布订阅，消除技术债务，实现真正的单一数据源

### 🏢 层级管理设计哲学 ⭐ **双机制合理性说明**

#### 设计原则：业务操作优先 + 运维工具备用
```yaml
主要机制 (智能级联更新):
  设计目标: 99%+的日常业务操作无需手动干预
  触发时机: 父组织编码/名称变更时自动执行
  数据一致性: 事务一致性保证，无时间窗口问题
  性能优化: 异步处理 + CTE递归优化 + 批量更新
  错误处理: 级联更新失败时业务操作自动回滚
  用户体验: 开发者和用户无需理解复杂的层级维护逻辑

备用机制 (手动层级刷新):
  设计目标: 处理<1%的异常情况和运维需求
  使用场景: 数据迁移、数据修复、系统异常恢复
  权限控制: 仅向运维人员开放，防止误操作
  操作记录: 完整的审计日志和影响评估
  安全机制: dry_run模式 + 强制确认 + 批量操作限制
```

#### 为什么需要双机制？ 🤔 **常见疑问解答**
```yaml
Q1: “智能级联更新已经可以处理所有情况，为什么还要手动刷新？”
A1: 实际中存在智能级联无法处理的边界情况:
    - 历史数据迁移后的层级重建 (数据没有触发正常的业务操作)
    - 数据库直接修改后的一致性恢复 (绕过API层)
    - 系统异常中断后的数据修复 (智能级联未正常执行)
    - 性能调优需求 (定期重建缓存索引)

Q2: “这不是功能重复吗？”
A2: 不是功能重复，而是职责分离:
    - 智能级联: 业务数据一致性维护 (日常运营)
    - 手动刷新: 系统维护和异常恢复 (运维工具)
    - 类比: 汽车的“自动驾驶”和“手动驾驶”不是功能重复

Q3: “这会增加API的复杂性吗？”  
A3: 对普通用户不会，对运维人员有价值:
    - 99%+的API用户只需要知道智能级联更新存在
    - <1%的运维场景在没有备用方案时会面临无解的困境
    - 权限控制确保运维工具不会被误用
```

#### 设计哲学对比 🎨 **架构思考**
```yaml
纯单一机制方案 (被否决):
  “只用智能级联更新”:
    ✓ API简洁性高
    ✗ 数据迁移时无法处理历史数据
    ✗ 数据库维护后无法恢复一致性
    ✗ 系统异常时缺乏强制修复能力

  “只用手动刷新”:
    ✓ 灵活性高，可处理所有情况
    ✗ 用户负担重，需要理解复杂逻辑
    ✗ 人为错误风险高
    ✗ 数据一致性风险高

双机制方案 (当前设计):
  “智能主导 + 手动备用”:
    ✓ 日常使用简单无脑
    ✓ 异常情况有对策
    ✓ 数据一致性和灵活性兼得
    ✗ API表面上显得复杂 (但实际使用简单)
```

#### 实际使用统计预期 📊 **数据驱动决策**
```yaml
API使用频率预期 (基于企业级系统经验):
  智能级联更新: 99.5%+ (日常业务操作)
  手动层级刷新: <0.5% (运维场景)
  一致性检查: 1-2次/月 (定期检查)

用户类型分布预期:
  业务开发者: 95% (只使用智能级联，无需关心手动刷新)
  系统运维人员: 5% (同时使用两种机制)
  关键洞察: 大多数用户永远不会接触到手动刷新API
```

## 🏗️ API端点总览 (企业级标准格式 v3.0)

### 查询操作 (GraphQL)
| 操作 | GraphQL Query/Mutation | 描述 | 认证 |
|------|------------------------|------|------|
| Query | `organizations(filter: OrgFilter)` | 获取组织单元列表(时态查询优化) | Bearer Token |
| Query | `organization(code: String!)` | 获取单个组织单元 | Bearer Token |
| Query | `organizationStats` | 获取组织统计信息 | Bearer Token |
| Query | `organizationAuditHistory(code: String!)` | 组织完整审计历史 | Bearer Token |
| Query | `auditLog(auditId: String!)` | 详细审计记录查询 | Bearer Token |
| Query | `organizationChangeAnalysis(code: String!)` | 跨版本变更分析 | Bearer Token |
| **Query** | **`organizationHierarchy(code: String!)`** | **获取组织完整层级路径信息** | Bearer Token |
| **Query** | **`organizationSubtree(code: String!)`** | **获取组织子树结构** | Bearer Token |
| **Query** | **`hierarchyStatistics`** | **层级分布统计信息** | Bearer Token |
| **Query** | **`hierarchyConsistencyCheck`** | **层级一致性检查报告** | Bearer Token |

### 命令操作 (REST API) ⭐ **统一命名规范 v3.2**

#### 标准CRUD操作
| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/v1/organization-units` | 创建组织单元 | Bearer Token |
| PUT | `/api/v1/organization-units/{code}` | 完全替换组织单元 (完整资源更新) | Bearer Token |
| PATCH | `/api/v1/organization-units/{code}` | 部分更新组织单元 (仅限常规数据字段) | Bearer Token |
| DELETE | `/api/v1/organization-units/{code}` | 删除组织单元 | Bearer Token |

#### 专用业务操作端点 🚨 **统一命名规范**
| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| **POST** | **`/api/v1/organization-units/{code}/suspend`** | **停用组织单元** | Bearer Token |
| **POST** | **`/api/v1/organization-units/{code}/activate`** | **激活组织单元** | Bearer Token |
| **POST** | **`/api/v1/organization-units/validate`** | **验证组织数据** | Bearer Token |

#### 系统运维端点 ⭐ **运维工具专用 - 严格CQRS合规**
| 方法 | 端点 | 描述 | 认证 | 权限要求 |
|------|------|------|------|----------|
| **POST** | **`/api/v1/organization-units/{code}/refresh-hierarchy`** | **手动刷新层级结构** | Bearer Token | `hr.organization.maintenance` |
| **POST** | **`/api/v1/organization-units/batch-refresh-hierarchy`** | **批量刷新层级结构** | Bearer Token | `hr.organization.maintenance` |

**重要说明**: 层级一致性检查现在统一使用GraphQL查询 `hierarchyConsistencyCheck`，符合严格的CQRS协议分离原则。

## 📊 数据模型

### 组织单元核心模型 (高级层级管理版 v7.0) ⭐
```json
{
  "code": "string (7位数字: 1000000-9999999, 主键)",
  "parentCode": "string (optional, 父级组织编码)",
  "tenantId": "uuid (租户ID)",
  "name": "string (组织名称, 最大255字符)",
  "unitType": "DEPARTMENT | COST_CENTER | COMPANY | PROJECT_TEAM",
  "status": "ACTIVE | INACTIVE (纯业务状态, 默认ACTIVE)",
  "isDeleted": "boolean (删除标记, 默认false)",
  "level": "number (层级, 1-17级, 默认1)",
  "hierarchyDepth": "number (层级深度缓存, 与level同步, 1-17级)",
  "codePath": "string (编码路径: /1000000/1000001/1000002, 最大2000字符)",
  "namePath": "string (名称路径: /高谷集团/爱治理办公室/技术部, 最大4000字符)",
  "sortOrder": "number (排序, 默认0)",
  "description": "string (optional, 描述)",
  "profile": "object (JSONB配置信息, 默认{})",
  "createdAt": "timestamp (记录创建时间)",
  "updatedAt": "timestamp (最后操作时间, 含义由operationType确定)",
  "operationType": "CREATE | UPDATE | SUSPEND | REACTIVATE | DELETE (操作类型, 只读字段)",
  "operatedBy": {
    "id": "uuid (操作人ID)",
    "name": "string (操作人姓名)"
  },
  "operationReason": "string (optional, 操作原因, 最大500字符)",
  "effectiveDate": "date (生效日期, 主键)",
  "endDate": "date (optional, 结束日期)",
  "isCurrent": "boolean (动态计算字段, 基于asOfDate参数判断是否在指定日期生效)",
  "isFuture": "boolean (动态计算字段, 基于asOfDate参数判断是否为未来生效记录)",
  "recordId": "uuid (记录唯一ID)"
}
```

### 操作类型枚举 ✨ **唯一性原则版 v8**
```yaml
CREATE: 新建操作 - updatedAt表示创建时间，默认status=ACTIVE
UPDATE: 变更操作 - updatedAt表示最后修改时间，由PUT/PATCH端点自动设置  
SUSPEND: 停用操作 - updatedAt表示停用时间，强制status=INACTIVE
REACTIVATE: 重新启用操作 - updatedAt表示启用时间，强制status=ACTIVE
DELETE: 删除操作 - updatedAt表示删除时间，强制isDeleted=true

# 重要说明：operationType完全由API端点自动确定，不接受用户输入
```

### 操作与状态强制绑定规则 ⭐ **新增 v7**
```yaml
# 强制绑定映射
SUSPEND操作 → 强制设置 status=INACTIVE (专用端点 /suspend)
REACTIVATE操作 → 强制设置 status=ACTIVE (专用端点 /activate)  
CREATE操作 → 默认设置 status=ACTIVE
DELETE操作 → 设置 isDeleted=true，status保持删除前状态

# 端点→operationType固定映射规则 ⭐ **唯一性原则强制执行**
所有operationType完全由API端点决定，用户无法指定此字段：

## CRUD操作端点映射
- POST /api/v1/organization-units → operationType=CREATE
- PUT /api/v1/organization-units/{code} → operationType=UPDATE  
- PATCH /api/v1/organization-units/{code} → operationType=UPDATE
- DELETE /api/v1/organization-units/{code} → operationType=DELETE

## 专用业务操作端点映射
- POST /api/v1/organization-units/{code}/suspend → operationType=SUSPEND
- POST /api/v1/organization-units/{code}/activate → operationType=REACTIVATE

## 业务数据验证端点映射 (不记录operationType)
- POST /api/v1/organization-units/validate → 数据验证操作，不修改数据

## 运维系统端点映射 ⭐ **运维工具专用** (不记录operationType)
- POST /api/v1/organization-units/{code}/refresh-hierarchy → 运维工具，手动修复层级不一致问题
- POST /api/v1/organization-units/batch-refresh-hierarchy → 运维工具，批量修复层级不一致问题

# 已移除的推导规则 (违反唯一性原则)
# ❌ 特殊操作检测: isDeleted变化 → operationType=DELETE (应使用DELETE端点)
# ❌ 状态变化推导: ACTIVE→INACTIVE=SUSPEND (应使用专用/suspend端点)
# ❌ 状态变化推导: INACTIVE→ACTIVE=REACTIVATE (应使用专用/activate端点)
```

### 数据库表结构详情

#### 主键设计 (优化后)
- **复合主键**: `(code, effectiveDate)` - 支持时态数据管理
- **唯一约束**: `recordId` - 每条记录全局唯一标识
- **动态当前版本**: 无物理约束，通过查询逻辑确定当前生效记录

#### 时态字段说明 (优化后 v6)
```yaml
effectiveDate: 记录生效日期，支持历史版本管理和未来生效
endDate: 记录结束日期，NULL表示无明确结束时间
isCurrent: 动态计算字段，表示该记录在指定日期(asOfDate)是否生效
  计算逻辑: effectiveDate <= asOfDate AND (endDate IS NULL OR endDate >= asOfDate) AND isDeleted = false
isFuture: 动态计算字段，表示该记录相对于指定日期(asOfDate)是否为未来生效
  计算逻辑: effectiveDate > asOfDate AND isDeleted = false

# 记录状态逻辑 (基于isCurrent + isFuture组合)
历史记录: isCurrent = false AND isFuture = false (已结束的记录)
当前记录: isCurrent = true AND isFuture = false (在指定日期生效的记录)  
未来记录: isCurrent = false AND isFuture = true (在指定日期之后才生效的记录)

# 时态状态完整矩阵
状态组合说明:
  isCurrent=true  + isFuture=false → 当前生效记录
  isCurrent=false + isFuture=true  → 未来生效记录  
  isCurrent=false + isFuture=false → 历史过期记录
  isCurrent=true  + isFuture=true  → 不可能的状态 (逻辑矛盾)

# 设计优化说明 ⭐
移除isTemporal字段的原因:
  - 功能冗余: 完全可由isCurrent和isFuture组合替代
  - 认知简化: 减少字段数量，降低API消费者理解成本  
  - 正交设计: 两个布尔字段的组合提供完整的时态状态表达
  - 查询优化: 直接基于effectiveDate和endDate计算，无需额外存储

移除path字段的原因:
  - 数据重复: 与codePath完全相同，违反DRY原则
  - 存储浪费: 每条记录存储两份相同数据
  - 维护复杂: 需要保持两个字段的同步一致性
  - 认知负荷: 给API消费者造成困惑，不清楚使用哪个字段
  - 设计简化: codePath语义更明确，与namePath形成清晰对比
```

#### 多版本审计信息管理机制 ⭐ **新增 v7**

#### 双重关联审计架构
```yaml
# 审计记录与时态记录的关联关系
business_entity_level:    # 业务实体层（组织维度）
  identifier: code        # 组织编码 (如: 1000001)
  purpose: 跨版本关联     # 关联同一组织的所有历史版本
  
version_record_level:     # 版本记录层（记录维度）  
  identifier: record_id   # 版本记录UUID
  purpose: 精确定位       # 关联具体的时态记录版本

# 审计记录结构
audit_record_structure:
  audit_id: 审计记录唯一标识
  business_entity_id: 组织编码 (跨版本关联)
  record_id: 具体版本记录ID (精确关联)
  version_sequence: 版本序号 (1,2,3,4...)
  operation: 操作类型
  before_data: 操作前完整数据快照
  after_data: 操作后完整数据快照
  field_changes: 字段级变更摘要

# 版本关系示例
组织实体(1000001):
  ├── 版本1 (record_id: uuid-v1) → 审计记录1 (CREATE)
  ├── 版本2 (record_id: uuid-v2) → 审计记录2 (UPDATE) 
  ├── 版本3 (record_id: uuid-v3) → 审计记录3 (SUSPEND)
  └── 版本4 (record_id: uuid-v4) → 审计记录4 (REACTIVATE)
```

#### 审计字段说明 (优化后 v7)
```yaml
# 核心审计字段
recordId: 全局唯一记录ID，用于跨版本关联和审计追踪
createdAt: 记录创建时间（数据记录物理创建时间）
updatedAt: 最后操作时间（业务操作逻辑时间，含义由operationType确定）
operationType: 操作类型（明确updatedAt的业务含义）
operatedBy: 操作人ID（责任追溯，记录谁执行了操作）
operationReason: 操作原因（可选，记录为什么执行操作）
isDeleted: 删除标记（独立于业务状态的删除状态）

# 操作审计三元组 ⭐ 核心设计
operationType + operatedBy + operationReason = 完整操作审计链
  - 做了什么操作 (operationType)
  - 谁执行的操作 (operatedBy)  
  - 为什么执行操作 (operationReason)

# 操作语义映射  
create: updatedAt = 创建时间, operatedBy = 创建人, operationReason = "业务扩展需要"
update: updatedAt = 修改时间, operatedBy = 修改人, operationReason = "预算调整和人员编制优化"
suspend: updatedAt = 停用时间, operatedBy = 停用人, status = inactive, operationReason = "业务调整"
activate: updatedAt = 启用时间, operatedBy = 启用人, status = active, operationReason = "恢复运营"
delete: updatedAt = 删除时间, operatedBy = 删除人, isDeleted = true, operationReason = "合并到其他部门"

# 字段职责清晰分工
时间审计: createdAt, updatedAt
操作审计: operationType, operatedBy, operationReason  
时态管理: effectiveDate, endDate, isCurrent, isFuture
状态管理: status, isDeleted
元数据: recordId, tenantId

# 审计查询能力
按组织查询: WHERE businessEntityId = '1000001' ORDER BY timestamp DESC
按版本查询: WHERE recordId = 'uuid-version-3'
按用户查询: WHERE operatedBy = 'user-uuid' ORDER BY timestamp DESC
按操作查询: WHERE operationType = 'suspend' AND riskLevel = 'HIGH'
变更分析: 连接相邻版本的beforeData/afterData进行对比分析
```

### 时态系统中的操作历史 (优化后 v3)
```yaml
# 示例：组织单元完整生命周期 (消除冗余字段后)
版本1: CREATE, status=ACTIVE, is_deleted=false, operated_by=user123, operation_reason="部门重组", updated_at=2025-01-01, is_current=false
版本2: SUSPEND, status=INACTIVE, is_deleted=false, operated_by=user456, operation_reason="业务调整", updated_at=2025-02-01, is_current=false  
版本3: REACTIVATE, status=ACTIVE, is_deleted=false, operated_by=user789, operation_reason="恢复运营", updated_at=2025-03-01, is_current=false
版本4: DELETE, status=ACTIVE, is_deleted=true, operated_by=user456, operation_reason="合并到其他部门", updated_at=2025-04-01, is_current=true

# 操作审计三元组示例
操作1: CREATE + user123 + "部门重组" = 完整的创建操作审计
操作2: SUSPEND + user456 + "业务调整" = 完整的停用操作审计
操作3: REACTIVATE + user789 + "恢复运营" = 完整的恢复操作审计
操作4: DELETE + user456 + "合并到其他部门" = 完整的删除操作审计

# 字段简化优势
- 消除冗余：移除change_reason，避免概念重叠
- 职责清晰：operation_reason承担唯一的原因记录职责
- 审计完整：操作审计三元组提供完整的操作轨迹
- 用户友好：无需困惑应该填写哪个原因字段

# 查询能力
活跃组织: WHERE is_deleted = false AND is_current = true
操作人追踪: WHERE operated_by = 'user456' ORDER BY updated_at DESC
操作类型统计: SELECT operation_type, COUNT(*) FROM organization_units GROUP BY operation_type
删除原因分析: SELECT operation_reason, COUNT(*) FROM organization_units WHERE operation_type = 'DELETE' GROUP BY operation_reason
```
```

### 单元类型枚举
```yaml
DEPARTMENT: 部门（常规业务部门）
COST_CENTER: 成本中心（财务管理单元）
COMPANY: 公司（法人实体）
PROJECT_TEAM: 项目团队（临时性组织）
```

### 状态枚举 (优化后 v7)
```yaml
# 业务状态 (status字段)
ACTIVE: 活跃状态 (默认值)
INACTIVE: 非活跃状态

# 删除标记 (is_deleted字段)
false: 正常状态 (默认值)
true: 已删除（软删除）
```

### 状态组合说明 (优化后 v7)
```yaml
# 正常业务状态
status=ACTIVE + is_deleted=false: 正常运行的组织
status=INACTIVE + is_deleted=false: 非活跃但存在的组织

# 删除状态 (保留删除前的业务状态)
status=ACTIVE + is_deleted=true: 删除前是活跃状态
status=INACTIVE + is_deleted=true: 删除前是非活跃状态

# 时态状态 (通过动态计算字段表达)
is_future=true: 计划中的组织 (未来生效)
is_current=true: 当前生效的组织
is_current=false + is_future=false: 历史组织

# 查询模式 (基于is_current + is_future组合)
活跃记录: WHERE is_deleted = false
已删除记录: WHERE is_deleted = true  
特定业务状态: WHERE status = 'ACTIVE' AND is_deleted = false
当前生效组织: WHERE is_current = true AND is_deleted = false
未来计划组织: WHERE is_future = true AND is_deleted = false  
历史过期组织: WHERE is_current = false AND is_future = false AND is_deleted = false
所有时态状态: WHERE is_deleted = false (包含当前、未来、历史记录)
```

## 🗃️ 数据库性能优化

### 索引系统 (优化后)
```yaml
# 主键和唯一索引
organization_units_pkey: 复合主键 (code, effective_date)
idx_org_units_record_id: 记录ID唯一索引

# 动态is_current查询优化索引 ⭐ 核心优化
idx_current_effective_optimized: 当前生效记录专用索引
  (tenant_id, code, effective_date DESC, end_date DESC NULLS LAST) 
  WHERE is_deleted = false
  
idx_current_date_range: 日期范围查询优化
  (tenant_id, effective_date, end_date) 
  WHERE is_deleted = false AND effective_date <= CURRENT_DATE

# 时态查询优化索引  
idx_org_temporal_core: 时态核心查询 
  (tenant_id, code, effective_date DESC, end_date DESC NULLS LAST)
idx_temporal_range_query: 时态范围查询 
  (code, effective_date, end_date) WHERE effective_date IS NOT NULL

# 业务查询优化
idx_org_units_parent_code: 父级编码查询
idx_org_units_tenant_status: 租户状态查询 (tenant_id, status, is_deleted)
idx_org_status_type: 状态类型查询 (tenant_id, status, unit_type, effective_date DESC)
idx_org_effective_date: 生效日期索引
idx_org_operation_type: 操作类型查询 (tenant_id, operation_type, updated_at DESC)

# 审计查询优化
idx_org_operated_by: 操作人查询 (operated_by, updated_at DESC)
idx_org_audit_trail: 审计追踪 (tenant_id, code, updated_at DESC, operation_type)

# 全文搜索优化
idx_org_units_name_gin: 组织名称GIN索引 (支持模糊搜索)

# 删除状态优化
idx_org_not_deleted: 非删除记录查询 (tenant_id, is_deleted, status) WHERE is_deleted = false
```

### 动态时态字段实现 (asOfDate参数化)
```yaml
# API参数
asOfDate: 指定日期参数 (默认今天)
  - 格式: YYYY-MM-DD
  - 默认值: CURRENT_DATE
  - 示例: ?asOfDate=2025-12-31

# 单条记录查询 (基于指定日期)
查询逻辑:
  SET @asOfDate = COALESCE(:asOfDate, CURRENT_DATE);
  
  SELECT *,
    -- 动态计算is_current
    (effective_date <= @asOfDate 
     AND (end_date IS NULL OR end_date >= @asOfDate) 
     AND is_deleted = false) as is_current,
     
    -- 动态计算is_future  
    (effective_date > @asOfDate 
     AND is_deleted = false) as is_future
     
  FROM organization_units
  WHERE tenant_id = ? AND code = ?
  ORDER BY effective_date DESC;

# 批量查询优化
WITH temporal_records AS (
  SELECT *,
    (effective_date <= @asOfDate 
     AND (end_date IS NULL OR end_date >= @asOfDate) 
     AND is_deleted = false) as is_current,
    (effective_date > @asOfDate 
     AND is_deleted = false) as is_future,
    ROW_NUMBER() OVER (
      PARTITION BY code 
      ORDER BY effective_date DESC
    ) as rn
  FROM organization_units 
  WHERE tenant_id = ?
    AND is_deleted = false
)
SELECT * FROM temporal_records 
WHERE is_current = true OR :include_future = true;

# GraphQL层实现
type Organization {
  # 基于asOfDate参数动态计算
  isCurrent: Boolean! 
  isFuture: Boolean!
}

# 查询参数
input OrganizationFilter {
  asOfDate: Date # 默认今天
  includeFuture: Boolean # 是否包含未来记录
  onlyFuture: Boolean # 只返回未来记录
}
```

### 数据验证规则体系 ⭐ **新增 v7**

#### 分层验证架构
```yaml
# 第1层：数据库硬约束 (不可违反)
database_constraints:
  时间逻辑约束:
    - effective_date <= COALESCE(end_date, '9999-12-31')
    - 同一code不能有时间重叠的记录
    - 每个code最多只能有一条无结束日期的记录
  
  不可变字段约束:
    - tenant_id: 绝对不可修改
    - code: 组织编码不可修改  
    - operation_type: 由API端点自动设置，用户不可指定或修改
    - 删除后不可恢复: is_deleted=true后不能改为false

# 第2层：触发器业务约束 (强制执行)
trigger_validation:
  状态操作一致性:
    - SUSPEND操作 → 强制status=INACTIVE
    - REACTIVATE操作 → 强制status=ACTIVE
    - DELETE操作 → 强制is_deleted=true
    
  时态序列验证:
    - 已删除记录后不能创建新版本
    - 状态转换必须符合操作语义
    - 防止不合理的操作序列

# 第3层：API层验证 (用户友好)
api_validation:
  父子关系时态约束:
    - 子组织生命周期必须在父组织生命周期内
    - 父组织删除策略: REJECT_WITH_CHILDREN (推荐)
    - 防止循环引用和层级深度检查
    
  复杂业务规则:
    - 操作序列合理性检查
    - 未来操作时间限制 (最多365天)
    - 批量操作一致性验证

# 第4层：应用层配置 (可定制)
configurable_rules:
  temporal_constraints:
    allow_time_gaps: false              # 是否允许时间间隙
    max_future_operations_days: 365     # 最远未来操作天数
    require_end_date_for_suspend: false # SUSPEND是否必须指定结束日期
    
  parent_child_rules:
    deletion_policy: "REJECT_WITH_CHILDREN"
    max_hierarchy_depth: 10
    allow_cross_tenant_reference: false
    
  field_validation:
    immutable_fields: ["tenant_id", "code"]
    restricted_fields: ["unit_type"]    # 需要特殊权限才能修改
    audit_required_fields: ["name", "parent_code"]
```

#### 验证API端点 ⭐ **唯一性原则保证**
```bash
# 数据验证专用端点
POST /api/v1/organization-units/validate
{
  "operation": "create|update|suspend|activate|delete",
  "data": {
    "code": "1000001",
    "status": "INACTIVE", 
    "effectiveDate": "2025-09-01"
  },
  "dryRun": true
}

# 验证响应示例
{
  "valid": true,
  "warnings": ["父组织将在6个月后过期"],
  "errors": [],
  "suggestions": ["建议设置结束日期以明确停用期限"],
  "validationDetails": {
    "temporalCheck": "PASS",
    "parentChildCheck": "PASS", 
    "operationSequenceCheck": "PASS"
  }
}
```

#### 校验逻辑唯一性保证 🚨 **重要**
```yaml
# 核心设计原则
唯一校验模块: 
  - validate端点与所有操作端点(POST/PUT/PATCH)必须调用同一个核心校验服务
  - 禁止在不同端点中重复实现校验逻辑
  - 校验规则变更时只需修改一处核心代码

实现要求:
  - 所有端点必须import和调用共享的ValidationService
  - 禁止在controller层直接实现校验逻辑
  - 必须通过unit test验证校验逻辑一致性

API一致性承诺:
  - validate端点返回"valid: true"的数据，在实际操作端点中必须能成功执行
  - 任何校验规则的修改都会同时影响所有相关端点
  - 绝对避免"预校验通过但实际操作失败"的情况
```

### 触发器系统
```yaml
auto_end_date_trigger: 自动管理结束日期
organization_units_change_trigger: 组织变更通知
set_org_unit_code: 组织单元编码自动生成
temporal_gap_auto_fill: 时态间隙自动填充
update_organization_units_updated_at: 自动更新时间戳
```

### 架构简化清理记录 ⭐ **PostgreSQL原生架构优化**

**已移除的CDC发布订阅配置** (v3.1 - 2025-08-23):
```yaml
# ❌ 已清理的冗余配置
# dbz_publication: Debezium CDC发布 - 已移除
# debezium_org_publication: 组织专用Debezium发布 - 已移除  
# debezium_publication: 通用Debezium发布 - 已移除
# organization_publication: 组织变更发布 - 已移除

# 清理原因:
清理收益:
  架构一致性: 符合单一PostgreSQL数据源架构
  性能提升: 消除无用WAL日志生成，减少I/O开销
  维护简化: 移除4个冗余配置，降低系统复杂度
  资源优化: 释放发布订阅占用的内存和存储空间
  认知简化: 消除开发者对无用配置的困惑

技术债务清理:
  - CDC管道已移除，相关发布订阅失去存在意义
  - 单一数据库架构无需跨系统数据同步
  - PostgreSQL原生查询性能已达极致(1.5-8ms)
  - 架构简化60%，移除非必要复杂性
```

## 🔍 API详细规范

### 🌐 统一企业级标准格式

**GraphQL查询端点**: `http://localhost:8090/graphql`  
**REST命令端点**: `http://localhost:9090/api/v1`

**统一响应格式** ⭐ **企业级标准信封**:
```json
// GraphQL查询响应 (查询操作) - 保持GraphQL标准
{
  "data": {
    "organizations": [...],
    "pagination": {
      "total": 8,
      "page": 1,
      "hasNext": false,
      "pageSize": 50
    },
    "temporal": {
      "asOfDate": "2025-08-23",
      "currentCount": 5,
      "futureCount": 2,
      "historicalCount": 1
    }
  }
}

// REST API统一响应信封 (命令操作) - 成功与错误结构一致
// 成功响应
{
  "success": true,
  "data": {...},
  "message": "Organization unit created successfully",
  "timestamp": "2025-08-23T10:30:00Z",
  "requestId": "req_abc123"
}

// 错误响应 - 统一信封结构
{
  "success": false,
  "error": {
    "code": "ORG_UNIT_NOT_FOUND", 
    "message": "Organization unit not found",
    "details": null
  },
  "timestamp": "2025-08-23T10:30:00Z",
  "requestId": "req_def456"
}
```

### 1. 获取组织单元列表 (GraphQL查询)

#### GraphQL Query
```graphql
query GetOrganizations($filter: OrganizationFilter, $pagination: PaginationInput) {
  organizations(filter: $filter, pagination: $pagination) {
    data {
      code
      parentCode
      tenantId
      name
      unitType
      status
      isDeleted
      level
      hierarchyDepth
      codePath
      namePath
      sortOrder
      description
      profile
      createdAt
      updatedAt
      operationType
      operatedBy
      operationReason
      effectiveDate
      endDate
      isCurrent
      isFuture
      recordId
    }
    pagination {
      total
      page
      hasNext
      pageSize
    }
    temporal {
      asOfDate
      currentCount
      futureCount
      historicalCount
    }
  }
}
```

#### 查询参数类型定义
```graphql
input OrganizationFilter {
  asOfDate: Date # 时态查询基准日期 (默认今天)
  includeFuture: Boolean # 是否包含未来生效记录 (默认false)
  onlyFuture: Boolean # 只返回未来生效记录 (默认false)
  unitType: UnitType # 单元类型过滤
  status: Status # 状态过滤
  parentCode: String # 父单元代码过滤
  searchText: String # 名称模糊搜索
}

input PaginationInput {
  page: Int # 页码 (默认1)
  pageSize: Int # 每页大小 (默认50, 最大1000)
}
```

#### GraphQL响应示例 (企业级标准格式 v3.0)
```json
{
  "data": {
    "organizations": {
      "data": [
        {
          "code": "1000001",
          "parentCode": "1000000",
          "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
          "name": "技术部",
          "unitType": "DEPARTMENT",
          "status": "ACTIVE",
          "isDeleted": false,
          "level": 2,
          "hierarchyDepth": 2,
          "codePath": "/1000000/1000001",
          "namePath": "/高谷集团/技术部",
          "sortOrder": 0,
          "description": "技术研发部门",
          "profile": {
            "type": "rd",
            "budget": 5000000,
            "costCenterCode": "CC001"
          },
          "createdAt": "2025-08-05T11:23:01.426455Z",
          "updatedAt": "2025-08-23T06:13:47.072807Z",
          "operationType": "UPDATE",
          "operatedBy": {
            "id": "456e7890-e89b-12d3-a456-426614174002",
            "name": "Li Si"
          },
          "operationReason": "预算调整和人员编制优化",
          "effectiveDate": "2025-08-05",
          "endDate": null,
          "isCurrent": true,
          "isFuture": false,
          "recordId": "123e4567-e89b-12d3-a456-426614174000"
        }
      ],
      "pagination": {
        "total": 8,
        "page": 1,
        "pageSize": 50,
        "hasNext": false
      },
      "temporal": {
        "asOfDate": "2025-08-23",
        "currentCount": 5,
        "futureCount": 2,
        "historicalCount": 1
      }
    }
  }
}
```

#### GraphQL查询示例
```graphql
# 查看当前生效的记录 (默认行为)
query {
  organizations {
    data { code name status isCurrent }
    temporal { asOfDate currentCount }
  }
}

# 查看明年1月1日的组织架构
query {
  organizations(filter: { asOfDate: "2026-01-01" }) {
    data { code name status isCurrent }
    temporal { asOfDate currentCount }
  }
}

# 包含未来生效的记录
query {
  organizations(filter: { includeFuture: true }) {
    data { code name status isCurrent isFuture effectiveDate }
    temporal { asOfDate currentCount futureCount }
  }
}

# 只看未来生效的记录
query {
  organizations(filter: { onlyFuture: true }) {
    data { code name status isFuture effectiveDate }
    temporal { asOfDate futureCount }
  }
}

# 查看计划停用的组织
query {
  organizations(filter: { includeFuture: true, status: INACTIVE }) {
    data { code name status operationType effectiveDate }
    temporal { asOfDate futureCount }
  }
}
```

### 2. 高级层级管理查询 (GraphQL) ⭐ **新增**

#### 2.1 获取组织完整层级路径信息
```graphql
query GetOrganizationHierarchy($code: String!, $tenantId: UUID!) {
  organizationHierarchy(code: $code, tenantId: $tenantId) {
    code
    name
    level
    hierarchyDepth
    codePath
    namePath
    parentChain
    childrenCount
    isRoot
    isLeaf
  }
}
```

#### 2.2 获取组织子树结构 ⭐ **唯一性原则指引**
```graphql
query GetOrganizationSubtree($code: String!, $tenantId: UUID!, $maxDepth: Int) {
  organizationSubtree(code: $code, tenantId: $tenantId, maxDepth: $maxDepth) {
    code
    name
    level
    hierarchyDepth
    codePath
    namePath
    children {
      code
      name
      level
      hierarchyDepth
      codePath
      namePath
    }
  }
}
```

#### 查询选择指引 🚨 **避免功能重叠**
```yaml
# 为避免API功能重叠，明确不同查询的使用场景：

获取直接子节点 (推荐使用):
  查询方式: organizations(filter: { parentCode: "1000000" })
  适用场景: 只需要一级子节点列表
  性能特点: 基于索引的简单过滤，性能最佳
  返回格式: 标准列表格式，支持分页和排序

获取多级子树 (专用场景):
  查询方式: organizationSubtree(code: "1000000", maxDepth: 3)
  适用场景: 需要获取多级层级结构(maxDepth ≥ 2)
  性能特点: 递归CTE查询，适合层级展示
  返回格式: 树形嵌套结构，保持层级关系

# 避免的重叠使用方式
❌ 不推荐: organizationSubtree(maxDepth: 1) - 与简单过滤功能重叠
✅ 推荐: organizations(filter: { parentCode }) - 直接子节点查询的唯一标准方式

# 设计原则
每种具体需求只提供一种最优的查询方式，避免让API消费者在功能重叠的选项中困惑选择。
```

#### 2.3 层级分布统计信息
```graphql
query GetHierarchyStatistics($tenantId: UUID!) {
  hierarchyStatistics(tenantId: $tenantId) {
    tenantId
    totalOrganizations
    maxDepth
    avgDepth
    depthDistribution {
      depth
      count
      percentage
    }
    rootOrganizations
    leafOrganizations
    integrityIssues {
      type
      count
      affectedCodes
    }
  }
}
```

#### 2.4 层级一致性检查 ⭐ **新增运维工具**
```graphql
query HierarchyConsistencyCheck($tenantId: UUID!, $checkMode: ConsistencyCheckMode) {
  hierarchyConsistencyCheck(tenantId: $tenantId, checkMode: $checkMode) {
    checkId
    tenantId
    executedAt
    executionTimeMs
    totalChecked
    issuesFound
    checkMode
    consistencyReport {
      pathMismatches {
        code
        expectedCodePath
        actualCodePath
        expectedNamePath
        actualNamePath
        severity
      }
      levelInconsistencies {
        code
        expectedLevel
        actualLevel
        parentCode
        reason
      }
      orphanedNodes {
        code
        name
        parentCode
        reason
      }
      circularReferences {
        affectedCodes
        circularPath
        severity
      }
      depthViolations {
        code
        currentDepth
        maxAllowedDepth
        parentChain
      }
      cacheInconsistencies {
        code
        fieldName
        cachedValue
        calculatedValue
        impactLevel
      }
    }
    repairSuggestions {
      issueType
      affectedCodes
      suggestedAction
      automatable
      riskLevel
    }
    healthScore
    recommendedActions
  }
}

enum ConsistencyCheckMode {
  FAST        # 仅检查缓存字段与计算字段的一致性
  DEEP        # 全面检查包括递归验证和交叉检查
  TARGETED    # 针对特定区域的定制检查
}
```

#### 层级管理GraphQL响应示例
```json
{
  "data": {
    "organizationHierarchy": {
      "code": "1000001",
      "name": "技术部",
      "level": 2,
      "hierarchyDepth": 2,
      "codePath": "/1000000/1000001",
      "namePath": "/高谷集团/技术部",
      "parentChain": ["1000000", "1000001"],
      "childrenCount": 3,
      "isRoot": false,
      "isLeaf": false
    },
    "organizationSubtree": [
      {
        "code": "1000001",
        "name": "技术部",
        "level": 2,
        "hierarchyDepth": 2,
        "codePath": "/1000000/1000001",
        "namePath": "/高谷集团/技术部",
        "children": [
          {
            "code": "1000011",
            "name": "前端开发组",
            "level": 3,
            "hierarchyDepth": 3,
            "codePath": "/1000000/1000001/1000011",
            "namePath": "/高谷集团/技术部/前端开发组"
          }
        ]
      }
    ],
    "hierarchyStatistics": {
      "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
      "totalOrganizations": 25,
      "maxDepth": 4,
      "avgDepth": 2.3,
      "depthDistribution": [
        { "depth": 1, "count": 3, "percentage": 12.0 },
        { "depth": 2, "count": 15, "percentage": 60.0 },
        { "depth": 3, "count": 6, "percentage": 24.0 },
        { "depth": 4, "count": 1, "percentage": 4.0 }
      ],
      "rootOrganizations": 3,
      "leafOrganizations": 12,
      "integrityIssues": {
        "type": "path_mismatch",
        "count": 0,
        "affectedCodes": []
      }
    }
  }
}
```

### 3. 获取单个组织单元 (GraphQL查询)

#### GraphQL Query
```graphql
query GetOrganization($code: String!, $asOfDate: Date) {
  organization(code: $code, asOfDate: $asOfDate) {
    code
    parentCode
    tenantId
    name
    unitType
    status
    isDeleted
    level
    sortOrder
    description
    profile
    createdAt
    updatedAt
    operationType
    operatedBy
    operationReason
    effectiveDate
    endDate
    isCurrent
    isFuture
    recordId
  }
}
```

#### GraphQL响应示例
```json
{
  "data": {
    "organization": {
      "code": "1000001",
      "parentCode": "1000000",
      "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
      "name": "技术部",
      "unitType": "DEPARTMENT",
      "status": "ACTIVE",
      "isDeleted": false,
      "level": 2,
      "codePath": "/1000000/1000001",
      "namePath": "/高谷集团/技术部",
      "sortOrder": 0,
      "description": "技术研发部门",
      "profile": {
        "type": "rd",
        "budget": 5000000,
        "costCenterCode": "CC001"
      },
      "createdAt": "2025-08-05T11:23:01.426455Z",
      "updatedAt": "2025-08-06T06:13:47.072807Z",
      "operationType": "UPDATE",
      "operatedBy": {
        "id": "456e7890-e89b-12d3-a456-426614174002",
        "name": "Li Si"
      },
      "operationReason": "预算调整和人员编制优化",
      "effectiveDate": "2025-08-05",
      "endDate": null,
      "isCurrent": true,
      "isFuture": false,
      "recordId": "123e4567-e89b-12d3-a456-426614174000"
    }
  }
}
```
### 3. 获取组织统计信息 (GraphQL查询)

#### GraphQL Query
```graphql
query GetOrganizationStats($asOfDate: Date) {
  organizationStats(asOfDate: $asOfDate) {
    total
    byType {
      type
      count
    }
    byStatus {
      status
      count
    }
    byLevel {
      level
      count
    }
    temporal {
      current
      future
      historical
    }
  }
}
```

#### GraphQL响应示例
```json
{
  "data": {
    "organizationStats": {
      "total": 8,
      "byType": [
        { "type": "COMPANY", "count": 1 },
        { "type": "DEPARTMENT", "count": 7 }
      ],
      "byStatus": [
        { "status": "ACTIVE", "count": 8 }
      ],
      "byLevel": [
        { "level": 1, "count": 1 },
        { "level": 2, "count": 7 }
      ],
      "temporal": {
        "current": 5,
        "future": 2,
        "historical": 1
      }
    }
  }
}
```

### 4. 创建组织单元 (REST命令)

#### REST API Endpoint
**`POST /api/v1/organization-units`**

#### 请求体示例
```json
{
  "name": "新部门",
  "unitType": "DEPARTMENT",
  "parentCode": "1000000",
  "description": "新创建的部门",
  "effectiveDate": "2025-08-23",
  "operationReason": "业务扩展需要"
}
```

#### REST响应示例 ⭐ **统一信封结构**
```json
{
  "success": true,
  "message": "Organization unit created successfully",
  "data": {
    "code": "1000008",
    "parentCode": "1000000",
    "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
    "name": "新部门",
    "unitType": "DEPARTMENT",
    "status": "ACTIVE",
    "isDeleted": false,
    "level": 2,
    "codePath": "/1000000/1000008",
    "namePath": "/高谷集团/新部门",
    "sortOrder": 0,
    "description": "新创建的部门",
    "profile": {},
    "createdAt": "2025-08-23T15:00:00Z",
    "updatedAt": "2025-08-23T15:00:00Z",
    "operationType": "CREATE",
    "operatedBy": {
      "id": "789e0123-e89b-12d3-a456-426614174003",
      "name": "Zhang San"
    },
    "operationReason": "业务扩展需要",
    "effectiveDate": "2025-08-23",
    "endDate": null,
    "isCurrent": true,
    "isFuture": false,
    "recordId": "456e7890-e89b-12d3-a456-426614174008"
  },
  "timestamp": "2025-08-23T15:00:00Z",
  "requestId": "req_create_1000008"
}
```

❌ **已废弃的旧示例** (违反标识符统一性原则)

**重要说明**: 此部分包含已废弃的UUID标识符用法，与当前标识符设计规范不符。

**问题分析**:
- 使用 `parent_unit_id` 而非标准的 `parentCode`
- 使用 `id` 而非标准的 `code`  
- 使用 `unit_type` 而非标准的 `unitType`
- 使用内部UUID而非对外编码

**正确的示例**: 请参考上述"4. 创建组织单元 (REST命令)"部分，该示例符合最新的标识符设计规范。
```

❌ **已废弃的旧示例** (违反标识符统一性原则)

**重要说明**: 此获取单个组织单元的示例使用了已废弃的UUID标识符。

**问题**: 
- 路径参数使用 `{id}` 而非标准的 `{code}`
- 响应使用UUID字段而非编码字段

**正确的方式**: 请参考上述GraphQL查询示例:
```graphql
query { 
  organization(code: "1000001") { 
    code parentCode name unitType status 
  } 
}
```
```

### 4. 完全替换组织单元 (PUT)

**`PUT /api/v1/organization-units/{code}`**

完全替换组织单元，必须提供完整的资源表示。未提供的字段将被重置为默认值。

#### 重要说明
- **完全替换语义**: 必须提供所有必需字段和可选字段
- **未提供字段**: 将被重置为默认值或null
- **幂等性**: 多次调用结果一致
- **适用场景**: 完整的资源重建或大量字段更新

#### 请求体 (完整资源表示)
```json
{
  "name": "技术研发部",
  "unitType": "DEPARTMENT",
  "parentCode": "1000000",
  "description": "负责产品研发、技术创新和系统架构设计",
  "status": "ACTIVE",
  "sortOrder": 0,
  "profile": {
    "budget": 6000000,
    "managerPositionCode": "POS-789e0123",
    "costCenterCode": "CC001",
    "headCountLimit": 60,
    "establishedDate": "2024-01-01"
  },
  "effectiveDate": "2025-08-23",
  "endDate": null,
  "operationReason": "组织架构全面重构"
}
```

#### 响应示例 ⭐ **已统一命名风格**
```json
{
  "code": "1000001",
  "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
  "unitType": "DEPARTMENT",
  "name": "技术研发部",
  "description": "负责产品研发、技术创新和系统架构设计",
  "parentCode": "1000000",
  "level": 2,
  "status": "ACTIVE",
  "profile": {
    "budget": 6000000,
    "managerPositionCode": "POS-1001",
    "costCenterCode": "CC001",
    "headCountLimit": 60
  },
  "createdAt": "2025-01-01T00:00:00Z",
  "updatedAt": "2025-08-04T10:45:00Z",
  "operationType": "UPDATE",
  "operationReason": "组织架构全面重构"
}
```

### 5. 部分更新组织单元 (PATCH)

**`PATCH /api/v1/organization-units/{code}`**

部分更新组织单元的常规数据字段，只更新提供的字段，其他字段保持不变。**不支持核心状态变更操作**。

#### 重要说明 (唯一性原则优化)
- **部分更新语义**: 只更新请求体中提供的字段
- **字段保持**: 未提供的字段保持现有值不变
- **职责边界**: 仅用于常规数据字段更新(name, description, profile等)
- **状态变更限制**: 不能用于status字段变更，必须使用专用端点
- **适用场景**: 单个或少量常规字段的精确更新

#### 请求体 (部分字段)
```json
{
  "name": "技术研发部",
  "description": "负责产品研发、技术创新和系统架构设计",
  "profile": {
    "budget": 6000000,
    "headCountLimit": 60
  },
  "operationReason": "预算调整和人员编制优化"
}
```

#### PATCH响应示例
```json
{
  "success": true,
  "message": "Organization unit partially updated successfully", 
  "data": {
    "code": "1000001",
    "name": "技术研发部",
    "unitType": "DEPARTMENT",
    "parentCode": "1000000",
    "description": "负责产品研发、技术创新和系统架构设计",
    "status": "ACTIVE",
    "profile": {
      "budget": 6000000,
      "managerPositionCode": "POS-1001",
      "costCenterCode": "CC001",
      "headCountLimit": 60,
      "establishedDate": "2024-01-01"
    },
    "operationType": "UPDATE",
    "operationReason": "预算调整和人员编制优化",
    "updatedAt": "2025-08-23T15:30:00Z"
  },
  "timestamp": "2025-08-23T15:30:00Z"
}
```

#### 状态变更操作边界 🚨 **重要说明**

**PATCH端点不能用于以下操作**:
```yaml
# ❌ 禁止的状态变更 (违反唯一性原则)
PATCH /api/v1/organization-units/1000001
{
  "status": "INACTIVE"  # ❌ 禁止：会导致功能重复
}

# ✅ 正确的状态变更方式 (唯一实现)
POST /api/v1/organization-units/1000001/suspend
{
  "operationReason": "业务调整需要",
  "effectiveDate": "2025-08-23"
}
```

**设计原理**:
- **唯一性原则**: 每种功能只有一种实现方式
- **意图明确**: 专用端点路径本身就表达了操作意图  
- **行为透明**: 避免"魔法"推导，减少意外行为
- **维护简化**: 业务逻辑集中在单一处理路径

### 5. 专用状态操作端点 ⭐ **唯一性原则实现**

**设计哲学**: 重要的业务操作应该有专门的、明确的API端点，避免通过通用更新端点的"智能推导"来实现，确保每种功能只有一种实现方式。

#### 5.1 停用组织单元
**`POST /api/v1/organization-units/{code}/suspend`**

专用停用端点，强制设置 `operationType=SUSPEND` 和 `status=INACTIVE`。

#### 请求体 ⭐ **已统一命名风格**
```json
{
  "operationReason": "业务调整需要",
  "effectiveDate": "2025-08-23",
  "endDate": null
}
```

#### 响应示例 ⭐ **已统一命名风格**
```json
{
  "code": "1000001",
  "status": "INACTIVE",
  "operationType": "SUSPEND",
  "operatedBy": {
    "id": "user-uuid",
    "name": "Wang Wu"
  },
  "operationReason": "业务调整需要", 
  "effectiveDate": "2025-08-23",
  "updatedAt": "2025-08-23T10:30:00Z",
  "isCurrent": true,
  "isFuture": false
}
```

#### 5.2 激活组织单元
**`POST /api/v1/organization-units/{code}/activate`**

专用激活端点，强制设置 `operationType=REACTIVATE` 和 `status=ACTIVE`。

#### 请求体 ⭐ **已统一命名风格**
```json
{
  "operationReason": "恢复业务运营",
  "effectiveDate": "2025-08-23",
  "endDate": null
}
```

#### 5.3 未来日期操作支持 ⭐ **重要特性**

**业务场景**: 计划在未来某个日期执行状态变更操作

**示例: 计划停用操作**
```bash
# 今天: 2025-08-23，计划 2025-09-01 停用组织
POST /api/v1/organization-units/1000001/suspend
{
  "operationReason": "业务重组，部门合并",
  "effectiveDate": "2025-09-01"
}

# 系统创建未来生效记录:
# - operationType = SUSPEND, status = INACTIVE
# - effectiveDate = 2025-09-01 
# - isCurrent = false, isFuture = true (相对今天)
```

**时态查询示例** ⭐ **已统一命名风格**
```bash
# 查看当前状态 (今天仍为活跃) - GraphQL查询
query { organization(code: "1000001", asOfDate: "2025-08-23") { status isCurrent } }
# 返回: status=ACTIVE, isCurrent=true

# 查看计划操作 - GraphQL查询
query { organization(code: "1000001", includeFuture: true) { operationType effectiveDate } }
# 返回: 包含未来SUSPEND记录

# 查看未来状态 (2025-09-01已停用) - GraphQL查询
query { organization(code: "1000001", asOfDate: "2025-09-01") { status isCurrent } }
# 返回: status=INACTIVE, isCurrent=true
```

**取消计划操作** ⭐ **已统一命名风格**
```bash
# 创建同日期的REACTIVATE记录覆盖SUSPEND计划
POST /api/v1/organization-units/1000001/activate
{
  "operationReason": "计划变更，继续运营",
  "effectiveDate": "2025-09-01"
}
```

#### 5.4 幂等性处理
- **停用已停用的组织**: 返回 200 OK，状态无变化
- **激活已激活的组织**: 返回 200 OK，状态无变化
- **操作已删除的组织**: 返回 409 Conflict 错误

### 6. PUT vs PATCH 使用指南 ⭐ **新增**

#### 选择标准
```yaml
# 使用PUT的情况
完全重建: 需要重新定义整个组织单元
大量字段变更: 超过50%的字段需要更新
标准化操作: 按模板批量标准化组织配置
幂等要求: 需要严格的幂等性保证

# 使用PATCH的情况  
少量字段更新: 1-3个字段的精确更新
增量修改: 在现有基础上进行微调
用户界面操作: 表单中的单个字段编辑
性能优化: 减少网络传输和处理开销
```

#### 语义对比示例
```yaml
# PUT - 完全替换
PUT /api/v1/organization-units/1000001
{
  "name": "新技术部",
  "unitType": "DEPARTMENT",
  "parentCode": "1000000",
  "description": "技术研发部门",
  "status": "ACTIVE",
  "sortOrder": 0,
  "profile": {},  # 重置为空配置
  "effectiveDate": "2025-08-23"
}
# 结果: profile被重置，所有字段都是新值

# PATCH - 部分更新
PATCH /api/v1/organization-units/1000001  
{
  "name": "新技术部",
  "profile": {
    "budget": 7000000  # 只更新预算
  }
}
# 结果: 只更新name和预算，其他字段保持不变
```

#### 错误处理差异
```yaml
PUT错误:
  - INCOMPLETE_RESOURCE: 缺少必需字段
  - INVALID_RESOURCE: 完整资源验证失败
  
PATCH错误:
  - INVALID_FIELD_UPDATE: 单个字段更新失败
  - READONLY_FIELD: 尝试更新只读字段
```

### 7. 删除组织单元

**`DELETE /api/v1/organization-units/{code}`**

删除组织单元，会检查关联约束。

#### 删除约束
- 不能删除有子单元的组织单元
- 不能删除有关联职位的组织单元
- 删除前需要清理所有依赖关系

#### 响应
- **204 No Content**: 删除成功
- **404 Not Found**: 组织单元不存在
- **409 Conflict**: 存在子单元或关联职位，无法删除

### 7. 操作审计接口 ⭐ **新增 v7**

#### 7.1 获取组织完整审计历史 (GraphQL查询)
使用 `organizationAuditHistory` GraphQL查询获取组织的完整操作审计历史，支持跨版本的变更追踪。

#### 查询参数 ⭐ **已统一camelCase命名**
```yaml
startDate: 开始日期 (YYYY-MM-DD, 可选)
endDate: 结束日期 (YYYY-MM-DD, 可选)  
operation: 操作类型过滤 (create|update|suspend|activate|delete, 可选)
userId: 操作人过滤 (UUID, 可选)
limit: 返回记录数 (默认50, 最大200)
```

#### 响应示例
```json
{
  "data": {
    "businessEntityId": "1000001",
    "entityName": "技术部", 
    "totalVersions": 4,
    "auditTimeline": [
      {
        "auditId": "aud_001",
        "versionSequence": 1,
        "operation": "create",
        "timestamp": "2025-01-01T10:00:00Z",
        "userName": "张三",
        "operationReason": "New technical department created",
        "changesSummary": {
          "operationSummary": "CREATE",
          "totalChanges": 8,
          "keyChanges": ["created new organization"]
        },
        "riskLevel": "low"
      },
      {
        "auditId": "aud_002",
        "versionSequence": 2, 
        "operation": "update",
        "timestamp": "2025-06-01T14:30:00Z",
        "userName": "李四",
        "operationReason": "Department name standardization",
        "changesSummary": {
          "operationSummary": "UPDATE", 
          "totalChanges": 2,
          "keyChanges": [
            "name: '技术部' → '技术研发部'",
            "description: updated"
          ]
        },
        "riskLevel": "low"
      }
    ]
  },
  "meta": {
    "totalAuditRecords": 2,
    "dateRange": {
      "earliest": "2025-01-01T10:00:00Z",
      "latest": "2025-06-01T14:30:00Z"
    },
    "operationsSummary": {
      "create": 1,
      "update": 1
    }
  }
}
```

#### 7.2 获取详细审计记录 (GraphQL查询)
使用 `auditLog` GraphQL查询获取单条审计记录的完整详细信息，包括操作前后数据快照和字段级变更分析。

#### 响应示例
```json
{
  "data": {
    "auditId": "aud_002",
    "businessEntityId": "1000001",
    "recordId": "uuid-version-2", 
    "versionSequence": 2,
    "operation": "update",
    "timestamp": "2025-06-01T14:30:00Z",
    "userInfo": {
      "userId": "user-uuid-456",
      "userName": "李四",
      "role": "HR Manager"
    },
    "operationContext": {
      "operationReason": "Department name standardization",
      "ipAddress": "192.168.1.100",
      "apiEndpoint": "PUT /api/v1/organization-units/1000001"
    },
    "dataChanges": {
      "beforeData": {
        "id": "1000001",
        "name": "技术部",
        "description": "负责技术研发工作",
        "status": "active"
      },
      "afterData": {
        "id": "1000001",
        "name": "技术研发部", 
        "description": "负责产品研发和技术创新",
        "status": "active"
      },
      "fieldChanges": {
        "totalChanges": 2,
        "modifiedFields": ["name", "description"],
        "fieldDetails": {
          "name": {
            "before": "技术部",
            "after": "技术研发部",
            "changeType": "MODIFIED",
            "impactLevel": "LOW"
          }
        }
      }
    },
    "impactAnalysis": {
      "businessImpact": "Department identifier updated", 
      "affectedUsersCount": 0,
      "riskLevel": "low"
    }
  }
}
```

#### 7.3 跨版本变更分析 (GraphQL查询)
使用 `organizationChangeAnalysis` GraphQL查询分析组织在指定版本范围内的所有变更，提供完整的变化轨迹。

#### 查询参数
```yaml
fromVersion: 起始版本序号 (integer, 可选, 默认1)
toVersion: 结束版本序号 (integer, 可选, 默认最新版本)
analysisType: 分析类型 (summary|detailed, 默认summary)
```

### 8. 层级一致性检查接口 ⭐ **GraphQL查询专用 - CQRS合规**

**🚨 重要**: 层级一致性检查现在严格遵循CQRS原则，统一使用GraphQL查询，移除了违规的REST端点。

#### 8.1 层级一致性全面检查 (GraphQL查询)
使用 `hierarchyConsistencyCheck` GraphQL查询进行层级数据一致性检测和报告。

#### GraphQL查询示例
```graphql
query HierarchyConsistencyCheck($tenantId: UUID!, $checkMode: ConsistencyCheckMode) {
  hierarchyConsistencyCheck(tenantId: $tenantId, checkMode: $checkMode) {
    checkId
    tenantId
    executedAt
    executionTimeMs
    totalChecked
    issuesFound
    checkMode
    consistencyReport {
      pathMismatches {
        code
        expectedCodePath
        actualCodePath
        expectedNamePath
        actualNamePath
        severity
      }
      levelInconsistencies {
        code
        expectedLevel
        actualLevel
        parentCode
        reason
      }
      orphanedNodes {
        code
        name
        parentCode
        reason
      }
      circularReferences {
        affectedCodes
        circularPath
        severity
      }
      depthViolations {
        code
        currentDepth
        maxAllowedDepth
        parentChain
      }
      cacheInconsistencies {
        code
        fieldName
        cachedValue
        calculatedValue
        impactLevel
      }
    }
    repairSuggestions {
      issueType
      affectedCodes
      suggestedAction
      automatable
      riskLevel
    }
    healthScore
    recommendedActions
  }
}
```

#### 查询参数说明
```yaml
tenantId: UUID! (租户ID, 必需)
checkMode: ConsistencyCheckMode (FAST|DEEP|TARGETED, 默认FAST)
targetCodes: [String!] (可选, TARGETED模式下指定的组织编码列表)
includeRepairSuggestions: Boolean (是否包含修复建议, 默认true)
maxDepth: Int (最大检查深度, 默认17)
```

#### 检查模式说明
```yaml
FAST模式:
  检查范围: 缓存字段与计算字段的一致性
  执行时间: < 5秒 (适合定期检查)
  检测内容: level与hierarchyDepth, codePath与namePath一致性
  适用场景: 日常健康检查, 监控报警

DEEP模式:
  检查范围: 递归验证、循环引用、父子关系一致性
  执行时间: 30秒 - 5分钟 (数据量依赖)
  检测内容: 所有FAST模式检查 + 递归路径验证 + 孤儿节点检测
  适用场景: 数据迁移后的全面检查, 重大故障恢复

TARGETED模式:
  检查范围: 针对指定组织及其子树的精准检查
  执行时间: < 30秒 (检查范围可控)
  检测内容: 指定节点的DEEP模式检查
  适用场景: 针对性问题排查, 特定区域数据修复
```

#### GraphQL响应示例
```json
{
  "data": {
    "hierarchyConsistencyCheck": {
      "checkId": "check_20250823_143052",
      "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
      "executedAt": "2025-08-23T14:30:52.123Z",
      "executionTimeMs": 2340,
      "totalChecked": 127,
      "issuesFound": 3,
      "checkMode": "DEEP",
      "consistencyReport": {
        "pathMismatches": [
          {
            "code": "1000015",
            "expectedCodePath": "/1000000/1000001/1000015",
            "actualCodePath": "/1000000/1000015",
            "expectedNamePath": "/高谷集团/技术部/前端组",
            "actualNamePath": "/高谷集团/前端组",
            "severity": "HIGH"
          }
        ],
        "levelInconsistencies": [
          {
            "code": "1000015",
            "expectedLevel": 3,
            "actualLevel": 2,
            "parentCode": "1000001",
            "reason": "Parent hierarchy depth calculation error"
          }
        ],
        "orphanedNodes": [],
        "circularReferences": [],
        "depthViolations": [],
        "cacheInconsistencies": [
          {
            "code": "1000015",
            "fieldName": "hierarchyDepth",
            "cachedValue": 2,
            "calculatedValue": 3,
            "impactLevel": "MEDIUM"
          }
        ]
      },
      "repairSuggestions": [
        {
          "issueType": "PATH_MISMATCH",
          "affectedCodes": ["1000015"],
          "suggestedAction": "Execute refresh-hierarchy for organization 1000015 to recalculate paths",
          "automatable": true,
          "riskLevel": "LOW"
        }
      ],
      "healthScore": 97.6,
      "recommendedActions": [
        "Execute targeted refresh for organization 1000015",
        "Monitor parent-child relationship changes for code 1000001",
        "Consider running FAST mode consistency check daily"
      ]
    }
  }
}
```

#### 8.2 层级一致性问题类型说明
```yaml
PATH_MISMATCH: codePath或namePath与实际父子关系不一致
LEVEL_INCONSISTENCY: level字段与实际层级不一致
ORPHANED_NODE: parentCode存在但父节点不存在或已删除
CIRCULAR_REFERENCE: 存在循环引用，如A→B→C→A
DEPTH_VIOLATION: 层级深度超过系统限制(17级)
CACHE_INCONSISTENCY: 缓存字段与计算结果不一致
```

#### 8.3 检查结果处理流程 🚨 **关键流程**
```yaml
检查结果分析:
  healthScore >= 98: 系统健康，无需干预
  healthScore 90-97: 轻微问题，建议定期检查
  healthScore 80-89: 中等问题，需要及时修复
  healthScore < 80: 严重问题，立即修复并排查原因

自动修复流程:
  1. 执行GraphQL一致性检查 (获取repairSuggestions)
  2. 过滤automatable=true且riskLevel=LOW的建议
  3. 逐一执行refresh-hierarchy REST操作
  4. 再次执行GraphQL检查确认修复结果
  5. 记录修复操作的完整日志

手动干预流程:
  1. 分析consistencyReport中的具体问题
  2. 根据suggestedAction指导执行修复
  3. 对于riskLevel=HIGH的问题做充分评估
  4. 必要时联系业务方确认数据修复范围
```

### 9. CoreHR兼容接口 ⭐ **严格CQRS合规重构**

#### 创建组织 (REST命令)
**`POST /api/v1/corehr/organizations`**

提供与前端CoreHR模块兼容的组织创建接口，映射到OrganizationUnit实体。

#### 获取组织列表和统计 (GraphQL查询) 🚨 **协议合规**
**CoreHR兼容性查询现在统一使用GraphQL**：

```graphql
# 替代原 GET /api/v1/corehr/organizations
query GetCoreHROrganizations($filter: OrganizationFilter) {
  organizations(filter: $filter) {
    data {
      code
      name
      unitType
      status
      parentCode
      level
      description
    }
    pagination { total page hasNext }
  }
}

# 替代原 GET /api/v1/corehr/organizations/stats  
query GetCoreHRStats {
  organizationStats {
    total
    byType { type count }
    byStatus { status count }
    byLevel { level count }
    temporal { current future historical }
  }
}
```

**兼容性响应格式示例**:
```json
{
  "data": {
    "organizationStats": {
      "total": 25,
      "byType": [
        {"type": "DEPARTMENT", "count": 15},
        {"type": "COST_CENTER", "count": 5},
        {"type": "COMPANY", "count": 2},
        {"type": "PROJECT_TEAM", "count": 3}
      ],
      "byStatus": [
        {"status": "ACTIVE", "count": 23},
        {"status": "INACTIVE", "count": 1}
      ],
      "hierarchyDepth": 4,
      "unitsWithoutParent": 2
    }
  }
}
```

**🔧 迁移指导**: 
- 原REST查询端点已移除以符合CQRS架构
- CoreHR前端应使用GraphQL客户端访问查询数据
- 命令操作继续使用REST API

## 🔄 REST标准合规声明 ⭐ **重要**

本 API 严格遵循pRFC 7231 HTTP标准中PUT和PATCH方法的语义定义：

### HTTP方法语义保证
```yaml
PUT方法:
  语义: 完全替换 (Complete Replacement)
  行为: 提供完整资源表示，未提供的字段将被重置
  幂等性: 多次调用相同请求产生相同结果
  安全性: 幂等操作，对服务器状态无副作用
  适用: 完整资源更新、数据标准化、系统初始化

PATCH方法:
  语义: 部分更新 (Partial Modification)
  行为: 只修改提供的字段，其他字段保持不变
  非幂等性: 多次调用可能产生不同结果
  适用: 单个字段更新、增量修改、用户界面操作
```

### 数据一致性保证
```yaml
PUT操作后:
  - 资源状态 = 请求体定义的完整状态
  - 未提供的可选字段 = 默认值或 null
  - 必需字段缺失 = 400 INCOMPLETE_RESOURCE 错误

PATCH操作后:
  - 请求中的字段 = 新值
  - 未提供的字段 = 保持原值不变
  - 只读字段更新 = 400 READONLY_FIELD 错误
```

### 兼容性声明
ℹ️ **重要**: 这是一个破坏性变更。如果您的客户端代码使用PUT方法进行部分更新，请更改为PATCH方法。

```yaml
# 迁移指南
旧的使用方式 (不符合标准):
  PUT /api/v1/organization-units/1000001
  { "name": "新名称" }  # 仅提供部分字段

新的正确方式:
  PATCH /api/v1/organization-units/1000001  
  { "name": "新名称" }  # 部分更新使用PATCH
  
  或者使用PUT进行完整替换:
  PUT /api/v1/organization-units/1000001
  {  # 必须提供完整资源
    "name": "新名称",
    "unitType": "DEPARTMENT",
    "parentCode": "1000000",
    "status": "ACTIVE",
    "profile": {},
    "effectiveDate": "2025-08-23"
  }
```

### 字段优化迁移指南 ⭐ **重要变更**

#### path字段移除 (v3.1)
**影响**: 移除冗余的 `path` 字段，统一使用 `codePath`

```yaml
# 旧版本API响应 (已废弃)
{
  "codePath": "/1000000/1000001",
  "namePath": "/高谷集团/技术部",
  "path": "/1000000/1000001"  # 与codePath重复
}

# 新版本API响应 (推荐)
{
  "codePath": "/1000000/1000001",  # 统一的编码路径
  "namePath": "/高谷集团/技术部"    # 人类可读的名称路径
}

# 客户端迁移建议
旧代码: const path = organization.path
新代码: const path = organization.codePath

# 兼容性处理 (服务端可选实现)
服务端可在响应中动态添加path字段作为codePath的别名:
{
  "codePath": "/1000000/1000001",
  "namePath": "/高谷集团/技术部",
  "path": "/1000000/1000001"  // 临时兼容，建议客户端迁移到codePath
}

# 优化收益
存储优化: 每条记录减少约100-2000字符的冗余存储
维护简化: 消除字段同步复杂性
认知简化: API消费者只需理解codePath一个概念

## 🏢 单元类型配置

### DEPARTMENT（部门）配置
```json
{
  "budget": "number (年度预算)",
  "managerPositionCode": "string (部门经理职位编码)",
  "costCenterCode": "string (成本中心代码)",
  "headCountLimit": "number (人员上限)",
  "establishedDate": "date (成立日期)"
}
```

### COST_CENTER（成本中心）配置
```json
{
  "costCenterCode": "string (成本中心代码)",
  "budgetPeriod": "string (预算周期)",
  "budgetAmount": "number (预算金额)",
  "responsibleManager": "string (责任经理)",
  "profitCenter": "string (利润中心)"
}
```

### COMPANY（公司）配置
```json
{
  "legalName": "string (法人名称)",
  "registrationNumber": "string (注册号)",
  "taxId": "string (税务登记号)",
  "registeredAddress": "string (注册地址)",
  "businessScope": "string (经营范围)"
}
```

### PROJECT_TEAM（项目团队）配置
```json
{
  "projectDuration": "string (项目周期)",
  "teamLead": "string (团队负责人)",
  "budgetAllocated": "number (分配预算)",
  "projectType": "string (项目类型)",
  "deliverables": "array (交付物清单)"
}
```

## 🔐 安全与认证

### 认证架构概览 ⭐ **企业级OAuth 2.0标准**

**认证流程**: OAuth 2.0 Client Credentials Flow (机器对机器访问)  
**令牌标准**: JWT (JSON Web Token)  
**传输方式**: HTTPS + Bearer Token  
**权限模型**: 基于权限的访问控制 (PBAC - Permission-Based Access Control)

### 🔑 OAuth 2.0 Client Credentials Flow

#### 第一步：客户端注册和凭证获取
```yaml
注册方式: 在管理面板创建API客户端
凭证组成: 
  - Client ID: 公开标识符 (如: a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6)
  - Client Secret: 私密凭证 (如: super-secret-string-that-is-very-long-and-secure)
权限分配: 根据业务需求分配具体权限范围
安全要求: Client Secret视为密码，妥善加密存储
```

#### 第二步：获取Access Token
**令牌端点**: `POST /oauth/token`

**请求示例**:
```bash
curl -X POST https://api.yourcompany.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6" \
  -d "client_secret=super-secret-string-that-is-very-long-and-secure"
```

**成功响应**:
```json
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "tokenType": "Bearer",
  "expiresIn": 3600,
  "scope": "org:read org:create org:update"
}
```

#### 第三步：使用Access Token访问API
**认证头部**: `Authorization: Bearer <access_token>`

**API调用示例**:
```bash
curl -X GET http://localhost:8090/graphql \
  -H "Authorization: Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{"query": "query { organizations { data { code name status } } }"}'
```

### 🎫 JWT结构标准 ⭐ **权限载荷规范**

#### JWT Header
```json
{
  "alg": "RS256",
  "typ": "JWT"
}
```

#### JWT Payload (权限载荷)
```json
{
  "iss": "https://api.yourcompany.com",
  "sub": "a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6",
  "aud": "organization-management-api",
  "exp": 1724408400,
  "iat": 1724404800,
  "permissions": [
    "org:read",
    "org:create", 
    "org:update",
    "org:suspend",
    "org:read:hierarchy",
    "org:read:audit"
  ],
  "tenantId": "987fcdeb-51a2-43d7-8f9e-123456789012",
  "clientName": "HR Integration System"
}
```

#### JWT字段说明
```yaml
标准声明 (RFC 7519):
  iss (Issuer): 令牌颁发者
  sub (Subject): 客户端ID  
  aud (Audience): 目标API服务
  exp (Expiration): 过期时间戳
  iat (Issued At): 颁发时间戳

业务声明:
  permissions: 权限数组 - 访问控制的唯一依据
  tenant_id: 租户ID - 多租户隔离
  client_name: 客户端名称 - 审计追踪用
```

### 🔒 基于权限的访问控制模型 (PBAC)

#### 权限命名规范
**模式**: `资源:操作[:子资源]`

#### 核心权限列表
```yaml
# 基础数据权限
org:read: 读取组织单元信息
org:create: 创建组织单元
org:update: 更新组织单元基本信息
org:delete: 删除组织单元

# 状态操作权限  
org:suspend: 停用组织单元
org:reactivate: 重新激活组织单元

# 层级管理权限
org:read:hierarchy: 读取组织层级结构
org:recalculate: 手动触发层级重计算

# 审计查询权限
org:read:audit: 读取审计历史记录
org:stats: 获取组织统计信息

# 数据验证权限
org:validate: 数据有效性验证

# 运维工具权限 (高级权限)
org:maintenance: 层级一致性检查和修复
org:batch-operations: 批量操作权限
```

#### 权限分组示例
```yaml
# 只读用户
readonly_user:
  - org:read
  - org:read:hierarchy
  - org:stats

# 标准编辑用户  
editor_user:
  - org:read
  - org:create
  - org:update
  - org:read:hierarchy
  - org:validate

# 管理员用户
admin_user:
  - org:read
  - org:create
  - org:update
  - org:delete
  - org:suspend
  - org:reactivate
  - org:read:hierarchy
  - org:read:audit
  - org:stats
  - org:validate

# 系统运维人员
maintenance_user:
  - org:maintenance
  - org:batch-operations
  - org:read:audit
  - org:recalculate
```

### 🛡️ API端点权限要求

#### GraphQL查询端点权限映射
```yaml
# 基础查询
organizations(filter: OrgFilter): org:read
organization(code: String!): org:read  
organizationStats: org:stats

# 层级查询
organizationHierarchy(code: String!): org:read:hierarchy
organizationSubtree(code: String!): org:read:hierarchy
hierarchyStatistics: org:read:hierarchy

# 审计查询
organizationAuditHistory(code: String!): org:read:audit
auditLog(auditId: String!): org:read:audit
organizationChangeAnalysis(code: String!): org:read:audit

# 一致性检查
hierarchyConsistencyCheck: org:maintenance
```

#### REST命令端点权限映射
```yaml
# 标准CRUD操作
POST /api/v1/organization-units: org:create
PUT /api/v1/organization-units/{code}: org:update
PATCH /api/v1/organization-units/{code}: org:update
DELETE /api/v1/organization-units/{code}: org:delete

# 专用业务操作
POST /api/v1/organization-units/{code}/suspend: org:suspend
POST /api/v1/organization-units/{code}/activate: org:reactivate
POST /api/v1/organization-units/validate: org:validate

# 运维工具操作 (需要特殊权限) 
POST /api/v1/organization-units/{code}/refresh-hierarchy: org:maintenance  
POST /api/v1/organization-units/batch-refresh-hierarchy: org:batch-operations
```

### 🚨 认证和授权错误处理

#### 认证错误响应
```json
// 401 Unauthorized - 缺少或无效Token
{
  "success": false,
  "error": {
    "code": "INVALID_TOKEN",
    "message": "Invalid or expired access token",
    "details": {
      "errorDescription": "The access token provided is invalid, expired, or malformed",
      "errorUri": "https://docs.api.yourcompany.com/errors/invalid-token"
    }
  },
  "timestamp": "2025-08-23T10:30:00Z",
  "requestId": "req_auth_error_001"
}

// 401 Unauthorized - Token过期
{
  "success": false, 
  "error": {
    "code": "TOKEN_EXPIRED",
    "message": "Access token has expired",
    "details": {
      "expiredAt": "2025-08-23T09:30:00Z",
      "errorDescription": "Please obtain a new access token using your client credentials"
    }
  },
  "timestamp": "2025-08-23T10:30:00Z",
  "requestId": "req_auth_error_002"
}
```

#### 授权错误响应
```json
// 403 Forbidden - 权限不足
{
  "success": false,
  "error": {
    "code": "INSUFFICIENT_PERMISSIONS", 
    "message": "Insufficient permissions to access this resource",
    "details": {
      "requiredPermissions": ["org:delete"],
      "currentPermissions": ["org:read", "org:update"],
      "resource": "/api/v1/organization-units/1000001",
      "action": "DELETE"
    }
  },
  "timestamp": "2025-08-23T10:30:00Z", 
  "requestId": "req_auth_error_003"
}

// 403 Forbidden - 租户权限错误
{
  "success": false,
  "error": {
    "code": "TENANT_ACCESS_DENIED",
    "message": "Access denied to specified tenant resources", 
    "details": {
      "requestedTenant": "tenant-uuid-456",
      "authorizedTenants": ["tenant-uuid-123"],
      "errorDescription": "Your credentials do not have access to the requested tenant"
    }
  },
  "timestamp": "2025-08-23T10:30:00Z",
  "requestId": "req_auth_error_004" 
}
```

### 🔧 快速上手指南

#### 开发者快速开始 (5分钟设置)

**Step 1: 获取API凭证**
1. 访问开发者控制台：`https://developer.yourcompany.com`
2. 创建新的API客户端应用
3. 记录 `Client ID` 和 `Client Secret`

**Step 2: 获取访问令牌**
```bash
# 保存凭证到环境变量  
export CLIENT_ID="your-client-id"
export CLIENT_SECRET="your-client-secret"

# 获取访问令牌
TOKEN_RESPONSE=$(curl -s -X POST https://api.yourcompany.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET")

# 提取访问令牌
ACCESS_TOKEN=$(echo $TOKEN_RESPONSE | jq -r '.access_token')
```

**Step 3: 测试API调用**
```bash
# 查询组织列表 (GraphQL)
curl -X POST http://localhost:8090/graphql \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"query": "query { organizations { data { code name status } pagination { total } } }"}'

# 创建组织 (REST API)
curl -X POST http://localhost:9090/api/v1/organization-units \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试部门",
    "unitType": "DEPARTMENT", 
    "parentCode": "1000000",
    "effectiveDate": "2025-08-23",
    "operationReason": "API测试创建"
  }'
```

#### 常见问题排查

**问题1**: Token获取失败
```bash
# 检查凭证是否正确
curl -v -X POST https://api.yourcompany.com/oauth/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET"
```

**问题2**: 权限不足错误
```bash
# 检查Token中的权限
echo $ACCESS_TOKEN | cut -d'.' -f2 | base64 -d | jq '.permissions'
```

**问题3**: API调用失败
```bash  
# 验证Token有效性
curl -H "Authorization: Bearer $ACCESS_TOKEN" \
  http://localhost:8090/health
```

### ⚙️ 权限检查中间件设计 ⭐ **技术实现指导**

#### 中间件架构设计
```yaml
权限检查流程:
  1. Token提取: 从Authorization头部提取Bearer Token
  2. Token验证: 验证JWT签名、过期时间、发行者等
  3. 权限解析: 从JWT payload中提取permissions数组
  4. 权限匹配: 检查是否包含当前端点所需的权限
  5. 租户验证: 验证tenant_id与请求资源的租户匹配
  6. 审计记录: 记录权限检查结果用于安全审计

架构组件:
  AuthenticationMiddleware: 负责Token验证和解析
  AuthorizationMiddleware: 负责权限检查和租户验证
  PermissionRegistry: 维护端点与权限的映射关系
  AuditLogger: 记录认证授权事件
```

#### 中间件实现示例 (Node.js/Express)
```javascript
// 认证中间件
const authenticateToken = async (req, res, next) => {
  try {
    // 1. 提取Token
    const authHeader = req.headers['authorization'];
    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      return res.status(401).json({
        success: false,
        error: {
          code: 'MISSING_AUTHORIZATION',
          message: 'Authorization header is missing or invalid'
        }
      });
    }

    const token = authHeader.substring(7);

    // 2. 验证JWT
    const decoded = jwt.verify(token, PUBLIC_KEY, {
      algorithms: ['RS256'],
      issuer: 'https://api.yourcompany.com',
      audience: 'organization-management-api'
    });

    // 3. 检查Token过期
    if (decoded.exp < Date.now() / 1000) {
      return res.status(401).json({
        success: false,
        error: {
          code: 'TOKEN_EXPIRED', 
          message: 'Access token has expired'
        }
      });
    }

    // 4. 附加到请求对象
    req.auth = {
      clientId: decoded.sub,
      permissions: decoded.permissions || [],
      tenantId: decoded.tenant_id,
      clientName: decoded.client_name
    };

    next();
  } catch (error) {
    return res.status(401).json({
      success: false,
      error: {
        code: 'INVALID_TOKEN',
        message: 'Invalid or malformed access token'
      }
    });
  }
};

// 权限检查中间件工厂
const requirePermission = (permission) => {
  return (req, res, next) => {
    // 1. 检查权限存在性
    if (!req.auth || !req.auth.permissions) {
      return res.status(403).json({
        success: false,
        error: {
          code: 'INSUFFICIENT_PERMISSIONS',
          message: 'No permissions available'
        }
      });
    }

    // 2. 检查具体权限
    if (!req.auth.permissions.includes(permission)) {
      return res.status(403).json({
        success: false,
        error: {
          code: 'INSUFFICIENT_PERMISSIONS',
          message: 'Insufficient permissions to access this resource',
          details: {
            required_permissions: [permission],
            current_permissions: req.auth.permissions,
            resource: req.originalUrl,
            action: req.method
          }
        }
      });
    }

    // 3. 审计记录
    auditLogger.info('Permission check passed', {
      clientId: req.auth.clientId,
      permission: permission,
      resource: req.originalUrl,
      method: req.method,
      tenantId: req.auth.tenantId
    });

    next();
  };
};

// 租户权限检查中间件
const requireTenantAccess = (req, res, next) => {
  const requestedTenant = req.params.tenantId || req.body.tenantId || req.query.tenantId;
  
  if (requestedTenant && requestedTenant !== req.auth.tenantId) {
    return res.status(403).json({
      success: false,
      error: {
        code: 'TENANT_ACCESS_DENIED',
        message: 'Access denied to specified tenant resources',
        details: {
          requested_tenant: requestedTenant,
          authorized_tenants: [req.auth.tenantId]
        }
      }
    });
  }
  
  next();
};
```

#### 路由权限配置示例
```javascript
// GraphQL端点 (查询权限)
app.use('/graphql', authenticateToken, requirePermission('org:read'));

// REST API端点 (命令权限)
app.post('/api/v1/organization-units', 
  authenticateToken, 
  requirePermission('org:create'), 
  requireTenantAccess,
  createOrganizationUnit
);

app.put('/api/v1/organization-units/:code', 
  authenticateToken, 
  requirePermission('org:update'),
  requireTenantAccess,
  updateOrganizationUnit
);

app.post('/api/v1/organization-units/:code/suspend', 
  authenticateToken, 
  requirePermission('org:suspend'),
  requireTenantAccess,
  suspendOrganizationUnit
);

// 运维工具端点 (特殊权限)
app.post('/api/v1/organization-units/:code/refresh-hierarchy', 
  authenticateToken, 
  requirePermission('org:maintenance'),
  requireTenantAccess,
  refreshHierarchy
);
```

#### 权限注册表设计
```javascript
// 权限映射注册表
const PERMISSION_REGISTRY = {
  // GraphQL查询权限
  'graphql:organizations': ['org:read'],
  'graphql:organization': ['org:read'],
  'graphql:organizationStats': ['org:stats'],
  'graphql:organizationHierarchy': ['org:read:hierarchy'],
  'graphql:organizationAuditHistory': ['org:read:audit'],
  'graphql:hierarchyConsistencyCheck': ['org:maintenance'],

  // REST命令权限
  'POST /api/v1/organization-units': ['org:create'],
  'PUT /api/v1/organization-units/:code': ['org:update'],
  'PATCH /api/v1/organization-units/:code': ['org:update'],
  'DELETE /api/v1/organization-units/:code': ['org:delete'],
  'POST /api/v1/organization-units/:code/suspend': ['org:suspend'],
  'POST /api/v1/organization-units/:code/activate': ['org:reactivate'],
  'POST /api/v1/organization-units/validate': ['org:validate'],

  // 运维工具权限 (CQRS合规 - 只有命令操作)
  'POST /api/v1/organization-units/:code/refresh-hierarchy': ['org:maintenance'],
  'POST /api/v1/organization-units/batch-refresh-hierarchy': ['org:batch-operations']
};

// 动态权限检查
const checkEndpointPermission = (method, path) => {
  const key = `${method} ${path}`;
  const requiredPermissions = PERMISSION_REGISTRY[key];
  
  if (!requiredPermissions) {
    throw new Error(`No permission mapping found for ${key}`);
  }
  
  return (req, res, next) => {
    const hasPermission = requiredPermissions.some(permission => 
      req.auth.permissions.includes(permission)
    );
    
    if (!hasPermission) {
      return res.status(403).json({
        success: false,
        error: {
          code: 'INSUFFICIENT_PERMISSIONS',
          message: 'Insufficient permissions to access this resource',
          details: {
            required_permissions: requiredPermissions,
            current_permissions: req.auth.permissions
          }
        }
      });
    }
    
    next();
  };
};
```

#### 安全审计日志
```javascript
// 审计日志配置
const auditLogger = winston.createLogger({
  level: 'info',
  format: winston.format.combine(
    winston.format.timestamp(),
    winston.format.json()
  ),
  transports: [
    new winston.transports.File({ 
      filename: 'security-audit.log',
      level: 'info'
    }),
    new winston.transports.Console({
      level: 'warn' // 只在控制台显示警告和错误
    })
  ]
});

// 审计事件记录
const auditAuthEvent = (eventType, details) => {
  auditLogger.info('Security audit event', {
    event_type: eventType,
    timestamp: new Date().toISOString(),
    client_id: details.clientId,
    tenant_id: details.tenantId,
    resource: details.resource,
    action: details.action,
    permissions_checked: details.permissionsChecked,
    result: details.result, // 'ALLOWED' | 'DENIED'
    ip_address: details.ipAddress,
    user_agent: details.userAgent
  });
};

// 集成到中间件中
const enhancedRequirePermission = (permission) => {
  return (req, res, next) => {
    const hasPermission = req.auth.permissions.includes(permission);
    
    // 记录审计日志
    auditAuthEvent('PERMISSION_CHECK', {
      clientId: req.auth.clientId,
      tenantId: req.auth.tenantId,
      resource: req.originalUrl,
      action: req.method,
      permissionsChecked: [permission],
      result: hasPermission ? 'ALLOWED' : 'DENIED',
      ipAddress: req.ip,
      userAgent: req.get('User-Agent')
    });
    
    if (!hasPermission) {
      return res.status(403).json({
        success: false,
        error: {
          code: 'INSUFFICIENT_PERMISSIONS',
          message: 'Insufficient permissions to access this resource'
        }
      });
    }
    
    next();
  };
};
```

### 🛠️ 安全最佳实践

#### 客户端安全要求
```yaml
凭证存储:
  - Client Secret必须加密存储
  - 使用环境变量或安全密钥管理系统
  - 禁止在代码中硬编码凭证
  - 定期轮换Client Secret

Token管理:
  - 监控Token过期时间，提前刷新
  - 实现Token缓存机制，避免频繁请求
  - 在应用关闭时主动清理内存中的Token
  - 记录Token使用日志以便审计

网络安全:
  - 所有API调用必须使用HTTPS
  - 验证服务器SSL证书
  - 使用最新的TLS版本 (TLS 1.3)
  - 实现请求重试和超时机制
```

#### 服务端安全配置
```yaml
Token安全:
  - 使用RS256算法签名JWT
  - 设置合理的Token过期时间 (推荐1-4小时)
  - 实现Token黑名单机制
  - 记录所有认证事件日志

权限控制:
  - 实施最小权限原则
  - 定期审计客户端权限分配
  - 监控异常权限使用模式
  - 支持权限的动态调整

监控告警:
  - 认证失败率超阈值告警
  - 异常权限访问模式检测
  - Token泄露风险监控
  - API调用频率异常告警
```

## 📈 性能指标

### 响应时间目标
```yaml
列表查询: < 200ms
单个查询: < 100ms
创建操作: < 300ms
更新操作: < 200ms
删除操作: < 100ms
```

### 查询限制
```yaml
默认限制: 50条记录
最大限制: 1000条记录
层级深度: 最大10层
```

## ❌ 错误处理

### 错误响应格式 ⭐ **统一信封结构**
```json
{
  "success": false,
  "error": {
    "code": "ORG_UNIT_NOT_FOUND",
    "message": "Organization unit not found",
    "details": null
  },
  "timestamp": "2025-08-04T10:30:00Z",
  "requestId": "req_12345678"
}
```

### 常用错误码

#### 认证相关错误码 (401 Unauthorized)
```yaml
INVALID_TOKEN: 无效或格式错误的访问令牌
TOKEN_EXPIRED: 访问令牌已过期  
TOKEN_MALFORMED: 令牌格式不正确或签名无效
MISSING_AUTHORIZATION: 缺少Authorization头部
INVALID_CLIENT_CREDENTIALS: 客户端凭证无效
TOKEN_REVOKED: 令牌已被撤销或在黑名单中
```

#### 授权相关错误码 (403 Forbidden)
```yaml
INSUFFICIENT_PERMISSIONS: 权限不足，无法访问指定资源
TENANT_ACCESS_DENIED: 租户权限拒绝，无法访问指定租户资源
OPERATION_NOT_PERMITTED: 当前权限不允许执行此操作
RESOURCE_ACCESS_DENIED: 特定资源访问被拒绝
MAINTENANCE_PERMISSION_REQUIRED: 需要运维权限才能执行此操作
```

#### 业务逻辑错误码 (400 Bad Request)
```yaml
INVALID_REQUEST: 请求格式错误
VALIDATION_ERROR: 数据验证失败
ORG_UNIT_NOT_FOUND: 组织单元不存在
INVALID_UNIT_TYPE: 无效的单元类型
PARENT_UNIT_NOT_FOUND: 父单元不存在
CIRCULAR_REFERENCE: 循环引用错误
HAS_CHILD_UNITS: 存在子单元，无法删除
HAS_ASSOCIATED_POSITIONS: 存在关联职位，无法删除
INCOMPLETE_RESOURCE: PUT请求缺少必需字段
INVALID_RESOURCE: 完整资源验证失败
INVALID_FIELD_UPDATE: PATCH单个字段更新失败
READONLY_FIELD: 尝试更新只读字段
READONLY_OPERATION_TYPE: 尝试指定或修改operationType字段
```

#### 系统错误码 (500 Internal Server Error)
```yaml
INTERNAL_ERROR: 服务器内部错误
DATABASE_ERROR: 数据库连接或查询错误
EXTERNAL_SERVICE_ERROR: 外部服务调用失败
CONFIGURATION_ERROR: 系统配置错误
```

#### OAuth 2.0 专用错误码
```yaml
invalid_request: OAuth请求参数无效或缺失
invalid_client: 客户端认证失败
invalid_grant: 提供的授权信息无效
unauthorized_client: 客户端未被授权使用该授权类型
unsupported_grant_type: 不支持的授权类型
invalid_scope: 请求的权限范围无效或超出允许范围
server_error: OAuth服务器内部错误
temporarily_unavailable: OAuth服务器暂时无法处理请求
```

## 📊 监控和可观测性 ⭐ **生产级运维保障**

### 核心监控指标 📈 **关键业务指标**
```yaml
# 业务指标 (Business Metrics)
业务操作成功率:
  - metric: organization_unit_operations_success_rate
  - labels: [operation_type, unit_type, tenant_id]
  - target: >99.9%
  - 说明: 创建、更新、删除、状态变更的成功率

智能级联更新成功率:
  - metric: hierarchy_cascade_update_success_rate  
  - labels: [cascade_type, affected_units_count]
  - target: >99.95%
  - 说明: 父组织编码/名称变更时的级联更新成功率

层级一致性检查结果:
  - metric: hierarchy_consistency_check_status
  - labels: [check_mode, tenant_id]
  - values: [consistent, inconsistent, error]
  - 说明: 定期一致性检查的结果状态

# 性能指标 (Performance Metrics)  
API响应时间:
  - metric: api_response_duration_seconds
  - labels: [method, endpoint, status_code]
  - p99_target: <500ms (查询), <2s (命令)
  
级联更新性能:
  - metric: cascade_update_duration_seconds
  - labels: [cascade_depth, affected_units_count]
  - p99_target: <5s (depth<5), <15s (depth≥5)

数据库查询性能:
  - metric: database_query_duration_seconds
  - labels: [query_type, table_name]
  - p99_target: <100ms
```

### 关键告警规则 🚨 **自动化监控**
```yaml
# 业务异常告警
层级一致性异常:
  severity: critical
  condition: hierarchy_consistency_check_status{values="inconsistent"} > 0
  duration: 5m
  description: "检测到层级数据不一致，需要立即排查"
  runbook: "检查codePath/namePath/level字段，执行手动刷新修复"

级联更新失败:
  severity: high  
  condition: hierarchy_cascade_update_success_rate < 0.99
  duration: 5m
  description: "智能级联更新成功率下降，可能影响数据一致性"
  runbook: "检查级联更新日志，确认业务操作是否正确回滚"

API成功率下降:
  severity: high
  condition: organization_unit_operations_success_rate < 0.999  
  duration: 2m
  description: "组织单元API操作成功率异常下降"
  runbook: "检查错误日志、数据库连接状态、服务健康状态"

# 性能异常告警  
API响应延迟:
  severity: warning
  condition: histogram_quantile(0.99, api_response_duration_seconds) > 0.5  # 查询
  duration: 5m  
  description: "API响应时间P99超过阈值"
  runbook: "检查数据库性能、缓存命中率、查询计划"

级联更新超时:
  severity: warning
  condition: histogram_quantile(0.99, cascade_update_duration_seconds) > 15
  duration: 2m
  description: "级联更新性能异常，可能影响用户体验"
  runbook: "检查递归查询性能、批量更新效率、数据库锁状态"
```

### 审计和日志规范 📝 **合规性保障**
```yaml
# 操作审计日志
业务操作日志:
  level: INFO
  fields: [timestamp, requestId, tenantId, userId, operation, resourceId, changes]
  example: |
    2025-08-23T10:30:00Z [INFO] requestId=req_12345678 tenantId=T001 
    userId=U123456 operation=UPDATE resourceId=HR-001 
    changes={"code":"HR-001-NEW","name":"Human Resources Updated"}

级联更新日志:
  level: INFO  
  fields: [timestamp, requestId, triggerOperation, cascadeType, affectedUnits, duration]
  example: |
    2025-08-23T10:30:05Z [INFO] requestId=req_12345678 
    triggerOperation=UPDATE_UNIT cascadeType=CODE_CHANGE 
    affectedUnits=15 duration=1.2s status=SUCCESS

手动刷新日志:
  level: WARN  # 运维操作需要特别关注
  fields: [timestamp, requestId, operatorId, refreshScope, dryRun, affectedUnits]
  example: |
    2025-08-23T15:45:00Z [WARN] requestId=req_87654321 
    operatorId=admin_001 refreshScope=TENANT refreshMode=FULL 
    dryRun=false affectedUnits=234 reason="Data migration recovery"

# 错误和异常日志
错误日志:
  level: ERROR
  fields: [timestamp, requestId, errorCode, errorMessage, stackTrace, context]
  retention: 90天
  alerting: 错误率>1%时触发告警

性能日志:
  level: DEBUG  
  fields: [timestamp, requestId, operation, duration, dbQueries, cacheHits]
  sampling: 10% (避免日志量过大)
  retention: 7天
```

### 健康检查和诊断 🏥 **系统状态监控**
```yaml
# 健康检查端点
基础健康检查:
  endpoint: GET /health
  response: {"status": "healthy", "timestamp": "2025-08-23T10:30:00Z"}
  checks: [database, cache, external_services]
  timeout: 5s
  frequency: 30s

深度健康检查:  
  endpoint: GET /health/deep
  response: |
    {
      "status": "healthy",
      "checks": {
        "database": {"status": "healthy", "responseTime": "15ms"},
        "hierarchyConsistency": {"status": "healthy", "lastCheck": "2025-08-23T10:00:00Z"},
        "cache": {"status": "healthy", "hitRate": "95.2%"}
      }
    }
  timeout: 30s
  frequency: 5分钟

# 系统诊断工具
层级一致性快速检查:
  用途: 运维人员快速确认系统状态  
  命令: curl -X POST "http://localhost:9090/api/v1/organization-units/hierarchy/check" 
        -d '{"mode": "QUICK", "dryRun": true}'
  响应时间: <10s
  输出: 一致性状态摘要和问题数量

性能基准测试:
  用途: 定期性能回归测试
  覆盖: API响应时间、级联更新性能、数据库查询效率
  频率: 每周自动执行
  阈值: 与历史基准比较，性能下降>20%时告警
```

### 链路追踪和调试 🔍 **端到端可观测性**
```yaml
# 分布式链路追踪
请求标识:
  requestId: 自动生成UUID，在所有日志和响应中包含
  传播: 跨服务调用时自动传播
  用途: 快速定位问题请求的完整调用链

关键Span标记:
  - api.request: API请求开始
  - database.query: 数据库查询  
  - hierarchy.cascade: 级联更新处理
  - validation.business: 业务规则验证
  - audit.log: 审计日志记录

调试模式:
  header: X-Debug-Mode: true
  效果: 返回详细的内部处理信息
  限制: 仅在非生产环境可用
  安全: 不包含敏感数据

# 性能剖析
慢查询监控:
  阈值: 数据库查询>100ms时记录详细执行计划
  输出: SQL语句、执行时间、影响行数、索引使用情况
  用途: 识别需要优化的查询

级联更新分析:
  指标: 触发频率、影响范围、执行时间分布
  可视化: Grafana仪表板展示级联更新趋势
  优化: 基于数据指导索引和查询优化
```

### 告警集成和响应 📢 **运维自动化**
```yaml
# 告警通知
Slack集成:
  channel: #org-hierarchy-alerts
  format: 结构化消息，包含告警级别、描述、Runbook链接
  静默: 支持按告警类型设置静默期

PagerDuty集成:
  severity: CRITICAL告警自动创建事件
  escalation: 15分钟内无响应时升级
  runbook: 自动附加对应的处理手册链接

# 自动响应
自动修复:
  场景: 层级一致性检查发现问题时
  条件: 问题单元数量<10且影响范围可控
  操作: 自动执行dry_run检查，安全时执行修复
  通知: 修复操作完成后发送报告

预防性维护:
  定期任务: 每日凌晨执行层级一致性检查
  性能监控: 自动收集性能基线数据
  容量规划: 基于历史数据预测资源需求
```

## 🧪 API测试示例

### 使用curl测试

#### 获取组织单元列表
```bash
curl -X GET "http://localhost:8080/api/v1/organization-units?unit_type=DEPARTMENT&limit=10" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

#### 创建组织单元 (POST)
```bash
curl -X POST "http://localhost:9090/api/v1/organization-units" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术部",
    "unitType": "DEPARTMENT",
    "parentCode": "1000000",
    "description": "负责产品研发和技术创新",
    "effectiveDate": "2025-08-23",
    "operationReason": "业务扩展需要"
  }'
```

#### 完全替换组织单元 (PUT)
```bash
curl -X PUT "http://localhost:9090/api/v1/organization-units/1000001" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术研发部",
    "unitType": "DEPARTMENT",
    "parentCode": "1000000",
    "description": "负责产品研发、技术创新和系统架构设计",
    "status": "ACTIVE",
    "sortOrder": 0,
    "profile": {
      "budget": 6000000,
      "costCenterCode": "CC001",
      "establishedDate": "2024-01-01"
    },
    "effectiveDate": "2025-08-23",
    "operationReason": "组织架构全面重构"
  }'
```

#### 部分更新组织单元 (PATCH)
```bash
curl -X PATCH "http://localhost:9090/api/v1/organization-units/1000001" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "技术研发部",
    "profile": {
      "budget": 6000000
    },
    "operationReason": "预算调整和人员编制优化"
  }'
```

#### 获取单个组织单元 (GraphQL)
```bash
curl -X POST "http://localhost:8090/graphql" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { organization(code: \"1000001\") { code name status unitType parentCode profile } }"
  }'
```

## 📚 最佳实践

### 1. 层级结构设计
- 合理规划组织层级，避免过深的嵌套
- 使用适当的单元类型区分不同性质的组织
- 预留足够的扩展空间

### 2. 配置管理
- 根据单元类型使用相应的profile配置
- 定期审查和更新配置信息
- 保持配置的一致性和完整性

### 3. 关联管理
- 创建组织单元前确认父单元存在
- 删除前检查所有关联关系
- 使用软删除保留历史数据

### 4. 层级更新机制使用指南 ⭐ **新增 - 双机制设计**

#### 4.1 智能级联更新 (主要机制)
```yaml
使用场景:
  日常业务操作: 组织创建、编辑、停用、激活等所有常规业务操作
  自动触发: PUT/PATCH/POST/DELETE 操作自动触发层级更新
  数据一致性: 保证父组织变更后子组织路径即时同步
  无需干预: 开发者和用户无需手动处理层级一致性

技术实现:
  触发时机: 业务操作的事务提交后立即执行
  性能优化: 异步处理 + CTE递归查询 + 批量更新
  事务一致: 与业务操作在同一事务中执行
  错误处理: 级联更新失败时业务操作自动回滚

API调用示例:
  # 正常业务操作 - 无需额外干预
  PATCH /api/v1/organization-units/1000001
  {
    "name": "新名称",  # 自动触发namePath级联更新
    "parentCode": "1000002"  # 自动触发codePath级联更新
  }
```

#### 4.2 手动层级刷新 (运维工具)
```yaml
使用场景:
  数据修复: 由于系统异常导致的层级不一致问题
  数据迁移: 历史数据从其他系统迁移后的路径重建
  数据库维护: 直接修改数据库后的一致性恢复
  性能调优: 定期重建缓存索引和路径优化
  故障恢复: 服务重启后的数据一致性检查和修复

访问控制:
  特殊权限: 需要 hr.organization.maintenance 权限
  运维人员专用: 不对普通业务用户开放
  操作记录: 记录所有手动刷新操作的时间和用户
  影响评估: 执行前需评估对系统性能的影响

API调用示例:
  # 单个组织的层级修复
  POST /api/v1/organization-units/1000001/refresh-hierarchy
  {
    "reason": "数据迁移后的路径修复",
    "force": false  # 是否强制刷新（忽略正在进行的操作）
  }
  
  # 批量刷新（谨慎使用）
  POST /api/v1/organization-units/batch-refresh-hierarchy
  {
    "tenantFilter": "tenant-uuid",
    "reason": "系统迁移后的全量路径重建",
    "dryRun": true  # 预演练模式，不实际执行
  }
```

#### 4.3 机制选择决策树 🚨 **关键指导**
```yaml
情况分析:
  日常业务操作: 使用正常API操作（POST/PUT/PATCH）→ 智能级联自动处理
  层级数据正常: 不需要任何额外操作 → 依赖智能级联即可
  数据不一致疑似: 先使用GraphQL查询验证问题 → 确认问题后使用手动刷新
  系统迁移/故障: 使用手动刷新 → 先dry_run验证，后正式执行
  数据库直接修改: 必须使用手动刷新 → 恢复API层和数据库层的一致性

禁止行为:
  ✗ 日常业务中使用手动刷新解决正常的层级更新需求
  ✗ 在没有验证问题的情况下使用手动刷新
  ✗ 跳过dry_run直接执行批量刷新操作
  ✗ 在高并发场景下执行手动刷新操作
```

### 5. 性能优化
- 使用适当的查询过滤条件
- 避免一次性查询大量数据
- 考虑缓存频繁访问的组织信息

### 5. API设计唯一性原则 ⭐ **新增**
```yaml
# 功能唯一性保证
每种业务功能只能有一种API实现方式:
  - 查询直接子节点: 只使用organizations(filter: {parentCode})
  - 查询多级子树: 只使用organizationSubtree(maxDepth ≥ 2)
  - 数据校验: validate端点与操作端点必须共享核心校验服务
  - 协议选择: 查询必须用GraphQL，命令必须用REST

# 开发规范
代码实现层面必须确保:
  - 校验逻辑集中在共享的ValidationService中
  - 不同查询方式有明确的性能和使用场景区分  
  - 严格遵循CQRS协议分离，不得混用
  - API文档中明确标注推荐的使用方式

# 维护承诺
API演进过程中保证:
  - 不引入功能重叠的新端点
  - 保持现有功能的唯一实现原则
  - 及时清理过时或冗余的API路径
```

### 6. 深层次一致性规范 ⭐ **新增v3.2**

#### 响应结构一致性原则 🚨 **企业级标准**
```yaml
# 统一信封模式 (Envelope Pattern)
所有API响应必须使用相同的顶层结构:

成功响应 (2xx):
  success: true
  data: {...}          # 实际业务数据
  message: "string"    # 英文描述信息
  timestamp: "ISO8601" # 响应时间戳
  requestId: "string"  # 请求追踪ID

错误响应 (4xx/5xx):  
  success: false
  error:               # 错误详情对象
    code: "ERROR_CODE"
    message: "English error message"
    details: null|object
  timestamp: "ISO8601" 
  requestId: "string"

# 设计收益
消除响应结构不一致: 客户端可使用统一的解析逻辑
提升开发者体验: 无需记忆不同端点的响应格式差异  
增强可观测性: requestId支持端到端链路追踪
标准化错误处理: 统一的错误信息结构便于统一处理
```

#### 数据模型一致性原则 🚨 **跨端点标准化**
```yaml
# 操作人信息统一结构
operatedBy字段必须统一使用对象格式:
  ✅ 标准格式:
    "operatedBy": {
      "id": "uuid-string",
      "name": "English Name" 
    }
  
  ❌ 已禁用格式:
    "operatedBy": "uuid-string"  # 缺少用户可读信息
    "operated_by": "..."         # 命名风格不一致

# 时态数据统一结构  
所有时态相关字段必须使用统一命名:
  effectiveDate, endDate: 生效和结束日期
  isCurrent, isFuture: 动态时态状态字段
  createdAt, updatedAt: 记录生命周期时间戳
  operationType: 操作类型 (CREATE|UPDATE|SUSPEND|REACTIVATE|DELETE)

# 审计数据统一结构
所有审计相关字段必须使用统一格式:
  auditId, recordId: 审计和记录标识符  
  operationReason: 操作原因描述
  businessEntityId: 业务实体关联标识
  changesSummary: 变更摘要信息对象
```

#### 协议使用一致性原则 🚨 **CQRS架构强制执行**
```yaml
# 查询操作 - GraphQL专用
协议选择: 只能使用GraphQL
端点地址: http://localhost:8090/graphql
适用场景: 
  - 数据检索和过滤
  - 复杂关联查询
  - 统计和分析查询
  - 时态数据查询

强制规则:
  - 查询操作绝对禁止使用REST GET端点
  - GraphQL Query专用于数据读取
  - 支持灵活的字段选择和过滤

# 命令操作 - REST专用  
协议选择: 只能使用REST API
端点地址: http://localhost:9090/api/v1/organization-units
适用场景:
  - 数据创建、更新、删除
  - 状态变更操作
  - 批量数据操作
  - 业务流程执行

强制规则:
  - 命令操作绝对禁止使用GraphQL Mutation
  - REST方法语义必须准确 (POST/PUT/PATCH/DELETE)
  - 每种业务操作只能有一个API端点实现
```

#### 语言和术语一致性原则 🚨 **国际化标准**
```yaml
# API响应消息统一语言
响应消息语言: 统一使用英文
错误消息: 英文描述 + 标准错误代码
成功消息: 简洁的英文确认信息
用户显示信息: 支持客户端本地化

# 术语标准化
协议内术语一致性:
  REST API: organization-units (避免与其他资源术语冲突)
  GraphQL: organizations, organization (简洁性优先)
  
业务对象术语统一:
  组织单元: organization unit
  操作类型: operation type  
  生效日期: effective date
  层级路径: hierarchy path

# 字段命名词汇表
标准词汇必须跨所有端点保持一致:
  code/parentCode: 业务编码引用
  name/description: 基础描述信息
  status/isDeleted: 状态管理字段
  createdAt/updatedAt: 时间戳字段
  operationType/operatedBy/operationReason: 操作审计三元组
```

#### 一致性维护最佳实践 📖 **开发团队规范**
```yaml
# 开发流程一致性检查
代码审查必检项:
  - JSON字段命名风格统一 (camelCase)
  - 响应结构符合统一信封模式
  - 操作人信息使用标准对象结构
  - API消息统一使用英文
  - 协议选择符合CQRS原则

# 自动化质量保证
集成测试覆盖:
  - 响应结构格式验证
  - 字段命名风格检查  
  - 错误响应格式一致性
  - 跨端点数据模型一致性

# 文档维护标准
API文档更新规则:
  - 所有示例必须符合最新一致性标准
  - 新增端点必须遵循既定模式
  - 已废弃内容必须明确标注
  - 一致性违反必须立即修正

# 向后兼容策略
迁移支持机制:
  - 服务端临时支持新旧两种格式
  - 明确标注废弃时间表和迁移指导
  - 提供渐进式迁移工具和文档
  - 确保客户端有充足的适配时间
```

### 7. API一致性设计规范 ⭐ **已整合至深层次规范**
```yaml
# 一致性是API设计质量的关键标准，确保API行为可预测，极大降低开发者学习成本

# 1. 命名风格一致性 🚨 强制执行
JSON字段命名标准: 
  ✅ 统一使用camelCase: parentCode, unitType, isDeleted, createdAt, operationType
  ❌ 禁止snake_case: parent_unit_id, unit_type, is_deleted, created_at, operation_type

路径参数命名标准:
  ✅ 统一使用{code}: /api/v1/organization-units/{code}
  ❌ 禁止{id}: /api/v1/organization-units/{id}

查询参数命名标准:
  ✅ 统一使用camelCase: ?unitType=DEPARTMENT&asOfDate=2025-08-23
  ❌ 禁止snake_case: ?unit_type=DEPARTMENT&as_of_date=2025-08-23

# 2. 资源术语一致性 🚨 协议内一致
核心资源标识:
  ✅ REST API: organization-units (避免与其他资源冲突)
  ✅ GraphQL: organizations, organization (简洁性优先)
  📝 说明: 跨协议术语差异可接受，关键是协议内保持一致

关系字段命名:
  ✅ 统一使用: parentCode (引用父级编码)
  ❌ 禁止混用: parent_unit_id, parent_id, parentId

# 3. 标识符引用一致性 🚨 强制执行
对外标识符:
  ✅ 统一使用业务编码: code (7位数字: 1000001)
  ❌ 禁止内部UUID: id, unit_id, organization_id

路径参数:
  ✅ 标准格式: /{code}
  ❌ 已废弃格式: /{id}, /{unit_id}

# 4. REST端点结构一致性 🔧 推荐遵循
实例级动作模式:
  ✅ 推荐格式: /{resource}/{code}/{action}
  📝 当前实现: /{code}/suspend, /{code}/activate, /{code}/refresh-hierarchy

集合级动作模式:  
  ✅ 推荐格式: /{resource}/{action}
  📝 当前实现: /validate, /batch-refresh-hierarchy

# 5. 兼容性迁移策略
文档清理计划:
  - 🔴 立即清理: 所有snake_case字段示例
  - 🟡 逐步统一: GraphQL术语标准化
  - 🟢 标注废弃: 旧标识符引用方式

客户端迁移支持:
  - 服务端暂时支持两种格式（新/旧）
  - 响应中优先返回新格式字段
  - 明确标注废弃时间表

# 6. 一致性检查清单 ✅ **开发必备**
新增API端点检查:
  □ JSON字段全部使用camelCase命名
  □ 路径参数使用{code}而非{id}
  □ 查询参数使用camelCase格式
  □ 响应结构符合企业级标准格式
  □ 协议选择正确(查询用GraphQL，命令用REST)
  
代码审查检查:
  □ 无snake_case字段出现在API响应中
  □ 标识符引用统一使用code/parentCode
  □ 操作相关字段使用operationType/operatedBy/operationReason
  □ 时态字段使用effectiveDate/endDate/isCurrent/isFuture
  □ 审计字段使用recordId/tenantId/createdAt/updatedAt

文档更新检查:
  □ 所有示例代码使用统一命名风格
  □ 新增端点遵循既定模式
  □ 已废弃内容明确标注
  □ 术语在协议内保持一致
  □ 错误代码和响应格式符合标准

### 8. 专用端点命名规范 ⭐ **新增 v3.2 - 统一业务操作端点**

#### 设计原则 🚨 **强制执行**
```yaml
# 核心设计哲学
业务操作专用化原则:
  - 复杂业务操作(如suspend、activate)必须使用专用端点
  - 避免通过通用CRUD端点的"智能推导"实现复杂业务逻辑
  - 确保操作意图明确、业务逻辑集中、审计完整

# 命名一致性标准
实例级操作模式: /{resource}/{code}/{action}
  ✅ 标准格式: /api/v1/organization-units/{code}/suspend
  📝 action使用简洁动词，优先单个动词
  📝 相对操作使用对称动词 (suspend ↔ activate)

集合级操作模式: /{resource}/{action}  
  ✅ 标准格式: /api/v1/organization-units/validate
  ✅ 批量操作: /api/v1/organization-units/batch-{action}
  📝 统一使用batch-前缀表示批量操作

# 动词选择标准
简洁性优先: activate > reactivate (更简洁对称)
语义清晰: refresh-hierarchy > recalculate-hierarchy (更直观)
一致性: 同类操作使用相同的动词模式
对称性: 相反操作使用对称动词对(suspend/activate, lock/unlock)
```

#### 当前端点规范化建议 📋 **渐进式改进**
```yaml
# 符合规范的端点 ✅
/suspend: 简洁动词，符合规范
/validate: 单一动词，符合规范

# 规范化端点调整 ✅ (项目初期直接修正)
调整说明:
/activate (与suspend形成对称的简洁动词)
/refresh-hierarchy (替换recalculate-hierarchy, 更简洁直观)
/batch-refresh-hierarchy (替换batch-hierarchy-update, 统一batch-前缀)

# 实施策略
直接修正方案:
  1. 项目处于初期阶段，直接使用规范化端点名称
  2. 所有新功能严格遵循统一命名规范
  3. API文档完全基于规范化端点编写
  4. 测试用例同步更新为规范化端点
```

#### 命名规范检查清单 ✅ **开发必备**
```yaml
# 新增专用端点必检项
□ 动词选择: 使用简洁的现在时动词
□ 对称性: 相反操作使用对称动词对
□ 一致性: 符合既定的实例级/集合级模式
□ 语义性: 端点路径直接表达业务意图
□ 简洁性: 避免冗长的复合动词短语

# 业务逻辑检查
□ 专用化: 复杂业务操作不通过通用CRUD实现
□ 集中化: 相同业务逻辑只在一个端点实现
□ 审计完整: 操作类型(operationType)与端点语义匹配
□ 时态支持: 支持effectiveDate未来操作
□ 错误处理: 幂等性和边界情况处理完备

# 文档规范
□ 端点描述: 简洁明确的业务操作描述
□ 参数说明: 清晰的请求体结构和必选/可选字段
□ 响应格式: 符合统一的企业级信封结构
□ 示例代码: 提供curl和客户端SDK使用示例
□ 错误代码: 列出可能的错误情况和处理建议
```

#### 新增端点开发指导 📖 **标准流程**
```yaml
# 端点设计流程
1. 业务分析: 确认是否需要专用端点(复杂业务逻辑 = 专用端点)
2. 命名设计: 遵循实例级/集合级模式，选择简洁对称的动词
3. 参数设计: operationReason必选，effectiveDate支持未来操作
4. 响应设计: 使用统一企业级信封，包含完整的operationType审计
5. 文档编写: 包含业务描述、参数说明、示例代码、错误处理

# 代码实现规范
controller层: 参数验证、权限检查、业务服务调用
service层: 核心业务逻辑、数据完整性保证、审计日志
repository层: 数据持久化、事务管理、并发控制
响应构建: 统一的成功/失败响应结构、requestId追踪

# 测试覆盖要求
单元测试: 业务逻辑覆盖率 > 90%
集成测试: 端到端API调用测试
边界测试: 幂等性、并发、异常情况
性能测试: 响应时间 < 200ms目标
兼容测试: 确保不破坏现有API合约
```

# 7. 一致性维护最佳实践 📖 **团队规范**
开发流程:
  - PR审查时强制检查命名一致性和专用端点规范合规性
  - 使用自动化工具验证JSON schema一致性
  - 定期审查和清理已废弃的API示例
  - 新功能开发前先检查现有模式

测试覆盖:
  - 集成测试验证请求响应格式一致性
  - 跨协议测试确保数据结构对应
  - 向后兼容性测试保护现有客户端
  - 性能测试包含不同命名风格的场景

工具支持:
  - 配置IDE/Editor检查camelCase命名
  - 使用API文档生成工具保持示例一致性
  - 建立命名规范的代码模板
  - 设置CI/CD检查API规范合规性
```

---

## 📚 版本更新日志

### v4.2 - 数据模型命名规范统一版 (2025-08-23) ⭐ **命名一致性完成**

#### 🔧 命名一致性全面修正
- **camelCase字段统一**: 所有API响应示例中的字段统一使用camelCase命名
- **profile对象标准化**: 修正单元类型配置中的snake_case字段为camelCase
- **OAuth 2.0响应规范**: 统一accessToken、tokenType、expiresIn等字段命名
- **错误响应标准化**: 错误详情中的字段统一使用camelCase命名
- **冗余字段移除**: 从GraphQL查询中移除已废弃的"path"字段

#### 📋 API一致性保证
- **JSON字段标准**: 100% camelCase命名，彻底消除snake_case混用
- **响应格式统一**: 企业级信封结构保持一致
- **文档内容自洽**: 移除所有与命名规范冲突的示例
- **开发体验提升**: 消除命名风格混乱导致的开发困惑

#### 🛠️ 兼容性说明
- **向前兼容**: 服务端暂时支持新旧两种格式
- **迁移指导**: 建议客户端优先使用camelCase格式
- **渐进式升级**: 旧格式将在v5.0版本中完全废弃
- **工具支持**: 提供自动化命名转换工具和IDE配置

### v4.1 - 严格CQRS协议合规版 (2025-08-23) 🚨 **架构修正**

#### 🔧 CQRS协议违反修正
- **移除违规REST查询端点**: 删除 `GET /api/v1/organization-units/hierarchy-consistency-check`
- **CoreHR兼容性重构**: 移除 `GET /api/v1/corehr/organizations` 和 `GET /api/v1/corehr/organizations/stats`
- **层级一致性检查统一**: 现在只通过 `hierarchyConsistencyCheck` GraphQL查询提供
- **协议唯一性强制执行**: 确保每种功能只有一种协议实现

#### 📋 架构一致性恢复
- **查询操作**: 100% GraphQL专用，无任何REST GET端点
- **命令操作**: 100% REST专用，无任何GraphQL Mutation
- **文档自相矛盾消除**: 修正"绝对禁止REST查询"与实际端点的冲突
- **API消费者困惑解决**: 消除协议选择的二义性

#### 🛠️ 兼容性迁移指导
- **CoreHR查询迁移**: 从REST转换为GraphQL查询示例
- **层级检查工具更新**: 运维工具使用GraphQL查询获取检查结果
- **权限映射清理**: 移除已删除端点的权限配置
- **文档一致性**: 所有章节现在严格符合CQRS原则

#### 💡 架构设计价值
- **设计纯净性**: 恢复CQRS架构的完整性和一致性
- **维护简化**: 消除双重实现的维护负担
- **开发体验**: 清晰的协议边界，无选择困惑
- **技术债务清理**: 避免架构不一致累积的长期问题

### v4.0 - 企业级安全认证标准版 (2025-08-23) ⭐ **重大安全改进**

#### 🔐 新增安全特性
- **OAuth 2.0 Client Credentials Flow**: 完整的企业级机器对机器认证流程
- **基于权限的访问控制 (PBAC)**: 细粒度权限模型替代简单角色模型  
- **JWT标准载荷规范**: 标准化的权限载荷结构和字段定义
- **权限检查中间件**: 完整的技术实现指导和代码示例
- **企业级错误处理**: 详细的认证授权错误码和响应格式

#### 📋 文档改进
- **5分钟快速上手指南**: 开发者友好的认证设置流程
- **安全最佳实践**: 客户端和服务端安全配置指导
- **审计日志设计**: 完整的安全事件记录和监控方案
- **常见问题排查**: Token获取、权限检查、API调用的故障排除

#### 🛡️ 权限体系完善
- **17个核心权限**: 从基础数据操作到运维工具的完整权限覆盖
- **4种用户角色**: 只读、编辑、管理员、运维人员的权限分组
- **端点权限映射**: 所有GraphQL和REST端点的权限要求明确化
- **租户隔离**: 多租户环境下的权限边界控制

#### 🔧 技术实现指导  
- **中间件架构**: 认证、授权、审计的模块化设计
- **权限注册表**: 端点与权限的映射管理机制
- **动态权限检查**: 灵活的权限验证工厂模式
- **安全审计日志**: Winston集成的结构化日志记录

#### 📈 开发体验提升
- **统一错误响应**: 标准化的错误码分类和详细错误信息
- **调试友好**: Token解析、权限检查的故障排除工具
- **实例代码**: 完整的Node.js/Express中间件实现示例
- **测试支持**: 认证流程的端到端测试脚本

### v3.2 - 专用端点命名规范统一版 (2025-08-23)
- **专用业务操作端点**: suspend/activate对称设计完成
- **深层次API一致性**: 企业级信封响应结构统一
- **数据模型标准化**: 时态数据/审计数据结构规范
- **命名风格统一**: 全面camelCase字段命名标准化

### v3.1 - PostgreSQL原生架构优化版 (2025-08-22)
- **架构简化**: 移除Neo4j依赖，实现单一PostgreSQL数据源
- **性能革命**: 查询响应时间从15-58ms优化至1.5-8ms
- **技术债务清理**: 移除134条冗余数据和4个无用发布订阅配置
- **CDC管道移除**: 消除复杂的数据同步机制

---

**制定者**: 系统架构师  
**安全顾问**: 企业安全团队  
**审核者**: 技术委员会  
**生效日期**: 2025-08-04  
**安全更新日期**: 2025-08-23  
**下次审查**: 2025-11-04