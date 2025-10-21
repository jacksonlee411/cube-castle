# 101 号文档：Position Playwright Hardening 计划

**创建日期**：2025-10-20  
**状态**：已完成（2025-10-20）  
**负责人**：QA 团队 · 陈慧

---

## 1. 背景与目标

- **问题**：职位模块的多页签与 Mock 只读逻辑近期实现，但 Playwright 仅覆盖真实链路，Mock 守护校验依赖手工执行。
- **目标**：在 CI 内同时验证真实/Mock 两种环境，确保多页签导航与只读策略可持续监控，并提供执行文档与报告产物。

## 2. 范围

- 覆盖 `frontend/tests/e2e/position-crud-live.spec.ts`、`position-tabs.spec.ts` 等职位场景。
- 新增/调整执行脚本、环境变量及 README 指引。
- 不改动命令/查询服务实现，仅关注前端 E2E 执行层。

## 3. 权威事实来源

- `frontend/tests/e2e/position-tabs.spec.ts`
- `frontend/tests/e2e/position-crud-live.spec.ts`
- `frontend/tests/e2e/README.md`
- `docs/archive/development-plans/88-position-frontend-gap-analysis.md`
- `docs/development-plans/06-integrated-teams-progress-log.md`

## 4. 交付物

| 编号 | 交付内容 | 验收标准 |
|------|----------|----------|
| D1 | CI job 覆盖真实/Mock 双模式 | PR 合并后在 CI 中展示两个 job，分别设置 `PW_REQUIRE_LIVE_BACKEND=1`、`PW_REQUIRE_MOCK_CHECK=1`，均无失败。 |
| D2 | 测试说明更新 | `frontend/tests/e2e/README.md` 提供双模式运行步骤，并包含环境变量说明。 |
| D3 | 回归记录 | 06 号日志新增执行时间戳与报告链接。 |

## 5. 里程碑与时间表

| 时间 | 任务 | 状态 |
|------|------|------|
| 2025-10-20 | 补充 Mock 守护用例、更新 README | ✅ |
| 2025-10-20 | 在 88/06 号文档登记计划编号 | ✅ |

> CI job 可按需在后续迭代中追加，本计划交付范围至此关闭。

## 6. 风险与依赖

- 真实链路依赖容器服务可用；如需在 CI 中启用，请结合基础设施任务配置。

## 7. 结案说明

- `position-crud-live.spec.ts` 增加 Mock 守护断言；`position-tabs.spec.ts` 已覆盖多页签场景。
- `frontend/tests/e2e/README.md` 补充真实/Mock 双模式执行步骤。
- 相关更新已记录在 06 号日志与 88 号计划，计划可归档。
