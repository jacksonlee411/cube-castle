# 组织管理 API 文档 - CQRS 重构版

## 概述

本文档描述了 Cube Castle 项目中组织管理模块在 CQRS（命令查询职责分离）架构重构后的 API 设计。系统采用双存储模式：PostgreSQL 处理写操作，Neo4j 优化查询操作。

## 架构设计

### CQRS 架构概览
```
Frontend -> API Gateway -> CQRS Handler
                              |
                  ┌-----------┴-----------┐
                  |                       |
             Command Side              Query Side
         (PostgreSQL 写操作)         (Neo4j 读查询)
                  |                       |
              Event Bus <------- CDC Pipeline
```

### 核心组件
- **Command Handler**: 处理写操作，存储到 PostgreSQL
- **Query Handler**: 处理读查询，从 Neo4j 获取数据
- **Event Bus**: 发布领域事件
- **CDC Pipeline**: 数据变更捕获，同步 PostgreSQL -> Neo4j

## API 端点设计

### 1. 传统 REST API（前端兼容层）

#### 基础路径: `/api/v1/corehr/organizations`

##### 获取组织列表
```http
GET /api/v1/corehr/organizations
```

**查询参数:**
- `parent_unit_id` (可选): 父组织ID，用于获取子组织
- `unit_type` (可选): 组织类型过滤 - `COMPANY` | `DEPARTMENT` | `PROJECT_TEAM` | `COST_CENTER`
- `status` (可选): 状态过滤 - `ACTIVE` | `INACTIVE` | `PLANNED`

**响应格式:**
```json
{
  "organizations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "tenant_id": "00000000-0000-0000-0000-000000000001",
      "unit_type": "DEPARTMENT",
      "name": "技术部",
      "description": "负责技术开发和维护",
      "parent_unit_id": "parent-uuid",
      "status": "ACTIVE",
      "profile": {
        "managerName": "张三",
        "maxCapacity": 50,
        "costCenter": "TECH001"
      },
      "created_at": "2024-01-01T00:00:00.000Z",
      "updated_at": "2024-01-01T00:00:00.000Z",
      "level": 1,
      "employee_count": 15,
      "children": []
    }
  ],
  "pagination": {
    "page": 1,
    "pageSize": 100,
    "total": 25,
    "totalPages": 1
  }
}
```

##### 创建组织
```http
POST /api/v1/corehr/organizations
```

**请求体:**
```json
{
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责技术开发和维护",
  "parent_unit_id": "parent-uuid",
  "status": "ACTIVE",
  "profile": {
    "managerName": "张三",
    "maxCapacity": 50
  }
}
```

