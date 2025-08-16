# 职位管理API完整规范 (双重标识系统版本)

**版本**: v1.1  
**创建日期**: 2025-08-04  
**最后更新**: 2025-08-04  
**状态**: 生产就绪  
**基于实现**: `PositionHandlerBusinessID` (双重标识系统)

## 📋 概述

本文档定义了Cube Castle项目中职位管理API的完整规范，采用**双重标识系统**：业务ID（1000000-9999999）+ UUID 混合标识策略，提供用户友好且技术健壮的职位管理功能。

## 🎯 核心特性

### 1. 双重标识系统架构
- **业务ID**: 1000000-9999999 范围，用户友好的数字标识
- **系统UUID**: 内部系统使用的全局唯一标识符
- **查询模式**: 默认业务ID，支持UUID查询切换
- **响应控制**: 通过参数控制UUID显示

### 2. 多态职位类型支持
- **FULL_TIME**: 全职职位，包含薪资范围、福利配置
- **PART_TIME**: 兼职职位，包含时薪、工作时长限制
- **CONTINGENT_WORKER**: 合同工，包含合同期限、时薪配置
- **INTERN**: 实习生，包含实习期限、津贴配置

### 3. 层级管理结构
- **管理者-下属关系**: 基于manager_position_id的层级结构
- **层级验证**: 防止循环引用和无效层级
- **级联查询**: 支持上级、下级职位信息查询

### 4. 关联查询优化
- **按需加载**: 通过查询参数控制关联数据加载
- **性能优化**: 避免N+1查询问题
- **缓存策略**: 关联实体信息缓存优化

## 🔗 API端点

### 基础路由
```
Base URL: /api/v1/positions
```

### 1. 创建职位

**POST** `/positions`

创建新的职位记录，系统自动生成业务ID。

#### 请求体
```json
{
  "position_type": "FULL_TIME",        // required: 职位类型
  "job_profile_id": "profile-uuid",    // required: 岗位模板UUID
  "department_id": "100001",           // required: 部门业务ID
  "manager_position_id": "1000000",    // optional: 管理者职位业务ID
  "status": "OPEN",                    // optional: 职位状态 (默认: OPEN)
  "budgeted_fte": 1.0,                // optional: 预算FTE (默认: 1.0)
  "details": {                         // optional: 类型特定配置
    "salary_range": {
      "min": 60000,
      "max": 90000,
      "currency": "CNY"
    },
    "benefits": ["health_insurance", "annual_leave"]
  }
}
```

#### 响应 (201 Created)
```json
{
  "id": "1000001",                     // 业务ID
  "tenant_id": "tenant-uuid",
  "position_type": "FULL_TIME",
  "job_profile_id": "profile-uuid",
  "department_id": "100001",           // 转换为业务ID
  "manager_position_id": "1000000",    // 管理者职位业务ID
  "status": "OPEN",
  "budgeted_fte": 1.0,
  "details": { ... },
  "created_at": "2025-08-04T10:00:00Z",
  "updated_at": "2025-08-04T10:00:00Z"
}
```

### 2. 获取职位详情

**GET** `/positions/{position_id}`

获取单个职位的详细信息，支持双重标识查询。

#### 路径参数
- `position_id`: 职位业务ID (1000000-9999999) 或 UUID

#### 查询参数
- `uuid_lookup=true`: 使用UUID查询模式（position_id应为UUID格式）
- `include_uuid=true`: 响应中包含系统UUID
- `with_department=true`: 包含部门信息
- `with_manager=true`: 包含管理者职位信息
- `with_incumbents=true`: 包含当前在职员工信息
- `with_direct_reports=true`: 包含直接下属职位信息

