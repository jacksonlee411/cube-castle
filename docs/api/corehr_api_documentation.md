# CoreHR API文档 | CoreHR API Documentation

**API版本 | API Version**: v1.7.0  
**更新日期 | Last Updated**: 2025年7月31日  
**状态 | Status**: 生产就绪 | Production Ready  

本文档描述Cube Castle CoreHR模块的API端点，所有API现在都基于真实数据库操作。  
*This document describes the Cube Castle CoreHR module API endpoints, all APIs now based on real database operations.*

---

## 🎯 重要变更说明 | Important Changes

### v1.7.0 Mock替换更新 | Mock Replacement Update
- **✅ 移除所有Mock实现**: 所有API端点现在直接操作PostgreSQL数据库
  *Removed all mock implementations: All API endpoints now directly operate on PostgreSQL database*
- **✅ 真实数据库验证**: 所有输入数据通过CoreHRValidationChecker进行真实验证
  *Real database validation: All input data validated through CoreHRValidationChecker*
- **✅ 事务性操作**: 创建、更新、删除操作都在数据库事务中执行
  *Transactional operations: Create, update, delete operations executed in database transactions*
- **✅ 事件记录**: 所有数据变更自动记录到outbox事件系统
  *Event logging: All data changes automatically logged to outbox event system*

---

## 🏢 员工管理API | Employee Management API

### GET /api/v1/corehr/employees
获取员工列表 | Get employee list

**请求参数 | Request Parameters**:
```json
{
  "page": 1,           // 页码 | Page number (default: 1)
  "page_size": 10,     // 每页大小 | Page size (default: 10, max: 100)
  "search": ""         // 搜索关键词 | Search keyword (optional)
}
```

**响应示例 | Response Example**:
```json
{
  "success": true,
  "data": {
    "employees": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "employee_number": "EMP001",
        "first_name": "张",
        "last_name": "三",
        "email": "zhang.san@example.com",  
        "phone_number": "13800138000",
        "position": "软件工程师",
        "department": "技术部",
        "hire_date": "2024-01-15",
        "manager_id": "550e8400-e29b-41d4-a716-446655440001",
        "status": "active",
        "created_at": "2024-01-15T09:00:00Z",
        "updated_at": "2024-01-15T09:00:00Z"
      }
    ],
    "total": 1,
    "page": 1,
    "page_size": 10
  }
}
```

**错误响应 | Error Response**:
```json
{
  "success": false,
  "error": {
    "code": "DATABASE_ERROR",
    "message": "service not properly initialized: repository is nil",
    "details": "数据库连接未正确初始化"
  }
}
```

### POST /api/v1/corehr/employees
创建新员工 | Create new employee

**请求体 | Request Body**:
```json
{
  "employee_number": "EMP002",     // 必填 | Required, unique
  "first_name": "李",              // 必填 | Required
  "last_name": "四",               // 必填 | Required  
  "email": "li.si@example.com",    // 必填 | Required, unique, valid email
  "phone_number": "13800138001",   // 可选 | Optional
  "position": "产品经理",           // 可选 | Optional
  "department": "产品部",          // 可选 | Optional
  "hire_date": "2024-02-01",       // 必填 | Required, YYYY-MM-DD format
  "manager_id": "550e8400-e29b-41d4-a716-446655440001" // 可选 | Optional
}
```

**验证规则 | Validation Rules**:
- `employee_number`: 必须唯一，长度1-50字符
  *Must be unique, 1-50 characters*
- `email`: 必须是有效邮箱格式且唯一
  *Must be valid email format and unique*
- `hire_date`: 必须是有效日期，不能是未来日期
  *Must be valid date, cannot be future date*
- `manager_id`: 如果提供，必须是有效的员工ID
  *If provided, must be valid employee ID*