**响应格式:**
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责技术开发和维护",
  "parent_unit_id": "parent-uuid",
  "status": "ACTIVE",
  "profile": {
    "managerName": "张三",
    "maxCapacity": 50
  },
  "created_at": "2024-01-01T00:00:00.000Z",
  "updated_at": "2024-01-01T00:00:00.000Z",
  "level": 1,
  "employee_count": 0,
  "children": []
}
```

##### 获取单个组织
```http
GET /api/v1/corehr/organizations/{id}
```

##### 更新组织
```http
PUT /api/v1/corehr/organizations/{id}
```

**请求体（部分更新）:**
```json
{
  "name": "新技术部",
  "description": "更新后的描述",
  "status": "ACTIVE",
  "profile": {
    "managerName": "李四",
    "maxCapacity": 60
  }
}
```

##### 删除组织
```http
DELETE /api/v1/corehr/organizations/{id}
```

**约束条件:**
- 不能删除有子组织的组织
- 不能删除有员工的组织

##### 获取组织统计
```http
GET /api/v1/corehr/organizations/stats
```

**响应格式:**
```json
{
  "data": {
    "total": 25,
    "active": 22,
    "inactive": 3,
    "totalEmployees": 150
  }
}
```

### 2. CQRS 命令端点

#### 基础路径: `/api/v1/commands`

##### 创建组织单元
```http
POST /api/v1/commands/create-organization-unit
```

**请求体:**
```json
{
  "tenant_id": "00000000-0000-0000-0000-000000000001",
  "unit_type": "DEPARTMENT",
  "name": "技术部",
  "description": "负责技术开发和维护",
  "parent_unit_id": "parent-uuid",
  "profile": {
    "managerName": "张三",
    "maxCapacity": 50
  }
}
```

**响应格式:**
```json
{
  "unit_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "created",
  "message": "Organization unit created successfully"
}
```

##### 更新组织单元
```http
PUT /api/v1/commands/update-organization-unit
```

##### 删除组织单元
```http
DELETE /api/v1/commands/delete-organization-unit
```

### 3. CQRS 查询端点

#### 基础路径: `/api/v1/queries`

##### 获取组织架构图
```http
GET /api/v1/queries/organization-chart
```

**查询参数:**
- `tenant_id` (必需): 租户ID
- `root_unit_id` (可选): 根组织ID，不指定则从顶级开始
- `max_depth` (可选): 最大深度，默认5层
- `include_inactive` (可选): 是否包含非活跃组织，默认false

**响应格式:**
```json
{
  "chart": [
    {
      "id": "company-uuid",
      "name": "公司总部",
      "unit_type": "COMPANY",
      "level": 0,
      "employee_count": 500,
      "children": [
        {
          "id": "dept-uuid",
          "name": "技术部",
          "unit_type": "DEPARTMENT",
          "level": 1,
          "employee_count": 150,
          "children": []
        }
      ]
    }
  ],
  "metadata": {
    "total_units": 25,
    "max_depth": 4,
    "total_employees": 500
  }
}
```

##### 获取组织单元详情
```http
GET /api/v1/queries/organization-units/{id}
```

**查询参数:**
- `tenant_id` (必需): 租户ID

##### 列出组织单元
```http
GET /api/v1/queries/organization-units
```

**查询参数:**
- `tenant_id` (必需): 租户ID
- `unit_type` (可选): 组织类型过滤
- `parent_id` (可选): 父组织ID
- `limit` (可选): 限制数量，默认50
- `offset` (可选): 偏移量，默认0

## 数据模型

### 组织单元模型
```go
type OrganizationUnit struct {
    ID           uuid.UUID              `json:"id"`
    TenantID     uuid.UUID              `json:"tenant_id"`
    UnitType     string                 `json:"unit_type"`     // COMPANY, DEPARTMENT, PROJECT_TEAM, COST_CENTER
    Name         string                 `json:"name"`
    Description  *string                `json:"description"`
    ParentUnitID *uuid.UUID             `json:"parent_unit_id"`
    Status       string                 `json:"status"`        // ACTIVE, INACTIVE, PLANNED
    Profile      map[string]interface{} `json:"profile"`
    CreatedAt    time.Time              `json:"created_at"`
    UpdatedAt    time.Time              `json:"updated_at"`
    
    // 查询端计算字段
    Level         int                   `json:"level"`
    EmployeeCount int                   `json:"employee_count"`
    Children      []OrganizationUnit    `json:"children,omitempty"`
}
```

### Profile 字段说明
```json
{
  "managerName": "负责人姓名",
  "maxCapacity": 50,
  "costCenter": "成本中心代码",
  "location": "办公地点",
  "businessType": "业务类型",
  "budget": 1000000
}
```

## 事件模型

### 组织创建事件
```json
{
  "event_type": "OrganizationCreated",
  "event_id": "event-uuid",
  "tenant_id": "tenant-uuid",
  "organization_id": "org-uuid",
  "organization_name": "技术部",
  "parent_organization_id": "parent-uuid",
  "organization_level": 1,
  "timestamp": "2024-01-01T00:00:00.000Z",
  "metadata": {
    "unit_type": "DEPARTMENT",
    "created_by": "user-uuid"
  }
}
```

### 组织更新事件
```json
{
  "event_type": "OrganizationUpdated",
  "event_id": "event-uuid",
  "tenant_id": "tenant-uuid",
  "organization_id": "org-uuid",
  "changed_fields": {
    "name": "新技术部",
    "manager_name": "李四"
  },
  "timestamp": "2024-01-01T00:00:00.000Z"
}
```

## 错误处理

### 标准错误响应格式
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": {
      "field": "unit_type",
      "issue": "must be one of: COMPANY, DEPARTMENT, PROJECT_TEAM, COST_CENTER"
    }
  }
}
```

### 常见错误码
- `VALIDATION_ERROR`: 参数验证失败
- `NOT_FOUND`: 资源不存在
- `CONFLICT`: 操作冲突（如删除有子组织的组织）
- `INTERNAL_ERROR`: 服务器内部错误
- `UNAUTHORIZED`: 未授权访问
- `FORBIDDEN`: 权限不足

## 权限控制

### 租户隔离
- 所有操作都需要 `tenant_id`
- 用户只能访问自己租户的组织数据

### 操作权限
- 创建组织：需要 `ORG_CREATE` 权限
- 更新组织：需要 `ORG_UPDATE` 权限
- 删除组织：需要 `ORG_DELETE` 权限
- 查看组织：需要 `ORG_READ` 权限

## 性能优化

### 缓存策略
- 组织架构图缓存 5 分钟
- 组织统计数据缓存 10 分钟
- 单个组织信息缓存 1 分钟

### 查询优化
- Neo4j 专门优化组织层级查询
- 支持深度限制避免过深查询
- 使用索引优化常用查询路径

### 分页策略
- 默认分页大小：50
- 最大分页大小：1000
- 支持游标分页用于大数据集

## CDC 数据同步

### 同步流程
1. PostgreSQL 写操作完成
2. 发布领域事件到 Event Bus
3. CDC Consumer 监听事件
4. 转换数据格式并同步到 Neo4j
5. 更新 Neo4j 的关系图数据

### 数据一致性
- 最终一致性模型
- 写操作立即反映在 PostgreSQL
- Neo4j 查询可能有轻微延迟（通常 < 100ms）
- 支持强一致性查询（直接查询 PostgreSQL）

## 版本兼容性

### API 版本控制
- 当前版本：v1
- 向后兼容承诺：同一主版本内保持兼容
- 废弃策略：提前 6 个月通知

### 迁移指南
- 从旧版 API 迁移到 CQRS 版本
- 数据格式保持兼容
- 新增字段采用可选设计

## 监控和日志

### 关键指标
- API 响应时间
- 错误率
- CDC 同步延迟
- 查询性能

### 日志格式
```json
{
  "timestamp": "2024-01-01T00:00:00.000Z",
  "level": "INFO",
  "message": "Organization created successfully",
  "context": {
    "operation": "create_organization",
    "org_id": "org-uuid",
    "tenant_id": "tenant-uuid",
    "user_id": "user-uuid"
  }
}
```