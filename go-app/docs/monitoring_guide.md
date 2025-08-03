# CQRS员工管理系统监控配置

## 监控指标定义

### 核心业务指标

#### 命令端指标
```yaml
employee_commands_total:
  type: counter
  description: "员工命令执行总数"
  labels: [command_type, status, tenant_id]

employee_command_duration_seconds:
  type: histogram
  description: "员工命令执行时间"
  labels: [command_type, tenant_id]
  buckets: [0.1, 0.5, 1.0, 2.0, 5.0]

employee_command_errors_total:
  type: counter
  description: "员工命令错误总数"
  labels: [command_type, error_code, tenant_id]
```

#### 查询端指标
```yaml
employee_queries_total:
  type: counter
  description: "员工查询执行总数"
  labels: [query_type, source, tenant_id]

employee_query_duration_seconds:
  type: histogram
  description: "员工查询执行时间"
  labels: [query_type, source, tenant_id]
  buckets: [0.05, 0.1, 0.2, 0.5, 1.0]

employee_query_cache_hits_total:
  type: counter
  description: "查询缓存命中总数"
  labels: [query_type, tenant_id]
```

#### 事件指标
```yaml
employee_events_published_total:
  type: counter
  description: "员工事件发布总数"
  labels: [event_type, status, tenant_id]

employee_events_consumed_total:
  type: counter
  description: "员工事件消费总数"
  labels: [event_type, status, consumer_id]

employee_event_processing_duration_seconds:
  type: histogram
  description: "事件处理时间"
  labels: [event_type, consumer_id]
  buckets: [0.01, 0.05, 0.1, 0.5, 1.0]
```

### 基础设施指标

#### PostgreSQL指标
```yaml
postgres_connections_active:
  type: gauge
  description: "PostgreSQL活跃连接数"

postgres_query_duration_seconds:
  type: histogram
  description: "PostgreSQL查询执行时间"
  buckets: [0.01, 0.05, 0.1, 0.5, 1.0, 2.0]

postgres_errors_total:
  type: counter
  description: "PostgreSQL错误总数"
  labels: [error_type]
```

#### Neo4j指标
```yaml
neo4j_connections_active:
  type: gauge
  description: "Neo4j活跃连接数"

neo4j_query_duration_seconds:
  type: histogram
  description: "Neo4j查询执行时间"
  buckets: [0.01, 0.05, 0.1, 0.5, 1.0]

neo4j_errors_total:
  type: counter
  description: "Neo4j错误总数"
  labels: [error_type]
```

## 告警规则

### 高优先级告警

#### 系统可用性
```yaml
- alert: CQRSSystemDown
  expr: up{job="cube-castle-api"} == 0
  for: 1m
  severity: critical
  summary: "CQRS员工管理系统不可用"

- alert: HighErrorRate
  expr: rate(employee_commands_total{status="error"}[5m]) > 0.1
  for: 2m
  severity: critical
  summary: "员工命令错误率过高"

- alert: DatabaseConnectionLoss
  expr: postgres_connections_active == 0 or neo4j_connections_active == 0
  for: 30s
  severity: critical
  summary: "数据库连接丢失"
```

#### 性能告警
```yaml
- alert: SlowCommandExecution
  expr: histogram_quantile(0.95, employee_command_duration_seconds) > 2
  for: 5m
  severity: warning
  summary: "员工命令执行缓慢"

- alert: SlowQueryExecution
  expr: histogram_quantile(0.95, employee_query_duration_seconds) > 1
  for: 5m
  severity: warning
  summary: "员工查询执行缓慢"

- alert: EventProcessingDelay
  expr: histogram_quantile(0.95, employee_event_processing_duration_seconds) > 1
  for: 3m
  severity: warning
  summary: "事件处理延迟过高"
```

### 中优先级告警

#### 数据一致性
```yaml
- alert: DataInconsistency
  expr: abs(postgres_employee_count - neo4j_employee_count) > 10
  for: 10m
  severity: warning
  summary: "PostgreSQL和Neo4j员工数据不一致"

- alert: EventConsumptionLag
  expr: employee_events_published_total - employee_events_consumed_total > 100
  for: 5m
  severity: warning
  summary: "事件消费滞后"
```

#### 资源使用
```yaml
- alert: HighMemoryUsage
  expr: process_resident_memory_bytes / 1024 / 1024 > 1024
  for: 5m
  severity: warning
  summary: "内存使用过高"

- alert: HighCPUUsage
  expr: rate(process_cpu_seconds_total[5m]) > 0.8
  for: 5m
  severity: warning
  summary: "CPU使用率过高"
```

## 仪表板配置

### 业务概览仪表板