#### 响应 (200 OK)
```json
{
  "id": "1000001",                     // 业务ID
  "uuid": "123e4567-e89b-12d3-a456-426614174000", // 当include_uuid=true时
  "tenant_id": "tenant-uuid",
  "position_type": "FULL_TIME",
  "job_profile_id": "profile-uuid",
  "department_id": "100001",           // 部门业务ID
  "manager_position_id": "1000000",    // 管理者职位业务ID
  "status": "FILLED",
  "budgeted_fte": 1.0,
  "details": {
    "salary_range": {
      "min": 60000,
      "max": 90000,
      "currency": "CNY"
    }
  },
  "created_at": "2025-08-04T10:00:00Z",
  "updated_at": "2025-08-04T10:00:00Z",
  
  // 扩展信息 (根据查询参数)
  "department": {                      // 当with_department=true时
    "id": "100001",                    // 部门业务ID
    "name": "Engineering Department",
    "unit_type": "DEPARTMENT"
  },
  "manager": {                        // 当with_manager=true时
    "id": "1000000",                  // 管理者职位业务ID
    "position_type": "FULL_TIME",
    "status": "FILLED"
  },
  "incumbents": [                     // 当with_incumbents=true时
    {
      "id": "12345",                  // 员工业务ID
      "first_name": "John",
      "last_name": "Doe",
      "email": "john.doe@company.com"
    }
  ],
  "direct_reports": [                 // 当with_direct_reports=true时
    {
      "id": "1000002",               // 下属职位业务ID
      "position_type": "FULL_TIME",
      "status": "OPEN"
    }
  ]
}
```

### 3. 更新职位

**PUT** `/positions/{position_id}`

更新职位信息，支持部分更新。

#### 路径参数
- `position_id`: 职位业务ID

#### 查询参数
- `include_uuid=true`: 响应中包含系统UUID

#### 请求体 (部分更新)
```json
{
  "job_profile_id": "new-profile-uuid", // optional
  "department_id": "100002",            // optional: 部门业务ID
  "manager_position_id": "1000005",     // optional: 管理者职位业务ID
  "status": "FILLED",                   // optional
  "budgeted_fte": 0.8,                 // optional
  "details": {                          // optional: 覆盖现有details
    "salary_range": {
      "min": 65000,
      "max": 95000,
      "currency": "CNY"
    }
  }
}
```

#### 响应 (200 OK)
返回更新后的完整职位信息，格式同获取职位详情响应。

### 4. 删除职位

**DELETE** `/positions/{position_id}`

删除职位记录，包含完整的约束检查。

#### 路径参数
- `position_id`: 职位业务ID

#### 响应 (204 No Content)
成功删除，无响应体。

#### 删除约束检查
系统会检查以下条件，如不满足则返回409错误：
- ❌ 职位有直接下属职位
- ❌ 职位有当前在职员工
- ❌ 职位有历史任职记录

### 5. 职位列表查询

**GET** `/positions`

分页查询职位列表，支持多维度过滤。

#### 查询参数
- `page=1`: 页码（默认: 1）
- `page_size=20`: 每页数量（默认: 20，最大: 100）
- `position_type=FULL_TIME`: 按职位类型过滤
- `status=OPEN`: 按状态过滤
- `department_id=100001`: 按部门过滤（部门业务ID）
- `include_uuid=true`: 包含系统UUID
- `with_department=true`: 包含部门信息

#### 响应 (200 OK)
```json
{
  "positions": [
    {
      "id": "1000001",
      "tenant_id": "tenant-uuid",
      "position_type": "FULL_TIME",
      "job_profile_id": "profile-uuid",
      "department_id": "100001",
      "status": "OPEN",
      "budgeted_fte": 1.0,
      "details": { ... },
      "created_at": "2025-08-04T10:00:00Z",
      "updated_at": "2025-08-04T10:00:00Z",
      "department": {                 // 当with_department=true时
        "id": "100001",
        "name": "Engineering",
        "unit_type": "DEPARTMENT"
      }
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total": 150,
    "total_pages": 8
  }
}
```

### 6. 职位统计信息

**GET** `/positions/stats`

