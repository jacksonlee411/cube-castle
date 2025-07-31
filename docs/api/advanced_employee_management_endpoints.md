# Advanced Employee Management API Endpoints | 高级员工管理API接口

**Last Updated**: 2025-07-31 15:00:00  
**Version**: v1.5.0  
**Target**: Week 3 Advanced Features Implementation

## 📋 Overview | 概述

This document describes the advanced employee management endpoints implemented in Week 3, including position assignment, employee lifecycle management, and comprehensive analytics capabilities.

本文档描述了第三周实现的高级员工管理接口，包括岗位分配、员工生命周期管理和综合分析功能。

## 🎯 Core Endpoints | 核心接口

### 1. Position Assignment Management | 岗位分配管理

Base URL: `/api/v1/assignments`

#### POST /assignments - Assign Position | 分配岗位
```http
POST /api/v1/assignments
Content-Type: application/json
Authorization: Bearer {token}

{
  "employee_id": "uuid",
  "position_id": "uuid", 
  "start_date": "2025-07-31T09:00:00Z",
  "end_date": "2025-12-31T17:00:00Z",
  "assignment_type": "REGULAR|INTERIM|ACTING|TEMPORARY|SECONDMENT",
  "assignment_reason": "Initial assignment for new hire",
  "fte_percentage": 1.0,
  "approved_by": "uuid",
  "work_arrangement": "ON_SITE|REMOTE|HYBRID"
}
```

**Response 200 OK**:
```json
{
  "success": true,
  "assignment_id": "uuid",
  "employee_id": "uuid",
  "position_id": "uuid", 
  "start_date": "2025-07-31T09:00:00Z",
  "assignment_type": "REGULAR",
  "previous_assignment_id": "uuid",
  "message": "Position assigned successfully"
}
```

#### POST /assignments/transfer - Transfer Employee | 员工调转
```http
POST /api/v1/assignments/transfer
Content-Type: application/json

{
  "employee_id": "uuid",
  "from_position_id": "uuid",
  "to_position_id": "uuid",
  "transfer_date": "2025-08-01T09:00:00Z",
  "transfer_reason": "Promotion to senior role",
  "approved_by": "uuid",
  "fte_percentage": 1.0,
  "work_arrangement": "HYBRID"
}
```

#### DELETE /assignments/{employeeId} - End Assignment | 结束分配
```http
DELETE /api/v1/assignments/{employeeId}
Content-Type: application/json

{
  "end_date": "2025-07-31T17:00:00Z",
  "reason": "Employee resignation"
}
```

#### GET /assignments/active - Get Active Assignments | 获取活跃分配
```http
GET /api/v1/assignments/active
```

**Response 200 OK**:
```json
{
  "assignments": [
    {
      "id": "uuid",
      "employee_id": "uuid",
      "position_id": "uuid",
      "start_date": "2025-07-01T09:00:00Z",
      "is_active": true,
      "assignment_type": "REGULAR",
      "fte_percentage": 1.0,
      "work_arrangement": "ON_SITE"
    }
  ],
  "total": 1
}
```

### 2. Employee Lifecycle Management | 员工生命周期管理

Base URL: `/api/v1/lifecycle`

#### POST /lifecycle/onboard - Employee Onboarding | 员工入职
```http
POST /api/v1/lifecycle/onboard
Content-Type: application/json

{
  "employee_type": "FULL_TIME|PART_TIME|CONTRACTOR|INTERN",
  "employee_number": "EMP001",
  "first_name": "张",
  "last_name": "三",
  "email": "zhang.san@company.com",
  "personal_email": "zhang.san@gmail.com",
  "phone_number": "+86 138 0000 0000",
  "hire_date": "2025-08-01T00:00:00Z",
  "employee_details": {
    "department": "Engineering",
    "level": "Senior"
  },
  "initial_position_id": "uuid",
  "assignment_start_date": "2025-08-01T09:00:00Z",
  "work_arrangement": "HYBRID",
  "fte_percentage": 1.0,
  "onboarding_manager": "uuid",
  "onboarding_notes": "New senior engineer joining the team",
  "probation_period_days": 90
}
```

