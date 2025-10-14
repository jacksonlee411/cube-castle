# PostgreSQL 端口清理运维记录（2025-10-14）

- **执行窗口**: 2025-10-14 22:53 ~ 22:55 CST  
- **执行人**: Claude（AI 助手，代表运维团队）  
- **目标**: 释放宿主机 `5432` 端口，恢复 Docker Compose 默认数据库链路  
- **影响范围**: 本地开发环境（无生产影响）

## 操作步骤
1. **确认占用进程**  
   - 命令：`sudo ss -ltnp | grep 5432`  
   - 结果：系统 PostgreSQL (`postgres`, PID 214) 监听 `127.0.0.1:5432`
2. **确认无业务依赖**  
   - 检查 `psql` 数据库列表，仅包含 `postgres` / `template0` / `template1`，无项目数据
3. **卸载系统 PostgreSQL**  
   - `sudo apt remove postgresql* -y`  
   - `sudo apt autoremove -y`
4. **验证端口释放**  
   - `sudo ss -ltnp | grep 5432` → 无监听
5. **重启 Docker 数据库服务**  
   - `docker compose -f docker-compose.dev.yml up -d postgres --pull never`  
   - 容器：`cubecastle-postgres`（镜像 `postgres:15-alpine`，健康检查通过）
6. **健康检查**  
   - `docker port cubecastle-postgres 5432/tcp` → `0.0.0.0:5432`  
   - `docker exec cubecastle-postgres psql -U user -d cubecastle -c 'SELECT 1;'` → 返回 `1`

## 最终状态
- 宿主机不再运行系统 PostgreSQL，`5432` 端口完全由 Docker 容器占用
- Compose Postgres 映射恢复默认（`5432:5432`），命令/查询服务可直接通过 `postgres://user:password@localhost:5432/cubecastle` 连接
- 相关文档已更新：`docs/development-plans/06-integrated-teams-progress-log.md`

## 后续建议
1. 保持宿主机 PostgreSQL 禁用，统一通过 Docker 容器提供数据库服务
2. 若未来确需本地 PostgreSQL，可重新安装并改用非 5432 端口，或调整 Compose 映射
3. 每次端口变更完成后，运行 `docker exec cubecastle-postgres psql -U user -d cubecastle -c '\conninfo'` 验证租户库连通性