获取职位统计数据，支持实时计算。

#### 响应 (200 OK)
```json
{
  "total_positions": 150,
  "total_budgeted_fte": 142.5,
  "by_type": {
    "FULL_TIME": 120,
    "PART_TIME": 15,
    "CONTINGENT_WORKER": 10,
    "INTERN": 5
  },
  "by_status": {
    "OPEN": 25,
    "FILLED": 115,
    "FROZEN": 8,
    "PENDING_ELIMINATION": 2
  }
}
```

## 📊 数据模型

### Position 对象 (双重标识版本)
```typescript
interface PositionBusinessID {
  id: string                      // 业务ID (1000000-9999999)
  uuid?: string                   // 系统UUID (可选，通过include_uuid控制)
  tenant_id: string              // 租户ID (暂时使用UUID)
  position_type: PositionType     // 职位类型
  job_profile_id: string         // 岗位模板UUID
  department_id: string          // 部门业务ID
  manager_position_id?: string   // 管理者职位业务ID (可选)
  status: PositionStatus         // 职位状态
  budgeted_fte: number           // 预算FTE
  details: Record<string, any>   // 多态配置
  created_at: string             // 创建时间 (ISO 8601)
  updated_at: string             // 更新时间 (ISO 8601)
  
  // 扩展信息 (可选)
  department?: DepartmentInfo     // 部门信息
  manager?: PositionInfo         // 管理者信息
  incumbents?: EmployeeInfo[]    // 在职员工
  direct_reports?: PositionInfo[] // 下属职位
}
```

### 枚举类型

#### PositionType (职位类型)
```typescript
enum PositionType {
  FULL_TIME = "FULL_TIME",              // 全职
  PART_TIME = "PART_TIME",              // 兼职
  CONTINGENT_WORKER = "CONTINGENT_WORKER", // 合同工
  INTERN = "INTERN"                     // 实习生
}
```

#### PositionStatus (职位状态)
```typescript
enum PositionStatus {
  OPEN = "OPEN",                        // 开放招聘
  FILLED = "FILLED",                    // 已填充
  FROZEN = "FROZEN",                    // 冻结
  PENDING_ELIMINATION = "PENDING_ELIMINATION" // 待裁撤
}
```

### 扩展信息类型

#### DepartmentInfo
```typescript
interface DepartmentInfo {
  id: string          // 部门业务ID
  name: string        // 部门名称
  unit_type: string   // 单元类型
}
```

#### PositionInfo
```typescript
interface PositionInfo {
  id: string              // 职位业务ID
  position_type: string   // 职位类型
  status: string          // 职位状态
}
```

#### EmployeeInfo
```typescript
interface EmployeeInfo {
  id: string        // 员工业务ID
  first_name: string
  last_name: string
  email: string
}
```

### 多态详细配置

#### FULL_TIME 配置示例
```json
{
  "salary_range": {
    "min": 50000,
    "max": 80000,
    "currency": "CNY"
  },
  "benefits": ["health_insurance", "annual_leave", "stock_options"],
  "work_schedule": "9_to_5",
  "remote_allowed": true
}
```

#### PART_TIME 配置示例
```json
{
  "hourly_rate": 100,
  "max_hours_per_week": 20,
  "flexible_schedule": true,
  "benefits": ["proportional_leave"]
}
```

#### CONTINGENT_WORKER 配置示例
```json
{
  "contract_duration": "12m",
  "hourly_rate": 150,
  "renewal_possible": true,
  "specialized_skills": ["react", "nodejs"]
}
```

#### INTERN 配置示例
```json
{
  "internship_duration": "3m",
  "stipend": 3000,
  "mentor_assigned": true,
  "learning_objectives": ["web_development", "agile_methods"]
}
```

## 🔄 业务规则

