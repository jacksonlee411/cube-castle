# 组织单元管理API规范

**版本**: v1.0  
**创建日期**: 2025-08-04  
**基于实际实现**: ✅ 已验证  
**状态**: 生产就绪

## 📋 概述

组织单元管理API提供完整的企业组织架构管理功能，支持层级结构、多种单元类型和灵活的配置选项，实现多租户隔离和完整的CRUD操作。

### 🏷️ 标识符设计说明 ⭐

**重要变更**: 本API采用全新的标识符命名策略，详见[ADR-006标识符命名策略](../architecture-decisions/ADR-006-identifier-naming-strategy.md)

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
```

### 核心特性
- **层级结构**: 支持父子关系的组织架构
- **多种类型**: 部门、成本中心、公司、项目团队等
- **多态配置**: 基于单元类型的动态配置
- **多租户隔离**: 严格的租户数据边界
- **关联管理**: 与职位和员工的关联关系

## 🏗️ API端点总览

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/organization-units` | 获取组织单元列表 | Bearer Token |
| POST | `/api/v1/organization-units` | 创建组织单元 | Bearer Token |
| GET | `/api/v1/organization-units/{code}` | 获取单个组织单元 | Bearer Token |
| PUT | `/api/v1/organization-units/{code}` | 更新组织单元 | Bearer Token |
| DELETE | `/api/v1/organization-units/{code}` | 删除组织单元 | Bearer Token |
| GET | `/api/v1/corehr/organizations` | CoreHR兼容接口 | Bearer Token |
| POST | `/api/v1/corehr/organizations` | CoreHR兼容接口 | Bearer Token |
| GET | `/api/v1/corehr/organizations/stats` | 获取组织统计信息 | Bearer Token |

## 📊 数据模型

### 组织单元核心模型
```json
{
  "code": "string (7位数字: 1000000-9999999)",
  "name": "string",
  "description": "string (optional)",
  "parent_code": "string (optional, 父级组织编码)",
  "unit_type": "DEPARTMENT | COST_CENTER | COMPANY | PROJECT_TEAM",
  "level": "number",
  "status": "ACTIVE | INACTIVE | PLANNED",
  "profile": {},
  "created_at": "2025-08-04T00:00:00Z",
  "updated_at": "2025-08-04T00:00:00Z"
}
```
```

### 单元类型枚举
```yaml
DEPARTMENT: 部门（常规业务部门）
COST_CENTER: 成本中心（财务管理单元）
COMPANY: 公司（法人实体）
PROJECT_TEAM: 项目团队（临时性组织）
```

### 状态枚举
```yaml
ACTIVE: 活跃状态
INACTIVE: 非活跃状态
PLANNED: 计划中（未正式启用）
```

## 🔍 API详细规范

### 1. 获取组织单元列表

**`GET /api/v1/organization-units`**

获取当前租户下的组织单元列表，支持层级过滤和分页。

#### 查询参数
```yaml
# 过滤参数
unit_type: 单元类型过滤
status: 状态过滤
parent_unit_id: 父单元ID过滤（UUID格式）

