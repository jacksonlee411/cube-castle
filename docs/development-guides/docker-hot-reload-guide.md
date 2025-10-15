# Docker 热重载开发指南

本指南说明如何在 Docker 容器内启用 Go 服务热重载（Air），实现与宿主机代码同步更新。默认情况下 `make run-dev` 使用发布模式镜像，如需动态调试可按以下步骤切换。

---

## 1. 前置准备

- 确保仓库根目录存在（当前终端处于 `cube-castle/`）。
- 已执行 `make docker-up` 或 `make run-dev`，容器网络及基础设施可用。
- 宿主机未安装占用 9090/8090/5432/6379 等端口的服务。

---

## 2. 启用热重载

```bash
# 仅在当前终端会话生效
export COMMAND_SERVICE_BUILD_TARGET=dev
export COMMAND_SERVICE_WORKDIR=/workspace/cmd/organization-command-service
export GRAPHQL_SERVICE_BUILD_TARGET=dev
export GRAPHQL_SERVICE_WORKDIR=/workspace/cmd/organization-query-service

docker compose -f docker-compose.dev.yml up -d --build rest-service graphql-service
```

### 说明
- `build.target=dev`：选择 Dockerfile 中的 Air 开发阶段。
- `working_dir` 指向挂载后的源码目录，使 Air 能读取宿主机文件。
- `docker-compose.dev.yml` 已默认挂载 `./:/workspace:delegated`，无需额外配置。

---

## 3. 热重载体验

- 修改 `cmd/organization-command-service/` 或 `cmd/organization-query-service/` 下的 Go 源码后，Air 会在容器内自动编译并重启服务。
- 运行日志可通过：
  ```bash
  docker compose -f docker-compose.dev.yml logs -f rest-service
  docker compose -f docker-compose.dev.yml logs -f graphql-service
  ```
- 健康检查：`curl http://localhost:9090/health`、`curl http://localhost:8090/health`

---

## 4. 退出热重载

```bash
docker compose -f docker-compose.dev.yml down
unset COMMAND_SERVICE_BUILD_TARGET COMMAND_SERVICE_WORKDIR
unset GRAPHQL_SERVICE_BUILD_TARGET GRAPHQL_SERVICE_WORKDIR
```

重新执行 `make run-dev` 即可恢复发布模式镜像。

---

## 5. 常见问题

- **端口被占用**：确认未在宿主机启动 PostgreSQL/Redis/Go 服务，如占用请卸载后重试。
- **Air 未触发重载**：检查容器日志，确保 `working_dir` 环境变量指向 `/workspace/...`；如在新终端运行，记得重新导出环境变量。
- **性能缓慢**：首次构建会下载依赖，后续增量构建时间会缩短；如无需热重载，可直接使用 `make run-dev`。
