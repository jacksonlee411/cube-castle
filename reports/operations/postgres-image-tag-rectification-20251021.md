# PostgreSQL 镜像标签整改记录（2025-10-21）

- **关联调查**：`docs/development-plans/106-postgres-image-tags-investigation.md`
- **执行主体**：架构组 · 平台运维协作小组
- **操作窗口**：2025-10-21 10:16（UTC+08）

## 执行步骤

1. 更新 `docker-compose.dev.yml` PostgreSQL 镜像为 `postgres:16-alpine`，与标准/E2E 配置保持一致。
2. 执行 `docker-compose -f docker-compose.dev.yml down && docker-compose -f docker-compose.dev.yml up -d`，确认 `cubecastle-postgres` 以新标签运行。
3. 通过 `docker exec cubecastle-postgres postgres --version` 验证数据库版本为 `PostgreSQL 16.9`。
4. 删除冗余标签 `docker rmi postgres:15-alpine`，保留单一官方标签。
5. 更新 106 号调查文档状态，并记录整改完成时间。

## 验证结果

- `docker ps` 显示 `cubecastle-postgres` 运行镜像 `postgres:16-alpine`，状态 `healthy`。
- `docker images postgres` 仅保留 `postgres:16-alpine` 标签。
- 106 号文档状态更新为“已完成（2025-10-21 UTC）”。

## 风险与回滚

- **风险**：容器重启期间服务短暂中断（~30 秒），无数据丢失风险。
- **回滚**：保留原 Compose 版本，可通过 `git checkout docker-compose.dev.yml` 并重新部署恢复原状态。