**成功响应 | Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "employee_number": "EMP002",
    "first_name": "李",
    "last_name": "四",
    "email": "li.si@example.com",
    "phone_number": "13800138001",
    "position": "产品经理",
    "department": "产品部", 
    "hire_date": "2024-02-01",
    "manager_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "active",
    "created_at": "2024-02-01T10:30:00Z",
    "updated_at": "2024-02-01T10:30:00Z"
  },
  "message": "员工创建成功，已记录employee.created事件"
}
```

### GET /api/v1/corehr/employees/{id}
获取员工详情 | Get employee details

**路径参数 | Path Parameters**:
- `id`: 员工UUID | Employee UUID

**响应示例 | Response Example**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_number": "EMP001",
    "first_name": "张",
    "last_name": "三",
    "email": "zhang.san@example.com",
    "phone_number": "13800138000",
    "position": "软件工程师",
    "department": "技术部",
    "hire_date": "2024-01-15",
    "manager_id": "550e8400-e29b-41d4-a716-446655440001",
    "status": "active",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-01-15T09:00:00Z"
  }
}
```

### PUT /api/v1/corehr/employees/{id}
更新员工信息 | Update employee information

**请求体 | Request Body**:
```json
{
  "first_name": "张",              // 可选 | Optional
  "last_name": "三丰",             // 可选 | Optional
  "email": "zhang.sanfeng@example.com", // 可选 | Optional, must be unique
  "phone_number": "13800138002",   // 可选 | Optional
  "position": "高级软件工程师",     // 可选 | Optional
  "department": "技术部",          // 可选 | Optional
  "manager_id": "550e8400-e29b-41d4-a716-446655440003" // 可选 | Optional
}
```

**成功响应 | Success Response**:
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "employee_number": "EMP001",
    "first_name": "张",
    "last_name": "三丰",
    "email": "zhang.sanfeng@example.com",
    "phone_number": "13800138002",
    "position": "高级软件工程师",
    "department": "技术部",
    "hire_date": "2024-01-15",
    "manager_id": "550e8400-e29b-41d4-a716-446655440003",
    "status": "active",
    "created_at": "2024-01-15T09:00:00Z",
    "updated_at": "2024-07-31T13:20:00Z"
  },
  "message": "员工信息更新成功，已记录employee.updated事件"
}
```

---

## 🏗️ 组织管理API | Organization Management API

### GET /api/v1/corehr/organizations  
获取组织列表 | Get organization list

**响应示例 | Response Example**:
```json
{
  "success": true,
  "data": {
    "organizations": [
      {
        "id": "660f8400-e29b-41d4-a716-446655440000",
        "name": "技术部",
        "code": "TECH001", 
        "parent_id": "660f8400-e29b-41d4-a716-446655440001",
        "level": 2,
        "created_at": "2024-01-01T08:00:00Z",
        "updated_at": "2024-01-01T08:00:00Z"
      },
      {
        "id": "660f8400-e29b-41d4-a716-446655440001",
        "name": "Cube Castle Inc.",
        "code": "CC001",
        "parent_id": null,
        "level": 1,
        "created_at": "2024-01-01T08:00:00Z", 
        "updated_at": "2024-01-01T08:00:00Z"
      }
    ],
    "total": 2
  }
}
```

### POST /api/v1/corehr/organizations
创建组织单位 | Create organization unit

**请求体 | Request Body**:
```json
{
  "name": "产品部",                // 必填 | Required
  "code": "PROD001",               // 必填 | Required, unique
  "parent_id": "660f8400-e29b-41d4-a716-446655440001", // 可选 | Optional
  "level": 2                       // 可选 | Optional, default: 1
}
```

**验证规则 | Validation Rules**:
- `name`: 必填，长度1-255字符
  *Required, 1-255 characters*
- `code`: 必填且唯一，长度1-50字符
  *Required and unique, 1-50 characters*
- `parent_id`: 如果提供，必须是有效的组织ID
  *If provided, must be valid organization ID*
- `level`: 必须是正整数
  *Must be positive integer*

### GET /api/v1/corehr/organizations/tree
获取组织层级树 | Get organization hierarchy tree

**响应示例 | Response Example**:
```json
{
  "success": true,
  "data": {
    "tree": [
      {
        "id": "660f8400-e29b-41d4-a716-446655440001",
        "name": "Cube Castle Inc.",
        "code": "CC001",
        "level": 1,
        "children": [
          {
            "id": "660f8400-e29b-41d4-a716-446655440000",
            "name": "技术部",
            "code": "TECH001",
            "level": 2,
            "children": []
          },
          {
            "id": "660f8400-e29b-41d4-a716-446655440002", 
            "name": "产品部",
            "code": "PROD001",
            "level": 2,
            "children": []
          }
        ]
      }
    ]
  }
}
```

---

## 🔧 错误处理 | Error Handling

### 错误类型 | Error Types

#### 1. 服务初始化错误 | Service Initialization Error
```json
{
  "success": false,
  "error": {
    "code": "SERVICE_NOT_INITIALIZED", 
    "message": "service not properly initialized: repository is nil",
    "details": "数据库连接未正确初始化，请检查数据库配置"
  }
}
```

#### 2. 验证错误 | Validation Error
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "输入数据验证失败",
    "details": {
      "email": "邮箱格式无效",
      "employee_number": "员工编号已存在"
    }
  }
}
```