**Response 201 Created**:
```json
{
  "success": true,
  "employee_id": "uuid",
  "employee_number": "EMP001",
  "initial_assignment_id": "uuid",
  "onboarding_event_id": "uuid",
  "message": "Employee onboarded successfully"
}
```

#### POST /lifecycle/offboard - Employee Offboarding | 员工离职
```http
POST /api/v1/lifecycle/offboard
Content-Type: application/json

{
  "employee_id": "uuid",
  "termination_date": "2025-08-31T00:00:00Z",
  "termination_reason": "Resignation for personal reasons",
  "termination_type": "VOLUNTARY|INVOLUNTARY|RETIREMENT|END_OF_CONTRACT",
  "last_working_date": "2025-08-30T17:00:00Z",
  "exit_interview_date": "2025-08-29T14:00:00Z",
  "final_pay_date": "2025-09-15T00:00:00Z",
  "offboarding_manager": "uuid",
  "notes": "Exit process completed successfully"
}
```

#### POST /lifecycle/promote - Employee Promotion | 员工晋升
```http
POST /api/v1/lifecycle/promote
Content-Type: application/json

{
  "employee_id": "uuid",
  "new_position_id": "uuid",
  "promotion_date": "2025-09-01T00:00:00Z",
  "promotion_reason": "Outstanding performance and leadership skills",
  "salary_adjustment": 15000.00,
  "approved_by": "uuid",
  "effective_date": "2025-09-01T09:00:00Z",
  "work_arrangement": "HYBRID",
  "fte_percentage": 1.0
}
```

#### POST /lifecycle/status-change - Change Employment Status | 变更雇佣状态
```http
POST /api/v1/lifecycle/status-change
Content-Type: application/json

{
  "employee_id": "uuid",
  "new_status": "ACTIVE|ON_LEAVE|TERMINATED|SUSPENDED|PENDING_START",
  "effective_date": "2025-08-01T00:00:00Z",
  "reason": "Medical leave of absence",
  "expected_return_date": "2025-10-01T00:00:00Z",
  "approved_by": "uuid",
  "notes": "Approved for 2-month medical leave"
}
```

### 3. Analytics and Reporting | 分析和报表

Base URL: `/api/v1/analytics`

#### GET /analytics/metrics - Organizational Metrics | 组织指标
```http
GET /api/v1/analytics/metrics
```

**Response 200 OK**:
```json
{
  "tenant_id": "uuid",
  "report_date": "2025-07-31T15:00:00Z",
  "total_employees": 150,
  "active_employees": 145,
  "total_positions": 160,
  "filled_positions": 145,
  "open_positions": 15,
  "employees_by_type": {
    "FULL_TIME": 120,
    "PART_TIME": 15,
    "CONTRACTOR": 10,
    "INTERN": 5
  },
  "employees_by_status": {
    "ACTIVE": 145,
    "ON_LEAVE": 3,
    "PENDING_START": 2
  },
  "positions_by_status": {
    "FILLED": 145,
    "OPEN": 15
  },
  "average_assignment_duration_days": 365.5,
  "turnover_metrics": {
    "terminations_this_month": 2,
    "terminations_this_quarter": 8,
    "terminations_this_year": 25,
    "hires_this_month": 5,
    "hires_this_quarter": 18,
    "hires_this_year": 35,
    "monthly_turnover_rate": 1.33,
    "quarterly_turnover_rate": 5.33,
    "annual_turnover_rate": 16.67
  },
  "assignment_metrics": {
    "total_assignments": 200,
    "active_assignments": 145,
    "assignments_by_type": {
      "REGULAR": 180,
      "INTERIM": 10,
      "TEMPORARY": 10
    },
    "average_assignment_length_days": 400.5,
    "promotions_this_year": 15,
    "transfers_this_year": 8,
    "assignment_trends": []
  }
}
```

#### GET /analytics/employees/{id}/history - Employee History | 员工历史
```http
GET /api/v1/analytics/employees/{employee_id}/history
```