# 分页参数
limit: 每页大小，默认50，最大1000
offset: 偏移量，默认0
```

#### 响应示例
```json
{
  "data": [
    {
      "id": "123e4567-e89b-12d3-a456-426614174000",
      "tenant_id": "987fcdeb-51a2-43d7-8f9e-123456789012",
      "unit_type": "DEPARTMENT",
      "name": "技术部",
      "description": "负责产品研发和技术创新",
      "parent_unit_id": "456e7890-e89b-12d3-a456-426614174001",
      "level": 2,
      "status": "ACTIVE",
      "profile": {
        "budget": 5000000,
        "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
        "cost_center_code": "CC001"
      },
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-08-04T00:00:00Z"
    },
    {
      "id": "234e5678-e89b-12d3-a456-426614174003",
      "tenant_id": "987fcdeb-51a2-43d7-8f9e-123456789012",
      "unit_type": "PROJECT_TEAM",
      "name": "AI项目组",
      "description": "人工智能产品研发团队",
      "parent_unit_id": "123e4567-e89b-12d3-a456-426614174000",
      "level": 3,
      "status": "ACTIVE",
      "profile": {
        "project_duration": "2025-01-01 to 2025-12-31",
        "team_lead": "张三",
        "budget_allocated": 2000000
      },
      "created_at": "2025-02-01T00:00:00Z",
      "updated_at": "2025-08-04T00:00:00Z"
    }
  ],
  "limit": 50,
  "offset": 0,
  "total": 2
}
```

### 2. 创建组织单元

**`POST /api/v1/organization-units`**

创建新的组织单元，支持层级关系和类型配置。

#### 请求体
```json
{
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责产品研发和技术创新",
  "parent_unit_id": "456e7890-e89b-12d3-a456-426614174001",
  "status": "ACTIVE",
  "profile": {
    "budget": 5000000,
    "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
    "cost_center_code": "CC001"
  }
}
```

#### 字段验证规则
```yaml
unit_type: 必需，枚举值验证
name: 必需，长度1-100字符
description: 可选，最大500字符
parent_unit_id: 可选，必须存在的UUID
status: 可选，默认ACTIVE
profile: 可选，基于unit_type验证
```

#### 响应示例
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "987fcdeb-51a2-43d7-8f9e-123456789012",
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责产品研发和技术创新",
  "parent_unit_id": "456e7890-e89b-12d3-a456-426614174001",
  "level": 2,
  "status": "ACTIVE",
  "profile": {
    "budget": 5000000,
    "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
    "cost_center_code": "CC001"
  },
  "created_at": "2025-08-04T10:30:00Z",
  "updated_at": "2025-08-04T10:30:00Z"
}
```

### 3. 获取单个组织单元

**`GET /api/v1/organization-units/{id}`**

根据UUID获取组织单元详细信息。

#### 路径参数
- `id`: 组织单元UUID

#### 响应示例
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "987fcdeb-51a2-43d7-8f9e-123456789012",
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责产品研发和技术创新",
  "parent_unit_id": "456e7890-e89b-12d3-a456-426614174001",
  "level": 2,
  "status": "ACTIVE",
  "profile": {
    "budget": 5000000,
    "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
    "cost_center_code": "CC001",
    "established_date": "2024-01-01",
    "head_count_limit": 50
  },
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-08-04T10:30:00Z"
}
```

### 4. 更新组织单元

**`PUT /api/v1/organization-units/{id}`**

更新组织单元信息，支持部分字段更新。

#### 请求体
```json
{
  "name": "技术研发部",
  "description": "负责产品研发、技术创新和系统架构设计",
  "status": "ACTIVE",
  "profile": {
    "budget": 6000000,
    "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
    "cost_center_code": "CC001",
    "head_count_limit": 60
  }
}
```

#### 响应示例
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "tenant_id": "987fcdeb-51a2-43d7-8f9e-123456789012",
  "unit_type": "DEPARTMENT",
  "name": "技术研发部",
  "description": "负责产品研发、技术创新和系统架构设计",
  "parent_unit_id": "456e7890-e89b-12d3-a456-426614174001",
  "level": 2,
  "status": "ACTIVE",
  "profile": {
    "budget": 6000000,
    "manager_position_id": "789e0123-e89b-12d3-a456-426614174002",
    "cost_center_code": "CC001",
    "head_count_limit": 60
  },
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-08-04T10:45:00Z"
}
```

### 5. 删除组织单元

**`DELETE /api/v1/organization-units/{id}`**

删除组织单元，会检查关联约束。

#### 删除约束
- 不能删除有子单元的组织单元
- 不能删除有关联职位的组织单元
- 删除前需要清理所有依赖关系

#### 响应
- **204 No Content**: 删除成功
- **404 Not Found**: 组织单元不存在
- **409 Conflict**: 存在子单元或关联职位，无法删除

### 6. CoreHR兼容接口

#### 获取组织列表
**`GET /api/v1/corehr/organizations`**

提供与前端CoreHR模块兼容的接口，映射到OrganizationUnit实体。

#### 创建组织
**`POST /api/v1/corehr/organizations`**

