# 241C – E2E 验收与可观测性证据登记

编号: 241C  
标题: Playwright 验收连跑（2 浏览器 × 3 轮）与 `[OBS]` 证据登记  
创建日期: 2025-11-15  
状态: 待实施  
上游关联: 241A（骨架合流）、241B（Hook 与门禁）、240（职位页面重构）、232/232T（P0 稳定）

---

## 1. 背景与目标

- 背景：241 主计划要求以 E2E 证据与可观测性日志作为关单门槛；当前仓库缺少 plan241 的日志/报告资产。  
- 目标：在骨架合流与门禁收口后，完成 Playwright 连跑（Chromium/Firefox 各 3 轮），登记 `[OBS]` 事件与 performance 标记日志，形成可审计的关单资产。

---

## 2. 范围与产物

- 连跑套件（至少）：
  - `frontend/tests/e2e/temporal-management-integration.spec.ts`（组织/职位入口与详情流转）
  - `frontend/tests/e2e/position-tabs.spec.ts`（职位详情页签切换与 timeline/versions 行为）
- 运行参数与门控：
  - `PW_OBS=1 VITE_OBS_ENABLED=true`（启用 `[OBS]` 事件输出）
  - 选用 CI 采集：`VITE_ENABLE_MUTATION_LOGS=true`（以 `logger.mutation` 输出，便于采集）
- 证据落盘：
  - E2E 控制台日志与事件：`logs/plan241/C/obs-{spec}-{browser}.log`
  - Playwright 报告：`frontend/playwright-report/**`
  - 附加：`logs/plan241/C/e2e-{spec}-{browser}.log`（运行摘要与失败快照路径）

---

## 3. 验收标准

1) 连跑通过：Chromium/Firefox 各 3 轮，失败率 0；如因环境波动失败，需附故障说明与重跑记录  
2) 事件命中：
  - 职位：`position.hydrate.start/.done`、`position.tab.change`、`position.version.select`、`position.version.export.*`、`position.graphql.error`（必要路径）
  - 组织：只要求 `performance.mark('obs:temporal:*')` 存在（不新增 `organization.*` 事件名）
3) 选择器与门禁：仅使用 `temporalEntitySelectors`；`guard:selectors-246` 与 ESLint 规则同时通过  
4) 资产完整：上述日志与报告落盘路径存在且可供审计

---

## 4. 执行步骤

1) 准备：`make docker-up && make run-dev && (cd frontend && npm run dev)`（如需）  
2) 环境：导出 `PW_OBS=1 VITE_OBS_ENABLED=true [VITE_ENABLE_MUTATION_LOGS=true]`  
3) 执行：`cd frontend && npm run test:e2e`（按浏览器矩阵与轮次执行）  
4) 采集：将控制台输出中以 `[OBS] ` 前缀的行写入 `logs/plan241/C/obs-*.log`；复制 Playwright 报告  
5) 登记：在本文件“执行记录”节追加本次执行时间戳、命令、浏览器矩阵与失败重跑摘要；必要时在 215 执行日志同步索引

---

## 5. 风险与缓解

| 风险 | 影响 | 概率 | 缓解 |
|---|---|---|---|
| 本地/CI 环境波动导致偶发失败 | 中 | 中 | 记录失败原因与重跑日志；保留 trace/video；必要时降为“2 轮+1 重跑”并登记 |
| 事件采集遗漏 | 低 | 中 | 在职位页内保留对关键事件的“直接注入”，骨架层为补充，不相互覆盖 |

---

## 6. 执行记录（滚动登记）

- 2025-11-__（待执行）：命令、矩阵、通过率、失败重跑说明与链接

---

## 7. 退出准则

- 连跑通过（矩阵满足）；事件/标记命中；门禁通过；证据已登记并可审计  
- 本文件链接登记至 241 主计划“阶段性结论”章节与 215 执行日志