**Response 200 OK**:
```json
{
  "employee": {
    "id": "uuid",
    "employee_number": "EMP001",
    "first_name": "张",
    "last_name": "三",
    "employment_status": "ACTIVE"
  },
  "assignment_history": [
    {
      "assignment_id": "uuid",
      "position_id": "uuid",
      "position_type": "INDIVIDUAL_CONTRIBUTOR",
      "department_id": "uuid",
      "start_date": "2025-01-15T09:00:00Z",
      "end_date": "2025-07-31T17:00:00Z",
      "duration_days": 197,
      "is_active": false,
      "assignment_type": "REGULAR",
      "fte_percentage": 1.0,
      "work_arrangement": "ON_SITE" 
    }
  ],
  "total_assignments": 2,
  "current_assignment": {
    "assignment_id": "uuid",
    "position_id": "uuid",
    "start_date": "2025-08-01T09:00:00Z",
    "is_active": true
  },
  "total_tenure_days": 197,
  "average_assignment_days": 197.0
}
```

#### GET /analytics/positions/{id}/history - Position History | 岗位历史
```http
GET /api/v1/analytics/positions/{position_id}/history
```

#### GET /analytics/assignments/history - Historical Assignments | 历史分配
```http
GET /api/v1/analytics/assignments/history?start_date=2025-01-01&end_date=2025-07-31&limit=50&offset=0
```

**Query Parameters**:
- `start_date`: Filter assignments starting after this date
- `end_date`: Filter assignments starting before this date  
- `employee_id`: Query specific employee's assignments
- `position_id`: Query specific position's assignments
- `limit`: Maximum results (default: 100)
- `offset`: Pagination offset (default: 0)

## 🔒 Authentication & Authorization | 认证与授权

All endpoints require:
所有接口都需要：

- **Bearer Token Authentication** | Bearer Token认证
- **Tenant Context** | 租户上下文 - `X-Tenant-ID` header
- **Role-Based Access** | 基于角色的访问控制:
  - `HR_MANAGER`: Full access to all endpoints
  - `MANAGER`: Access to team member operations  
  - `EMPLOYEE`: Read-only access to own data

## 📊 Error Handling | 错误处理

### Standard Error Response | 标准错误响应
```json
{
  "error": "Validation failed",
  "details": [
    {
      "field": "employee_id",
      "message": "Employee not found"
    }
  ],
  "code": "VALIDATION_ERROR",
  "timestamp": "2025-07-31T15:00:00Z"
}
```

### Common Error Codes | 常见错误码
- `400 Bad Request`: Invalid request payload | 无效的请求数据
- `401 Unauthorized`: Missing or invalid authentication | 缺少或无效的认证信息
- `403 Forbidden`: Insufficient permissions | 权限不足
- `404 Not Found`: Resource not found | 资源未找到
- `409 Conflict`: Business rule violation | 业务规则冲突
- `503 Service Unavailable`: Database connection issues | 数据库连接问题

## 🔄 Business Rules | 业务规则

### Position Assignment Rules | 岗位分配规则
1. **Employee Status Check** | 员工状态检查: Only ACTIVE employees can be assigned
2. **Conflict Resolution** | 冲突解决: Existing assignments automatically ended
3. **Position Capacity** | 岗位容量: Positions can have multiple assignments based on type
4. **Transaction Safety** | 事务安全: All operations are atomic

### Lifecycle Management Rules | 生命周期管理规则
1. **Onboarding Process** | 入职流程: Creates employee + optional position assignment
2. **Offboarding Cleanup** | 离职清理: Ends all active assignments and updates statuses
3. **Status Transitions** | 状态转换: Validates legal status changes
4. **Promotion Logic** | 晋升逻辑: Handled as position transfers with event tracking

## 📈 Performance Considerations | 性能考虑

- **Pagination** | 分页: Large result sets automatically paginated
- **Database Transactions** | 数据库事务: All complex operations use transactions  
- **Caching Strategy** | 缓存策略: Analytics results cached for 15 minutes
- **Index Optimization** | 索引优化: Query performance optimized for common patterns

## 🔗 Related Documentation | 相关文档

- [Employee Model Design](../architecture/employee_model_design.md)
- [Database Schema Changes](../architecture/database_schema_week3.md)
- [Business Logic Implementation](../architecture/advanced_features_design.md)
- [Week 3 Implementation Report](../reports/week3_implementation_report.md)

---

**Next Review**: 2025-08-31 15:00:00