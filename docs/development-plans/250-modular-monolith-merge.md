# Plan 250 - 模块化单体合流与边界治理

文档编号: 250  
标题: 模块化单体合流与边界治理（来源：202 计划拆分）  
创建日期: 2025-11-15  
版本: v1.0  
关联计划: 202-CQRS混合架构深度分析与演进建议, 203/204/206, 215  
状态: ✅ 已完成（合流与治理门禁就绪；E2E/性能基线首轮在 CI 跑并回填 215）

---

## 1. 背景与目标

来源自 202 计划“从混合走向清晰分层”的总目标。本子计划聚焦于合流到“模块化单体”的工程化落地与边界治理：统一项目结构、明确聚合边界、收敛跨层命名、建立守卫与回归路径，支撑后续 CQRS、契约治理与流水线简化。

目标：
- 统一项目结构（cmd/*/、internal/*、pkg/*、docs/*）与命名；
- 明确领域聚合边界与共享层（internal/organization/* 范式推广）；
- 建立最低可用的守卫与回归脚手架（lint、arch-validator、document-sync）；
- 将合流与重构的验收证据“只登记不复制”，保持单一事实来源。

## 2. 交付物
- 目录结构与命名约定文档（引用 203/204/AGENTS）；
- 最小守卫与校验脚本（ESLint 架构守卫、document-sync、API compliance）；
- 合流基线清单与风险登记（docs/development-plans/250-* 路径内）；
- 验收记录与日志（logs/plan250/*）。

## 3. 依赖与约束
- 依赖：203/204 的结构与时间表，AGENTS.md 强约束（Docker 强制、契约唯一来源）；
- 约束：不得复制规范正文；仅引用权威链接。

## 3.1 运行模式与回退开关（治理）
- 默认运行模式：单体合流进程（单端口 9090）为唯一支持的运行模式（CI 与本地）
- 回退开关：仅限本地排障短期启用 `ENABLE_LEGACY_DUAL_SERVICE=true`（默认 false）
  - CI 门禁：工作流检测该变量，一旦为 true 则阻断
  - 回退登记：启用时须在 215 登记 `// TODO-TEMPORARY(YYYY-MM-DD):` 原因与截止期（一个迭代内收敛）
- 端口约束：不得修改 docker-compose*.yml 的容器端口映射；冲突时按 AGENTS 卸载宿主服务
  - 临时方案管控（AGENTS 对齐）：所有临时开关/豁免均需使用 `// TODO-TEMPORARY(YYYY-MM-DD):` 标注并在 215/06 登记，超期由 CI 门禁阻断（参考 Plan 258/Plan 253 的白名单机制）

## 3.2 旧入口与发布治理
- 旧双进程入口（如 `cmd/hrms-server/command/main.go`、`cmd/hrms-server/query/main.go`）仅保留为测试/工具用途
- CI/发布阶段不得构建或发布旧入口的独立可执行物；仅发布合流后的单一二进制
- 文档与开发者速查（README/Quick Reference）更新为“默认单体进程”；旧指令标注为“仅本地排障（禁 CI）”
- 面向未来的最佳实践：
  - 构建隔离：使用 Makefile/CI 仅编译单一 `cmd/hrms-server` 目标；将旧入口移出构建矩阵或加 build tag 约束（例如 `//go:build legacy`）
  - 目录隔离：如需长期保留，迁移到 `tools/legacy/` 并在 README 标注不可发布
  - CI 保证：新增“唯一二进制门禁”（见下）与“禁 legacy build tag”检查

## 3.3 准入门禁引用（避免重复造轮子）
- 环境/流水线通用门禁（端口映射、镜像标签、冷启动计时）不在本计划重复定义，统一引用 Plan 253 的门禁作为本计划准入条件：
  - Compose 端口映射变更检测（5432/6379/9090/8090）
  - PostgreSQL/Redis 禁止使用 `latest` 标签（固定版本）
  - 冷启动/就绪时间记录并登记 215（首轮 CI）

## 4. 验收标准
- 目录/命名/边界与 203/AGENTS 一致（脚本守卫通过）；
- 架构与文档同步校验通过（document-sync、architecture-validator）；
- CI 中新增/复用的守卫任务全绿；
- 产物与证据登记完整（logs/plan250/*）；
- 单一进程暴露 `/api/v1/*`、`/graphql`、`/health`、`/metrics`，功能等同性；
- CI 禁止双进程路径（变量门禁生效）。
- 221 必跑：`make test-db` 通过（引用 Plan 221），登记 `logs/plan250/test-db-*.log`
- E2E 烟测：最小集（引用 232/241/244 的关键路径）Chromium/Firefox 各 1 轮通过并登记 `logs/plan250/e2e-*.log`  
  - 建议引用测试文件（见 215）：`frontend/tests/e2e/smoke-org-detail.spec.ts`、`frontend/tests/e2e/temporal-header-status-smoke.spec.ts`、`frontend/tests/e2e/temporal-management-integration.spec.ts`
- JWKS/JWT/多租户链路：`/.well-known/jwks.json` 200；`X-Tenant-ID` 透传/权限校验在 REST/GraphQL 各抽样 1 条一致（登记 `logs/plan250/jwks-*.json`、`logs/plan250/tenant-check-*.log`）
- 性能/资源基线：引用 204 的性能指标与 performance/ 脚本，合流前/后同负载 P95/P99 延迟不升高 > 10%、RSS 内存不升高 > 15%（登记 `logs/plan250/perf-*.json`）  
  - 建议脚本：`performance/performance_test.sh`（或同目录基线脚本）；如缺失则补充后统一由 204 定义阈值

## 5. 执行与证据
- 执行路径：对照 203 的“结构与边界”章节 → 调整工程 → 落地守卫脚本 → 登记证据；
- 证据：logs/plan250/*.log、.json（包含脚本输出与校验摘要）。

## 6. 风险与回滚
- 风险：历史路径残留/命名不一致 → 以守卫门禁兜底，必要时分支回滚；
- 回滚：仅限本地启用 `ENABLE_LEGACY_DUAL_SERVICE=true`，CI 禁止；215 登记回退与恢复时间。
- 触发条件：SLO 退化（延迟/错误率超阈）、健康/指标异常、门禁失败
- 恢复步骤：提交恢复脚本（占位路径 `scripts/recovery/250-*.sh`），按步骤切回合流或回退到上一个稳定版本并在 215 登记

## 7. 门禁
- CI 变量门禁：检测 `ENABLE_LEGACY_DUAL_SERVICE`，为 true 则失败
- 文档同步门禁：document-sync 与 architecture-validator 全绿方可合并
- 唯一二进制门禁：构建产物仅允许单一服务二进制；检测重复 main/服务产物即失败  
  - 实现提示（示例）：
    - 统计 main 包入口：`rg -n \"^func main\\(\\)\" ./cmd | wc -l` 应为 1（或仅允许 `cmd/hrms-server`）  
    - 或在构建产物目录计数二进制数量=1
- 端口监听门禁：合流后不得监听 8090；检测到 8090 监听则失败
- 组合门禁复用：引用 Plan 253 的“compose 端口映射/镜像标签/冷启动记录”门禁作为本计划准入条件

## 8. 执行清单（引用 253/221/215）
- 253（流水线门禁）：
  - compose 端口映射与镜像标签门禁全绿；首轮冷启动/就绪计时登记 215
- 221（集成基座）：
  - `make test-db` 流程通过；Goose up/down 循环日志落盘
- 215（执行登记）：
  - 合流命令与日志路径、E2E/性能/链路验证产物、回退登记与恢复脚本链接

## 9. 运行态自检（面向未来）
- 启动时进行自检：若检测到 8090 端口监听或冲突，则拒绝启动并输出建议（合流仅使用 9090）
- 可选：自检“重复进程”与“绑定冲突”，输出排障指引（指向 215 的运行日志路径与常见问题）

---

维护者: 架构组（与 203/204 同步评审）
