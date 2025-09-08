# 🚀 Cube Castle 生产环境部署和运维手册

> 重要变更公告（2025-09-07）
> 
> - 项目已完成“PostgreSQL 原生化”：彻底移除 Phoenix/CDC、Neo4j、Kafka 相关组件与流程。
> - 本手册中如有涉及 Neo4j/Kafka/CDC 的历史描述与架构图，均仅作历史参考；以 README/CLAUDE.md/AGENTS.md 与 Makefile 的最新规范为准。
> - 统一入口命令：
>   - 基础设施（PostgreSQL+Redis）：`make docker-up`
>   - 后端启动（命令9090 + GraphQL8090）：`make run-dev`
>   - 前端启动：`make frontend-dev`
>   - 状态查看：`make status`

## 📋 项目概述

**Cube Castle**是一个基于现代化CQRS架构的人力资源管理系统，已完成时态管理API升级项目，具备企业级生产环境部署能力。

### 🏆 核心成果
- ✅ PostgreSQL 原生查询替代 Neo4j，简化为单一数据源
- ✅ GraphQL 查询 1.5–8ms（详见 README 性能节）
- ✅ 架构简化约 60%，移除 CDC 同步复杂性
- ✅ 监控栈可选（Prometheus/Grafana），脚本化启动

---

## 🏗️ 系统架构

### 核心服务架构（历史示意）
```
┌─────────────────┐    ┌─────────────────┐
│   前端应用      │    │   监控面板      │
│  (Port 3000)    │    │  (Grafana)      │
└─────────┬───────┘    └─────────────────┘
          │
┌─────────▼───────┐    ┌─────────────────┐
│  查询服务       │    │   命令服务      │
│ (GraphQL:8090)  │    │ (REST:9090)     │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
┌─────────▼───────┐    ┌─────────▼───────┐
│   （已移除）    │    │  PostgreSQL     │
│  Neo4j/Kafka    │    │  （单一数据源） │
└─────────────────┘    └─────────────────┘
          │                      │
          └──────┬─────────────────┘
                 ▼
       ┌─────────────────┐
       │  CDC同步服务    │
       │ (Kafka+Redis)   │
       └─────────────────┘
```

### 数据流（现行）
- 命令流：前端 → REST API(9090) → PostgreSQL（单一数据源）
- 查询流：前端 → GraphQL(8090) → PostgreSQL（原生查询 + 索引优化）
- 缓存：Redis（精确失效，按需）
- 监控：Prometheus → Grafana（可选）

---

## 🚀 部署指南

### 前置要求（PostgreSQL 原生）
- Docker & Docker Compose
- Go 1.23+
- Node.js 18+
- PostgreSQL 16+
- Redis 7.x

### 快速启动

#### 1. 启动基础设施
```bash
cd /home/shangmeilin/cube-castle
make docker-up   # 仅 PostgreSQL + Redis
```

## 🔁 CI/CD 流程（概览）

- 工作流: `.github/workflows/consistency-guard.yml`
- 触发条件:
  - push: 任意分支（branches: "**"），含 tag（tags: "*")
  - pull_request: 任意目标分支（branches: "**"）
  - workflow_dispatch: 手动触发
  - release: published/created/edited/prereleased
