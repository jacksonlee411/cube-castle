# 65号文档：工具与验证体系巩固计划（Phase 4）

**版本**: v0.1  
**创建日期**: 2025-10-12  
**维护团队**: 全栈工程师（单人执行）  
**状态**: 规划中  
**关联计划**: 60号系统级质量重构总计划、61号执行计划（第四阶段）  
**参考文档**: `docs/development-plans/60-execution-tracker.md`、`docs/archive/development-plans/63-front-end-query-plan.md`、`docs/development-plans/64-phase-3-acceptance-draft.md`

---

## 1. 背景与目标

随着 Phase 3 前端 API/Hooks/配置整治完成，系统质量重构进入最后阶段。Phase 4（Week 9-10）聚焦“工具与验证体系巩固”，核心目标：

1. **统一 Temporal / Validation 工具链**：避免前后端分别维护校验逻辑，减少重复维护。
2. **完善审计追踪数据**：保证 `cmd/organization-command-service/internal/audit` 相关模型在任意场景都能输出完整 DTO。
3. **强化 CI 守护**：在既有 `architecture-validator`、`validate-metrics` 基础上新增契约、审计、文档归档三类质量门禁，确保后续演进可持续。

---

## 2. 范围

### 2.1 主要影响模块
- 后端命令服务：`cmd/organization-command-service/internal/validators/*`、`internal/audit/*`、`internal/services/temporal/*`
- 前端验证层：`frontend/src/shared/validation/*`、`frontend/src/shared/types/*`
- 质量脚本与 CI：`scripts/quality/*`、`.github/workflows/*`、`Makefile`、`package.json` 脚本项
- 文档更新：`docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/06-integrated-teams-progress-log.md`、`docs/development-plans/60-execution-tracker.md`

### 2.2 非范围内容
- 不重新设计业务契约（REST / GraphQL 以 `docs/api` 为唯一真源）。
- 不涉及数据库结构变更；若需要新增字段，将另行立项。
- 不重写 Playwright / Vitest 测试，只在必要时补充验证覆盖。

---

## 3. 时间线 (预计 2 周)

| 周次 | 里程碑 | 关键产出 |
|------|--------|----------|
| Week 9 | Temporal & Validation 工具统一、审计 DTO 梳理 | 统一校验入口、审计 DTO 草案、回归测试集 |
| Week 10 | CI 守护任务落地、验收总结 | 新增质量脚本、GitHub Actions 任务、Phase 4 验收草稿 |

---

## 4. 详细任务

### 4.1 Week 9：工具与数据模型统一

1. **梳理现有校验逻辑**
   - 盘点后端 `internal/validators/business.go` 与前端 `frontend/src/shared/validation/schemas.ts` 差异。
   - 输出差异清单（字段、约束、错误码），记录于 `reports/validation/phase4-diff.md`（新建）。

2. **构建统一的 Validation 底座**
   - 在后端创建 `internal/validators/ruleset.go`（示例命名）定义可复用的规则集，导出供 REST Handler、Temporal 服务复用。
   - 前端通过生成脚本或共享 JSON（例如 `shared/contracts/validation-rules.json`）拉取规则，更新 `schemas.ts`。
   - 更新 `frontend/src/shared/utils/validation.ts`（如不存在则新建），封装统一的错误消息映射。

3. **Temporal 工具折叠**
   - 清理历史脚本/工具（如 `frontend/tests/e2e/utils/*`、`scripts/dev/*` 中与 Temporal 相关的临时脚本），保留单一入口。
   - 在 `Makefile` / `package.json` 中提供 `make temporal-validate`、`npm run validate:temporal` 等命令入口。

4. **审计 DTO 完整化**
   - 统一 `audit.AuditEvent` 映射，新增 DTO（例如 `internal/audit/dto.go`）确保 `resourceId`、`actorId`、`changes` 字段在认证 / 系统事件场景下也有值。
   - 更新 `cmd/organization-command-service/internal/audit/logger.go`、`internal/utils/metrics.go` 相关调用，确认 `audit_logs` 表兼容。
   - 补充 `tests/e2e/auth_flow_e2e_test.go` 或新增 Go 单测覆盖开发模式登录、刷新、异常场景。

5. **同步文档**
   - 更新 `docs/reference/03-API-AND-TOOLS-GUIDE.md` 增加“统一校验工具”与“审计 DTO”章节。
   - 在 `docs/development-plans/06-integrated-teams-progress-log.md` 中新增验证结果记录行。

### 4.2 Week 10：CI 守护与验收

