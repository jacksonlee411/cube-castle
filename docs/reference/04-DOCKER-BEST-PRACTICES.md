# Docker 容器化部署最佳实践

## 1. 强制原则（摘自 CLAUDE.md）
- 所有服务、数据库、中间件必须通过 Docker Compose 运行。
- 禁止在宿主机直接安装 PostgreSQL、Redis、Temporal 等组件。
- 遇到端口冲突时必须卸载宿主服务，严禁修改容器端口映射。

## 2. 开发流程

### 2.1 启动 / 停止服务
```bash
make run-dev                       # 启动 postgres + redis + rest + graphql
docker compose -f docker-compose.dev.yml down  # 停止全部容器
```

### 2.2 查看日志
```bash
docker compose -f docker-compose.dev.yml logs -f rest-service graphql-service
```

### 2.3 进入容器调试
```bash
docker exec -it cubecastle-rest sh
docker exec -it cubecastle-graphql sh
```

## 3. 配置说明

### 3.1 环境变量
- 容器内服务：`DATABASE_URL=postgres://user:password@postgres:5432/cubecastle?sslmode=disable`
- 宿主机工具：`DATABASE_URL=postgres://user:password@localhost:5432/cubecastle?sslmode=disable`

### 3.2 端口映射
- PostgreSQL: `localhost:5432 -> postgres:5432`
- Redis: `localhost:6379 -> redis:6379`
- REST API: `localhost:9090 -> rest-service:9090`
- GraphQL API: `localhost:8090 -> graphql-service:8090`

### 3.3 热重载（可选）
```bash
export COMMAND_SERVICE_BUILD_TARGET=dev
export COMMAND_SERVICE_WORKDIR=/workspace/cmd/organization-command-service
export GRAPHQL_SERVICE_BUILD_TARGET=dev
export GRAPHQL_SERVICE_WORKDIR=/workspace/cmd/organization-query-service
docker compose -f docker-compose.dev.yml up -d --build rest-service graphql-service
```
- 退出：`docker compose -f docker-compose.dev.yml down` 及 `unset` 上述变量。
- 详情见 `docs/development-guides/docker-hot-reload-guide.md`。

---

**提示**：所有 `localhost` 端点均由容器映射提供，发现端口占用时务必卸载宿主服务，严禁以修改容器端口方式绕过。
