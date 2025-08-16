# 员工管理API规范

**版本**: v1.0  
**创建日期**: 2025-08-04  
**基于实际实现**: ✅ 已验证  
**状态**: 生产就绪

## 📋 概述

员工管理API提供完整的员工生命周期管理功能，支持业务ID和UUID双重查询模式，实现多租户隔离和丰富的关联查询。

### 核心特性
- **双重标识**: 支持业务ID（1-99999999）和UUID查询
- **多租户隔离**: 严格的租户数据边界控制
- **关联查询**: 支持职位、组织和管理层级关联
- **生命周期管理**: 完整的员工入职到离职流程
- **类型多态**: 基于员工类型的详细信息配置

## 🏗️ API端点总览

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/employees` | 获取员工列表 | Bearer Token |
| POST | `/api/v1/employees` | 创建新员工 | Bearer Token |
| GET | `/api/v1/employees/{employee_id}` | 获取单个员工 | Bearer Token |
| PUT | `/api/v1/employees/{employee_id}` | 更新员工信息 | Bearer Token |
| DELETE | `/api/v1/employees/{employee_id}` | 删除员工 | Bearer Token |
| GET | `/api/v1/employees/stats` | 获取员工统计信息 | Bearer Token |
| GET | `/api/v1/employees/validate` | 验证业务ID格式 | Bearer Token |

## 📊 数据模型

### 员工核心模型
```json
{
  "id": "uuid",
  "business_id": "string (1-99999999)",
  "tenant_id": "uuid",
  "employee_type": "FULL_TIME | PART_TIME | CONTRACTOR | INTERN",
  "employee_number": "string",
  "first_name": "string",
  "last_name": "string",
  "email": "string",
  "personal_email": "string (optional)",
  "phone_number": "string (optional)",
  "current_position_id": "uuid (optional)",
  "employment_status": "ACTIVE | ON_LEAVE | TERMINATED | SUSPENDED | PENDING_START",
  "hire_date": "2025-08-04T00:00:00Z",
  "termination_date": "2025-08-04T00:00:00Z (optional)",
  "employee_details": {},
  "created_at": "2025-08-04T00:00:00Z",
  "updated_at": "2025-08-04T00:00:00Z"
}
```

### 员工类型枚举
```yaml
FULL_TIME: 全职员工
PART_TIME: 兼职员工
CONTRACTOR: 合同工
INTERN: 实习生
```

### 就业状态枚举
```yaml
ACTIVE: 在职
ON_LEAVE: 休假
TERMINATED: 离职
SUSPENDED: 停职
PENDING_START: 待入职
```

## 🔍 API详细规范

### 1. 获取员工列表

**`GET /api/v1/employees`**

获取当前租户下的员工列表，支持分页、过滤和关联查询。

#### 查询参数
```yaml
# 分页参数
page: 页码，默认1
page_size: 每页大小，默认20，最大100

# 搜索参数
search: 搜索关键词（姓名、邮箱、员工号）

# 过滤参数
status: 就业状态过滤
employee_type: 员工类型过滤
organization_id: 组织ID过滤（业务ID格式）

