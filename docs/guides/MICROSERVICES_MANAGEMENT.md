# Cube Castle 微服务管理指南

> 重要变更公告（2025-09-07）
>
> - 项目已完成“PostgreSQL 原生化”，现行核心后端仅包含两个服务：
>   - `organization-command-service`（REST：9090）
>   - `organization-query-service`（GraphQL：8090，直接查询 PostgreSQL）
> - 旧的多微服务拓扑（API Gateway、Neo4j、Kafka、Sync Service 等）已废弃，仅保留历史参考。
> - 本地开发与启动统一使用 Makefile：`make docker-up`、`make run-dev`、`make frontend-dev`、`make status`。

## 概述

Cube Castle项目采用CQRS微服务架构，包含多个独立的服务组件。本指南介绍如何使用统一的管理脚本来管理这些微服务。

## 当前架构（PostgreSQL 原生）

### 核心后端服务
- `organization-command-service`（端口 9090）：命令端（REST），写入 PostgreSQL
- `organization-query-service`（端口 8090）：查询端（GraphQL），直接从 PostgreSQL 读取

### 基础设施（最小依赖）
- PostgreSQL（5432）
- Redis（6379）

## 开发与启动（统一入口）

```bash
# 启动基础设施（PostgreSQL + Redis）
make docker-up

# 启动后端（命令 9090 + GraphQL 8090）
make run-dev

# 启动前端
make frontend-dev

# 查看状态
make status
```

## API 调用路径（现行）

- GraphQL 查询：`http://localhost:8090/graphql`（GraphiQL：`/graphiql`）
- 命令操作（REST）：`http://localhost:9090/api/v1/organization-units`

## 服务启动顺序（建议）

1. 基础设施（PostgreSQL、Redis）→ `make docker-up`
2. 命令服务 → `make run-dev`（命令服务随命令同时启动）
3. 查询服务 → `make run-dev`（查询服务随命令同时启动）
4. 前端（可选）→ `make frontend-dev`

## 健康检查

每个服务都提供健康检查端点：
- `http://localhost:{PORT}/health`

管理脚本会自动进行健康检查并显示状态。

## 日志文件

所有服务的日志文件位于：
- `cmd/{service-name}/logs/{service-name}.log`

例如：
- `cmd/organization-api-gateway/logs/organization-api-gateway.log`
- `cmd/organization-graphql-service/logs/organization-graphql-service.log`

## 故障排查

### 1. 服务启动失败
```bash
# 检查服务状态
./scripts/microservices-manager.sh status

# 查看服务日志
tail -f cmd/{service-name}/logs/{service-name}.log
```

### 2. 端口占用问题
```bash
# 检查端口占用
lsof -i :{PORT}

# 终止占用进程
kill -9 {PID}
```

### 3. 数据库连接问题
确保Docker容器正常运行：
```bash
docker-compose ps
```

### 4. 前端删除操作失败
确认API网关(8000)和命令服务器(9090)都在运行：
```bash
curl http://localhost:8000/health
curl http://localhost:9090/health
```

## 开发最佳实践

1. **启动开发环境**:
   ```bash
   # 启动基础设施
   docker-compose up -d
   
   # 启动微服务
   ./scripts/microservices-manager.sh start
   ```

2. **代码更改后重新部署**:
   ```bash
   # 重新编译并重启
   ./scripts/microservices-manager.sh build
   ./scripts/microservices-manager.sh restart
   ```

3. **停止开发环境**:
   ```bash
   # 停止微服务
   ./scripts/microservices-manager.sh stop
   
   # 停止基础设施（可选）
   docker-compose down
   ```

## 监控和运维

- 可选启用监控脚本：`./scripts/start-monitoring.sh`、`./scripts/test-monitoring.sh`
- 确保基础设施服务正常运行（`make status` 查看）

## 架构优势（现行）

1. 查询与命令分离（GraphQL/REST），但统一 PostgreSQL 单一数据源
2. 无 CDC/双数据库，同步复杂性归零，数据一致性更强
3. Redis 作为精确失效缓存，性能可控
4. 监控脚本完备，可选接入 Prometheus/Grafana

## 待优化项

1. 实施分布式追踪（Jaeger/Zipkin）
2. 集成Prometheus监控指标
3. 实施自动重启和故障恢复
4. 添加服务发现机制
5. 实施配置中心统一管理
