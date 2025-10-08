# Plan 23 - Plan16 P0 稳定化方案

## 背景概述
- 依据 `reports/iig-guardian/e2e-test-results-20251008.md`，最新复测显示 156 个用例中仅 72 个通过（46.2%），**相较 2025-10-02 基线（06 文档记载架构契约 6/6 全绿、业务流程部分通过）出现显著退化**，大量用例因框架误判后端不可用而进入 Mock 模式。
- `docs/development-plans/06-integrated-teams-progress-log.md` 已将 Plan16 中的 P0 阻塞项限定为「E2E 测试修复 ≥90% 通过率」「补齐 plan16-phase* Git 标签」「同步 16 系列文档事实」。
- `docs/reference/16-code-smell-analysis-and-improvement-plan.md` 与 `docs/reference/16-REVIEW-SUMMARY.md` 需与最新交付状态保持一致，当前尚未完成最终同步。

## 事实来源与一致性校验
- ✅ 测试数据：`reports/iig-guardian/e2e-test-results-20251008.md`（唯一存储本轮复测指标）。
- ✅ 进度对齐：`docs/development-plans/06-integrated-teams-progress-log.md`（Plan16 当前优先级声明）。
- ✅ 归档清单：`reports/iig-guardian/plan16-archive-readiness-checklist-20251008.md`（M1-M5 必修项与本计划目标一致）。
- 每项行动完成后需更新上述事实源之一，并在 `docs/development-plans/06-integrated-teams-progress-log.md` 记录一致性校验结果；最终需回填归档清单勾选状态。

## 目标
1. 恢复真实后端 E2E 测试能力，使通过率≥90%。
2. 补齐 Plan16 P0 约定的 Git 标签并推送远端。
3. 同步 16 系列文档与进度日志，确保事实唯一性、最新性。

## 工作拆解

### 0. 前置验证（责任：QA，预计 0.5 天）
- 执行 `make status` 确认命令/查询服务与前端均为 200 状态。
- 手动验证健康检查：`curl http://localhost:9090/health` 与 `curl http://localhost:8090/health` 必须返回 200。
- 校验 `.cache/dev.jwt` 存在且 scope 至少包含 `org:read org:create org:update`，必要时执行 `make jwt-dev-mint` 重新签发。
- 仅在上述全部通过后进入下一步骤。

### 1. 修复 E2E 健康检测逻辑（责任：测试平台组，预计 2–3 天）
- 复核 `frontend/tests` 内健康检查实现，定位触发 Mock 模式的条件（网络探测、超时、环境变量）。
- 对照 `curl http://localhost:9090/health` 与 `curl http://localhost:8090/health` 的实际返回，修正误判逻辑或延长超时。
- 更新相关配置/脚本，并在 `reports/iig-guardian/e2e-test-results-YYYYMMDD.md` 记录修复后的探测结果。
- 预留缓冲：若定位到后端服务启动时序、端口冲突或鉴权异常等问题，需额外安排 1 天与相关团队协同排查。

### 2. 回归执行与报告归档（责任：测试平台组，预计 0.5 天）
- 运行 `npm run test:e2e`（确保 `PW_JWT`、`PW_TENANT_ID` 注入有效）。
- 核对 Playwright 日志确认未触发 Mock 模式，并对 CRUD、GraphQL 契约等核心流程进行人工复核。
- 达到通过率≥90% 后，生成新的报告并覆盖/追加至 `reports/iig-guardian/e2e-test-results-YYYYMMDD.md`，同步更新 `frontend/playwright-report/` 与 `frontend/test-results/` 归档。
- 将通过率、失败样本、Mock 模式状态写入 `docs/development-plans/06-integrated-teams-progress-log.md`，并更新归档清单 M1 状态。

### 3. 补齐 Plan16 Git 标签（责任：交付负责人，预计 0.5 天）
- 结合 `docs/development-plans/06-integrated-teams-progress-log.md` 与 `git log --oneline --since="2025-10-01" --until="2025-10-10"`，定位 Phase1（handlers 拆分完成）、Phase2（弱类型清零归档）、Phase3（CQRS 验证完成或 E2E 修复收尾）对应的关键提交。
- 分别创建 `plan16-phase1-completed`、`plan16-phase2-completed`、`plan16-phase3-completed` 标签，并保留提交哈希记录至进度日志。
- 推送前执行 `git tag -l | grep plan16` 校验避免重复；推送后更新归档清单 M2 状态。

### 4. 文档一致性同步（责任：架构文档组，预计 1 天）
- 更新 `docs/reference/16-code-smell-analysis-and-improvement-plan.md`：
  - 补充 main.go 拆分成果（入口 13 行 + `internal/app/*` 模块化）与残余橙灯策略（temporal 系列 5 文件列入 P2 拆分计划）。
  - 新增 E2E 验收小节，引用通过率≥90% 的报告与核心用例截图/链接。
- 更新 `docs/reference/16-REVIEW-SUMMARY.md`：
  - 将弱类型治理状态调整为“173→0 已归档（参考 Plan21）”。
  - 写明 E2E 测试状态（2025-10-XX 复测 ≥90% 通过，核心 CRUD 100%）。
  - 补充 Phase0-3 时间线与标签指向的提交哈希。
- 更新 `docs/development-plans/06-integrated-teams-progress-log.md`：
  - 将 P0 待办全部标记为 ✅ 完成，并附测试报告与标签记录链接。
  - 新增“Plan16 归档完成”条目，注明责任人/日期/归档清单链接，更新 M3-M5 状态。

## 风险与缓解
- **健康检测仍可能误判**：预留手动兜底（直接检查 Playwright 配置、增加日志），必要时临时关闭 Mock 模式开关进行验证。
- **测试环境波动**：在执行前运行 `make status` 与健康检查命令确认服务全绿。
- **标签推送冲突**：在创建前使用 `git tag -l | grep plan16` 校验，避免重复。
- **Mock 模式兜底说明**：执行 `E2E_MOCK_MODE=false npm run test:e2e` 或在 Playwright 配置中移除相关 env，确保即使检测失败也能强制走真实后端路径以辅助定位。

## 验收标准
- E2E 复测报告显示整体通过率≥90%，**核心业务流程（CRUD、GraphQL 契约）用例 100% 通过**，并附 `frontend/playwright-report/index.html`、`frontend/test-results/` 归档路径。
- Playwright 输出中无 “⚠️ 启用 E2E Mock 模式” 警告，确认未启用 Mock 模式。
- 远端仓库可查询到三枚 `plan16-phase*-completed` 标签，且指向经进度日志确认的关键提交哈希。
- `docs/reference/16-code-smell-analysis-and-improvement-plan.md`、`docs/reference/16-REVIEW-SUMMARY.md`、`docs/development-plans/06-integrated-teams-progress-log.md` 记录与测试结果一致，并联动更新 `reports/iig-guardian/plan16-archive-readiness-checklist-20251008.md` M1-M5 为 ✅。
- `docs/development-plans/06-integrated-teams-progress-log.md` 的 P0 阻塞项全部标记为完成，且附上验收证据链接与责任人信息。
