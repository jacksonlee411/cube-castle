# 230 – Position 模块功能对齐核对（2025-11-07）

## 1. 功能矩阵（实现 vs 测试）
| 功能点 | 实现事实来源 | 当前实现状态 | Playwright/REST 覆盖 | 结论 |
| --- | --- | --- | --- | --- |
| 创建 / 替换职位 (POST `/api/v1/positions`, PUT `/api/v1/positions/{code}`) | `docs/api/openapi.yaml:1656-1725`；路由与处理器 `internal/organization/handler/position_handler.go:58-167` | 命令服务已提供 Create/Replace，并支持 `If-Match`/`X-Idempotency-Key` 验证 | `frontend/tests/e2e/position-crud-full-lifecycle.spec.ts:30-205` 覆盖 Step1 (Create) 与 Step3 (Update)，断言 201/200 及 GraphQL 可见性 | ✅ 可在真实数据上执行；若 Job Catalog 缺失则按 230 计划脚本修复后再跑测试 |
| 填充 / 空缺 / 删除职位 (POST `/fill`、`/vacate`、`/events`) | `docs/api/openapi.yaml:1764-1815`；处理器 `internal/organization/handler/position_handler.go:170-215` + `226-244` | Fill、Vacate、Event 路径由 PositionHandler 暴露，服务层在运行日志中可见 | `frontend/tests/e2e/position-crud-full-lifecycle.spec.ts:206-366` 覆盖 Step4/5/6，串联填充→空缺→删除生命周期 | ✅ 功能已交付且测试串联验证；需要确保 `testAssignmentId` 来源于 Fill 响应 |
| Position Version API (POST `/api/v1/positions/{code}/versions`) | `docs/api/openapi.yaml:1726-1763`；处理器 `internal/organization/handler/position_handler.go:121-169` | 服务端具备版本插入入口，但目前仅在操作历史中调用，未在 UI 直接暴露 | `position-crud-full-lifecycle` 用例 Step7 仅检查版本列表 UI（`frontend/tests/e2e/position-crud-full-lifecycle.spec.ts:369-389`），未直接调用 `/versions` | ⚠️ 需新增专门的 REST/Playwright 步骤或在现有用例中调用 `/versions` 后再校验 GraphQL，避免接口回归被遗漏 |
| Transfer / Assignments 子路由 | 处理器 `internal/organization/handler/position_handler.go:226-370` 提供 `/transfer` 与 `/assignments/*` | Handler 已暴露所有子路由，但当前前端 UI 未启用 Transfer / assignment CRUD，后台日志亦无覆盖 | Playwright 未涉及相关路径；`frontend/src/features/positions` 也无对应交互 | ⚠️ 暂视为未交付功能，若测试需要断言请先标记 `// TODO-TEMPORARY` 并在 219T 追踪，不得直接以失败视作缺陷 |

## 2. 测试映射与风险
- **已实现且有测试**：Create/Replace、Fill/Vacate/Event 的 REST → GraphQL 流程，现由 `tests/e2e/position-crud-full-lifecycle.spec.ts` 串联验证；若 Job Catalog 参考数据缺失，可用 `database/migrations/20251107123000_230_job_catalog_oper_fix.sql` + `scripts/diagnostics/check-job-catalog.sh` 恢复后再跑。
- **已实现但测试缺失**：Position Version API、Transfer、Assignments 列表/编辑。建议新增针对 `/versions` 的 REST 步骤（或在 Step3 后插入版本再验证版本列表），并在 UI 启用 Transfer/Assignment 之前将相关测试标记为待交付。
- **尚未实现**：无（以上功能在 handler 层已暴露，只是 UI/测试未覆盖）。如未来确认业务尚未交付，请在计划文档中记录豁免范围。

## 3. 行动指引
1. **测试前置校验**：运行 `bash scripts/diagnostics/check-job-catalog.sh`，确保 `OPER` 参考数据处于 ACTIVE 状态；否则执行 `make db-migrate-all` 引入 230 迁移。
2. **用例调整**：若 Playwright 需要验证 `Transfer`/`Assignments` 等尚未上线的能力，请先与后端确认交付节奏，再通过 `test.skip` + `// TODO-TEMPORARY(230)` 说明原因。
3. **证据归档**：功能验证日志（REST 响应、GraphQL 截图）与本文件一同归档，供 219T 报告引用。
