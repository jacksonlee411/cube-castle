# Plan 230F – 职位管理功能对齐与测试映射

**文档编号**: 230F  
**母计划**: Plan 230 – Position CRUD 参考数据修复计划  
**前置计划**: 230D（E2E 执行完成）  
**负责人**: QA + 前端 + 产品

---

## 1. 背景与目标

- 在 230 原计划 3A.7 中要求对照契约与实现，确认 Position 功能交付范围与 Playwright 测试断言保持一致。  
- 219E/219T 期间多次出现“测试覆盖尚未交付的功能”或 “UI data-testid 漂移” 的问题，需要系统化记录并建立“功能 → 测试”映射表。  
- 230F 的目标是产出《Position 模块 readiness 表》与测试映射，确保未来 E2E 断言有据可依。

---

## 2. 范围

1. 对照 `docs/api/openapi.yaml`（命令端）与 `frontend/src/features/positions` 当前实现，列出已交付/未交付功能（创建、编辑、权限校验、版本控制等）。  
2. 为每一项功能映射到 Playwright/REST/GraphQL 测试（例如 `tests/e2e/position-crud-full-lifecycle.spec.ts`, `tests/e2e/organization-validator/*.spec.ts`），指出断言内容与覆盖范围。  
3. 若测试覆盖尚未交付的功能，需在测试代码中加 `// TODO-TEMPORARY:` 或在计划中记录预期上线时间。  
4. 产出文档 `logs/230/position-module-readiness.md`（或同目录 Markdown），作为唯一事实来源，并在 230 主计划与 219E 中引用。  
5. 若发现后端缺陷或前端缺口，需同步给对应计划（例如 222、219B），并在 readiness 表中注明责任与 ETA。

---

## 3. 任务

| 步骤 | 描述 | 输出 |
| --- | --- | --- |
| F1 | 阅读 `docs/api/openapi.yaml` 相关 Position 接口与 `docs/reference/02-IMPLEMENTATION-INVENTORY.md`，列出功能清单 | readiness 表草稿 |
| F2 | 浏览 `frontend/src/features/positions`、`frontend/tests/e2e` 中的实现与测试，记录 data-testid、依赖数据 | readiness 表补充 |
| F3 | 将功能与测试映射写入 `logs/230/position-module-readiness.md` 表格，列明状态（✅/⏳/❌）、测试引用、责任人 | 文档 |
| F4 | 对于未交付但被测试覆盖的项，向相关负责人建 issue 或在 `docs/development-plans/06-integrated-teams-progress-log.md` 中登记，并在测试代码加注释（如属于 230 范围则回到 230B/D） | issue/日志/代码注释 |
| F5 | 将 readiness 表链接同步到 219E 文档“Position + Assignment 数据恢复计划”部分 | 文档引用 |

---

## 4. 依赖

- 230D 的 E2E 日志提供最新的测试覆盖范围。  
- `docs/api/openapi.yaml`、`docs/reference/02-IMPLEMENTATION-INVENTORY.md` 为唯一事实来源。  
- 前端代码库需保持与本仓库同步，以读取最新实现。

---

## 5. 验收标准

1. `logs/230/position-module-readiness.md` 包含功能 × 测试映射表，且引用所有相关日志/测试路径。  
2. 至少一次在测试代码或计划文档中更新了 TODO / 责任人，确保未交付功能不再导致 E2E 误报。  
3. 219E 文档引用 readiness 表链接，并在 2.6 章节中说明现状。  
4. 若发现新的缺口，已登记在 06-progress log 或对应计划；230F 本身不关闭缺陷，但必须完成记录与责任转交。

---

> 唯一事实来源：`docs/api/openapi.yaml`、`frontend/src/features/positions/*`、`tests/e2e/position-crud-full-lifecycle.spec.ts`、`logs/230/position-crud-playwright-*.log`。  
> 更新时间：2025-11-07。
