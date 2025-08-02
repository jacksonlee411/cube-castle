# 工作流系统API使用指南

## 概述

本文档描述了工作流系统的REST API接口，包括工作流管理、定义管理和事件查询功能。

## 基础信息

**Base URL**: `http://localhost:8080/api/v1`
**Content-Type**: `application/json`
**认证**: Bearer Token (待实现)

## API响应格式

所有API响应都遵循统一格式：

```json
{
  "success": true,
  "data": {
    // 实际响应数据
  },
  "error": "错误信息 (仅在success=false时)",
  "message": "操作消息 (可选)"
}
```

## 工作流管理API

### 1. 启动工作流

**POST** `/workflows`

启动一个新的工作流实例。

**请求体**:
```json
{
  "tenant_id": "uuid",
  "workflow_type": "EmployeeOnboarding",
  "initiated_by": "uuid", 
  "context": {
    "employee_id": "EMP_001",
    "employee_name": "张三",
    "department": "技术部"
  },
  "correlation_id": "CORR_001"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "workflow_instance": {
      "id": "uuid",
      "tenant_id": "uuid",
      "workflow_type": "EmployeeOnboarding",
      "current_state": "BACKGROUND_CHECK",
      "initiated_by": "uuid",
      "correlation_id": "CORR_001",
      "started_at": "2025-07-30T04:30:00Z"
    },
    "message": "工作流启动成功"
  }
}
```

### 2. 查询工作流实例

**GET** `/workflows?tenant_id={tenant_id}&workflow_type={type}&state={state}&limit={limit}&offset={offset}`

查询工作流实例列表。

**查询参数**:
- `tenant_id` (必需): 租户ID
- `workflow_type` (可选): 工作流类型过滤
- `state` (可选): 状态过滤
- `limit` (可选): 限制数量，默认20，最大100
- `offset` (可选): 偏移量，默认0

**响应**:
```json
{
  "success": true,
  "data": {
    "instances": [
      {
        "id": "uuid",
        "workflow_type": "EmployeeOnboarding",
        "current_state": "COMPLETED",
        "steps": [...]
      }
    ],
    "total": 25,
    "limit": 20,
    "offset": 0
  }
}
```

### 3. 获取工作流实例详情

**GET** `/workflows/{id}?tenant_id={tenant_id}`

获取特定工作流实例的详细信息。

**响应**:
```json
{
  "success": true,
  "data": {
    "workflow_instance": {
      "id": "uuid",
      "workflow_type": "EmployeeOnboarding",
      "current_state": "IN_PROGRESS", 
      "state_history": [
        {
          "state": "STARTED",
          "timestamp": "2025-07-30T04:30:00Z",
          "triggered_by": "uuid"
        }
      ],
      "context": {...},
      "steps": [...]
    }
  }
}
```

### 4. 添加工作流步骤

**POST** `/workflows/{id}/steps`

为工作流实例添加新的步骤。

**请求体**:
```json
{
  "step_name": "额外验证步骤",
  "step_type": "MANUAL",
  "assigned_to": "uuid",
  "input_data": {
    "priority": "high",
    "description": "需要人工验证"
  },
  "due_date": "2025-08-01T12:00:00Z"
}
```

### 5. 完成工作流步骤

**POST** `/workflows/steps/{stepId}/complete`

标记工作流步骤为完成状态。

**请求体**:
```json
{
  "output_data": {
    "result": "completed",
    "notes": "验证通过",
    "verified_by": "系统"
  },
  "completed_by": "uuid"
}
```

### 6. 跳过工作流步骤

**POST** `/workflows/steps/{stepId}/skip`

跳过某个工作流步骤。

**请求体**:
```json
{
  "reason": "业务需要跳过此步骤",
  "skipped_by": "uuid"
}
```

### 7. 获取待处理步骤

**GET** `/workflows/steps/pending?tenant_id={tenant_id}&assigned_to={user_id}&limit={limit}`

获取当前用户或租户的待处理步骤列表。

**响应**:
```json
{
  "success": true,
  "data": {
    "pending_steps": [
      {
        "id": "uuid",
        "step_name": "BACKGROUND_CHECK",
        "step_type": "MANUAL",
        "workflow_instance_id": "uuid",
        "assigned_to": "uuid",
        "due_date": "2025-08-02T12:00:00Z",
        "input_data": {...}
      }
    ],
    "count": 5
  }
}
```

## 工作流定义API

### 1. 列出工作流定义

**GET** `/workflow-definitions`

获取所有可用的工作流定义。

**响应**:
```json
{
  "success": true,
  "data": {
    "definitions": [
      {
        "name": "EmployeeOnboarding",
        "description": "员工入职工作流",
        "version": "1.0",
        "states": [...],
        "transitions": {...}
      }
    ],
    "count": 2
  }
}
```

### 2. 获取工作流定义

**GET** `/workflow-definitions/{name}`

获取特定工作流定义的详细信息。

