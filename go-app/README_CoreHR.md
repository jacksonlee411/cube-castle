# CoreHR API 文档

## 概述

CoreHR API 是 Cube Castle 项目的人力资源管理模块，提供了完整的员工管理和组织架构管理功能。

## 功能特性

### ✅ 已实现功能

1. **员工管理**
   - ✅ 获取员工列表（支持分页和搜索）
   - ✅ 创建新员工
   - ✅ 获取员工详情
   - ✅ 更新员工信息
   - ✅ 删除员工（软删除）

2. **组织管理**
   - ✅ 获取组织列表
   - ✅ 获取组织树结构

3. **数据验证**
   - ✅ 请求参数验证
   - ✅ 邮箱格式验证
   - ✅ 员工编号唯一性检查
   - ✅ 状态值验证

4. **错误处理**
   - ✅ 统一的错误响应格式
   - ✅ 用户友好的错误消息
   - ✅ 适当的 HTTP 状态码

## API 端点

### 基础 URL
```
http://localhost:8080/api/v1/corehr
```

### 员工管理端点

#### 1. 获取员工列表
```http
GET /api/v1/corehr/employees
```

**查询参数：**
- `page` (可选): 页码，默认为 1
- `page_size` (可选): 每页数量，默认为 20，最大 100
- `search` (可选): 搜索关键词，支持姓名、邮箱、员工编号搜索

**响应示例：**
```json
{
  "employees": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "employee_number": "EMP001",
      "first_name": "张",
      "last_name": "三",
      "email": "zhangsan@example.com",
      "phone_number": "13800138000",
      "hire_date": "2023-01-15",
      "status": "active",
      "created_at": "2023-01-15T00:00:00Z",
      "updated_at": "2023-01-15T00:00:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_pages": 1,
    "has_next": false,
    "has_prev": false
  },
  "total_count": 1
}
```

#### 2. 创建员工
```http
POST /api/v1/corehr/employees
```

**请求体：**
```json
{
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三",
  "email": "zhangsan@example.com",
  "hire_date": "2023-01-15",
  "phone_number": "13800138000",
  "position_id": "550e8400-e29b-41d4-a716-446655440001",
  "organization_id": "550e8400-e29b-41d4-a716-446655440002"
}
```

**必填字段：**
- `employee_number`: 员工编号（唯一）
- `first_name`: 姓名
- `last_name`: 姓氏
- `email`: 邮箱地址
- `hire_date`: 入职日期

**可选字段：**
- `phone_number`: 电话号码
- `position_id`: 职位ID
- `organization_id`: 组织ID

#### 3. 获取员工详情
```http
GET /api/v1/corehr/employees/{employee_id}
```

#### 4. 更新员工
```http
PUT /api/v1/corehr/employees/{employee_id}
```

**请求体：**
```json
{
  "first_name": "新姓名",
  "last_name": "新姓氏",
  "email": "newemail@example.com",
  "phone_number": "13900139000",
  "status": "active"
}
```

#### 5. 删除员工
```http
DELETE /api/v1/corehr/employees/{employee_id}
```

**注意：** 删除操作为软删除，将员工状态设置为 `inactive`

### 组织管理端点

#### 1. 获取组织列表
```http
GET /api/v1/corehr/organizations
```

**响应示例：**
```json
{
  "organizations": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "技术部",
      "code": "TECH",
      "level": 1,
      "parent_id": null,
      "status": "active",
      "created_at": "2023-01-15T00:00:00Z",
      "updated_at": "2023-01-15T00:00:00Z"
    }
  ],
  "total_count": 1
}
```

#### 2. 获取组织树
```http
GET /api/v1/corehr/organizations/tree
```

**响应示例：**
```json
{
  "tree": [
    {
      "id": "550e8400-e29b-41d4-a716-446655440003",
      "name": "技术部",
      "code": "TECH",
      "level": 1,
      "children": [
        {
          "id": "550e8400-e29b-41d4-a716-446655440004",
          "name": "前端组",
          "code": "FRONTEND",
          "level": 2
        }
      ]
    }
  ]
}
```

## 错误处理

### 错误响应格式
```json
{
  "error": {
    "message": "错误描述",
    "status": 400
  },
  "timestamp": "2023-01-15T10:30:00Z"
}
```

### 常见错误状态码

- `400 Bad Request`: 请求参数错误或验证失败
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突（如员工编号已存在）
- `500 Internal Server Error`: 服务器内部错误