#### 3. 数据库错误 | Database Error  
```json
{
  "success": false,
  "error": {
    "code": "DATABASE_ERROR",
    "message": "数据库操作失败",
    "details": "ERROR: duplicate key value violates unique constraint"
  }
}
```

#### 4. 资源未找到错误 | Resource Not Found Error
```json
{
  "success": false, 
  "error": {
    "code": "RESOURCE_NOT_FOUND",
    "message": "员工记录未找到",
    "details": "ID为550e8400-e29b-41d4-a716-446655440000的员工不存在"
  }
}
```

### HTTP状态码 | HTTP Status Codes
- `200 OK`: 请求成功 | Request successful
- `201 Created`: 资源创建成功 | Resource created successfully  
- `400 Bad Request`: 请求参数错误 | Invalid request parameters
- `404 Not Found`: 资源未找到 | Resource not found
- `422 Unprocessable Entity`: 验证失败 | Validation failed
- `500 Internal Server Error`: 服务器内部错误 | Internal server error
- `503 Service Unavailable`: 服务不可用（数据库连接问题）| Service unavailable (database connection issues)

---

## 🚀 性能指标 | Performance Metrics

### API响应时间 | API Response Times
- **员工列表查询 | Employee List Query**: 平均 7.32ms
- **员工创建 | Employee Creation**: 平均 8.28ms（包含事务和事件记录）
- **组织查询 | Organization Query**: 平均 <10ms
- **错误处理 | Error Handling**: 平均 153ns

### 数据库操作 | Database Operations
- **连接池管理 | Connection Pool**: 自动管理，支持高并发
  *Automatic management, supports high concurrency*
- **事务处理 | Transaction Processing**: ACID保证，自动回滚
  *ACID guaranteed, automatic rollback*
- **事件记录 | Event Logging**: 异步处理，不影响API响应时间
  *Asynchronous processing, no impact on API response time*

---

## 🔐 安全和认证 | Security and Authentication

### 数据验证 | Data Validation
- **输入清理 | Input Sanitization**: 所有输入数据自动清理和验证
  *All input data automatically sanitized and validated*
- **SQL注入防护 | SQL Injection Protection**: 使用参数化查询
  *Uses parameterized queries*
- **XSS防护 | XSS Protection**: 输出数据自动转义
  *Output data automatically escaped*

### 多租户支持 | Multi-tenant Support
- **数据隔离 | Data Isolation**: 严格的租户间数据隔离
  *Strict inter-tenant data isolation*
- **租户验证 | Tenant Validation**: 每个请求验证租户权限
  *Tenant permissions validated for each request*

---

**API文档版本 | API Documentation Version**: v1.0  
**最后更新 | Last Updated**: 2025-07-31 13:25:00  
**下次更新计划 | Next Update Scheduled**: 随版本发布更新  

**技术支持 | Technical Support**: 
- 详细错误信息已在响应中提供 | Detailed error information provided in responses
- 性能监控端点: `GET /metrics/http` | Performance monitoring endpoint
- 健康检查端点: `GET /health/detailed` | Health check endpoint