### 职位创建规则
1. **业务ID生成**: 系统自动生成1000000-9999999范围内的唯一业务ID
2. **部门验证**: department_id必须是有效的组织单元业务ID
3. **管理者验证**: manager_position_id（如提供）必须是有效的职位业务ID
4. **类型验证**: 根据position_type验证details字段的结构和内容
5. **租户隔离**: 所有关联实体必须属于同一租户
6. **FTE验证**: budgeted_fte必须在0-5之间

### 职位更新规则
1. **可更新字段**: job_profile_id, department_id, manager_position_id, status, budgeted_fte, details
2. **不可更新字段**: id, tenant_id, position_type, created_at
3. **关联验证**: 更新的关联ID必须存在且属于同一租户
4. **状态转换**: 遵循职位生命周期状态转换规则
5. **层级验证**: 防止管理者层级循环引用

### 职位删除规则
1. **层级约束**: 有下属职位的职位不能删除
2. **在职约束**: 有当前在职员工的职位不能删除
3. **历史约束**: 有历史任职记录的职位不能删除
4. **级联影响**: 删除前需要处理所有依赖关系

## 🔒 安全和权限

### 认证授权
- **Bearer Token**: 所有API调用需要有效的JWT token
- **租户隔离**: 严格的多租户数据隔离，用户只能访问所属租户数据
- **权限控制**: 基于角色的操作权限验证

### 权限级别
- `positions:read` - 查看职位信息
- `positions:write` - 创建和更新职位
- `positions:delete` - 删除职位
- `positions:stats` - 查看统计信息

### 数据验证
- **输入验证**: 严格的请求参数和数据格式验证
- **业务ID验证**: 完整的业务ID格式和范围验证
- **关联验证**: 确保所有关联实体的有效性
- **SQL注入防护**: 使用参数化查询防止SQL注入攻击

## ⚡ 性能规范

### 响应时间目标
- 单个职位查询: < 100ms
- 职位列表查询: < 200ms
- 职位创建: < 300ms
- 职位更新: < 250ms
- 职位删除: < 200ms
- 统计查询: < 500ms

### 优化策略
- **索引优化**: business_id, department_id, manager_position_id, status字段建立索引
- **查询优化**: 避免N+1查询问题，使用预加载和批量查询
- **缓存策略**: 部门信息转换结果缓存，关联查询结果缓存
- **分页优化**: 高效的分页查询和计数优化

### 资源使用
- **内存使用**: 关联查询时控制内存使用，避免大量数据加载
- **数据库连接**: 优化连接池使用，避免连接泄漏
- **并发处理**: 支持1000+ QPS的并发处理能力

## 🚨 错误处理

### 标准错误格式
```json
{
  "error": "VALIDATION_ERROR",
  "message": "Invalid request data",
  "details": {
    "field": "position_type",
    "value": "INVALID_TYPE",
    "constraint": "must be one of: FULL_TIME, PART_TIME, CONTINGENT_WORKER, INTERN"
  },
  "timestamp": "2025-08-04T10:00:00Z",
  "request_id": "req_12345678"
}
```

### 错误代码映射

| HTTP状态码 | 错误代码 | 说明 |
|-----------|---------|------|
| 400 | `VALIDATION_ERROR` | 请求数据验证失败 |
| 400 | `INVALID_BUSINESS_ID` | 无效的业务ID格式 |
| 400 | `INVALID_UUID` | 无效的UUID格式 |
| 401 | `UNAUTHORIZED` | 认证失败 |
| 403 | `PERMISSION_DENIED` | 权限不足 |
| 404 | `POSITION_NOT_FOUND` | 职位不存在 |
| 409 | `HAS_SUBORDINATES` | 职位有下属，无法删除 |
| 409 | `HAS_CURRENT_INCUMBENTS` | 职位有在职员工，无法删除 |
| 409 | `HAS_OCCUPANCY_HISTORY` | 职位有历史记录，无法删除 |
| 500 | `INTERNAL_ERROR` | 服务器内部错误 |

## 🧪 测试用例