#### 获取组织统计
**`GET /api/v1/corehr/organizations/stats`**

统计信息包括总数、按类型分布、按状态分布等。

```json
{
  "total_units": 25,
  "active_units": 23,
  "by_type": {
    "DEPARTMENT": 15,
    "COST_CENTER": 5,
    "COMPANY": 2,
    "PROJECT_TEAM": 3
  },
  "by_status": {
    "ACTIVE": 23,
    "INACTIVE": 1,
    "PLANNED": 1
  },
  "hierarchy_depth": 4,
  "units_without_parent": 2
}
```

## 🏢 单元类型配置

### DEPARTMENT（部门）配置
```json
{
  "budget": "number (年度预算)",
  "manager_position_id": "uuid (部门经理职位)",
  "cost_center_code": "string (成本中心代码)",
  "head_count_limit": "number (人员上限)",
  "established_date": "date (成立日期)"
}
```

### COST_CENTER（成本中心）配置
```json
{
  "cost_center_code": "string (成本中心代码)",
  "budget_period": "string (预算周期)",
  "budget_amount": "number (预算金额)",
  "responsible_manager": "string (责任经理)",
  "profit_center": "string (利润中心)"
}
```

### COMPANY（公司）配置
```json
{
  "legal_name": "string (法人名称)",
  "registration_number": "string (注册号)",
  "tax_id": "string (税务登记号)",
  "registered_address": "string (注册地址)",
  "business_scope": "string (经营范围)"
}
```

### PROJECT_TEAM（项目团队）配置
```json
{
  "project_duration": "string (项目周期)",
  "team_lead": "string (团队负责人)",
  "budget_allocated": "number (分配预算)",
  "project_type": "string (项目类型)",
  "deliverables": "array (交付物清单)"
}
```

## 🔐 安全与认证

### 认证方式
```yaml
类型: Bearer Token (JWT)
头部: Authorization: Bearer <token>
必需: 所有API端点都需要认证
```

### 权限控制
```yaml
读取权限: hr.organization.read
创建权限: hr.organization.create
更新权限: hr.organization.update
删除权限: hr.organization.delete
统计权限: hr.organization.stats
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

### 错误响应格式
```json
{
  "error": "ORG_UNIT_NOT_FOUND",
  "message": "Organization unit not found",
  "details": null,
  "timestamp": "2025-08-04T10:30:00Z",
  "request_id": "req_12345678"
}
```

### 常用错误码
```yaml
INVALID_REQUEST: 请求格式错误
VALIDATION_ERROR: 数据验证失败
ORG_UNIT_NOT_FOUND: 组织单元不存在
INVALID_UNIT_TYPE: 无效的单元类型
PARENT_UNIT_NOT_FOUND: 父单元不存在
CIRCULAR_REFERENCE: 循环引用错误
HAS_CHILD_UNITS: 存在子单元，无法删除
HAS_ASSOCIATED_POSITIONS: 存在关联职位，无法删除
UNAUTHORIZED: 未授权访问
FORBIDDEN: 权限不足
INTERNAL_ERROR: 服务器内部错误
```

## 🧪 API测试示例

### 使用curl测试

#### 获取组织单元列表
```bash
curl -X GET "http://localhost:8080/api/v1/organization-units?unit_type=DEPARTMENT&limit=10" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

#### 创建组织单元
```bash
curl -X POST "http://localhost:8080/api/v1/organization-units" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "unit_type": "DEPARTMENT",
    "name": "技术部",
    "description": "负责产品研发和技术创新",
    "status": "ACTIVE",
    "profile": {
      "budget": 5000000,
      "cost_center_code": "CC001"
    }
  }'
```

#### 获取单个组织单元
```bash
curl -X GET "http://localhost:8080/api/v1/organization-units/123e4567-e89b-12d3-a456-426614174000" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
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

### 4. 性能优化
- 使用适当的查询过滤条件
- 避免一次性查询大量数据
- 考虑缓存频繁访问的组织信息

---

**制定者**: 系统架构师  
**审核者**: 技术委员会  
**生效日期**: 2025-08-04  
**下次审查**: 2025-11-04