### 常见错误消息

- `"employee number is required"`: 员工编号不能为空
- `"invalid email format"`: 邮箱格式不正确
- `"employee not found"`: 员工不存在
- `"employee number EMP001 already exists"`: 员工编号已存在
- `"validation failed"`: 参数验证失败

## 测试

### 使用测试页面

1. 启动服务器：
```bash
cd go-app
go run cmd/server/main.go
```

2. 打开测试页面：
```
http://localhost:8080/test.html
```

3. 在测试页面中可以：
   - 测试所有 API 端点
   - 查看请求和响应
   - 验证错误处理

### 使用 curl 测试

#### 获取员工列表
```bash
curl -X GET "http://localhost:8080/api/v1/corehr/employees?page=1&page_size=10"
```

#### 创建员工
```bash
curl -X POST "http://localhost:8080/api/v1/corehr/employees" \
  -H "Content-Type: application/json" \
  -d '{
    "employee_number": "EMP001",
    "first_name": "张",
    "last_name": "三",
    "email": "zhangsan@example.com",
    "hire_date": "2023-01-15"
  }'
```

#### 获取员工详情
```bash
curl -X GET "http://localhost:8080/api/v1/corehr/employees/550e8400-e29b-41d4-a716-446655440000"
```

#### 更新员工
```bash
curl -X PUT "http://localhost:8080/api/v1/corehr/employees/550e8400-e29b-41d4-a716-446655440000" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "新姓名",
    "email": "newemail@example.com"
  }'
```

#### 删除员工
```bash
curl -X DELETE "http://localhost:8080/api/v1/corehr/employees/550e8400-e29b-41d4-a716-446655440000"
```

## 数据库结构

### 员工表 (corehr.employees)
```sql
CREATE TABLE corehr.employees (
    id UUID PRIMARY KEY,
    employee_number VARCHAR(50) UNIQUE NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone_number VARCHAR(20),
    hire_date DATE NOT NULL,
    position_id UUID,
    organization_id UUID,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### 组织表 (corehr.organizations)
```sql
CREATE TABLE corehr.organizations (
    id UUID PRIMARY KEY,
    tenant_id UUID NOT NULL,
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL,
    parent_id UUID,
    level INTEGER DEFAULT 1,
    status VARCHAR(20) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

## 技术实现

### 架构模式
- **分层架构**: Controller -> Service -> Repository
- **依赖注入**: 通过构造函数注入依赖
- **错误处理**: 统一的错误处理机制

### 主要组件

1. **Service 层** (`internal/corehr/service.go`)
   - 业务逻辑处理
   - 参数验证
   - 数据转换

2. **Repository 层** (`internal/corehr/repository.go`)
   - 数据库操作
   - SQL 查询
   - 数据映射

3. **HTTP 处理器** (`cmd/server/main.go`)
   - 路由处理
   - 请求解析
   - 响应格式化

### 数据验证

- 使用字符串匹配进行邮箱格式验证
- 员工编号唯一性检查
- 状态值枚举验证
- 必填字段检查

### 错误处理策略

- 区分业务错误和系统错误
- 不暴露内部错误信息给客户端
- 提供有意义的错误消息
- 使用适当的 HTTP 状态码

## 后续改进计划

### 🔄 待实现功能

1. **高级搜索**
   - 按部门搜索
   - 按职位搜索
   - 按入职日期范围搜索

2. **组织管理**
   - 创建组织
   - 更新组织
   - 删除组织

3. **职位管理**
   - 职位 CRUD 操作
   - 职位层级管理

4. **权限控制**
   - 基于角色的访问控制
   - API 认证和授权

5. **数据导入导出**
   - Excel 导入员工数据
   - 员工数据导出

6. **审计日志**
   - 操作记录
   - 变更历史

### 🛠️ 技术改进

1. **数据验证**
   - 使用 validator 库进行更严格的验证
   - 自定义验证规则

2. **缓存**
   - Redis 缓存热门数据
   - 查询结果缓存

3. **性能优化**
   - 数据库索引优化
   - 查询性能调优

4. **监控和日志**
   - 结构化日志
   - 性能监控
   - 错误追踪

## 贡献指南

1. Fork 项目
2. 创建功能分支
3. 提交更改
4. 推送到分支
5. 创建 Pull Request

## 许可证

本项目采用 MIT 许可证。 