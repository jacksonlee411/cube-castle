# 组织单元API监控配置

## 📊 Prometheus监控配置

### prometheus.yml
```yaml
global:
  scrape_interval: 15s
  evaluation_interval: 15s

rule_files:
  - "organization_api_rules.yml"

scrape_configs:
  - job_name: 'organization-units-api'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 10s
    scrape_timeout: 5s

alerting:
  alertmanagers:
    - static_configs:
        - targets:
          - alertmanager:9093
```

### Grafana仪表板配置
```json
{
  "dashboard": {
    "title": "组织单元API v2.0 监控",
    "panels": [
      {
        "title": "API响应时间",
        "type": "graph",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job=\"organization-units-api\"}[5m]))",
            "legendFormat": "P95响应时间"
          }
        ]
      },
      {
        "title": "QPS (每秒请求数)",
        "type": "graph", 
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"organization-units-api\"}[5m])",
            "legendFormat": "{{method}} {{endpoint}}"
          }
        ]
      },
      {
        "title": "错误率",
        "type": "singlestat",
        "targets": [
          {
            "expr": "rate(http_requests_total{job=\"organization-units-api\",status=~\"5..\"}[5m]) / rate(http_requests_total{job=\"organization-units-api\"}[5m]) * 100",
            "legendFormat": "错误率 %"
          }
        ]
      }
    ]
  }
}
```

## 🚨 告警规则

### organization_api_rules.yml
```yaml
groups:
  - name: organization-api-alerts
    rules:
      - alert: HighResponseTime
        expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket{job="organization-units-api"}[5m])) > 0.1
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "API响应时间过高"
          description: "P95响应时间超过100ms，当前值: {{ $value }}s"

      - alert: HighErrorRate
        expr: rate(http_requests_total{job="organization-units-api",status=~"5.."}[5m]) / rate(http_requests_total{job="organization-units-api"}[5m]) > 0.05
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "API错误率过高"
          description: "错误率超过5%，当前值: {{ $value | humanizePercentage }}"

      - alert: ServiceDown
        expr: up{job="organization-units-api"} == 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "API服务不可用"
          description: "组织单元API服务已停止响应"

      - alert: DatabaseConnectionFailure
        expr: database_connections_failed_total{job="organization-units-api"} > 0
        for: 1m
        labels:
          severity: critical
        annotations:
          summary: "数据库连接失败"
          description: "数据库连接失败次数: {{ $value }}"
```

## 📈 指标收集

### 关键指标
- **http_requests_total**: HTTP请求总数
- **http_request_duration_seconds**: HTTP请求延迟
- **database_query_duration_seconds**: 数据库查询时间
- **database_connections_active**: 活跃数据库连接数
- **organization_units_total**: 组织单元总数
- **api_cache_hits_total**: 缓存命中次数

### 业务指标
- **organization_queries_by_code**: 按编码查询次数
- **organization_list_queries**: 列表查询次数  
- **organization_stats_queries**: 统计查询次数
- **invalid_code_requests**: 无效编码请求次数

## 🔍 日志监控

### 日志格式
```json
{
  "timestamp": "2025-08-05T20:10:01+08:00",
  "level": "info",
  "method": "GET",
  "path": "/api/v1/organization-units/1000000",
  "status": 200,
  "duration": "2.5ms",
  "ip": "192.168.1.100",
  "user_agent": "React/18.0",
  "organization_code": "1000000",
  "tenant_id": "3b99930c-4dc6-4cc9-8e4d-7d960a931cb9"
}
```

### 日志分析查询
```bash
# 查看最近的API调用
tail -f /var/log/organization-units-api/access.log | jq .

# 统计各端点调用次数
cat access.log | jq -r '.path' | sort | uniq -c | sort -nr

# 查找慢查询
cat access.log | jq 'select(.duration > "50ms")'

# 分析错误请求
cat access.log | jq 'select(.status >= 400)'
```

## 📱 健康检查

### 应用健康检查
```bash
#!/bin/bash
# health_check.sh

API_URL="http://localhost:8080"
HEALTH_ENDPOINT="$API_URL/health"

# 检查API健康状态
response=$(curl -s -w "%{http_code}" -o /tmp/health_response $HEALTH_ENDPOINT)
http_code="${response: -3}"

if [ "$http_code" -eq 200 ]; then
    version=$(cat /tmp/health_response | jq -r '.version')
    timestamp=$(cat /tmp/health_response | jq -r '.timestamp')
    echo "✅ API健康检查通过 - 版本: $version, 时间: $timestamp"
    exit 0
else
    echo "❌ API健康检查失败 - HTTP状态码: $http_code"
    exit 1
fi
```

### 数据库健康检查
```bash
#!/bin/bash
# db_health_check.sh

DB_HOST="localhost"
DB_PORT="5432"
DB_NAME="cubecastle"
DB_USER="user"

# 检查数据库连接
if PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT 1;" > /dev/null 2>&1; then
    # 检查组织单元表
    count=$(PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c "SELECT COUNT(*) FROM organization_units;")
    echo "✅ 数据库健康检查通过 - 组织单元数量: $count"
    exit 0
else
    echo "❌ 数据库连接失败"
    exit 1
fi
```

## 🎯 SLA监控

### 服务水平目标
- **可用性**: 99.9% (每月停机时间 < 43分钟)
- **响应时间**: P95 < 100ms
- **错误率**: < 0.1%
- **吞吐量**: > 1000 RPS

### SLA仪表板
```yaml
# sla_dashboard.yml
dashboard:
  title: "SLA监控仪表板"
  time_range: "7d"
  
  panels:
    - title: "可用性 (7天)"
      query: "avg_over_time(up{job='organization-units-api'}[7d]) * 100"
      target: 99.9
      
    - title: "P95响应时间 (24小时)"
      query: "histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[24h]))"
      target: 0.1
      
    - title: "错误率 (24小时)"
      query: "rate(http_requests_total{status=~'5..'}[24h]) / rate(http_requests_total[24h]) * 100"
      target: 0.1
```

## 🔧 监控部署脚本

### setup_monitoring.sh
```bash
#!/bin/bash

echo "🔧 配置监控系统..."

# 创建监控目录
mkdir -p monitoring/{prometheus,grafana,alertmanager}

# 部署Prometheus
docker run -d \
  --name prometheus \
  -p 9090:9090 \
  -v $(pwd)/monitoring/prometheus:/etc/prometheus \
  prom/prometheus:latest

# 部署Grafana
docker run -d \
  --name grafana \
  -p 3000:3000 \
  -e "GF_SECURITY_ADMIN_PASSWORD=admin123" \
  grafana/grafana:latest

# 部署Alertmanager
docker run -d \
  --name alertmanager \
  -p 9093:9093 \
  -v $(pwd)/monitoring/alertmanager:/etc/alertmanager \
  prom/alertmanager:latest

echo "✅ 监控系统部署完成"
echo "📊 Prometheus: http://localhost:9090"
echo "📈 Grafana: http://localhost:3000 (admin/admin123)"
echo "🚨 Alertmanager: http://localhost:9093"
```

---

> 📊 **监控系统配置完成！**  
> 提供全方位的API性能监控和告警机制