# 查询选项
include_uuid: 是否包含UUID，默认false
with_position: 是否包含职位信息，默认false
with_organization: 是否包含组织信息，默认false
with_manager: 是否包含管理者信息，默认false
```

#### 响应示例
```json
{
  "employees": [
    {
      "business_id": "123456",
      "employee_number": "EMP001",
      "first_name": "张",
      "last_name": "三",
      "email": "zhang.san@company.com",
      "employment_status": "ACTIVE",
      "employee_type": "FULL_TIME",
      "hire_date": "2025-01-01T00:00:00Z",
      "current_position": {
        "position_title": "高级软件工程师",
        "department": "技术部",
        "location": "北京"
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

### 2. 创建员工

**`POST /api/v1/employees`**

创建新员工记录，支持关联职位和组织信息。

#### 请求体
```json
{
  "first_name": "张",
  "last_name": "三",
  "email": "zhang.san@company.com",
  "personal_email": "zhang.san@gmail.com",
  "phone_number": "+86-13812345678",
  "hire_date": "2025-08-04",
  "employee_type": "FULL_TIME",
  "position_id": "654321",
  "organization_id": "100001",
  "manager_id": "789123",
  "employee_details": {
    "salary_grade": "L6",
    "probation_period": 3,
    "work_location": "北京"
  }
}
```

#### 响应示例
```json
{
  "business_id": "123456",
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三",
  "email": "zhang.san@company.com",
  "employment_status": "ACTIVE",
  "employee_type": "FULL_TIME",
  "created_at": "2025-08-04T10:30:00Z"
}
```

### 3. 获取单个员工

**`GET /api/v1/employees/{employee_id}`**

根据业务ID或UUID获取员工详细信息。

#### 路径参数
- `employee_id`: 员工业务ID（默认）或UUID

#### 查询参数
```yaml
uuid_lookup: 是否使用UUID查询，默认false
include_uuid: 是否在响应中包含UUID，默认false
with_position: 是否包含职位信息，默认false
with_organization: 是否包含组织信息，默认false
with_manager: 是否包含管理者信息，默认false
```

#### 响应示例
```json
{
  "business_id": "123456",
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三",
  "email": "zhang.san@company.com",
  "personal_email": "zhang.san@gmail.com",
  "phone_number": "+86-13812345678",
  "employment_status": "ACTIVE",
  "employee_type": "FULL_TIME",
  "hire_date": "2025-01-01T00:00:00Z",
  "current_position": {
    "position_title": "高级软件工程师",
    "department": "技术部",
    "job_level": "L6",
    "location": "北京"
  },
  "organization": {
    "name": "技术部",
    "unit_type": "DEPARTMENT"
  },
  "manager": {
    "business_id": "789123",
    "first_name": "李",
    "last_name": "四",
    "position_title": "技术总监"
  },
  "employee_details": {
    "salary_grade": "L6",
    "probation_period": 3,
    "work_location": "北京"
  },
  "created_at": "2025-01-01T09:00:00Z",
  "updated_at": "2025-08-04T10:30:00Z"
}
```

### 4. 更新员工信息

**`PUT /api/v1/employees/{employee_id}`**

更新员工信息，支持部分字段更新。

#### 请求体
```json
{
  "personal_email": "new.email@gmail.com",
  "phone_number": "+86-13987654321",
  "employee_details": {
    "salary_grade": "L7",
    "work_location": "上海"
  }
}
```

#### 响应示例
```json
{
  "business_id": "123456",
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三",
  "email": "zhang.san@company.com",
  "personal_email": "new.email@gmail.com",
  "phone_number": "+86-13987654321",
  "employment_status": "ACTIVE",
  "employee_type": "FULL_TIME",
  "updated_at": "2025-08-04T10:45:00Z"
}
```

### 5. 删除员工

**`DELETE /api/v1/employees/{employee_id}`**

删除员工记录（软删除，标记为已删除状态）。

#### 响应
- **204 No Content**: 删除成功
- **404 Not Found**: 员工不存在
- **409 Conflict**: 员工有关联数据，无法删除

### 6. 获取员工统计信息

**`GET /api/v1/employees/stats`**

获取员工统计信息，包括总数、分类统计等。

#### 响应示例
```json
{
  "total_employees": 150,
  "active_employees": 145,
  "by_type": {
    "FULL_TIME": 120,
    "PART_TIME": 15,
    "CONTRACTOR": 8,
    "INTERN": 7
  },
  "by_status": {
    "ACTIVE": 145,
    "ON_LEAVE": 3,
    "TERMINATED": 2,
    "SUSPENDED": 0,
    "PENDING_START": 0
  },
  "recent_hires": 5,
  "terminations_this_month": 1
}
```

### 7. 验证业务ID

**`GET /api/v1/employees/validate`**

验证员工业务ID格式是否正确。

#### 查询参数
```yaml
business_id: 待验证的业务ID
```

#### 响应示例
```json
{
  "business_id": "123456",
  "valid": true
}
```

## 🔐 安全与认证

### 认证方式
```yaml
类型: Bearer Token (JWT)
头部: Authorization: Bearer <token>
必需: 所有API端点都需要认证
```

### 租户隔离
```yaml
机制: JWT令牌中的tenant_id
作用: 确保数据访问限制在租户范围内
验证: 每个请求都验证租户权限
```

### 权限控制
```yaml
读取权限: hr.employee.read
创建权限: hr.employee.create
更新权限: hr.employee.update
删除权限: hr.employee.delete
统计权限: hr.employee.stats
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

### 分页限制
```yaml
默认页大小: 20
最大页大小: 100
最大查询记录: 10000
```

## ❌ 错误处理

### 错误响应格式
```json
{
  "error": "EMPLOYEE_NOT_FOUND",
  "message": "Employee not found",
  "details": null,
  "timestamp": "2025-08-04T10:30:00Z",
  "request_id": "req_12345678"
}
```

### 常用错误码
```yaml
INVALID_REQUEST: 请求格式错误
VALIDATION_ERROR: 数据验证失败
EMPLOYEE_NOT_FOUND: 员工不存在
EMPLOYEE_ALREADY_EXISTS: 员工已存在
UNAUTHORIZED: 未授权访问
FORBIDDEN: 权限不足
INTERNAL_ERROR: 服务器内部错误
```

## 🧪 API测试示例

### 使用curl测试

#### 获取员工列表
```bash
curl -X GET "http://localhost:8080/api/v1/employees?page=1&page_size=10&with_position=true" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

#### 创建员工
```bash
curl -X POST "http://localhost:8080/api/v1/employees" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "张",
    "last_name": "三",
    "email": "zhang.san@company.com",
    "hire_date": "2025-08-04",
    "employee_type": "FULL_TIME"
  }'
```

#### 获取单个员工
```bash
curl -X GET "http://localhost:8080/api/v1/employees/123456?with_position=true&with_organization=true" \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json"
```

## 📚 最佳实践

### 1. 查询优化
- 使用业务ID进行常规查询，性能更优
- 需要UUID时使用`uuid_lookup=true`参数
- 合理使用关联查询参数，避免过度查询

### 2. 数据创建
- 创建前验证关联的职位和组织ID
- 使用适当的员工类型和详细配置
- 确保邮箱地址唯一性

### 3. 错误处理
- 检查响应状态码和错误信息
- 实现适当的重试机制
- 记录错误日志便于排查

### 4. 性能考虑
- 使用分页查询大量数据
- 缓存频繁访问的员工信息
- 监控API响应时间

---

**制定者**: 系统架构师  
**审核者**: 技术委员会  
**生效日期**: 2025-08-04  
**下次审查**: 2025-11-04