- 强制守护（Enforce=ON）:
  - 前端 REST 查询守护（GraphQL-only 查询约束）
  - cmd/* 配置守护（CORS 硬编码/端口/内联 JWT 配置）
- 本地自检命令:
  - `bash scripts/ci/check-permissions.sh`
  - `bash scripts/ci/check-rest-queries.sh`
  - `bash scripts/ci/check-hardcoded-configs.sh` （`ENFORCE=1` 可模拟强制）

#### 2. 启动核心服务
```bash
# 一键后端启动（命令 9090 + PostgreSQL 原生 GraphQL 8090）
make run-dev
```

#### 3. 启动前端（可选）
```bash
make frontend-dev
```

#### 4. 验证部署
```bash
# 健康检查
curl http://localhost:9090/health
curl http://localhost:8090/health

# API测试
curl -X POST http://localhost:8090/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ organizations { code name } }"}'
```

---

## 🔧 运维操作

### 服务管理

#### 健康检查
```bash
# 核心服务状态
curl http://localhost:9090/health    # 命令服务
curl http://localhost:8090/health    # 查询服务

# 数据库连接测试
PGPASSWORD=password psql -h localhost -U user -d cubecastle -c "SELECT 1;"
```

#### 性能监控
```bash
# 查看实时指标
curl http://localhost:9090/metrics   # Prometheus指标
curl http://localhost:8090/metrics   # GraphQL服务指标

# 缓存性能
redis-cli info stats | grep hit_rate
```

#### 日志管理
```bash
# 查看服务日志
tail -f /tmp/command-service.log
tail -f /tmp/query-service.log
tail -f /tmp/sync-service.log

# 应用程序日志级别: info, warn, error
export LOG_LEVEL=info
```

### 故障排除

#### 常见问题

**1. 服务无法启动**
- 检查端口占用: `netstat -tulpn | grep :9090`
- 检查数据库连接: `docker ps | grep postgres`
- 查看错误日志: `tail -f /tmp/*.log`

**2. API响应缓慢**
- 检查缓存命中率: `redis-cli info stats`
- 查看数据库连接池: PostgreSQL慢查询日志
- 监控内存使用: `docker stats`

**3. 数据不一致**
- 当前为单一数据源（PostgreSQL），无 CDC 同步。请检查应用写入事务与查询 SQL。

**4. 前端无法访问**
- 检查CORS配置: `CORS_ALLOWED_ORIGINS=http://localhost:3000`
- 验证API连通性: `curl http://localhost:9090/health`
- 检查前端构建: `npm run build`

### 备份和恢复

#### 数据库备份
```bash
# PostgreSQL备份
PGPASSWORD=password pg_dump -h localhost -U user -d cubecastle > backup_$(date +%Y%m%d).sql
```

#### 配置备份
```bash
# 备份配置文件
tar -czf config_backup_$(date +%Y%m%d).tar.gz \
  .env.production \
  docker-compose.yml \
  monitoring/
```

---

## 📊 API文档

### GraphQL API (查询操作)

**端点**: http://localhost:8090/graphql  
**GraphiQL界面**: http://localhost:8090/graphiql

#### 核心查询
```graphql
# 查询所有组织
query {
  organizations {
    code
    name 
    unitType
    status
    level
    path
  }
}

# 查询特定组织
query {
  organization(code: "1000001") {
    code
    name
    parentCode
    children {
      code
      name
    }
  }
}

# 组织统计信息
query {
  organizationStats {
    totalCount
    activeCount
    departmentCount
    teamCount
  }
}
```

### REST API (命令操作)

**端点**: http://localhost:9090/api/v1/organization-units

#### 核心操作
```bash
# 创建组织
POST /api/v1/organization-units
{
  "name": "新部门",
  "unit_type": "DEPARTMENT", 
  "parent_code": "1000001",
  "description": "部门描述"
}

# 更新组织
PUT /api/v1/organization-units/1000001
{
  "name": "更新后的名称",
  "description": "新的描述"
}

# 删除组织
DELETE /api/v1/organization-units/1000001

# 查询单个组织 (兼容性)
GET /api/v1/organization-units/1000001
```

#### 时态管理API (扩展功能)
```bash
# 时间点查询
GET /api/v1/organization-units/1000001?as_of_date=2025-01-01

# 历史版本查询
GET /api/v1/organization-units/1000001/history

# 创建变更事件
POST /api/v1/organization-units/1000001/events
{
  "event_type": "UPDATE",
  "effective_date": "2025-09-01",
  "change_data": {
    "name": "新名称",
    "description": "变更描述"
  },
  "change_reason": "部门重组"
}
```

---

## 📈 监控指标

### 核心性能指标
- **API响应时间**: <100ms (目标)
- **缓存命中率**: >90% (目标)
- **CDC同步延迟**: <300ms
- **错误率**: <0.1%
- **服务可用性**: >99.9%

### 监控访问
- **Prometheus**: http://localhost:9090 (如果启用)
- **前端监控面板**: http://localhost:3000/monitoring
- **Kafka UI**: http://localhost:8081
- **Neo4j Browser**: http://localhost:7474

### 告警规则
- API响应时间超过500ms
- 缓存命中率低于85%
- 错误率超过1%
- 服务无响应超过1分钟
- 数据库连接失败

---

## 🔒 安全配置

### 生产环境安全检查清单
- [ ] 修改默认数据库密码
- [ ] 配置API访问控制 (CORS)
- [ ] 启用HTTPS (生产环境)
- [ ] 配置防火墙规则
- [ ] 定期备份数据
- [ ] 监控异常访问
- [ ] 更新安全补丁

### 环境变量配置
```bash
# 数据库安全
DATABASE_URL=postgres://user:strong_password@localhost:5432/cubecastle

# API安全
CORS_ALLOWED_ORIGINS=https://your-domain.com
API_RATE_LIMIT=1000
SESSION_TIMEOUT=3600

# 日志安全
LOG_LEVEL=info  # 生产环境避免debug级别
SENSITIVE_DATA_MASKING=true
```

---

## 📞 支持联系

### 技术支持
- **项目仓库**: /home/shangmeilin/cube-castle
- **文档路径**: /DOCS2/
- **监控配置**: /monitoring/
- **API文档**: /docs/api/

### 故障上报
1. 收集错误日志和系统状态
2. 记录重现步骤和环境信息
3. 执行基础故障排除步骤
4. 提供监控指标和性能数据

---

## 🎯 最佳实践

### 开发建议
- 使用GraphQL进行所有查询操作
- 使用REST API进行所有命令操作  
- 避免跨协议混用（保持架构一致性）
- 合理利用Redis缓存提升性能
- 监控CDC同步状态确保数据一致性

### 运维建议
- 定期检查服务健康状态
- 监控关键性能指标
- 及时清理日志文件
- 保持数据备份最新
- 定期更新系统依赖

---

**文档版本**: v1.0-Production  
**最后更新**: 2025-08-10  
**系统状态**: 🚀 **生产环境就绪**