#### 核心KPI面板
```json
{
  "title": "员工管理系统 - 业务概览",
  "panels": [
    {
      "title": "每日员工操作量",
      "type": "stat",
      "targets": [
        {
          "expr": "increase(employee_commands_total[24h])",
          "legendFormat": "命令执行"
        },
        {
          "expr": "increase(employee_queries_total[24h])",
          "legendFormat": "查询执行"
        }
      ]
    },
    {
      "title": "系统健康状态",
      "type": "stat",
      "targets": [
        {
          "expr": "up{job=\"cube-castle-api\"}",
          "legendFormat": "系统可用性"
        }
      ]
    }
  ]
}
```

#### 性能趋势面板
```json
{
  "title": "性能趋势",
  "type": "graph",
  "targets": [
    {
      "expr": "histogram_quantile(0.95, employee_command_duration_seconds)",
      "legendFormat": "命令响应时间 P95"
    },
    {
      "expr": "histogram_quantile(0.95, employee_query_duration_seconds)",
      "legendFormat": "查询响应时间 P95"
    }
  ]
}
```

### 技术运维仪表板

#### 数据库性能面板
```json
{
  "title": "数据库性能监控",
  "panels": [
    {
      "title": "PostgreSQL连接数",
      "type": "graph",
      "targets": [
        {
          "expr": "postgres_connections_active",
          "legendFormat": "活跃连接"
        }
      ]
    },
    {
      "title": "Neo4j查询性能",
      "type": "graph",
      "targets": [
        {
          "expr": "histogram_quantile(0.95, neo4j_query_duration_seconds)",
          "legendFormat": "查询时间 P95"
        }
      ]
    }
  ]
}
```

#### 事件流监控面板
```json
{
  "title": "事件流监控",
  "panels": [
    {
      "title": "事件发布速率",
      "type": "graph",
      "targets": [
        {
          "expr": "rate(employee_events_published_total[5m])",
          "legendFormat": "{{event_type}}"
        }
      ]
    },
    {
      "title": "事件消费延迟",
      "type": "graph",
      "targets": [
        {
          "expr": "employee_events_published_total - employee_events_consumed_total",
          "legendFormat": "待处理事件数"
        }
      ]
    }
  ]
}
```

## 日志配置

### 结构化日志格式
```json
{
  "timestamp": "2024-03-20T10:30:00Z",
  "level": "INFO",
  "service": "cube-castle-api",
  "component": "cqrs",
  "operation": "employee_command",
  "command_type": "hire_employee",
  "employee_id": "uuid",
  "tenant_id": "uuid",
  "duration_ms": 150,
  "status": "success",
  "trace_id": "abc123",
  "message": "Employee hired successfully"
}
```

### 日志级别配置
```yaml
logging:
  level: info
  format: json
  outputs:
    - type: file
      path: /var/log/cube-castle/app.log
      rotation:
        max_size: 100MB
        max_age: 30d
        max_backups: 10
    - type: elasticsearch
      url: http://elasticsearch:9200
      index: cube-castle-logs
```

## 监控最佳实践

### 1. 监控策略
- **Golden Signals**: 延迟、流量、错误、饱和度
- **USE Method**: 利用率、饱和度、错误
- **RED Method**: 速率、错误、持续时间

### 2. 告警策略
- **分层告警**: 业务 → 应用 → 基础设施
- **告警抑制**: 避免告警风暴
- **智能路由**: 按严重程度路由到不同团队

### 3. 仪表板设计
- **角色导向**: 不同角色看不同仪表板
- **层次结构**: 概览 → 详细 → 诊断
- **实时更新**: 重要指标实时刷新

### 4. 性能基准
```yaml
SLA目标:
  命令响应时间: P95 < 2s, P99 < 5s
  查询响应时间: P95 < 500ms, P99 < 1s
  系统可用性: > 99.9%
  事件处理延迟: P95 < 100ms

性能基准:
  每秒处理命令: > 100 TPS
  每秒处理查询: > 1000 QPS
  并发用户数: > 1000
  数据同步延迟: < 1s
```

## 故障排查手册

### 常见问题诊断

#### 命令执行缓慢
1. 检查PostgreSQL连接池状态
2. 分析慢查询日志
3. 检查事务锁等待
4. 验证索引使用情况

#### 查询性能下降
1. 检查Neo4j连接状态
2. 分析查询执行计划
3. 检查缓存命中率
4. 验证索引策略

#### 数据不一致
1. 检查事件发布状态
2. 验证事件消费进度
3. 对比数据库记录数
4. 检查同步日志

### 应急响应流程
1. **立即响应** (< 5分钟)
2. **影响评估** (< 15分钟)
3. **临时修复** (< 30分钟)
4. **根本原因分析** (< 2小时)
5. **永久修复** (< 24小时)
6. **事后回顾** (< 1周)