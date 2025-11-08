# 219T4 – 验证复盘与报告回填子方案

## 1. 背景
- 在 219T1~219T3 完成后，需要重新执行 REST/性能/E2E 全量测试，并将结果回填至 219T 主报告与 `219E-e2e-validation.md`。
- 此子方案负责组织最终验证、更新文档与汇报状态。

## 2. 目标
1. 统一安排重新执行的命令：`scripts/e2e/org-lifecycle-smoke.sh`、`scripts/perf/rest-benchmark.sh`（新版）、`npm run test:e2e`。
2. 整理日志与报告，生成 2025-11-XX 新档案。
3. 更新 219T 主文档“下一步”章节为已完成，并将关键数据同步到 219E 计划。

## 3. 步骤
| 序号 | 操作 | 说明 |
| --- | --- | --- |
| R1 | 确认 219T1~219T3 都已合入 main/feature 分支，环境无残留 | 使用 `git log`、`make status` |
| R2 | 运行全量脚本，收集日志 | 目标：`logs/219E/org-lifecycle-*.log`、`logs/219E/perf-rest-*.log`、`frontend/test-results/` |
| R3 | 更新 219T 报告章节（REST/性能/Playwright）与 219E 计划 | 标记旧问题已解决，附新数据 |
| R4 | 在 `docs/development-plans/06-integrated-teams-progress-log.md` 记录阶段验收结论 | 与 Plan 06 负责人同步 |

## 4. 交付物
- 新一版 219T 报告与 219E 计划。
- 若仍有失败项，附阻塞说明与后续迭代计划。

## 5. 验收标准
1. 三类测试全部可重复，日志齐备并上传至 `logs/219E/`。
2. 219T 主文档“下一步”事项标记为完成，并追加“回填结果”章节。
3. Plan 06 进度记录同步更新。

---

> 唯一事实来源：`docs/development-plans/219T-e2e-validation-report.md`、`docs/development-plans/219E-e2e-validation.md`、`logs/219E/`。  
> 更新时间：2025-11-07。