### 功能测试
- [x] 双重标识系统测试
- [x] 多态职位类型测试
- [x] 关联查询功能测试
- [x] 业务规则验证测试
- [x] 删除约束检查测试

### 性能测试
- [x] 1000个职位列表查询 < 200ms
- [x] 并发100个请求响应时间 < 500ms
- [x] 关联查询性能测试
- [x] 统计查询性能测试

### 安全测试
- [x] 业务ID注入攻击防护
- [x] 跨租户访问防护
- [x] 权限验证测试
- [x] 输入验证测试

## 📈 监控指标

### 关键指标
- **业务ID使用率**: 业务ID vs UUID查询的比例
- **关联查询使用情况**: 各种关联查询参数的使用统计
- **错误率分布**: 各种错误类型的发生频率
- **性能指标**: 各端点的响应时间分布

### 告警阈值
- 响应时间 > 1s: 警告
- 错误率 > 1%: 警告
- 错误率 > 5%: 严重
- 业务ID使用率 < 60%: 关注

## 🔧 实施状态

### 已完成 ✅
- 双重标识系统实现
- 多态职位类型支持
- 完整的CRUD操作
- 关联查询功能
- 统计功能
- 业务规则验证
- 性能优化

### API兼容性
- **向后兼容**: 支持原有UUID查询方式
- **渐进迁移**: 允许客户端逐步迁移到业务ID
- **双模式支持**: 同时支持业务ID和UUID查询

## 📚 使用示例

### TypeScript客户端示例
```typescript
import { positionsApi } from '@/lib/api/positions'

// 使用业务ID查询职位（默认模式）
const position = await positionsApi.getPosition('1000001', {
  with_department: true,
  with_incumbents: true
})

// 使用UUID查询职位（兼容模式）  
const positionByUUID = await positionsApi.getPosition(
  '123e4567-e89b-12d3-a456-426614174000',
  { uuid_lookup: true }
)

// 创建职位（系统自动生成业务ID）
const newPosition = await positionsApi.createPosition({
  position_type: 'FULL_TIME',
  job_profile_id: 'profile-uuid',
  department_id: '100001',  // 部门业务ID
  manager_position_id: '1000000', // 管理者职位业务ID
  details: {
    salary_range: { min: 60000, max: 90000, currency: 'CNY' }
  }
})
```

### cURL示例
```bash
# 使用业务ID获取职位（默认模式）
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/v1/positions/1000001?with_department=true"

# 使用UUID获取职位（兼容模式）
curl -H "Authorization: Bearer <token>" \
     "http://localhost:8080/api/v1/positions/123e4567-e89b-12d3-a456-426614174000?uuid_lookup=true"

# 创建职位
curl -X POST \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer <token>" \
     -d '{
       "position_type": "FULL_TIME",
       "job_profile_id": "profile-uuid",
       "department_id": "100001",
       "details": {"salary_range": {"min": 60000, "max": 90000, "currency": "CNY"}}
     }' \
     "http://localhost:8080/api/v1/positions"
```

## 🔄 版本历史

### v1.1 (2025-08-04) - 双重标识系统版本
- ✅ 实现双重标识系统（业务ID + UUID）
- ✅ 支持多态职位类型配置
- ✅ 添加关联查询支持
- ✅ 实现统计功能
- ✅ 完成业务规则验证
- ✅ 性能优化和缓存策略

### v1.0 (2025-08-04) - 基础版本
- ✅ 基础CRUD功能
- ✅ UUID标识系统
- ✅ 租户隔离
- ✅ 基本权限控制

### 兼容性承诺
- 向后兼容性保证：继续支持UUID查询方式
- 渐进迁移路径：客户端可以逐步迁移到业务ID
- 废弃功能通知：重大变更提前3个月通知

---

**维护者**: Cube Castle开发团队  
**技术负责人**: 系统架构师  
**文档状态**: 生产就绪  
**下次审核**: 2025-11-04