**响应**:
```json
{
  "success": true,
  "data": {
    "definition": {
      "name": "EmployeeOnboarding",
      "description": "员工入职工作流",
      "version": "1.0",
      "states": [
        {
          "name": "BACKGROUND_CHECK",
          "type": "MANUAL",
          "timeout": "72h",
          "metadata": {
            "priority": "high",
            "description": "进行背景调查"
          },
          "required": ["employee_id", "documents"]
        }
      ],
      "transitions": {
        "BACKGROUND_CHECK": [
          {
            "to_state": "DOCUMENTATION",
            "condition": "background_check_passed"
          }
        ]
      },
      "variables": {
        "employee_id": {
          "type": "string",
          "required": true,
          "description": "员工ID"
        }
      }
    }
  }
}
```

### 3. 导出工作流定义

**GET** `/workflow-definitions/{name}/export`

导出工作流定义为JSON文件。

**响应**: 直接下载JSON文件

### 4. 导入工作流定义

**POST** `/workflow-definitions/import`

导入新的工作流定义。

**请求体**: 完整的工作流定义JSON

## 事件查询API

### 1. 查询事件

**GET** `/events?tenant_id={tenant_id}&event_type={type}&entity_type={entity_type}&start_date={date}&end_date={date}&limit={limit}&offset={offset}`

查询业务流程事件。

**查询参数**:
- `tenant_id` (必需): 租户ID
- `event_type` (可选): 事件类型过滤
- `entity_type` (可选): 实体类型过滤
- `entity_id` (可选): 实体ID过滤
- `correlation_id` (可选): 关联ID过滤
- `start_date` (可选): 开始日期 (RFC3339)
- `end_date` (可选): 结束日期 (RFC3339)
- `limit` (可选): 限制数量，默认50，最大200
- `offset` (可选): 偏移量，默认0

### 2. 按关联ID查询事件

**GET** `/events/by-correlation/{correlationId}?tenant_id={tenant_id}`

获取特定关联ID的所有相关事件。

### 3. 按实体查询事件

**GET** `/events/by-entity/{entityType}/{entityId}?tenant_id={tenant_id}`

获取特定实体的所有事件。

### 4. 事件统计

**GET** `/events/statistics?tenant_id={tenant_id}&start_date={date}&end_date={date}`

获取事件统计信息。

**响应**:
```json
{
  "success": true,
  "data": {
    "statistics": {
      "total_events": 1250,
      "events_by_type": {
        "Workflow.Started": 125,
        "WorkflowStep.Completed": 890,
        "WorkflowStep.Skipped": 45
      },
      "events_by_status": {
        "COMPLETED": 1100,
        "PENDING": 50,
        "FAILED": 100
      },
      "events_by_day": {
        "2025-07-29": 450,
        "2025-07-30": 800
      }
    },
    "period": {
      "start_date": "2025-07-29T00:00:00Z",
      "end_date": "2025-07-30T23:59:59Z"
    }
  }
}
```

## 错误处理

### 常见错误代码

- `400 Bad Request`: 请求参数错误或格式不正确
- `401 Unauthorized`: 认证失败
- `403 Forbidden`: 权限不足
- `404 Not Found`: 资源不存在
- `409 Conflict`: 资源冲突 (如重复创建)
- `500 Internal Server Error`: 服务器内部错误

### 错误响应示例

```json
{
  "success": false,
  "error": "工作流定义不存在: NonExistentWorkflow",
  "data": null
}
```

## 使用示例

### JavaScript (Fetch API)

```javascript
// 启动工作流
const startWorkflow = async () => {
  const response = await fetch('/api/v1/workflows', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer your-token'
    },
    body: JSON.stringify({
      tenant_id: 'tenant-uuid',
      workflow_type: 'EmployeeOnboarding',
      initiated_by: 'user-uuid',
      context: {
        employee_id: 'EMP_001',
        employee_name: '张三',
        department: '技术部'
      }
    })
  });
  
  const result = await response.json();
  if (result.success) {
    console.log('工作流启动成功:', result.data.workflow_instance.id);
  }
};

// 查询待处理步骤
const getPendingSteps = async (tenantId, userId) => {
  const response = await fetch(
    `/api/v1/workflows/steps/pending?tenant_id=${tenantId}&assigned_to=${userId}`,
    {
      headers: {
        'Authorization': 'Bearer your-token'
      }
    }
  );
  
  const result = await response.json();
  return result.data.pending_steps;
};
```

### cURL 示例

```bash
# 启动工作流
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "tenant_id": "tenant-uuid",
    "workflow_type": "EmployeeOnboarding", 
    "initiated_by": "user-uuid",
    "context": {
      "employee_id": "EMP_001",
      "employee_name": "张三",
      "department": "技术部"
    }
  }'

# 查询工作流定义
curl -X GET http://localhost:8080/api/v1/workflow-definitions/EmployeeOnboarding \
  -H "Authorization: Bearer your-token"

# 完成工作流步骤
curl -X POST http://localhost:8080/api/v1/workflows/steps/{stepId}/complete \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer your-token" \
  -d '{
    "output_data": {"result": "completed"},
    "completed_by": "user-uuid"
  }'
```

## 最佳实践

1. **错误处理**: 始终检查响应中的 `success` 字段
2. **分页**: 使用 `limit` 和 `offset` 进行分页查询
3. **过滤**: 利用查询参数减少不必要的数据传输
4. **关联ID**: 使用有意义的 `correlation_id` 便于事件追踪
5. **超时**: 设置合适的HTTP请求超时时间
6. **重试**: 对于网络错误实施指数退避重试策略