1. **新增质量脚本**
   - `scripts/quality/lint-contract.js`：校验生成的契约 JSON 是否与 `docs/api` 一致（结合 `shared/contracts/`）。
   - `scripts/quality/lint-audit.js`：检测审计 DTO / 数据库字段是否缺失（可调用 Go 程序 `cmd/tools/audit-lint/main.go`）。
   - `scripts/quality/doc-archive-check.js`：校验 `docs/development-plans/` 与 `docs/archive/development-plans/` 的计划状态一致性。

2. **CI 集成**
   - 在 `.github/workflows/quality.yml`（若不存在则新建）中新增 `lint-contract`、`lint-audit`、`doc-archive-check` 三个 job。
   - 本地 Makefile/NPM Scripts 增补快捷命令：`make lint-contract`、`make lint-audit`、`npm run lint:docs` 等。

3. **回归验证**
   - 执行 `make test`、`npm run test`、`npm run test:e2e:smoke`、`npm run build:analyze`，确保 Phase 3 基线不回退。
   - 复核 `reports/implementation-inventory.json`，确认新增工具未破坏现有导出。

4. **验收文档**
   - 起草 Phase 4 验收文档 `docs/development-plans/66-phase-4-acceptance-draft.md`（占位名称，执行过程中创建）。
   - 更新 `docs/development-plans/60-execution-tracker.md`，标记第四阶段进度。
   - 在 `docs/development-plans/06-integrated-teams-progress-log.md` 添加 Week 9-10 验收结果表格。

---

## 5. 验收标准

1. **校验工具统一**
   - 前后端共享同一份 Validation 规则（通过契约文件或脚本生成），重复逻辑清零。
   - `npm run validate:temporal` 与 `make temporal-validate` 输出一致、通过。

2. **审计记录完整**
   - 任意 `AUTH`、`SYSTEM`、`USER` 事件的 `resource_id` 不为 NULL，`changes` / `beforeData` / `afterData` 字段完整。
   - `tests/e2e/auth_flow_e2e_test.go` 引入断言校验审计内容。

3. **CI 守护生效**
   - GitHub Actions 新增三个 job 全绿。
   - 本地 `make lint-contract && make lint-audit && npm run lint:docs` 返回 0。

4. **文档同步**
   - `docs/reference/03-API-AND-TOOLS-GUIDE.md`、`docs/development-plans/06-integrated-teams-progress-log.md`、`docs/development-plans/60-execution-tracker.md` 均记录 Phase 4 成果。

---

## 6. 依赖与前提

- Phase 3（文档参考：`docs/archive/development-plans/63-front-end-query-plan.md`、`docs/development-plans/64-phase-3-acceptance-draft.md`）已完成并归档。
- CI 管道具备新增 job 的权限（需提前在仓库设置中确认）。
- `scripts/quality/` 目录已有现成脚本模板（如 `architecture-validator.js`）可复用。

---

## 7. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 前后端 Validation 规则冲突 | 中 | 建立自动对比脚本，若差异超出白名单则阻断 CI |
| 审计日志历史数据缺失 | 中 | 引入兼容迁移脚本（仅在需要时），新版逻辑允许回写缺失字段 |
| CI Job 耗时增加导致流水线变慢 | 低 | 将新 job 与现有 job 并行执行，必要时调高缓存利用率 |
| 工具统一影响既有脚本 | 中 | 在 Week 9 完成后立即回归 `scripts/quality/iig-guardian.js`、`validate-metrics.sh`，确保输出未变 |

---

## 8. 交付物

- 统一的 Validation 规则文件与生成脚本
- 审计 DTO / Logger 更新代码与测试
- 新增质量脚本（`lint-contract`、`lint-audit`、`doc-archive-check`）
- GitHub Actions 工作流更新
- Phase 4 验收草案（66 号文档，占位）
- 更新后的参考文档与执行跟踪条目

---

## 9. 验收流程

1. 按验收标准逐项执行自检，将结果登记至 06 号文档表格。
2. 完成 Phase 4 验收草案并提交评审。
3. 评审通过后：更新 60 号执行跟踪 → 将本计划归档（移至 `docs/archive/development-plans/`）。

---

## 10. 附录 / 参考资料

- `cmd/organization-command-service/internal/audit/logger.go` — 审计日志实现
- `cmd/organization-command-service/internal/validators/business.go` — 当前业务校验逻辑
- `frontend/src/shared/validation/schemas.ts` — 前端校验 Schema
- `scripts/quality/` — 现有质量工具目录
- `.github/workflows/*` — 质量守护 CI 配置

> 后续执行中的新增脚本与文档，请遵循 `docs/DOCUMENT-MANAGEMENT-GUIDELINES.md`，确保唯一事实来源与命名一致性。
