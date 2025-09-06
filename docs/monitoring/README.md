# Prometheus + Grafana 监控系统使用指南

## 🚀 快速开始

### 1️⃣ **一键启动监控系统**

```bash
# 切换到项目根目录
cd /home/shangmeilin/cube-castle

# 启动监控系统
./scripts/start-monitoring.sh

# 验证系统运行
./scripts/test-monitoring.sh
```

### 2️⃣ **访问监控界面**

| 服务 | 地址 | 用途 |
|------|------|------|
| **Grafana** | http://localhost:3001 | 📊 数据可视化仪表板 |
| **Prometheus** | http://localhost:9090 | 🔍 指标查询和告警 |
| **AlertManager** | http://localhost:9093 | 🚨 告警管理 |

**Grafana登录信息**:
- 用户名: `admin`
- 密码: `cube-castle-2025`

## 📊 **核心监控指标**

### **SLO关键指标 (ADR-008合规)**

```yaml
成功率指标:
  • activate_success_total / activate_requests_total ≥ 99.9%
  • suspend_success_total / suspend_requests_total ≥ 99.9%

延迟指标:  
  • histogram_quantile(0.95, activate_duration_seconds_bucket) ≤ 150ms
  • histogram_quantile(0.95, suspend_duration_seconds_bucket) ≤ 150ms

合规指标:
  • deprecated_endpoint_used_total = 0 (零容忍)
  • audit_write_success_total / audit_write_attempts_total = 100%
```

### **业务监控指标**

```yaml
API请求:
  • api_requests_total - 按方法/端点/状态分组的请求计数
  • api_request_duration_seconds - 请求延迟分布

权限系统:
  • permission_check_success_total - 权限检查成功次数
  • permission_check_duration_seconds - 权限检查延迟

系统健康:
  • up - 服务存活状态
  • organization_active_count - 当前活跃组织数
```

## 🎯 **使用场景**

### **场景1: 日常SLO监控**

1. **打开Grafana仪表板**:
   ```
   http://localhost:3001
   → Dashboard → 组织启停API - SLO监控仪表板
   ```

2. **关注关键指标**:
   - ✅ 成功率SLO状态 (绿色=正常，红色=违规)
   - ⚡ P95延迟趋势图 (应≤150ms) 
   - 🚫 弃用端点访问计数 (应为0)
   - 💰 错误预算余额 (绿色=充足)

3. **SLO违规处理**:
   - 成功率<99.9%: 检查错误日志，分析失败原因
   - 延迟>150ms: 检查数据库性能，优化查询
   - 弃用端点访问>0: 追踪客户端来源，推进迁移

### **场景2: API性能分析**

1. **查看请求分布**:
   ```
   Prometheus查询: sum(rate(api_requests_total[5m])) by (endpoint)
   ```

2. **分析慢查询**:
   ```
   延迟>1秒的请求: api_request_duration_seconds_bucket{le="1"} 
   ```

3. **错误率分析**:
   ```
   4xx错误率: sum(rate(api_requests_total{status=~"4.."}[5m])) by (endpoint)
   ```

### **场景3: ADR-008合规监控**

1. **检查弃用端点访问**:
   ```
   Prometheus查询: deprecated_endpoint_used_total
   ```

2. **审计完整性监控**:
   ```
   审计失败: audit_write_failures_total
   审计成功率: audit_write_success_total / audit_write_attempts_total
   ```

3. **权限系统健康**:
   ```
   权限检查延迟: histogram_quantile(0.99, permission_check_duration_seconds_bucket)
   ```

## 🚨 **告警处理**

### **P0级告警 (立即处理)**
- **AuditWriteFailureSLOViolation**: 审计日志写入失败
  - 影响: 数据完整性威胁
  - 处理: 立即检查审计系统，确保operatedBy字段记录

### **P1级告警 (2分钟内响应)**
- **ActivateAPISuccessRateSLOViolation**: 启用API成功率<99.9%
- **SuspendAPISuccessRateSLOViolation**: 停用API成功率<99.9%
  - 影响: 核心业务功能受损
  - 处理: 检查API服务状态、数据库连接、权限系统

