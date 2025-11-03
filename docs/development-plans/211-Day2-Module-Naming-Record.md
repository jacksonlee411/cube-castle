# 211 · Day2 模块命名确认记录

**关联计划**: `docs/development-plans/211-phase1-module-unification-plan.md`  
**执行负责人**: Codex（全栈）  
**记录日期**: 2025-11-03  

---

## 1. 目的与范围
- 落实 Plan 211 Day2 “模块命名” 交付物，明确统一后模块名称及引用策略。
- 为 Day3 go.mod 合并与 Day4-5 目录迁移提供决策依据，确保资源唯一性与跨层一致。

## 2. 基线信息
- 现有五个 `go.mod`：根目录、`cmd/hrms-server/command/`、`cmd/hrms-server/query/`、`pkg/health/`、`shared/`，以及 `go.work` 工作区。
- 根模块声明 `module cube-castle-deployment-test`；子模块各自维护依赖，存在跨层 import 混用（例如 `organization-command-service/...`、`cube-castle-deployment-test/internal/...`）。
- `reports/phase1-module-unification.md` 已记录资产盘点及脚本执行时间戳。

## 3. 决议事项
1. **根模块命名**：统一调整为 `module cube-castle`，Go 版本对齐 Plan 204 约定（go1.22.x；2025-11-03 执行阶段因 `github.com/jackc/pgx/v5` 等依赖要求已临时提升至 `go1.24.0`，待 Steering 复核后回写 Plan 204）。
2. **子模块依赖归并**：收敛五个 `go.mod` 至单一根模块；移除 `go.work`，避免平行事实来源。
3. **包路径规范**：命令/查询服务在迁移完成前暂采用 `cube-castle/cmd/<service>`，共享代码统一 `cube-castle/internal/*`、`cube-castle/shared/*`、`cube-castle/pkg/health`.
4. **文档同步**：所有命名与路径调整须更新相关计划、日志与引用文档，确保事实来源唯一。

## 4. 后续行动
- Day3：依据本决议编制 go.mod 合并策略并准备批量替换脚本；记录执行前后校验日志。
- Day4-5：执行路径迁移与 CI 清理，提交变更及测试证据至 `reports/phase1-module-unification.md` / `reports/phase1-regression.md`。
- 日志同步：每次关键动作后更新 `docs/development-plans/06-integrated-teams-progress-log.md` 与上述执行日志。
