# 06号文档：集成团队进展日志（2025-10-14 更新）

## 一、阶段进展概览
- **数据库基线修复**：`008_temporal_management_schema.sql` 已补充幂等建表逻辑，配合 `docker exec ... psql < database/migrations/*.sql` 验证可在全新环境一次性执行，旧版 `organization_unit_versions`、时间线事件等遗留表彻底移除。
- **Stage1 迁移交付**：`043_create_positions_and_job_catalog.sql` 已在容器内执行，记录于 `reports/database/migration-043-stage1-20251014.log`；租户隔离巡检 `docs/development-plans/81-tenant-isolation-checks.sql` 输出归档 `reports/architecture/tenant-isolation-check-stage1-20251014.log`（空集）。
- **代码检查**：`go test ./cmd/organization-command-service/...`、`go test ./cmd/organization-query-service/...` 均通过；`organization-command-service/internal/audit` 新增 `ResourceTypeJobCatalog`、`ResourceTypePosition` 保持审计常量完备。

## 二、近期完成
- ✅ 更新 `008/024/037` 迁移脚本，消除历史版本表依赖并跳过演示数据缺失时的 reparent 脚本。
- ✅ 追加运维记录 `reports/operations/postgresql-port-cleanup-20251014.md`，明确 Docker-only 原则与端口清理步骤。
- ✅ 刷新 `docs/development-plans/82-position-management-stage1-implementation-plan.md` 里程碑（M1/M5）、验收项，并注明迁移/巡检证据。

## 三、待办事项
1. **集成测试**  
   - 覆盖 `/api/v1/positions*`、`/api/v1/job-*` REST 流程（创建→版本→状态），收集日志或脚本输出，勾选 82 号计划和 81 号契约第 10 节剩余项。
2. **GraphQL 验证**  
   - 通过 `positions`、`positionTimeline`、`positionHeadcountStats` 等查询验证分页、过滤、租户隔离；记录请求/响应示例。
3. **文档同步**  
   - 更新 `docs/development-plans/81-position-api-contract-update-plan.md` 余下勾选项；确认 `docs/reference/02-IMPLEMENTATION-INVENTORY.md` 与实际接口保持一致。
4. **Docker 合规整改（参考 83 号计划）**  
   - Phase 1：Makefile 与 `.env` 默认路径迁移至 Docker，保留 `run-dev-debug` 调试模式但加警告。  
   - Phase 2：更新快速参考文档、脚本提示，并在 CI 增加 docker-compliance 检查。  
   - Phase 3：落地热重载方案、最佳实践文档。

## 四、风险与提醒
- **Docker 首选**：遵循 CLAUDE.md/AGENTS.md 强制原则，后续开发须通过 `docker compose` 启动服务；宿主调试仅限 `run-dev-debug`。
- **Stage2 准备**：命令/查询 REST & GraphQL 测试未完成前，请勿合入前端集成；需有明确验收日志。
- **计划引用**：后续进展和证据需回填 `docs/development-plans/82-...`、`docs/development-plans/83-...` 确保计划闭环。

## 五、快速检查清单
- [x] 全量迁移在空库验证通过。
- [x] Stage1 迁移及租户巡检日志归档。
- [ ] REST 集成测试覆盖职位与 Job Catalog。
- [ ] GraphQL 查询验证并记录。
- [ ] 81 号契约第 10 节余项勾选。
- [ ] Docker 合规 Phase 1 修复（Makefile/.env/README 等）。