### **P2级告警 (5分钟内响应)**
- **DeprecatedEndpointUsageSLOViolation**: 检测到弃用端点访问
  - 影响: 合规性违规
  - 处理: 分析410响应来源，推进客户端迁移

### **P3级告警 (监控，非紧急)**
- **HighActivateLatency**: 启用API延迟>150ms
- **HighSuspendLatency**: 停用API延迟>150ms
  - 影响: 用户体验下降
  - 处理: 性能优化，查询调优

## 🔧 **故障排除**

### **监控系统问题**

```bash
# 检查容器状态
docker compose -f monitoring/docker-compose.monitoring.yml ps

# 查看服务日志
docker compose -f monitoring/docker-compose.monitoring.yml logs prometheus
docker compose -f monitoring/docker-compose.monitoring.yml logs grafana

# 重启监控系统
docker compose -f monitoring/docker-compose.monitoring.yml restart
```

### **数据采集问题**

```bash
# 检查Prometheus targets
curl http://localhost:9090/api/v1/targets

# 检查API服务metrics端点
curl http://localhost:9090/metrics

# 手动查询指标
curl "http://localhost:9090/api/v1/query?query=up"
```

### **Grafana仪表板问题**

```bash
# 检查数据源连接
curl -u admin:cube-castle-2025 http://localhost:3001/api/datasources

# 仪表板管理
# 1. 访问 http://localhost:3001
# 2. Dashboards → 已自动加载 “组织启停API - SLO监控仪表板”
# 3. 如需手动导入/备份，可使用文件 monitoring/grafana/dashboards/slo-dashboard.json
```

## 📈 **高级用法**

### **自定义查询示例**

```yaml
# 错误率趋势
sum(rate(api_requests_total{status=~"5.."}[5m])) / sum(rate(api_requests_total[5m])) * 100

# 最慢的API端点
topk(5, histogram_quantile(0.95, sum(rate(api_request_duration_seconds_bucket[5m])) by (endpoint, le)))

# 弃用端点访问来源分析  
sum(deprecated_endpoint_used_total) by (client_id, user_agent)

# 权限拒绝率
sum(rate(api_requests_total{status="403"}[5m])) / sum(rate(api_requests_total[5m])) * 100
```

### **仪表板定制**

1. **复制现有仪表板**:
   - Grafana → Dashboard → 组织启停API SLO → Settings → Save As

2. **添加新面板**:
   - Add Panel → 选择Prometheus数据源 → 输入查询

3. **设置告警**:
   - Panel → Alert → Create Alert Rule

## 📝 **维护任务**

### **日常维护**
- ✅ 检查SLO仪表板状态 (每日)
- ✅ 处理活跃告警 (实时)
- ✅ 审查错误预算消耗 (周度)

### **定期维护**
- 🔄 更新仪表板配置 (月度)
- 🗂️ 清理历史数据 (Prometheus默认保留30天)
- 📊 分析SLO趋势，优化阈值 (季度)

---

## 🆘 **快速参考**

| 需求 | 操作 |
|------|------|
| **启动监控** | `./scripts/start-monitoring.sh` |
| **验证状态** | `./scripts/test-monitoring.sh` |
| **查看日志** | `docker compose -f monitoring/docker-compose.monitoring.yml logs -f` |
| **重启监控** | `docker compose -f monitoring/docker-compose.monitoring.yml restart` |
| **停止监控** | `docker compose -f monitoring/docker-compose.monitoring.yml down` |

**紧急联系**: 如遇到监控系统问题，请检查Docker服务状态和配置文件完整性。
Linux注意事项:
- 监控容器已内置 `host.docker.internal:host-gateway` 映射，Prometheus 可抓取宿主机上运行的 API（localhost:9090）。如仍抓取失败，请确认 API 已暴露 `GET /metrics` 并在宿主机监听 9090。
